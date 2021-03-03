package services

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"math"
	"sync"
)

const (
	TaskNew      = "task_new"
	TaskFinished = "task_finished"
)

type schedulerServices struct {
	repo repositories.TaskRepository
	mb   repositories.MessageBrokerRepository
}

func NewSchedulerService(repo repositories.TaskRepository, mb repositories.MessageBrokerRepository) SchedulerService {
	return &schedulerServices{
		repo: repo,
		mb:   mb,
	}
}

func (s schedulerServices) ReadMessages() {
	log.Debugf("Starting read message from %s", TaskFinished)
	finishedTask := make(chan interface{})
	forever := make(chan bool, 1)

	go s.mb.ReadMessage(finishedTask, TaskFinished)
	go func() {
		for msg := range finishedTask {
			m := msg.(amqp.Delivery)
			var task models.Task
			err := json.Unmarshal(m.Body, &task)
			if err != nil {
				log.Error(err)
			}

			log.Debugf("Updating finished task %s", task.Id.String())
			err = s.repo.Update(&task)
			if err != nil {
				log.Error(err)
			}
			err = m.Ack(false)
			if err != nil {
				log.Error(err)
			}
			if task.Kind == models.TaskSplit {
				s.createTranscodeTaskFromSplitTask(&task)
			} else if task.Kind == models.TaskTranscode {
				//	check if previously is split task, then merge,
				//	else, create dash task
			}

		}
	}()

	<-forever
}

func (s schedulerServices) createTranscodeTaskFromSplitTask(task *models.Task) {
	var wg sync.WaitGroup
	for _, vid := range task.TaskSplit.VideoList {
		wg.Add(1)
		go func(v models.Video, w *sync.WaitGroup) {
			err := s.CreateTranscodeTask(&v)
			if err != nil {
				log.Error(err)
			}
			w.Done()
		}(vid, &wg)
	}

	wg.Wait()
}

func (s schedulerServices) CreateSplitTask(video *models.Video) error {
	//split by size in Byte
	var sizePerVid int
	var sizeLeft int
	var minSize = 10240 << 10 // 10 MB

	// if video size less than min file size, forward to transcode task
	if video.Size < minSize {
		err := s.CreateTranscodeTask(video)
		if err != nil {
			return err
		}
		return nil
	} else {
		//split per 10 MB files
		sizePerVid = minSize
		sizeLeft = video.Size % minSize
	}

	task := models.Task{
		Kind: models.TaskSplit,
		TaskSplit: models.SplitTask{
			Video:       *video,
			TargetChunk: int(math.Ceil(float64(video.Size) / float64(minSize))),
			SizePerVid:  sizePerVid,
			SizeLeft:    sizeLeft,
		},
		Status: models.TaskQueued,
	}
	//save to db
	err := s.repo.Add(&task)
	if err != nil {
		log.Fatal(err)
	}
	//publish
	err = s.mb.Publish(task, TaskNew)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// Transcode video input to all representation, each of video repr as task
func (s schedulerServices) CreateTranscodeTask(video *models.Video) error {

	target := []map[string]interface{}{
		{"res": "640x360", "br": 400000},
		{"res": "960x540", "br": 800000},
		{"res": "1280x720", "br": 1500000},
	}
	var taskList []*models.Task

	for _, t := range target {
		taskList = append(taskList, &models.Task{
			Kind: models.TaskTranscode,
			TaskTranscode: models.TranscodeTask{
				Video:          *video,
				TargetRes:      t["res"].(string),
				TargetBitrate:  t["br"].(int),
				TargetEncoding: "",
			},
			Status: models.TaskQueued,
		})
	}
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	for _, task := range taskList {
		wg.Add(1)
		go func(t *models.Task, w *sync.WaitGroup) {
			err := s.repo.Add(t)
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

func (s schedulerServices) DeleteTask(taskId string) error {
	panic("implement me")
}
