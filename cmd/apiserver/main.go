package main

import (
	"fmt"
	inspectorApi "github.com/luqmansen/gosty/apiserver/api/video"
	"github.com/luqmansen/gosty/apiserver/pkg"
	"github.com/luqmansen/gosty/apiserver/pkg/util/config"
	"github.com/luqmansen/gosty/apiserver/pkg/util/health"
	"github.com/luqmansen/gosty/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/apiserver/services"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	cfg := config.LoadConfig(".")
	cfg.DebugConfig()

	vidRepo, err := mongo.NewVideoRepository(cfg.Database)
	if err != nil {
		log.Fatalf(err.Error())
	}

	taskRepo, err := mongo.NewTaskRepository(cfg.Database)
	if err != nil {
		log.Fatalf(err.Error())
	}

	workerRepo, err := mongo.NewWorkerRepository(cfg.Database)
	if err != nil {
		log.Fatalf(err.Error())
	}

	rabbit := rabbitmq.NewRabbitMQRepo(cfg.MessageBroker.GetMessageBrokerUri())

	schedulerSvc := services.NewSchedulerService(taskRepo, vidRepo, rabbit)
	workerSvc := services.NewWorkerService(workerRepo, rabbit)

	go schedulerSvc.ReadMessages()
	go workerSvc.ReadMessage()

	insSvc := services.NewVideoService(vidRepo, schedulerSvc)
	insHandler := inspectorApi.NewInspectorHandler(cfg, insSvc)

	go health.InitHealthCheck(cfg)

	r := inspectorApi.Routes(insHandler)
	port := pkg.GetEnv("PORT", "8000")
	log.Infof("apiserver running on pod %s, listening to %s", os.Getenv("HOSTNAME"), port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		log.Println(err.Error())
	}
}
