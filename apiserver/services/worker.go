package services

import "github.com/luqmansen/gosty/apiserver/models"

type WorkerService interface {
	//Poll message from message broker
	ReadMessage()
	Create() error
	Get(workerName string) models.Worker
	GetIdle() models.Worker
	Update(workerName string) models.Worker
	Terminate(workerName string) error
}
