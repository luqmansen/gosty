package worker

import (
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/pkg/util/config"
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
	messageBroker repositories.MessageBrokerRepository
	worker        *models.Worker
	config        *config.Configuration
}

func NewWorkerService(mb repositories.MessageBrokerRepository, conf *config.Configuration) Services {
	return &workerSvc{
		messageBroker: mb,
		worker: &models.Worker{
			WorkerPodName: viper.GetString("HOSTNAME"),
			Status:        models.WorkerStatusIdle,
			UpdatedAt:     time.Now(),
		},
		config: conf,
	}
}

func (s workerSvc) GetWorkerInfo() *models.Worker {
	return s.worker
}
