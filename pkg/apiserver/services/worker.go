package services

import (
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/patrickmn/go-cache"
	"github.com/r3labs/sse/v2"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/client-go/kubernetes"
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
	k8sClient  *kubernetes.Clientset
}

type WorkerService interface {
	//ReadMessage will read all message from message broker
	ReadMessage()
	Scale(replicaNum int32) (*autoscalingv1.Scale, error)
	Get(workerName string) models.Worker
	GetAll() ([]*models.Worker, error)
	Update(workerName string) models.Worker
	Terminate(workerName string) error
}
