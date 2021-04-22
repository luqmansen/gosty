package main

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	"github.com/luqmansen/gosty/pkg/worker"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
	"time"
)

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
	go processNewTask(newTaskData, workerSvc, mb)

	go worker.InitHealthCheck(cfg)
	log.Printf("Worker running. To exit press CTRL+C")
	<-forever

}

func processNewTask(newTaskData chan interface{}, workerSvc worker.Services, mq repositories.MessageBrokerRepository) {
	//Todo: refactor ack and publish part of this loop
	for t := range newTaskData {
		msg := t.(amqp.Delivery)
		var task models.Task
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			log.Error(err)
		}

		switch taskKind := task.Kind; taskKind {
		case models.TaskSplit:
			notifyApiServer(mq, workerSvc, task)
			err = workerSvc.ProcessTaskSplit(&task)
			if err != nil {
				log.Error(err)
			}
			if err == nil {
				if err = msg.Ack(false); err != nil {
					log.Error(err)
				}
				if err = mq.Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
					log.Error(err)
				}
				notifyApiServer(mq, workerSvc, models.Task{})
			}

		case models.TaskTranscode:
			notifyApiServer(mq, workerSvc, task)
			switch txType := task.TaskTranscode.TranscodeType; txType {
			case models.TranscodeVideo:
				notifyApiServer(mq, workerSvc, task)
				err = workerSvc.ProcessTaskTranscodeVideo(&task)
				if err != nil {
					log.Error(err)
				}
				if err == nil {
					if err = msg.Ack(false); err != nil {
						log.Error(err)
					}
					if err = mq.Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
						log.Error(err)
					}
					notifyApiServer(mq, workerSvc, models.Task{})
				}

			case models.TranscodeAudio:
				notifyApiServer(mq, workerSvc, task)
				err = workerSvc.ProcessTaskTranscodeAudio(&task)
				if err != nil {
					log.Error(err)
				}
				if err == nil {
					if err = msg.Ack(false); err != nil {
						log.Error(err)
					}
					if err = mq.Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
						log.Error(err)
					}
					notifyApiServer(mq, workerSvc, models.Task{})
				}

			}
		case models.TaskDash:
			notifyApiServer(mq, workerSvc, task)
			err = workerSvc.ProcessTaskDash(&task)
			if err != nil {
				log.Error(err)
			}
			if err == nil {
				if err = msg.Ack(false); err != nil {
					log.Error(err)
				}
				if err = mq.Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
					log.Error(err)
				}
				notifyApiServer(mq, workerSvc, models.Task{})
			}

		case models.TaskMerge:
			notifyApiServer(mq, workerSvc, task)
			err = workerSvc.ProcessTaskMerge(&task)
			if err != nil {
				log.Error(err)
			}
			if err == nil {
				if err = msg.Ack(false); err != nil {
					log.Error(err)
				}
				if err = mq.Publish(&task, services.MessageBrokerQueueTaskFinished); err != nil {
					log.Error(err)
				}
				notifyApiServer(mq, workerSvc, models.Task{})
			}
		default:
			log.Error("No task kind found")
			if err = msg.Nack(false, true); err != nil {
				log.Error(err)
			}
		}

	}
}

func notifierDecorator(
	mb repositories.MessageBrokerRepository,
	workerSvc worker.Services,
	task models.Task,
	f func(task *models.Task) error) error {
	notifyApiServer(mb, workerSvc, task)
	err := f(&task)
	notifyApiServer(mb, workerSvc, models.Task{})
	return err
}

func notifyApiServer(mb repositories.MessageBrokerRepository, workerSvc worker.Services, task models.Task) {
	w := workerSvc.GetWorkerInfo()
	w.UpdatedAt = time.Now()

	//check if task is empty
	if task == (models.Task{}) {
		w.Status = models.WorkerStatusIdle
		w.WorkingOn = ""

	} else {
		w.Status = models.WorkerStatusWorking
		w.WorkingOn = task.Id.String()
	}

	if err := mb.Publish(w, services.WorkerAssigned); err != nil {
		log.Error(err)
	}
}
