package services

import "github.com/luqmansen/gosty/apiserver/model"

type WorkerService interface {
	Create() error
	Get(workerId uint) model.Worker
	GetIdle() model.Worker
	Update(workerId uint) model.Worker
	Terminate(workerId uint) error
}
