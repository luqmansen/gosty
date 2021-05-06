package worker

import (
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Services interface {
	GetWorkerInfo() *models.Worker
	GetMessageBroker() repositories.Messenger
	ProcessTaskDash(task *models.Task) error
	ProcessTaskSplit(task *models.Task) error
	ProcessTaskTranscodeVideo(task *models.Task) error
	ProcessTaskMerge(task *models.Task) error
	ProcessTaskTranscodeAudio(task *models.Task) error
}

const (
	TmpPath = "tmp-worker"
)

type Svc struct {
	messageBroker repositories.Messenger
	// TODO [$609358567c9cf10008f9351b]:  implement this storage repository
	storage repositories.StorageRepository
	worker  *models.Worker
	config  *config.Configuration
}

func NewWorkerService(mb repositories.Messenger, conf *config.Configuration) Services {
	return &Svc{
		messageBroker: mb,
		worker: &models.Worker{
			Id:            primitive.NewObjectID(),
			WorkerPodName: viper.GetString("HOSTNAME"),
			Status:        models.WorkerStatusReady,
			UpdatedAt:     time.Now(),
		},
		config: conf,
	}
}

func (s *Svc) GetWorkerInfo() *models.Worker {
	return s.worker
}

func (s *Svc) GetMessageBroker() repositories.Messenger {
	return s.messageBroker
}
