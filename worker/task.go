package worker

import (
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/spf13/viper"
	"time"
)

type Services interface {
	GetWorkerInfo() *models.Worker
	ProcessTaskDash(task *models.Task) error
	ProcessTaskSplit(task *models.Task) error
	ProcessTaskTranscodeVideo(task *models.Task) error
	ProcessTaskTranscodeAudio(task *models.Task) error
}

type workerSvc struct {
	mb repositories.MessageBrokerRepository
	w  models.Worker
}

func NewWorkerService(mb repositories.MessageBrokerRepository) Services {
	return &workerSvc{
		mb: mb,
		w: models.Worker{
			WorkerPodName: viper.GetString("hostname"),
			Status:        models.WorkerStatusIdle,
			UpdatedAt:     time.Now(),
		},
	}
}

func (s workerSvc) GetWorkerInfo() *models.Worker {
	return &s.w
}
