package worker

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

func (s workerSvc) ProcessTaskTranscodeVideo(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/%s", wd, TmpPath)

	inputPath := fmt.Sprintf("%s/%s", workdir, task.TaskTranscode.Video.FileName)

	origFileName := strings.Split(task.TaskTranscode.Video.FileName, ".")
	newFileName := fmt.Sprintf("%s_%s.%s", origFileName[0], task.TaskTranscode.TargetRes, origFileName[1])

	outputPath := fmt.Sprintf("%s/%s", workdir, newFileName)
	url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), task.TaskTranscode.Video.FileName)
	err := pkg.Download(inputPath, url)
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
	addArgs := []string{
		"-c:v", "libx264",
		"-x264opts", "keyint=24:min-keyint=24:no-scenecut",
		// https://stackoverflow.com/questions/60368162/conversion-failed-2-frames-left-in-the-queue-on-closing-ffmpeg
		"-max_muxing_queue_size", "9999",
		"-bufsize", strconv.Itoa(2 * task.TaskTranscode.TargetBitrate),
		"-vf", fmt.Sprintf("scale=-2:%s", strings.Split(task.TaskTranscode.TargetRes, "x")[1]),
		"-tune", "zerolatency",
		//Note temporary solution: currently we copy the whole audio to same video,
		//since there is a problem for generating DASH with audio via cmd exec
		"-c:a", "copy",
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

	file, _ := os.Open(outputPath)
	values := map[string]io.Reader{"file": file}
	url = fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), newFileName)
	if err = pkg.Upload(url, values); err != nil {
		log.Error(err)
		return err
	}

	vidRes := strings.Split(task.TaskTranscode.TargetRes, "x")
	width, _ := strconv.Atoi(vidRes[0])
	height, _ := strconv.Atoi(vidRes[1])

	probeResult, err := fluentffmpeg.Probe(outputPath)

	format := probeResult["format"].(map[string]interface{})
	duration, err := strconv.ParseFloat(format["duration"].(string), 32)
	if err != nil {
		log.Error(err)
	}
	size, err := strconv.ParseInt(format["size"].(string), 10, 32)
	if err != nil {
		log.Error(err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error)

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		if err = os.Remove(outputPath); err != nil {
			log.Error(err)
			errCh <- err
		}
		w.Done()
	}(&wg)

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		if err = os.Remove(inputPath); err != nil {
			log.Error(err)
			errCh <- err
		}
		w.Done()
	}(&wg)

	result := &models.Video{
		FileName: newFileName,
		Size:     size,
		Bitrate:  task.TaskTranscode.TargetBitrate,
		Duration: float32(duration),
		Width:    width,
		Height:   height,
	}

	select {
	case err = <-errCh:
		return err
	default:
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskTranscode.ResultVideo = result
		return nil
	}
}

func (s workerSvc) ProcessTaskTranscodeAudio(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/%s", wd, TmpPath)

	inputPath := fmt.Sprintf("%s/%s", workdir, task.TaskTranscode.Video.FileName)

	origFileName := strings.Split(task.TaskTranscode.Video.FileName, ".")
	newFileName := fmt.Sprintf("%s.mp3", origFileName[0])

	outputPath := fmt.Sprintf("%s/%s", workdir, newFileName)
	url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), task.TaskTranscode.Video.FileName)
	err := pkg.Download(inputPath, url)
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

	url = fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), newFileName)
	file, _ := os.Open(outputPath)
	values := map[string]io.Reader{"file": file}
	if err = pkg.Upload(url, values); err != nil {
		log.Error(err)
		return err
	}
	var wg sync.WaitGroup
	errCh := make(chan error)

	wg.Add(1)
	go func() {
		if err = os.Remove(outputPath); err != nil {
			log.Error(err)
			errCh <- err
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		if err = os.Remove(inputPath); err != nil {
			log.Error(err)
			errCh <- err
		}
		wg.Done()
	}()

	select {
	case err = <-errCh:
		return err
	default:
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskTranscode.ResultAudio = &models.Audio{
			FileName: newFileName,
			Bitrate:  128000, // TODO: add bitrate transcode variation
		}
		return nil
	}
}
