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

	rabbitClient := rabbitmq.NewRabbitMQConn(cfg.MessageBroker.GetMessageBrokerUri())
	mb := rabbitmq.NewRepository(cfg.MessageBroker.GetMessageBrokerUri(), rabbitClient)
	workerSvc := worker.NewWorkerService(mb, cfg)

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
	go w.workerStatusNotifier()

	go worker.InitHealthCheck(cfg, rabbitClient)
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
		wrk.notifyApiServer(&task)
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
				wrk.notifyApiServer(nil)
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
					wrk.notifyApiServer(nil)
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
					wrk.notifyApiServer(nil)
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
				wrk.notifyApiServer(nil)
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
				wrk.notifyApiServer(nil)
			}
		default:
			wrk.notifyApiServer(nil)
			log.Error("No task kind found")
			if err = msg.Nack(false, true); err != nil {
				log.Error(err)
			}
		}
	}
}

//notifyApiServer will notify ApiServer when task is assigned
//to a worker, this will update the task and the worker status
func (wrk *svc) notifyApiServer(task *models.Task) {
	w := wrk.GetWorkerInfo()
	w.UpdatedAt = time.Now()

	//check if task is empty
	if task == nil {
		w.Status = models.WorkerStatusReady
		w.WorkingOn = ""

	} else {
		w.Status = models.WorkerStatusWorking
		w.WorkingOn = task.OriginVideo.FileName

		//only update task status if task is actually assigned to a worker
		task.Status = models.TaskStatusOnProgress
		task.Worker = w.WorkerPodName
		if err := wrk.GetMessageBroker().Publish(task, services.MessageBrokerQueueTaskUpdateStatus); err != nil {
			log.Error(err)
		}
	}

	if err := wrk.GetMessageBroker().Publish(w, services.WorkerAssigned); err != nil {
		log.Error(err)
	}
}

// Notify availability of worker every 3 second to message broker
// ApiServer will consume the message and update the worker status
func (wrk *svc) workerStatusNotifier() {
	w := wrk.GetWorkerInfo()

	//this part will only executed at worker boot initiation
	if err := wrk.GetMessageBroker().Publish(w, services.WorkerNew); err != nil {
		log.Error(err)
	}
	for {
		w.UpdatedAt = time.Now()
		if err := wrk.GetMessageBroker().Publish(w, services.WorkerStatus); err != nil {
			log.Error(err)
		}
		time.Sleep(3 * time.Second)
	}
}
