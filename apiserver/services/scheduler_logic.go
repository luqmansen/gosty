package services

import (
	"encoding/json"
	"fmt"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"math"
	"strings"
	"sync"
	"time"
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

func NewSchedulerService(
	taskRepo repositories.TaskRepository,
	videoRepo repositories.VideoRepository,
	mb repositories.MessageBrokerRepository,
) SchedulerService {
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
	//TODO: Refactor this repetitive message ack
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
				// get by file name
				//TODO : Find one and update
				toUpdate, err := s.videoRepo.GetOneByName(task.TaskTranscode.Video.FileName)
				if err != nil {
					log.Error(err)
					break
				}

				if task.TaskTranscode.TranscodeType == models.TranscodeAudio {
					toUpdate.Audio = task.TaskTranscode.ResultAudio

				} else {
					toUpdate.Video = append(toUpdate.Video, task.TaskTranscode.ResultVideo)
				}
				if err = s.videoRepo.Update(toUpdate); err != nil {
					log.Error(err)
					break
				}

				//	check if previously is split task, merge,
				if task.PrevTask == models.TaskSplit {
					s.CreateMergeTask(&task)
				} else { // must be a video with small size (prevTask == TaskNew)
					err = s.CreateDashTask(task.TaskTranscode.Video)
					if err != nil {
						log.Error(errors.Wrap(err, "services.Scheduler.ReadMessages"))
						break
					}
				}

				err = msg.Ack(false)
				if err != nil {
					log.Error(err)
					break
				}
			case models.TaskMerge:
				if err := s.CreateMergeTask(&task); err != nil {
					log.Error(err)
					break
				}

			case models.TaskDash:
				file := strings.Split(task.TaskDash.ListVideo[0].FileName, "_")
				toUpdate, err := s.videoRepo.GetOneByName(fmt.Sprintf("%s.mp4", file[0]))
				if err != nil {
					log.Error(err)
				}
				toUpdate.DashFile = task.TaskDash.ResultDash
				if err = s.videoRepo.Update(toUpdate); err != nil {
					log.Error(err)
				}
				err = msg.Ack(false)
				if err != nil {
					log.Error(err)
				}

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

//For transcode audio
//Currently audio transcoding is disabled, original audio is embedded on audio,
//see note on task_transcode line ~60
func (s schedulerServices) createTranscodeAudioTask(video *models.Video) error {
	task := models.Task{
		Kind: models.TaskTranscode,
		TaskTranscode: &models.TranscodeTask{
			TranscodeType: models.TranscodeAudio,
			Video:         video,
		},
		Status:        models.TaskQueued,
		TaskSubmitted: time.Now(),
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
	var minSize int64 = 1024000 << 10 // 10 MB

	// if video size less than min file size, forward to transcode task
	if video.Size < minSize {
		err := s.CreateTranscodeTask(video)
		if err != nil {
			return err
		}

		//if err := s.createTranscodeAudioTask(video); err != nil {
		//	log.Error(err)
		//	return err
		//}
		return nil
	} else {
		//split per 10 MB files
		sizePerVid = minSize
		sizeLeft = video.Size % minSize
	}

	//if err := s.createTranscodeAudioTask(video); err != nil {
	//	log.Error(err)
	//	return err
	//}

	task := models.Task{
		Kind: models.TaskSplit,
		TaskSplit: &models.SplitTask{
			Video:       video,
			TargetChunk: int(math.Ceil(float64(video.Size) / float64(minSize))),
			SizePerVid:  sizePerVid,
			SizeLeft:    sizeLeft,
		},
		PrevTask:      models.TaskNew,
		Status:        models.TaskQueued,
		TaskSubmitted: time.Now(),
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
// this function might be transcoding a part video (video with _00X name)
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

	//TODO: only choose target res below the original video
	target := []map[string]interface{}{
		{"res": "256x144", "br": 80_000},
		//{"res": "426x240", "br": 300_000},
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
			PrevTask:      prevTask,
			Status:        models.TaskQueued,
			TaskSubmitted: time.Now(),
		})
	}

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	for _, task := range taskList {
		wg.Add(1)
		go func(t *models.Task) {
			wg.Done()

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

		}(task)

	}
	select {
	case err := <-errChan:
		return err
	default:
		wg.Wait()
		return nil
	}
}

func (s schedulerServices) CreateDashTask(video *models.Video) error {
	// get all video resolution and audio
	video, err := s.videoRepo.GetOneByName(video.FileName)
	if err != nil {
		log.Error(errors.Wrap(err, "services.Scheduler.CreateDashTask"))
		return err
	}

	//Disable this, see not on worker/task_transcode ~60
	//check if audio already transcoded
	//if video.Audio == nil{
	//	log.Debug("Audio still empty")
	//	return nil
	//}
	//Todo: check of available video representation
	if len(video.Video) != 1 { // number of available video representation
		log.Debug("Video transcoding haven't finished")
		return nil
	}

	task := &models.Task{
		Kind: models.TaskDash,
		TaskDash: &models.DashTask{
			ListVideo: video.Video,
			ListAudio: []*models.Audio{video.Audio},
		},
		Status:        models.TaskQueued,
		TaskSubmitted: time.Now(),
	}

	err = s.taskRepo.Add(task)
	if err != nil {
		log.Error(errors.Wrap(err, "services.Scheduler.CreateDashTask"))
		return err
	}

	err = s.mb.Publish(task, TaskNew)
	if err != nil {
		log.Error(errors.Wrap(err, "services.Scheduler.CreateDashTask"))
		return err
	}

	return nil
}

// Merge task previously must be a split task, so to task parameter is
// task with TaskTranscode struct filled
func (s schedulerServices) CreateMergeTask(task *models.Task) error {
	panic("implement me")
}

func (s schedulerServices) DeleteTask(taskId string) error {
	panic("implement me")
}
