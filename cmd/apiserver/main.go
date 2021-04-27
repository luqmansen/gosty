package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/luqmansen/gosty/pkg/apiserver/api"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"net/http"
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

	// for development purposes
	r := server.GetRouter()
	dropEverything(r, cfg)

	server.Serve()
}

func dropEverything(router *chi.Mux, cfg *config.Configuration) {
	router.Get("/drop", func(writer http.ResponseWriter, request *http.Request) {
		c, _ := mongo.NewMongoClient(cfg.Database.GetDatabaseUri(), cfg.Database.Timeout)
		writer.Write([]byte(fmt.Sprintf("Dropping %s\n", "db")))
		err := c.Database("gosty").Drop(context.Background())
		if err != nil {
			writer.Write([]byte(fmt.Sprintf("Error dropping %s: %s", "db", err)))
		}

		conn, err := amqp.Dial(cfg.MessageBroker.GetMessageBrokerUri())
		if err != nil {
			log.Fatalf("Failed to connect to rabbitmq: %s", err.Error())
		}
		ch, err := conn.Channel()
		if err != nil {
			log.Error()
		}

		queue := []string{services.MessageBrokerQueueTaskUpdateStatus, services.MessageBrokerQueueTaskFinished,
			services.MessageBrokerQueueTaskNew, services.WorkerStatus, services.WorkerAssigned, services.WorkerNew}

		for _, q := range queue {
			writer.Write([]byte(fmt.Sprintf("Dropping %s\n", q)))
			_, err = ch.QueuePurge(q, true)
			if err != nil {
				writer.Write([]byte(fmt.Sprintf("Error dropping %s: %s", q, err)))
			}
		}

		writer.Write([]byte("DROP SUCCESS"))
	})
}
