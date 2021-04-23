package worker

import (
	"bytes"
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s Svc) ProcessTaskSplit(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/%s", wd, TmpPath)

	filePath := fmt.Sprintf("%s/%s", workdir, task.TaskSplit.Video.FileName)
	url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), task.TaskSplit.Video.FileName)
	err := util.Download(filePath, url)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task id: %s", task.Id.Hex())
	//dockerVol := fmt.Sprintf("%s:/work/", workdir)
	cmd := exec.Command(
		"MP4Box", "-splits", strconv.Itoa(int(task.TaskSplit.SizePerVid)/1024), // Split size in KB
		task.TaskSplit.Video.FileName)

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
	if len(tempFiles) == 0 {
		return errors.New("worker.ProcessTaskSplit: no files found, something wrong")
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
			url := fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), fileName)
			if err = util.Upload(url, values); err != nil {
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
			Size:     task.OriginVideo.Size,
			Bitrate:  task.OriginVideo.Bitrate,
			Duration: task.OriginVideo.Duration,
			Width:    task.OriginVideo.Width,
			Height:   task.OriginVideo.Height,
		})
	}

	select {
	case err = <-errCh:
		return err
	default:
		wg.Wait()
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskSplit.SplitedVideo = videoList
		return nil
	}
}