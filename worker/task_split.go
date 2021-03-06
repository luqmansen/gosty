package main

import (
	"bytes"
	"fmt"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/pkg"
	"github.com/luqmansen/gosty/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type splitTaskService struct {
	mb   repositories.MessageBrokerRepository
	file *os.File
	msg  []byte
}

func NewSplitTaskService(mb repositories.MessageBrokerRepository) *splitTaskService {
	return &splitTaskService{
		mb:   mb,
		file: nil,
		msg:  nil,
	}

}

func processTaskSplit(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/worker/tmp", wd)

	filePath := fmt.Sprintf("%s/%s", workdir, task.TaskSplit.Video.FileName)
	url := fmt.Sprintf("%s/files/%s", viper.GetString("fs_host"), task.TaskSplit.Video.FileName)
	err := pkg.Download(filePath, url)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task id: %s", task.Id.Hex())
	dockerVol := fmt.Sprintf("%s:/work/", workdir)
	cmd := exec.Command(
		"docker", "run", "--rm", "-v", dockerVol,
		"sambaiz/mp4box", "-splits", strconv.Itoa(task.TaskSplit.SizePerVid/1024), // Split size in KB
		task.TaskSplit.Video.FileName)
	log.Debug(cmd.String())

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = workdir

	err = cmd.Run()
	if err != nil {
		log.Error(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	log.Infof("Processing done %s: ", out.String())

	files, err := ioutil.ReadDir(workdir)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Debugf("Reading directory %s: ", workdir)

	var tempFiles []string
	for _, f := range files {
		name := strings.Split(strings.Split(f.Name(), ".")[0], "_")[0]
		if name == strings.Split(task.TaskSplit.Video.FileName, ".")[0] {
			tempFiles = append(tempFiles, f.Name())
		}

	}

	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	for _, file := range tempFiles[1:] { // skip original file
		wg.Add(1)
		go func(fileName string, w *sync.WaitGroup) {
			filePath := fmt.Sprintf("%s/%s", workdir, fileName)
			fileReader, err := os.Open(filePath)
			if err != nil {
				log.Error(err)
				errCh <- err
				return
			}

			values := map[string]io.Reader{"file": fileReader}
			url := fmt.Sprintf("%s/upload?filename=%s", viper.GetString("fs_host"), fileName)
			log.Debugf("Sending file to %s", url)
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
			w.Done()
		}(file, &wg)
	}

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		log.Debugf("removing %s", filePath)
		err = os.Remove(filePath)
		if err != nil {
			log.Error(err)
			errCh <- err
		}
		w.Done()
	}(&wg)

	var videoList []*models.Video
	for _, file := range tempFiles[1:] {
		videoList = append(videoList, &models.Video{
			FileName: file,
		})
	}

	select {
	case err = <-errCh:
		return err
	default:
		wg.Wait()
		task.TaskDuration = time.Since(start)
		task.CompletedAt = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskSplit.SplitedVideo = videoList
		return nil
	}
}
