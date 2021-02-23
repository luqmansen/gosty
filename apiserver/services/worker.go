package services

import "github.com/luqmansen/gosty/apiserver/models"

type WorkerService interface {
	Create() error
	Get(workerId uint) models.Worker
	GetIdle() models.Worker
	Update(workerId uint) models.Worker
	Terminate(workerId uint) error
}
