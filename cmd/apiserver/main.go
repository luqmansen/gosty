package main

import (
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/api"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	cfg := config.LoadConfig(".")
	util.DebugStruct(*cfg)

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

	rabbit := rabbitmq.NewRepository(cfg.MessageBroker.GetMessageBrokerUri())

	schedulerSvc := services.NewSchedulerService(taskRepo, vidRepo, rabbit)
	workerSvc := services.NewWorkerService(workerRepo, rabbit)
	videoSvc := services.NewVideoService(vidRepo, schedulerSvc)

	go schedulerSvc.ReadMessages()
	go workerSvc.ReadMessage()
	go util.InitHealthCheck(cfg)

	videoRestHandler := api.NewVideoHandler(cfg, videoSvc)
	workerRestHandler := api.NewWorkerHandler(workerSvc)
	schedulerRestHandler := api.NewSchedulerHandler(schedulerSvc)

	r := api.NewRouter()
	api.AddWorkerRoutes(r, workerRestHandler)
	api.AddVideoRoutes(r, videoRestHandler)
	api.AddSchedulerRoutes(r, schedulerRestHandler)

	port := util.GetEnv("PORT", "8000")
	log.Infof("apiserver running on pod %s, listening to %s", os.Getenv("HOSTNAME"), port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), r)
	if err != nil {
		log.Println(err.Error())
	}
}
