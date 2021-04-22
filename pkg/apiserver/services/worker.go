package services

import "github.com/luqmansen/gosty/pkg/apiserver/models"

type WorkerService interface {
	//Poll message from message broker
	ReadMessage()
	Create() error
	Get(workerName string) models.Worker
	GetAll() ([]*models.Worker, error)
	Update(workerName string) models.Worker
	Terminate(workerName string) error
}
