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
	"io/ioutil"
	"net/http"
)

func main() {
	cfg := config.LoadConfig(".")
	util.DebugStruct(*cfg)

	mongoClient, err := mongo.NewMongoClient(cfg.Database.GetDatabaseUri(), cfg.Database.Timeout)
	if err != nil {
		log.Fatalf("failed to init mongo mongoClient: %s", err.Error())
	}

	vidRepo := mongo.NewVideoRepository(cfg.Database, mongoClient)
	taskRepo := mongo.NewTaskRepository(cfg.Database, mongoClient)
	workerRepo := mongo.NewWorkerRepository(cfg.Database, mongoClient)

	rabbitClient := rabbitmq.NewRabbitMQConn(cfg.MessageBroker.GetMessageBrokerUri())
	rabbit := rabbitmq.NewRepository(cfg.MessageBroker.GetMessageBrokerUri(), rabbitClient)
	sseServer := sse.New()
	sseServer.CreateStream(services.WorkerHTTPEventStream)
	sseServer.CreateStream(services.TaskHTTPEventStream)

	schedulerSvc := services.NewSchedulerService(taskRepo, vidRepo, rabbit, sseServer)
	workerSvc := services.NewWorkerService(workerRepo, rabbit, sseServer)
	videoSvc := services.NewVideoService(vidRepo, schedulerSvc)

	go schedulerSvc.ReadMessages()
	go workerSvc.ReadMessage()
	go util.InitHealthCheck(cfg, mongoClient, rabbitClient)

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
		_, _ = writer.Write([]byte(fmt.Sprintf("Dropping %s\n", "db")))
		err := c.Database("gosty").Drop(context.Background())
		if err != nil {
			_, _ = writer.Write([]byte(fmt.Sprintf("Error dropping %s: %s", "db", err)))
		}

		conn, err := amqp.Dial(cfg.MessageBroker.GetMessageBrokerUri())
		if err != nil {
			log.Fatalf("Failed to connect to rabbitmq: %s", err.Error())
		}
		if conn == nil {
			_, _ = writer.Write([]byte("Failed drop, connection is nil"))
			return
		}
		ch, err := conn.Channel()
		if err != nil {
			log.Error()
		}
		if ch == nil {
			_, _ = writer.Write([]byte("Failed drop, channel is nil"))
			return
		}

		queue := []string{services.MessageBrokerQueueTaskUpdateStatus, services.MessageBrokerQueueTaskFinished,
			services.MessageBrokerQueueTaskNew, services.WorkerStatus, services.WorkerAssigned, services.WorkerNew}

		for _, q := range queue {
			_, _ = writer.Write([]byte(fmt.Sprintf("Dropping %s\n", q)))
			_, err = ch.QueuePurge(q, true)
			if err != nil {
				_, _ = writer.Write([]byte(fmt.Sprintf("Error dropping %s: %s", q, err)))
			}
		}

		resp, err := http.Get(cfg.FileServer.GetFileServerUri() + "/drop")
		if err != nil {
			log.Fatalln(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			writer.Write([]byte("failed to drop file server data\n"))
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		writer.Write(body)

		_, _ = writer.Write([]byte("DROP SUCCESS"))
	})
}
