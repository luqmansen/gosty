package main

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/pkg"
	"github.com/luqmansen/gosty/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/apiserver/services"
	"github.com/luqmansen/gosty/worker"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

func main() {
	pkg.InitConfig()
	mq := rabbitmq.NewRabbitMQRepo(viper.GetString("mb"))
	workerSvc := worker.NewWorkerService(mq)

	go func() {
		if err := mq.Publish(workerSvc.GetWorkerInfo(), services.WorkerNew); err != nil {
			log.Error(err)
		}
	}()

	log.Infof("Worker %s started", workerSvc.GetWorkerInfo().WorkerPodName)

	newTaskData := make(chan interface{})
	defer close(newTaskData)

	go mq.ReadMessage(newTaskData, services.TaskNew)

	forever := make(chan bool)
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
					if err = mq.Publish(&task, services.TaskFinished); err != nil {
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
						if err = mq.Publish(&task, services.TaskFinished); err != nil {
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
						if err = mq.Publish(&task, services.TaskFinished); err != nil {
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
					if err = mq.Publish(&task, services.TaskFinished); err != nil {
						log.Error(err)
					}
				}

			case models.TaskMerge:
				panic("Not implemented")
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
