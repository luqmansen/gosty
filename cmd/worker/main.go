package main

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	"github.com/luqmansen/gosty/pkg/worker"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
	"time"
)

type svc struct {
	worker.Services
}

func main() {
	cfg := config.LoadConfig(".")
	forever := make(chan bool)

	mb := rabbitmq.NewRepository(cfg.MessageBroker.GetMessageBrokerUri())
	workerSvc := worker.NewWorkerService(mb, cfg)

	go func() {
		if err := mb.Publish(workerSvc.GetWorkerInfo(), services.WorkerNew); err != nil {
			log.Error(err)
		}
	}()

	//todo: this initiation should be handled by storage service
	if _, err := os.Stat(worker.TmpPath); os.IsNotExist(err) {
		err = os.Mkdir(worker.TmpPath, 0700)
		if err != nil {
			log.Error(err)
		}
	}

	log.Infof("Worker %s started", workerSvc.GetWorkerInfo().WorkerPodName)

	newTaskData := make(chan interface{})
	defer close(newTaskData)
	go mb.ReadMessage(newTaskData, services.MessageBrokerQueueTaskNew)

	w := svc{workerSvc}
	go w.processNewTask(newTaskData)

	go worker.InitHealthCheck(cfg)
	log.Printf("Worker running. To exit press CTRL+C")
	<-forever

}

func (wrk *svc) processNewTask(newTaskData chan interface{}) {
	//Todo: refactor ack and publish part of this loop
	for t := range newTaskData {
		msg := t.(amqp.Delivery)
		var task models.Task
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			log.Error(err)
		}
		wrk.notifyApiServer(task)
		switch taskKind := task.Kind; taskKind {
		case models.TaskSplit:
			err = wrk.ProcessTaskSplit(&task)
			if err != nil {
				log.Error(err)
			}
			if err == nil {
				if err = msg.Ack(false); err != nil {
					log.Error(err)
				}
				if err = wrk.GetMessageBroker().Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
					log.Error(err)
				}
				wrk.notifyApiServer(models.Task{})
			}

		case models.TaskTranscode:
			switch txType := task.TaskTranscode.TranscodeType; txType {
			case models.TranscodeVideo:
				err = wrk.ProcessTaskTranscodeVideo(&task)
				if err != nil {
					log.Error(err)
				}
				if err == nil {
					if err = msg.Ack(false); err != nil {
						log.Error(err)
					}
					if err = wrk.GetMessageBroker().Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
						log.Error(err)
					}
					wrk.notifyApiServer(models.Task{})
				}

			case models.TranscodeAudio:
				err = wrk.ProcessTaskTranscodeAudio(&task)
				if err != nil {
					log.Error(err)
				}
				if err == nil {
					if err = msg.Ack(false); err != nil {
						log.Error(err)
					}
					if err = wrk.GetMessageBroker().Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
						log.Error(err)
					}
					wrk.notifyApiServer(models.Task{})
				}

			}
		case models.TaskDash:
			err = wrk.ProcessTaskDash(&task)
			if err != nil {
				log.Error(err)
			}
			if err == nil {
				if err = msg.Ack(false); err != nil {
					log.Error(err)
				}
				if err = wrk.GetMessageBroker().Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
					log.Error(err)
				}
				wrk.notifyApiServer(models.Task{})
			}

		case models.TaskMerge:
			err = wrk.ProcessTaskMerge(&task)
			if err != nil {
				log.Error(err)
			}
			if err == nil {
				if err = msg.Ack(false); err != nil {
					log.Error(err)
				}
				if err = wrk.GetMessageBroker().Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
					log.Error(err)
				}
				wrk.notifyApiServer(models.Task{})

			}
		default:
			wrk.notifyApiServer(models.Task{})
			log.Error("No task kind found")
			if err = msg.Nack(false, true); err != nil {
				log.Error(err)
			}
		}
	}
}

func (wrk *svc) notifyApiServer(task models.Task) {
	w := wrk.GetWorkerInfo()
	w.UpdatedAt = time.Now()

	//check if task is empty
	if task == (models.Task{}) {
		w.Status = models.WorkerStatusIdle
		w.WorkingOn = ""

	} else {
		w.Status = models.WorkerStatusWorking
		w.WorkingOn = task.Id.Hex()
	}

	if err := wrk.GetMessageBroker().Publish(w, services.WorkerAssigned); err != nil {
		log.Error(err)
	}
}
