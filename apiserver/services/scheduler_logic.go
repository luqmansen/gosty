package services

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"math"
	"time"
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
			json.Unmarshal(m.Body, &task)
			log.Debugf("Updating finished task %s", task.Id.String())
			err := s.repo.Update(&task)
			if err != nil {
				log.Error(err)
			}
			err = m.Ack(false)
			if err != nil {
				log.Error(err)
			}
		}
	}()

	<-forever
}

func (s schedulerServices) CreateSplitTask(video *models.Video) error {
	//split by size in Byte
	var sizePerVid int
	var sizeLeft int
	var minSize = 1024 << 10 // 10 MB

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
		Status:       models.TaskQueued,
		CompletedAt:  time.Time{},
		TaskDuration: 0,
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

func (s schedulerServices) DeleteTask(taskId string) error {
	panic("implement me")
}

func (s schedulerServices) CreateTranscodeTask(video *models.Video) error {
	panic("implement me")
}
