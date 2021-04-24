package main

import (
	"github.com/luqmansen/gosty/pkg/apiserver/api"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
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
	sseServer := sse.New()

	schedulerSvc := services.NewSchedulerService(taskRepo, vidRepo, rabbit)
	workerSvc := services.NewWorkerService(workerRepo, rabbit, sseServer)
	videoSvc := services.NewVideoService(vidRepo, schedulerSvc)

	go schedulerSvc.ReadMessages()
	go workerSvc.ReadMessage()
	go util.InitHealthCheck(cfg)

	videoRestHandler := api.NewVideoHandler(cfg, videoSvc)
	workerRestHandler := api.NewWorkerHandler(workerSvc)
	schedulerRestHandler := api.NewSchedulerHandler(schedulerSvc)

	//create sse stream event for every service
	sseServer.CreateStream(services.WorkerHTTPEventStream)

	port := util.GetEnv("PORT", "8000")
	server := api.NewServer(port, "8000")
	server.AddWorkerRoutes(workerRestHandler)
	server.AddVideoRoutes(videoRestHandler)
	server.AddSchedulerRoutes(schedulerRestHandler)
	server.AddEventStreamServer(sseServer)
	server.AddEventStreamRoute()

	server.Serve()
}
