package worker

import (
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *Svc) ProcessTaskTranscodeVideo(task *models.Task) error {
	start := time.Now()

	originalFilePath := fmt.Sprintf("%s/%s", workdir, task.TaskTranscode.Video.FileName)
	origFileName := strings.Split(task.TaskTranscode.Video.FileName, ".") // remove filename full path
	transcodedFileName := fmt.Sprintf("%s_%s.%s", origFileName[0], task.TaskTranscode.TargetRes, origFileName[1])
	outputPath := fmt.Sprintf("%s/%s", workdir, transcodedFileName) // full path for transcoded file

	log.Debugf("Processing task %s,  id: %s", models.TASK_NAME_ENUM[task.Kind], task.Id.Hex())

	url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), task.TaskTranscode.Video.FileName)
	err := util.Download(originalFilePath, url)
	if err != nil {
		log.Error(err)
		return err
	}

	cmd := fluentffmpeg.NewCommand("").
		InputPath(originalFilePath).
		OutputFormat("mp4").
		//Resolution(task.TaskTranscode.TargetRes). // the default is only set aspect ration, not scaling :(
		//VideoBitRate(task.TaskTranscode.TargetBitrate).
		OutputPath(outputPath).
		Overwrite(true).
		Build()

	//too bad this wrapper doesn't support additional arguments
	output := cmd.Args[len(cmd.Args)-1]
	cmd.Args = cmd.Args[:len(cmd.Args)-1]

	// Commented to change the transcoding command same with morph, so the comparison
	// can be more accurate
	//addArgs := []string{
	//	"-c:v", "libx264",
	//	"-x264opts", "keyint=24:min-keyint=24:no-scenecut",
	//	// https://stackoverflow.com/questions/60368162/conversion-failed-2-frames-left-in-the-queue-on-closing-ffmpeg
	//	"-max_muxing_queue_size", "9999",
	//	"-bufsize", strconv.Itoa(2 * task.TaskTranscode.TargetBitrate),
	//	"-vf", fmt.Sprintf("scale=-2:%s", strings.Split(task.TaskTranscode.TargetRes, "x")[1]),
	//	"-tune", "zerolatency",
	//	//Note temporary solution: currently we copy the whole audio to same video,
	//	//since there is a problem for generating DASH with audio via cmd exec
	//	"-c:a", "copy",
	//	output}

	//This is the command is used originally by morph
	//TODO: change the transcoding argument to be configurable
	addArgs := []string{"-strict", "-2", "-s", task.TaskTranscode.TargetRes, output}

	for _, v := range addArgs {
		cmd.Args = append(cmd.Args, v)
	}

	log.Debug(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Errorf("Transcode error: %s", err)
		return err
	}

	errCh := make(chan error, 1)
	waitCh := make(chan struct{}, 1)
	doneChan := make(chan bool, 2)

	go func() {
		file, _ := os.Open(outputPath)
		values := map[string]io.Reader{"file": file}
		url = fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), transcodedFileName)
		if err = util.Upload(url, values); err != nil {
			log.Errorf("Failed to upload %s: %s", transcodedFileName, err)
			errCh <- err
		}
		doneChan <- true
	}()

	var transcodedVideoResult *models.Video
	go func() {
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

		vidRes := strings.Split(task.TaskTranscode.TargetRes, "x")
		width, _ := strconv.Atoi(vidRes[0])
		height, _ := strconv.Atoi(vidRes[1])

		transcodedVideoResult = &models.Video{
			FileName: transcodedFileName,
			Size:     size,
			Bitrate:  task.TaskTranscode.TargetBitrate,
			Duration: float32(duration),
			Width:    width,
			Height:   height,
		}

		doneChan <- true
	}()

	go func() {
		// this will block until two previous function done
		<-doneChan
		<-doneChan
		log.Debugf("removing output file %s", output)
		if err = os.Remove(outputPath); err != nil {
			log.Error(err) // Currently don't care about the error
		}
		close(waitCh)
	}()

	go func() {
		// didn't have guarantee that this function will be
		// executed before parent function exit, but most likely will
		log.Debugf("removing original file %s", originalFilePath)
		if err = os.Remove(originalFilePath); err != nil {
			log.Error(err) // Currently don't care about the error
		}
	}()

	select {
	case err := <-errCh:
		task.Status = models.TaskStatusFailed
		return err
	case <-waitCh:
		task.TaskStarted = start
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskTranscode.ResultVideo = transcodedVideoResult
		return nil
	}
}

func (s Svc) ProcessTaskTranscodeAudio(task *models.Task) error {
	start := time.Now()
	wd, _ := os.Getwd()
	workdir := fmt.Sprintf("%s/%s", wd, TmpPath)

	inputPath := fmt.Sprintf("%s/%s", workdir, task.TaskTranscode.Video.FileName)

	origFileName := strings.Split(task.TaskTranscode.Video.FileName, ".")
	newFileName := fmt.Sprintf("%s.mp3", origFileName[0])

	outputPath := fmt.Sprintf("%s/%s", workdir, newFileName)
	url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), task.TaskTranscode.Video.FileName)
	err := util.Download(inputPath, url)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task %s,  id: %s", models.TASK_NAME_ENUM[task.Kind], task.Id.Hex())

	cmd := fluentffmpeg.NewCommand("").
		InputPath(inputPath).
		AudioCodec("aac").
		OutputPath(outputPath).
		Overwrite(true).
		Build()

	output := cmd.Args[len(cmd.Args)-1]
	cmd.Args = cmd.Args[:len(cmd.Args)-1]
	addArgs := []string{"-vn", "-b:a", "128k", output}

	for _, v := range addArgs {
		cmd.Args = append(cmd.Args, v)
	}

	log.Debug(cmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Errorf("Transcode error: %s", err)
		return err
	}

	url = fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), newFileName)
	file, _ := os.Open(outputPath)
	values := map[string]io.Reader{"file": file}
	if err = util.Upload(url, values); err != nil {
		log.Error(err)
		return err
	}
	var wg sync.WaitGroup
	errCh := make(chan error)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err = os.Remove(outputPath); err != nil {
			log.Error(err)
			errCh <- err
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err = os.Remove(inputPath); err != nil {
			log.Error(err)
			errCh <- err
		}
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
