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
	GetMessageBroker() repositories.MessageBrokerRepository
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
	messageBroker repositories.MessageBrokerRepository
	//todo: implement this storage repository
	storage repositories.StorageRepository
	worker  *models.Worker
	config  *config.Configuration
}

func NewWorkerService(mb repositories.MessageBrokerRepository, conf *config.Configuration) Services {
	return &Svc{
		messageBroker: mb,
		worker: &models.Worker{
			Id:            primitive.NewObjectID(),
			WorkerPodName: viper.GetString("HOSTNAME"),
			Status:        models.WorkerStatusIdle,
			UpdatedAt:     time.Now(),
		},
		config: conf,
	}
}

func (s *Svc) GetWorkerInfo() *models.Worker {
	return s.worker
}

func (s *Svc) GetMessageBroker() repositories.MessageBrokerRepository {
	return s.messageBroker
}
