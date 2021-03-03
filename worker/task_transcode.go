package main

import (
	"bytes"
	"fmt"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/pkg"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"time"
)

func processTaskTranscode(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/worker/tmp", wd)

	inputPath := fmt.Sprintf("%s/%s", workdir, task.TaskTranscode.Video.FileName)
	newFileName := fmt.Sprintf("%s_%s", task.TaskTranscode.Video.FileName, task.TaskTranscode.TargetRes)
	outputPath := fmt.Sprintf("%s/%s", workdir, newFileName)
	err := pkg.Download(inputPath, "http://localhost:8001/files/"+task.TaskTranscode.Video.FileName)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task id: %s", task.Id.Hex())
	outBuff := &bytes.Buffer{}
	err = fluentffmpeg.NewCommand("").
		InputPath(inputPath).
		OutputFormat("mp4").
		Resolution(task.TaskTranscode.TargetRes).
		VideoBitRate(task.TaskTranscode.TargetBitrate).
		OutputPath(outputPath).
		OutputLogs(outBuff).
		Overwrite(true).
		Run()
	if err != nil {
		log.Errorf("Transcode error: %s", err)
	}
	out, _ := ioutil.ReadAll(outBuff)
	log.Debug(string(out))

	url := fmt.Sprintf("http://localhost:8001/upload?filename=%s", newFileName)
	log.Debugf("Sending file to %s", url)

	file, _ := os.Open(outputPath)
	defer file.Close()

	values := map[string]io.Reader{"video": file}
	if err = pkg.Upload(url, values); err != nil {
		log.Error(err)
		return err
	}
	if err = os.Remove(outputPath); err != nil {
		log.Error(err)
		return err
	}

	task.TaskDuration = time.Since(start)
	task.CompletedAt = time.Now()
	task.Status = models.TaskStatusDone
	return nil
}
