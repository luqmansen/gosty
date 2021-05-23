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
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"net/http"
)

var gitCommit string

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
	go rabbit.ResourcesWatcher()
	sseServer := sse.New()
	sseServer.CreateStream(services.WorkerHTTPEventStream)
	sseServer.CreateStream(services.TaskHTTPEventStream)

	schedulerSvc := services.NewSchedulerService(taskRepo, vidRepo, rabbit, sseServer)
	workerSvc := services.NewWorkerService(workerRepo, rabbit, sseServer)
	videoSvc := services.NewVideoService(vidRepo, schedulerSvc)

	go schedulerSvc.ReadMessages()
	go workerSvc.ReadMessage()
	go util.InitHealthCheck(cfg, mongoClient, rabbitClient, cfg.MessageBroker.GetMessageBrokerUri())

	videoRestHandler := api.NewVideoHandler(cfg, videoSvc)
	workerRestHandler := api.NewWorkerHandler(workerSvc)
	schedulerRestHandler := api.NewSchedulerHandler(schedulerSvc)

	port := util.GetEnv("PORT", "8000")
	router := api.NewRouter(schedulerRestHandler, workerRestHandler, videoRestHandler)
	server := api.NewServer(port, "0.0.0.0", router)
	server.AddEventStreamRoute(sseServer)

	// for development purposes
	r := server.GetRouter()
	dropEverythingRoute(r, cfg, mongoClient, rabbitClient)
	getVersion(r)

	server.Serve()
}

func dropEverythingRoute(router *chi.Mux, cfg *config.Configuration, mongoClient *mongo2.Client, rabbitConn *amqp.Connection) {
	router.Get("/drop", func(writer http.ResponseWriter, request *http.Request) {

		//drop mongo collection
		_, _ = writer.Write([]byte(fmt.Sprintf("Dropping %s\n", "db")))
		err := mongoClient.Database("gosty").Drop(context.Background())
		if err != nil {
			_, _ = writer.Write([]byte(fmt.Sprintf("Error dropping %s: %s", "db", err)))
		}

		//drop rabbitmq queue
		ch, err := rabbitConn.Channel()
		if err != nil {
			log.Error()
		}
		if ch == nil {
			writer.Write([]byte("Failed drop, channel is nil"))
			return
		}

		queue := []string{
			services.MessageBrokerQueueTaskUpdateStatus,
			services.MessageBrokerQueueTaskFinished,
			services.MessageBrokerQueueTaskNew,
			//services.WorkerStatus,
			services.WorkerAssigned,
			services.WorkerNew,
		}

		for _, q := range queue {
			_, _ = writer.Write([]byte(fmt.Sprintf("Dropping %s\n", q)))
			_, err = ch.QueuePurge(q, true)
			if err != nil {
				writer.Write([]byte(fmt.Sprintf("Error dropping %s: %s", q, err)))
			}
		}
		err = ch.Close()
		if err != nil {
			writer.Write([]byte(fmt.Sprintf("Error close connection: %s", err)))
		}

		//drop all on fileserver
		resp, err := http.Get(cfg.FileServer.GetFileServerUri() + "/drop")
		if err != nil {
			log.Error(err)
		}
		if resp != nil {
			if resp.StatusCode != http.StatusNoContent {
				writer.Write([]byte("failed to drop file server data\n"))
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
			}
			writer.Write(body)
		}

		_, _ = writer.Write([]byte("DROP SUCCESS"))

	})
}

func getVersion(router *chi.Mux) {
	router.Get("/version", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(gitCommit))
	})
}
