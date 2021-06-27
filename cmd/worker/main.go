package main

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/luqmansen/gosty/pkg/worker"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"
)

type svc struct {
	worker.Services
}

var gitCommit string

func main() {
	cfg := config.LoadConfig(".")
	forever := make(chan bool)

	rabbitClient := rabbitmq.NewRabbitMQConn(cfg.MessageBroker.GetMessageBrokerUri())
	mb := rabbitmq.NewRepository(cfg.MessageBroker.GetMessageBrokerUri(), rabbitClient)
	workerSvc := worker.NewWorkerService(mb, cfg)

	// TODO [#3]:  Storage repository initiation should be handled by storage service
	if _, err := os.Stat(worker.TmpPath); os.IsNotExist(err) {
		err = os.Mkdir(worker.TmpPath, 0700)
		if err != nil {
			log.Error(err)
		}
	}

	log.Infof("Worker version %s", gitCommit)
	go workerSvc.RegisterWorker()

	newTaskData := make(chan interface{})
	defer close(newTaskData)
	go mb.ReadMessage(newTaskData, services.MessageBrokerQueueTaskNew, true)

	w := svc{workerSvc}
	go w.processNewTask(newTaskData)
	go worker.InitHealthCheck(cfg, rabbitClient, cfg.MessageBroker.GetMessageBrokerUri())

	// This is endpoint is for api server to check availability
	// of this worker and this worker pod name
	go func() {
		http.HandleFunc("/", getSystemInfo())
		http.HandleFunc("/debug", debugGoroutine())

		err := http.ListenAndServe(":8088", nil)
		if err != nil {
			log.Error(err)
		}
	}()

	log.Printf("Worker running. To exit press CTRL+C")
	<-forever

}

func (wrk *svc) processNewTask(newTaskData chan interface{}) {
	// TODO [#4]:  refactor ack and publish part of this loop
	for t := range newTaskData {
		msg := t.(amqp.Delivery)
		var task models.Task
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			log.Errorf("%s: %s", util.GetCaller(), err)
		}
		wrk.notifyApiServer(&task)
		switch taskKind := task.Kind; taskKind {
		case models.TaskSplit:
			err = wrk.ProcessTaskSplit(&task)
			if err != nil {
				log.Error(err)
				if err := msg.Nack(false, true); err != nil {
					log.Errorf("failed to requeue failed task, id: %s", task.Id.String())
				}
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
					if err := msg.Nack(false, true); err != nil {
						log.Errorf("failed to requeue failed task, id: %s", task.Id.String())
					}
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
					if err := msg.Nack(false, true); err != nil {
						log.Errorf("failed to requeue failed task, id: %s", task.Id.String())
					}
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
		case models.TaskMerge:
			err = wrk.ProcessTaskMerge(&task)
			if err != nil {
				log.Error(err)
				if err := msg.Nack(false, true); err != nil {
					log.Errorf("failed to requeue failed task, id: %s", task.Id.String())
				}
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

		case models.TaskDash:
			err = wrk.ProcessTaskDash(&task)
			if err != nil {
				log.Error(err)
				if err := msg.Nack(false, true); err != nil {
					log.Errorf("failed to requeue failed task, id: %s", task.Id.String())
				}
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

func getSystemInfo() func(w http.ResponseWriter, request *http.Request) {
	containerHostname, _ := os.Hostname()
	return func(w http.ResponseWriter, request *http.Request) {
		info := map[string]interface{}{
			"hostname": containerHostname,
		}
		data, _ := json.Marshal(info)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func debugGoroutine() func(w http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, true)
		writer.Write(buf)
	}
}
