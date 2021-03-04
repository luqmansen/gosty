package main

import (
	"encoding/json"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/apiserver/services"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	mq := rabbitmq.NewRabbitMQRepo("amqp://guest:guest@localhost:5672/")

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
				err = processTaskSplit(&task)
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
					err = processTaskTranscodeVideo(&task)
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
					err = processTaskTranscodeAudio(&task)
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

	log.Printf("Worker started. To exit press CTRL+C")
	<-forever

}
