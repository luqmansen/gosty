package worker

import (
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func (s *Svc) ProcessTaskDash(task *models.Task) error {
	log.Debugf("Processing task %s,  id: %s", models.TASK_NAME_ENUM[task.Kind], task.Id.Hex())
	start := time.Now()

	errCh := make(chan error)
	waitCh := make(chan struct{})
	wg := sync.WaitGroup{}

	go func() {
		for _, video := range task.TaskDash.ListVideo {
			wg.Add(1)
			go func(vidName string) {
				defer wg.Done()

				inputPath := fmt.Sprintf("%s/%s", workdir, vidName)
				url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), vidName)
				log.Debug(inputPath)

				downloadAndVerify := func() error {
					err := util.Download(inputPath, url)
					if err != nil {
						log.Errorf("worker.processTaskDash, url: %s, inputpath: %s, err: %s", url, inputPath, err)
						return err
					}
					cmd := exec.Command("MP4Box", "-info", inputPath)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err = cmd.Run()
					if err != nil {
						log.Error(err)
						os.Remove(inputPath)
						return err
					}
					return nil
				}

				if err := backoff.Retry(downloadAndVerify, backoff.NewConstantBackOff(1*time.Second)); err != nil {
					errCh <- err
				}

			}(video.FileName)
		}
		wg.Wait() //need to make sure all files downloaded
		close(waitCh)
	}()
	select {
	case err := <-errCh:
		return err
	case <-waitCh:
		log.Debug("Downloading all files done, processing...")
	}

	//list absolute path of all files
	var fileList []string
	for _, v := range task.TaskDash.ListVideo {
		fileList = append(fileList, fmt.Sprintf("%s/%s", workdir, v.FileName))
	}

	origFileName := strings.Split(strings.Split(task.TaskDash.ListVideo[0].FileName, ".")[0], "_")[0]
	cmd := exec.Command("bash", "script/dash.sh",
		fmt.Sprintf("%s.mpd", fmt.Sprintf("%s/%s", workdir, origFileName)),
		strings.Join(fileList, " "),
	)
	log.Debug(cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return err
	}

	//Removing source file
	for _, file := range fileList {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			if err = os.Remove(f); err != nil {
				// *Currently* I didn't really care if removing failed
				// This might result source file being sent again to fileserver
				// I'll handle this later
				log.Errorf("Error removing source file %s: %s", f, err)
			}
		}(file)
	}

	files, err := ioutil.ReadDir(workdir)
	if err != nil {
		log.Fatal(err)
		return err
	}

	var dashResult []string
	for _, f := range files {
		if strings.Contains(f.Name(), "dash") || strings.Contains(f.Name(), "mpd") {
			// all dash result will have *_dashXX filename format or .mpd extension
			dashResult = append(dashResult, f.Name())
		}
	}
	waitCh = make(chan struct{})
	go func() {
		defer close(waitCh)
		defer wg.Wait()

		for _, file := range dashResult {
			wg.Add(1)
			go func(fileName string) {
				defer wg.Done()

				filePath := fmt.Sprintf("%s/%s", workdir, fileName)
				fileReader, err := os.Open(filePath)
				if err != nil {
					log.Errorf("error opening dash result: %s", err)
					errCh <- err
					return
				}

				values := map[string]io.Reader{"file": fileReader}
				url := fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), fileName)
				if err = util.Upload(url, values); err != nil {
					log.Errorf("Error uploading file %s: %s", fileName, err)
					errCh <- err
					return
				}
				if err = os.Remove(filePath); err != nil {
					log.Errorf("Error removing dash result file after uploading %s: %s", fileName, err)
					// TODO: handle cleanup file to remove afterward if error happen
					// Currently I don't care if error happen, this shouldn't affect the
					// current task processing
					// errCh <- err
					return
				}
			}(file)
		}
	}()

	select {
	case err = <-errCh:
		return err
	case <-waitCh:
		task.TaskStarted = start
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskDash.ResultDash = dashResult
	}
	wg.Wait() // this only wait for remove source file which most likely already done
	return nil
}
