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
	finishedTaskData := make(chan interface{})
	go mq.ReadMessage(newTaskData, services.TaskNew)
	go mq.ReadMessage(finishedTaskData, services.TaskFinished)

	forever := make(chan bool)
	go func() {
		var task models.Task

		for t := range newTaskData {
			msg := t.(amqp.Delivery)
			err := json.Unmarshal(msg.Body, &task)
			if err != nil {
				log.Error(err)
			}
			if task.Kind == models.TaskSplit {
				err = processTaskSplit(&task)
				if err != nil {
					log.Error(err)
				}
			} else if task.Kind == models.TaskTranscode {

			} else if task.Kind == models.TaskMerge {

			} else {
				log.Error("No task kind found")
			}

			if err == nil {
				if err = msg.Ack(true); err != nil {
					log.Error(err)
				}
			}
			if err = mq.Publish(&task, services.TaskFinished); err != nil {
				log.Error(err)
			}
		}
	}()

	log.Printf("Worker started. To exit press CTRL+C")
	<-forever

}
