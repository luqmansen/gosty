package services

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"math"
	"strings"
	"sync"
)

const (
	TaskNew      = "task_new"
	TaskFinished = "task_finished"
)

type schedulerServices struct {
	taskRepo  repositories.TaskRepository
	videoRepo repositories.VideoRepository
	mb        repositories.MessageBrokerRepository
}

func NewSchedulerService(taskRepo repositories.TaskRepository, videoRepo repositories.VideoRepository, mb repositories.MessageBrokerRepository) SchedulerService {
	return &schedulerServices{
		taskRepo:  taskRepo,
		videoRepo: videoRepo,
		mb:        mb,
	}
}

func (s schedulerServices) ReadMessages() {
	log.Debugf("Starting read message from %s", TaskFinished)
	finishedTask := make(chan interface{})
	forever := make(chan bool, 1)

	go s.mb.ReadMessage(finishedTask, TaskFinished)
	go func() {
		for t := range finishedTask {
			msg := t.(amqp.Delivery)
			var task models.Task
			err := json.Unmarshal(msg.Body, &task)
			if err != nil {
				log.Error(err)
			}

			log.Debugf("Updating finished task %s", task.Id.String())
			err = s.taskRepo.Update(&task)
			if err != nil {
				log.Error(err)
			}

			switch taskKind := task.Kind; taskKind {
			case models.TaskSplit:
				s.createTranscodeTaskFromSplitTask(&task)

			case models.TaskTranscode:
				//	check if previously is split task, merge,
				//if task.Pre	//if err := s.videoRepo.AddMany(task.TaskSplit.SplitedVideo); err != nil{
				//				//	log.Error(err)
				//				//}vTask == models.TaskSplit {
				//
				//} else {
				//	s.CreateDashTask(task.TaskTranscode.Video)
				//}
				//	else, create dash task

				// get by file name
				//TODO : Find one and update
				toUpdate, err := s.videoRepo.GetOneByName(task.TaskTranscode.Video.FileName)
				if err != nil {
					log.Error(err)
				}

				if task.TaskTranscode.TranscodeType == models.TranscodeAudio {
					toUpdate.Audio = task.TaskTranscode.ResultAudio

				} else {
					toUpdate.Video = append(toUpdate.Video, task.TaskTranscode.ResultVideo)
				}
				if err = s.videoRepo.Update(toUpdate); err != nil {
					log.Error(err)
				}

				err = msg.Ack(false)
				if err != nil {
					log.Error(err)
				}
			case models.TaskMerge:
				//	create dash task
			}

		}
	}()

	<-forever
}

func (s schedulerServices) createTranscodeTaskFromSplitTask(task *models.Task) {
	var wg sync.WaitGroup
	for _, vid := range task.TaskSplit.SplitedVideo {
		wg.Add(1)
		go func(v *models.Video, w *sync.WaitGroup) {
			err := s.CreateTranscodeTask(v)
			if err != nil {
				log.Error(err)
			}
			w.Done()
		}(vid, &wg)
	}

	wg.Wait()
}

func (s schedulerServices) createTranscodeAudioTask(video *models.Video) error {
	task := models.Task{
		Kind: models.TaskTranscode,
		TaskTranscode: &models.TranscodeTask{
			TranscodeType: models.TranscodeAudio,
			Video:         video,
		},
		Status: models.TaskQueued,
	}
	err := s.taskRepo.Add(&task)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = s.mb.Publish(task, TaskNew)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (s schedulerServices) CreateSplitTask(video *models.Video) error {
	//split by size in Byte
	var sizePerVid int64
	var sizeLeft int64
	var minSize int64 = 10240 << 10 // 10 MB

	// if video size less than min file size, forward to transcode task
	if video.Size < minSize {
		err := s.CreateTranscodeTask(video)
		if err != nil {
			return err
		}
		if err := s.createTranscodeAudioTask(video); err != nil {
			log.Error(err)
			return err
		}
		return nil
	} else {
		//split per 10 MB files
		sizePerVid = minSize
		sizeLeft = video.Size % minSize
	}

	// Must transcode audio, else the video will have no audio
	if err := s.createTranscodeAudioTask(video); err != nil {
		log.Error(err)
		return err
	}

	task := models.Task{
		Kind: models.TaskSplit,
		TaskSplit: &models.SplitTask{
			Video:       video,
			TargetChunk: int(math.Ceil(float64(video.Size) / float64(minSize))),
			SizePerVid:  sizePerVid,
			SizeLeft:    sizeLeft,
		},
		PrevTask: models.TaskNew,
		Status:   models.TaskQueued,
	}

	err := s.taskRepo.Add(&task)
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = s.mb.Publish(task, TaskNew)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

// Transcode video input to all representation, each of video represent as task
func (s schedulerServices) CreateTranscodeTask(video *models.Video) error {
	// check if previously from task split
	// file name on task split contain random-name_001
	fileName := strings.Split(video.FileName, "_")
	var prevTask models.TaskKind
	if len(fileName) == 1 {
		prevTask = models.TaskNew
	} else {
		prevTask = models.TaskSplit
	}

	target := []map[string]interface{}{
		{"res": "256x144", "br": 80_000},
		{"res": "426x240", "br": 300_000},
		//{"res": "640x360", "br": 400_000},
		//{"res": "854x480", "br": 500_000},
		//{"res": "1280x720", "br": 1_500_000},
		//{"res": "1920x1080", "br": 3_000_000},
		//{"res": "2560x1440", "br": 6_000_000},
		//{"res": "3840x2160", "br": 13_000_000},
		//{"res": "7680x4320", "br": 20_000_000},
	}
	var taskList []*models.Task

	for _, t := range target {
		taskList = append(taskList, &models.Task{
			Kind: models.TaskTranscode,
			TaskTranscode: &models.TranscodeTask{
				TranscodeType:  models.TranscodeVideo,
				Video:          video,
				TargetRes:      t["res"].(string),
				TargetBitrate:  t["br"].(int),
				TargetEncoding: "",
			},
			PrevTask: prevTask,
			Status:   models.TaskQueued,
		})
	}
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for _, task := range taskList {
		wg.Add(1)
		go func(t *models.Task, w *sync.WaitGroup) {
			err := s.taskRepo.Add(t)
			if err != nil {
				log.Error(err)
				errChan <- err
				return
			}
			err = s.mb.Publish(t, TaskNew)
			if err != nil {
				log.Error(err)
				errChan <- err
				return
			}
			wg.Done()
		}(task, &wg)

	}
	select {
	case err := <-errChan:
		return err
	default:
		wg.Wait()
		return nil
	}
}

func (s schedulerServices) CreateMergeTask(video *models.Video) error {
	panic("implement me")
}

func (s schedulerServices) CreateDashTask(video *models.Video) error {
	// get all video resolution and audio
	//videoList := s.videoRepo.Find(video.FileName)

	return nil
}

func (s schedulerServices) DeleteTask(taskId string) error {
	panic("implement me")
}
