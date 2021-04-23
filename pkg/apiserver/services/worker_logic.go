package services

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

const (
	WorkerNew      = "worker_new"
	WorkerAssigned = "worker_assigned"
	WorkerStatus   = "worker_status"
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
	log.Debug("Starting read message from queue")
	forever := make(chan bool, 1)

	newWorker := make(chan interface{})
	go wrk.mb.ReadMessage(newWorker, WorkerNew)
	go wrk.workerStateUpdate(newWorker, "added")

	workerAssigned := make(chan interface{})
	go wrk.mb.ReadMessage(workerAssigned, WorkerAssigned)
	go wrk.workerStateUpdate(workerAssigned, "assigned")

	workerAvailable := make(chan interface{})
	go wrk.mb.ReadMessage(workerAvailable, WorkerStatus)
	go wrk.workerStateUpdate(workerAvailable, "updated")
	go wrk.workerWatcher()

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
		if err := wrk.workerRepo.Upsert(&worker); err != nil {
			log.Error(err)
		} else {
			log.Debugf("Worker %s %s", worker.WorkerPodName, action)
			if err = msg.Ack(false); err != nil {
				log.Error(err)
			}
		}
	}
}

func (wrk workerServices) workerWatcher() {
	for {
		workerList, _ := wrk.GetAll()
		for _, worker := range workerList {
			if time.Since(worker.UpdatedAt) > 4*time.Second {
				worker.Status = models.WorkerStatusTerminated
			}

			if err := wrk.workerRepo.Upsert(worker); err != nil {
				log.Error(err)
			}
		}
		time.Sleep(4 * time.Second)
	}
}

func (wrk workerServices) Create() error {
	panic("implement me")
}

func (wrk workerServices) Get(workerName string) models.Worker {
	panic("implement me")
}

func (wrk workerServices) GetAll() ([]*models.Worker, error) {
	w, err := wrk.workerRepo.GetAll(100)
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
