package services

import "github.com/luqmansen/gosty/pkg/apiserver/models"

type WorkerService interface {
	//ReadMessage will read all message from message broker
	ReadMessage()
	Create() error
	Get(workerName string) models.Worker
	GetAll() ([]*models.Worker, error)
	Update(workerName string) models.Worker
	Terminate(workerName string) error
}
