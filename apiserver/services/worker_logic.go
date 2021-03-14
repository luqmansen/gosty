package services

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	WorkerNew = "worker_new"
)

type workerServices struct {
	workerRepo repositories.WorkerRepository
	mb         repositories.MessageBrokerRepository
}

func NewWorkerService(
	workerRepo repositories.WorkerRepository,
	mb repositories.MessageBrokerRepository,
) WorkerService {
	return &workerServices{
		mb:         mb,
		workerRepo: workerRepo,
	}
}

func (wrk workerServices) ReadMessage() {
	log.Debugf("Starting read message from %s", WorkerNew)
	forever := make(chan bool, 1)

	newWorker := make(chan interface{})
	go wrk.mb.ReadMessage(newWorker, WorkerNew)
	go func() {
		for w := range newWorker {
			msg := w.(amqp.Delivery)
			var worker models.Worker
			err := json.Unmarshal(msg.Body, &worker)
			if err != nil {
				log.Error(err)
			}
			log.Debugf("New worker %s added", worker.WorkerPodName)
			if err := wrk.workerRepo.Upsert(&worker); err != nil {
				log.Error(err)
			}
			if err == nil {
				if err = msg.Ack(false); err != nil {
					log.Error(err)
				}
			}
		}
	}()

	<-forever
}

func (wrk workerServices) Create() error {
	panic("implement me")
}

func (wrk workerServices) Get(workerName string) models.Worker {
	panic("implement me")
}

func (wrk workerServices) GetIdle() models.Worker {
	panic("implement me")
}

func (wrk workerServices) Update(workerName string) models.Worker {
	panic("implement me")
}

func (wrk workerServices) Terminate(workerName string) error {
	panic("implement me")
}
