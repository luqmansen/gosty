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
	"strconv"
	"strings"
	"sync"
	"time"
)

func processTaskTranscodeVideo(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/worker/tmp", wd)

	inputPath := fmt.Sprintf("%s/%s", workdir, task.TaskTranscode.Video.FileName)

	origFileName := strings.Split(task.TaskTranscode.Video.FileName, ".")
	newFileName := fmt.Sprintf("%s_%s.%s", origFileName[0], task.TaskTranscode.TargetRes, origFileName[1])

	outputPath := fmt.Sprintf("%s/%s", workdir, newFileName)
	err := pkg.Download(inputPath, "http://localhost:8001/files/"+task.TaskTranscode.Video.FileName)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task id: %s", task.Id.Hex())
	outBuff := &bytes.Buffer{}
	cmd := fluentffmpeg.NewCommand("").
		InputPath(inputPath).
		OutputFormat("mp4").
		Resolution(task.TaskTranscode.TargetRes). // the default is only set aspect ration, not scaling :(
		VideoBitRate(task.TaskTranscode.TargetBitrate).
		OutputPath(outputPath).
		OutputLogs(outBuff).
		Overwrite(true).
		Build()

	//too bad this wrapper doesn't support additional arguments
	output := cmd.Args[len(cmd.Args)-1]
	cmd.Args = cmd.Args[:len(cmd.Args)-1]
	addArgs := []string{"-an",
		"-c:v", "libx264",
		"-x264opts", "keyint=24:min-keyint=24:no-scenecut",
		// https://stackoverflow.com/questions/60368162/conversion-failed-2-frames-left-in-the-queue-on-closing-ffmpeg
		"-max_muxing_queue_size", "9999",
		"-bufsize", strconv.Itoa(2 * task.TaskTranscode.TargetBitrate),
		"-vf", fmt.Sprintf("scale=-2:%s", strings.Split(task.TaskTranscode.TargetRes, "x")[1]),
		output}

	for _, v := range addArgs {
		cmd.Args = append(cmd.Args, v)
	}

	log.Debug(cmd)
	err = cmd.Run()
	if err != nil {
		log.Errorf("Transcode error: %s", err)
		out, _ := ioutil.ReadAll(outBuff)
		log.Debug(string(out))
		return err
	}

	url := fmt.Sprintf("http://localhost:8001/upload?filename=%s", newFileName)
	log.Debugf("Sending file to %s", url)

	file, _ := os.Open(outputPath)
	defer file.Close()

	values := map[string]io.Reader{"file": file}
	if err = pkg.Upload(url, values); err != nil {
		log.Error(err)
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err = os.Remove(outputPath); err != nil {
			log.Error(err)
			return
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		if err = os.Remove(inputPath); err != nil {
			log.Error(err)
			return
		}
		wg.Done()
	}()

	task.TaskDuration = time.Since(start)
	task.CompletedAt = time.Now()
	task.Status = models.TaskStatusDone
	return nil
}

func processTaskTranscodeAudio(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/worker/tmp", wd)

	inputPath := fmt.Sprintf("%s/%s", workdir, task.TaskTranscode.Video.FileName)

	origFileName := strings.Split(task.TaskTranscode.Video.FileName, ".")
	newFileName := fmt.Sprintf("%s.m4a", origFileName[0])

	outputPath := fmt.Sprintf("%s/%s", workdir, newFileName)
	err := pkg.Download(inputPath, "http://localhost:8001/files/"+task.TaskTranscode.Video.FileName)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task id: %s", task.Id.Hex())
	outBuff := &bytes.Buffer{}
	cmd := fluentffmpeg.NewCommand("").
		InputPath(inputPath).
		AudioCodec("aac").
		OutputPath(outputPath).
		OutputLogs(outBuff). // provide a io.Writer
		Overwrite(true).
		Build()

	output := cmd.Args[len(cmd.Args)-1]
	cmd.Args = cmd.Args[:len(cmd.Args)-1]
	addArgs := []string{"-vn", "-b:a", "128k", output}

	for _, v := range addArgs {
		cmd.Args = append(cmd.Args, v)
	}

	log.Debug(cmd)
	err = cmd.Run()
	if err != nil {
		log.Errorf("Transcode error: %s", err)
		out, _ := ioutil.ReadAll(outBuff)
		log.Debug(string(out))
		return err
	}

	url := fmt.Sprintf("http://localhost:8001/upload?filename=%s", newFileName)
	log.Debugf("Sending file to %s", url)

	file, _ := os.Open(outputPath)
	defer file.Close()

	values := map[string]io.Reader{"file": file}
	if err = pkg.Upload(url, values); err != nil {
		log.Error(err)
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err = os.Remove(outputPath); err != nil {
			log.Error(err)
			return
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		if err = os.Remove(inputPath); err != nil {
			log.Error(err)
			return
		}
		wg.Done()
	}()

	task.TaskDuration = time.Since(start)
	task.CompletedAt = time.Now()
	task.Status = models.TaskStatusDone
	return nil
}
