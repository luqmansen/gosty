package worker

import (
	"bytes"
	"fmt"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/pkg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func (s workerSvc) ProcessTaskDash(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/tmp", wd)

	errCh := make(chan error)
	var wg sync.WaitGroup

	for _, video := range task.TaskDash.ListVideo {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()

			log.Debug(os.Getwd())

			inputPath := fmt.Sprintf("%s/%s", workdir, v)
			url := fmt.Sprintf("%s/files/%s", viper.GetString("fs_host"), v)
			log.Debug(inputPath)
			err := pkg.Download(inputPath, url)
			if err != nil {
				log.Errorf("worker.processTaskDash: %s", err)
				errCh <- err
			}
		}(video.FileName)
	}
	wg.Wait() //need to make sure all files downloaded

	var fileList []string
	for _, v := range task.TaskDash.ListVideo {
		fileList = append(fileList, fmt.Sprintf("%s/%s", workdir, v.FileName))
	}

	log.Debugf("Processing dash task id: %s", task.Id.Hex())
	origFileName := strings.Split(strings.Split(task.TaskDash.ListVideo[0].FileName, ".")[0], "_")[0]

	cmd := exec.Command("bash", "dash.sh",
		fmt.Sprintf("%s.mpd", fmt.Sprintf("%s/%s", workdir, origFileName)),
		strings.Join(fileList, " "),
	)
	log.Debug(cmd.String())

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = wd

	err := cmd.Run()
	if err != nil {
		log.Error(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}

	//Removing source file
	for _, file := range fileList {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			if err = os.Remove(f); err != nil {
				log.Error(err)
			}
		}(file)
	}

	var dashResult []string
	files, err := ioutil.ReadDir(workdir)
	if err != nil {
		log.Fatal(err)
		return err
	}
	for _, f := range files {
		dashResult = append(dashResult, f.Name())
		log.Debug(f.Name())
	}

	for _, file := range dashResult {
		go func(fileName string) {
			filePath := fmt.Sprintf("%s/%s", workdir, fileName)
			fileReader, err := os.Open(filePath)
			if err != nil {
				log.Error(err)
				errCh <- err
				return
			}

			values := map[string]io.Reader{"file": fileReader}
			url := fmt.Sprintf("%s/upload?filename=%s", viper.GetString("fs_host"), fileName)
			if err = pkg.Upload(url, values); err != nil {
				log.Error(err)
				errCh <- err
				return
			}
			if err = os.Remove(filePath); err != nil {
				log.Error(err)
				errCh <- err
				return
			}
		}(file)
	}

	select {
	case err = <-errCh:
		log.Error(err)
		return err
	default:
		wg.Wait()
		task.TaskDuration = time.Since(start)
		task.CompletedAt = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskDash.ResultDash = dashResult
		return nil
	}
}
