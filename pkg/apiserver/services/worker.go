package services

import (
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/patrickmn/go-cache"
	"github.com/r3labs/sse/v2"
)

const (
	WorkerNew      = "worker_new"
	WorkerAssigned = "worker_assigned"
	WorkerStatus   = "worker_status"
)

const (
	WorkerHTTPEventStream = "worker"
)

type workerServices struct {
	workerRepo repositories.WorkerRepository
	mb         repositories.Messenger
	sse        *sse.Server
	cache      *cache.Cache
}

type WorkerService interface {
	//ReadMessage will read all message from message broker
	ReadMessage()
	Create() error
	Get(workerName string) models.Worker
	GetAll() ([]*models.Worker, error)
	Update(workerName string) models.Worker
	Terminate(workerName string) error
}
