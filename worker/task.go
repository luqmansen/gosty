package worker

import (
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
)

type Services interface {
	ProcessTaskDash(task *models.Task) error
	ProcessTaskSplit(task *models.Task) error
	ProcessTaskTranscodeVideo(task *models.Task) error
	ProcessTaskTranscodeAudio(task *models.Task) error
}

type taskSvc struct {
	mb repositories.MessageBrokerRepository
}

func NewWorkerService(mb repositories.MessageBrokerRepository) Services {
	return &taskSvc{
		mb: mb,
	}
}
