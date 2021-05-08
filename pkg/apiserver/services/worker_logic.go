package services

import (
	"encoding/json"
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"net/http"
	"sync"
	"time"
)

func NewWorkerService(
	workerRepo repositories.WorkerRepository,
	mb repositories.Messenger,
	sse *sse.Server,
) WorkerService {
	return &workerServices{
		mb:         mb,
		workerRepo: workerRepo,
		sse:        sse,
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
		// TODO [#10]:  add bulk upsert to reduce write
		if err := wrk.workerRepo.Upsert(&worker); err != nil {
			log.Errorf("Worker %s failed to %s, err %s", worker.WorkerPodName, action, err)
		} else {
			if action == "added" {
				log.Infof("Worker %s, IP: %s added", worker.WorkerPodName, worker.IpAddress)
			}
			if err = msg.Ack(false); err != nil {
				log.Error(err)
			}
			wrk.publishWorkerEvent()
		}
	}
}

func (wrk workerServices) workerWatcher() {
	var wg sync.WaitGroup
	var workerRetryAttempt sync.Map
	for {
		// TODO [#18]: cache worker.GetAll() function
		// Cache get all worker function and expire the cache
		// when new worker registered
		workerList, _ := wrk.GetAll()
		for _, worker := range workerList {
			wg.Add(1)
			go func(w *models.Worker) {
				defer wg.Done()

				if w.Status == models.WorkerStatusTerminated {
					return
				}

				retry, _ := workerRetryAttempt.LoadOrStore(w.WorkerPodName, 0)

				if retry.(int) > 5 {
					// TODO [#19]: Also remove worker from db if retry failed > 5
					log.Infof("Ping to to ip %s worker %s failed >5 times, setting worker to terminated...", w.IpAddress, w.WorkerPodName)
					w.Status = models.WorkerStatusTerminated
					w.WorkingOn = ""
					if err := wrk.workerRepo.Upsert(w); err != nil {
						log.Errorf("Failed to upsert worker %s, err: %s", w.WorkerPodName, err)
					}

					return
				}
				// 8087 is worker's health check port
				client := http.Client{Timeout: 4 * time.Second}
				resp, err := client.Get(fmt.Sprintf("http://%s:8087/live", w.IpAddress))
				if err != nil {
					log.Errorf("Failed to ping ip %s worker %s on attempt no %d, error: %s",
						w.IpAddress, w.WorkerPodName, retry, err)
					workerRetryAttempt.Store(w.WorkerPodName, retry.(int)+1)
				}

				if resp != nil && resp.StatusCode == http.StatusOK {
					w.Status = models.WorkerStatusReady
					//reset retry attempt
					workerRetryAttempt.Store(w.WorkerPodName, 0)
				}
				w.UpdatedAt = time.Now()

				if err := wrk.workerRepo.Upsert(w); err != nil {
					log.Errorf("Failed to upsert worker %s, err: %s", w.WorkerPodName, err)
				}
			}(worker)
		}
		time.Sleep(4 * time.Second)
		wg.Wait()
		wrk.publishWorkerEvent()
	}
}

func (wrk workerServices) publishWorkerEvent() {
	allWorker, err := wrk.GetAll()
	if err != nil {
		log.Error(err)
	}

	resp, err := json.Marshal(allWorker)
	if err != nil {
		log.Error(err)
	}
	// TODO [#11]:  only publish 1 data for every worker update
	// add parameter to this function and only publish worker changes
	// instead off all worker result
	wrk.sse.Publish(WorkerHTTPEventStream, &sse.Event{
		Data: resp,
	})
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
