package worker

import (
	"bytes"
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *Svc) ProcessTaskMerge(task *models.Task) error {
	start := time.Now()

	log.Debugf("Processing task %s,  id: %s", models.TASK_NAME_ENUM[task.Kind], task.Id.Hex())

	if err := s.downloadFileList(task); err != nil {
		return err
	}

	//list all file with absolute path
	var fileList []string
	for _, v := range task.TaskMerge.ListVideo {
		fileList = append(fileList, fmt.Sprintf("%s/%s", workdir, v.FileName))
	}
	if len(fileList) == 0 {
		return errors.New("no file to merge")
	}

	sort.Slice(fileList, func(i, j int) bool {
		// Below function is to extract the number from string like this
		// /app/tmpworker/9cbcd3f38f0f4f339b551f2c38c9bad6-767812846-10_854x480
		// to just get the "10"
		a := strings.Split(strings.Split(strings.Split(fileList[i], "/")[3], "-")[2], "_")[0]
		b := strings.Split(strings.Split(strings.Split(fileList[j], "/")[3], "-")[2], "_")[0]
		n, _ := strconv.Atoi(a)
		m, _ := strconv.Atoi(b)

		return n < m
	})

	//create FIFOs for every video with format: absolute/path/filename_00X_WxH
	var namedPipeList []string
	for _, f := range fileList {
		namedPipeList = append(namedPipeList, strings.Split(f, ".")[0])
	}
	if len(namedPipeList) == 0 {
		return errors.New("worker.merge: no named pipelist created")
	}

	//What below code does is, split filename from
	// "/path/to/tmpworker/filename-alsofilename-5_256x144.mp4"
	// to  path/to/tmpworker/filename-alsofilename_256x144.mp4"
	splitName := strings.Split(fileList[0], "-")
	ext := strings.Split(strings.Split(splitName[2], "-")[0], "_")
	outputFilePath := fmt.Sprintf("%s_%s", splitName[0], ext[1])

	//merging using concat protocol + named pipe since currently we only support mp4
	//https://trac.ffmpeg.org/wiki/Concatenate#protocol
	if err := createPipeFile(namedPipeList); err != nil {
		return err
	}

	if err := s.concatOperation(fileList, namedPipeList, outputFilePath); err != nil {
		return err
	}

	//upload the result
	wg := sync.WaitGroup{}
	waitCh := make(chan struct{})
	errCh := make(chan error)
	go func() {
		defer close(waitCh)
		defer wg.Wait()

		fileReader, err := os.Open(outputFilePath)
		if err != nil {
			log.WithField("worker", s.worker.WorkerPodName).
				Error(errors.Wrap(err, "error opening output"))
			errCh <- err
		}

		values := map[string]io.Reader{"file": fileReader}
		fileName := strings.Split(outputFilePath, "/")
		url := fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), fileName[len(fileName)-1])
		if err = util.Upload(url, values); err != nil {
			log.WithField("url", url).
				WithField("worker", s.worker.WorkerPodName).
				Error("Error uploading file %s: %s", outputFilePath, err)
			errCh <- err

		}
		if err = os.Remove(outputFilePath); err != nil {
			log.Errorf("Error removing result file after uploading %s: %s", outputFilePath, err)
			// Currently I don't care about this part if removing file is failed
			//errCh <- err
			//return
		}
	}()

	//Removing source file
	log.Debug("removing source file")
	for _, f := range fileList {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			if err := os.Remove(file); err != nil {
				log.Errorf("Error removing source file %s: %s", file, err)
			}
		}(f)
	}

	//Removing named pipe file
	// TODO [#13]:  (improvement) make the pipe generic name (eg: temp1)
	// and can be reused to next process, might reducing io if the
	// created pipe file is a lot
	log.Debug("removing pipe file")
	for _, f := range namedPipeList {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			if err := os.Remove(file); err != nil {
				log.Errorf("Error removing source file %s: %s", file, err)
			}
		}(f)
	}

	fileName := strings.Split(outputFilePath, "/")
	result := &models.Video{
		FileName: fileName[len(fileName)-1],
	}

	wg.Wait()

	select {
	case err := <-errCh:
		task.Status = models.TaskStatusFailed
		// possible error is when uploading file failed
		return err
	case <-waitCh:
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskMerge.Result = result
		return nil
	}
}

func createPipeFile(namedPipeList []string) error {
	cmd := exec.Command("mkfifo", namedPipeList...)
	log.Debug(cmd.String())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = wd
	err := cmd.Run()
	if err != nil {
		log.Errorf("%s: Error making pipe file, %s ", err.Error(), stderr.String())
	}
	return err
}

func (s *Svc) concatOperation(fileList, namedPipeList []string, outputFilePath string) error {
	wg := sync.WaitGroup{}
	errCh := make(chan error)
	waitCh := make(chan struct{})

	//write every mp4 file to named pipe concurrently
	for idx, f := range fileList {
		wg.Add(1)
		go func(filename, pipeFile string) {
			defer wg.Done()

			args := fmt.Sprintf("-y -i %s -c copy -bsf:v h264_mp4toannexb -f mpegts %s",
				filename, pipeFile)
			splitArgs := strings.Split(args, " ")
			cmd := exec.Command("ffmpeg", splitArgs...)
			log.Debug(cmd.String())
			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			cmd.Dir = wd
			err := cmd.Run()
			if err != nil {
				log.Errorf("%s: Error writing to pipe for %s, %s ", err.Error(), stderr.String(), filename)
				errCh <- err
			}
		}(f, namedPipeList[idx])
	}

	// this part that writing byte from piped list to actual file output
	go func() {
		defer close(waitCh)
		defer wg.Wait()
		//this stupid exec somehow can't find the file, must run the command line via bash
		cmd := exec.Command("bash", "script/concat.sh", strings.Join(namedPipeList, "|"), outputFilePath)
		log.Debug(cmd.String())
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		cmd.Dir = wd
		err := cmd.Run()
		fmt.Println(cmd.Dir)
		if err != nil {
			log.Errorf("%s: Error concating pipe, %s", err.Error(), stderr.String())
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-waitCh:
		return nil
	}
}

func (s Svc) downloadFileList(task *models.Task) error {
	wg := sync.WaitGroup{}
	errCh := make(chan error)
	waitCh := make(chan struct{})

	go func() {
		defer close(waitCh)
		defer wg.Wait() //need to make sure all files are downloaded

		for _, video := range task.TaskMerge.ListVideo {
			wg.Add(1)
			go func(vidName string) {
				defer wg.Done()

				inputPath := fmt.Sprintf("%s/%s", workdir, vidName)
				url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), vidName)
				err := util.Download(inputPath, url)
				if err != nil {
					log.Errorf("worker.processTaskMerge: %s", err)
					errCh <- err
				}
			}(video.FileName)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-waitCh:
		return nil
	}
}
