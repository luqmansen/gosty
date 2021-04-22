package services

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	WorkerNew      = "worker_new"
	WorkerAssigned = "worker_assigned"
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
	go wrk.workerStateUpdate(newWorker, "added")

	updateWorkerStatus := make(chan interface{})
	go wrk.mb.ReadMessage(updateWorkerStatus, WorkerAssigned)
	go wrk.workerStateUpdate(updateWorkerStatus, "assigned")

	<-forever
}

func (wrk workerServices) workerStateUpdate(workerQueue chan interface{}, action string) {
	for w := range workerQueue {
		msg := w.(amqp.Delivery)
		var worker models.Worker
		err := json.Unmarshal(msg.Body, &worker)
		if err != nil {
			log.Error(err)
		}
		log.Debugf("Worker %s %s", worker.WorkerPodName, action)
		if err := wrk.workerRepo.Upsert(&worker); err != nil {
			log.Error(err)
		} else {
			if err = msg.Ack(false); err != nil {
				log.Error(err)
			}
		}
	}
}

func (wrk workerServices) Create() error {
	panic("implement me")
}

func (wrk workerServices) Get(workerName string) models.Worker {
	panic("implement me")
}

func (wrk workerServices) GetAll() ([]*models.Worker, error) {
	w, err := wrk.workerRepo.GetAll(12)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return w, nil
}

func (wrk workerServices) Update(workerName string) models.Worker {
	panic("implement me")
}

func (wrk workerServices) Terminate(workerName string) error {
	panic("implement me")
}
