package main

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/pkg/util/config"
	"github.com/luqmansen/gosty/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/apiserver/services"
	"github.com/luqmansen/gosty/worker"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
)

func main() {
	cfg := config.LoadConfig(".")

	mq := rabbitmq.NewRabbitMQRepo(cfg.MessageBroker.GetMessageBrokerUri())
	workerSvc := worker.NewWorkerService(mq, cfg)

	go func() {
		if err := mq.Publish(workerSvc.GetWorkerInfo(), services.WorkerNew); err != nil {
			log.Error(err)
		}
	}()

	go worker.InitHealthCheck(cfg)

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

	//Todo: also publish to api server when worker start
	// working on a task and after finishing a task
	go mq.ReadMessage(newTaskData, services.MessageBrokerQueueTaskNew)

	forever := make(chan bool)
	//Todo: refactor ack and publish part of this loop
	go func() {

		for t := range newTaskData {
			msg := t.(amqp.Delivery)
			var task models.Task
			err := json.Unmarshal(msg.Body, &task)
			if err != nil {
				log.Error(err)
			}

			switch taskKind := task.Kind; taskKind {
			case models.TaskSplit:
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
				}

			case models.TaskTranscode:
				switch txType := task.TaskTranscode.TranscodeType; txType {
				case models.TranscodeVideo:
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
					}

				case models.TranscodeAudio:
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
					}

				}
			case models.TaskDash:
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
				}

			case models.TaskMerge:
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
				}
			default:
				log.Error("No task kind found")
				if err = msg.Nack(false, true); err != nil {
					log.Error(err)
				}
			}

		}
	}()

	log.Printf("Worker running. To exit press CTRL+C")
	<-forever

}
