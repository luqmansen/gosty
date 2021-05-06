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
	"strings"
	"sync"
	"time"
)

func (s Svc) ProcessTaskMerge(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/%s", wd, TmpPath)

	errCh := make(chan error)
	var wg sync.WaitGroup

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
	wg.Wait() //need to make sure all files downloaded

	//list all file with absolute path
	var fileList []string
	for _, v := range task.TaskMerge.ListVideo {
		fileList = append(fileList, fmt.Sprintf("%s/%s", workdir, v.FileName))
	}
	if len(fileList) == 0 {
		return errors.New("no file to merge")
	}

	//create FIFOs for every video with format: absolute/path/filename_00X_WxH
	var namedPipeList []string
	for _, f := range fileList {
		namedPipeList = append(namedPipeList, strings.Split(f, ".")[0])
	}

	//output name will be absolute/path/original_file_name_WxH.mp4
	splitName := strings.Split(fileList[0], "_")
	outputFilePath := fmt.Sprintf("%s_%s", splitName[0], splitName[2])

	//merging using concat protocol + named pipe
	//since currently we only support mp4
	//https://trac.ffmpeg.org/wiki/Concatenate#protocol
	func() {
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
	}()

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
			}
		}(f, namedPipeList[idx])
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
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
		}
	}()

	wg.Wait() // make sure everything is done

	//upload the result
	wg.Add(1)
	go func() {
		defer wg.Done()

		fileReader, err := os.Open(outputFilePath)
		if err != nil {
			log.Errorf("error opening output: %s", err)
			errCh <- err
			return
		}

		values := map[string]io.Reader{"file": fileReader}
		fileName := strings.Split(outputFilePath, "/")
		url := fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), fileName[len(fileName)-1])
		if err = util.Upload(url, values); err != nil {
			log.Errorf("Error uploading file %s: %s", outputFilePath, err)
			errCh <- err
			return
		}
		if err = os.Remove(outputFilePath); err != nil {
			log.Errorf("Error removing result file after uploading %s: %s", outputFilePath, err)
			errCh <- err
			return
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
	// TODO [$609358567c9cf10008f9351a]:  (improvement) make the pipe generic name (eg: temp1)
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

	select {
	case err := <-errCh:
		return err
	default:
		wg.Wait()
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskMerge.Result = result
		return nil
	}
}
