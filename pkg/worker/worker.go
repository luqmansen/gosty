package worker

import (
	"fmt"
	"os"
	"time"

	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Services interface {
	GetWorkerInfo() *models.Worker
	//RegisterWorker will register worker to API Server
	RegisterWorker()
	GetMessageBroker() repositories.Messenger
	ProcessTaskDash(task *models.Task) error
	ProcessTaskSplit(task *models.Task) error
	ProcessTaskTranscodeVideo(task *models.Task) error
	ProcessTaskMerge(task *models.Task) error
	ProcessTaskTranscodeAudio(task *models.Task) error
}

const (
	TmpPath = "tmpworker"
)

var (
	wd      string
	workdir string
)

type Svc struct {
	messageBroker repositories.Messenger
	// TODO [#14]:  implement this storage repository
	worker *models.Worker
	config *config.Configuration
}

func NewWorkerService(mb repositories.Messenger, conf *config.Configuration) Services {
	containerHostname, _ := os.Hostname()
	wd, _ = os.Getwd()
	workdir = fmt.Sprintf("%s/%s", wd, TmpPath)

	return &Svc{
		messageBroker: mb,
		worker: &models.Worker{
			Id:            primitive.NewObjectID(),
			WorkerPodName: containerHostname,
			IpAddress:     getWorkerHost(),
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

// RegisterWorker will execute every 30 second to regularly register.
// This is current workaround to prevent worker from falsely marked as
// terminated when actually there is a network partition during
// api server check
func (s *Svc) RegisterWorker() {
	// TODO: worker status will stale on working if worker interrupted when working
	// need to add flag to worker to check whether worker is still working or not
	for {
		if w := s.GetWorkerInfo(); w != nil {
			if err := s.GetMessageBroker().Publish(w, services.WorkerNew); err != nil {
				log.Errorf("Failed to registering worker %s to apiserver, ip: %s", w.WorkerPodName, w.IpAddress)
			}
		}
		time.Sleep(30 * time.Second)
	}
}

func getWorkerHost() string {
	_, found := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if found {
		// this means we're on k8s env
		return viper.GetString("POD_IP")
	} else {
		// could be inside docker container
		containerHostname, _ := os.Hostname()
		return containerHostname
	}
}
