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

	client, err := mongo.NewMongoClient(cfg.Database.GetDatabaseUri(), cfg.Database.Timeout)
	if err != nil {
		log.Fatalf("failed to init mongo client: %s", err.Error())
	}

	vidRepo := mongo.NewVideoRepository(cfg.Database, client)
	taskRepo := mongo.NewTaskRepository(cfg.Database, client)
	workerRepo := mongo.NewWorkerRepository(cfg.Database, client)

	rabbit := rabbitmq.NewRepository(cfg.MessageBroker.GetMessageBrokerUri())
	sseServer := sse.New()
	sseServer.CreateStream(services.WorkerHTTPEventStream)
	sseServer.CreateStream(services.TaskHTTPEventStream)

	schedulerSvc := services.NewSchedulerService(taskRepo, vidRepo, rabbit, sseServer)
	workerSvc := services.NewWorkerService(workerRepo, rabbit, sseServer)
	videoSvc := services.NewVideoService(vidRepo, schedulerSvc)

	go schedulerSvc.ReadMessages()
	go workerSvc.ReadMessage()
	go util.InitHealthCheck(cfg)

	videoRestHandler := api.NewVideoHandler(cfg, videoSvc)
	workerRestHandler := api.NewWorkerHandler(workerSvc)
	schedulerRestHandler := api.NewSchedulerHandler(schedulerSvc)

	port := util.GetEnv("PORT", "8000")
	server := api.NewServer(port, "0.0.0.0")
	server.AddWorkerRoutes(workerRestHandler)
	server.AddVideoRoutes(videoRestHandler)
	server.AddSchedulerRoutes(schedulerRestHandler)
	server.AddEventStreamServer(sseServer)
	server.AddEventStreamRoute()

	server.Serve()
}
