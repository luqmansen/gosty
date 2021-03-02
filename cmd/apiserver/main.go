package main

import (
	"fmt"
	inspectorApi "github.com/luqmansen/gosty/apiserver/api/inspector"
	"github.com/luqmansen/gosty/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/apiserver/services"
	"os"

	log "github.com/sirupsen/logrus"
	"net/http"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}
func main() {
	mongoUri := "mongodb://username:password@localhost:27017/gosty?authSource=admin"
	vidRepo, err := mongo.NewVideoRepository(mongoUri, "gosty", 2)
	if err != nil {
		log.Fatalf(err.Error())
	}
	taskRepo, err := mongo.NewTaskRepository(mongoUri, "gosty", 2)
	if err != nil {
		log.Fatalf(err.Error())
	}

	mb := rabbitmq.NewRabbitMQRepo("amqp://guest:guest@localhost:5672/")

	schedulerSvc := services.NewSchedulerService(taskRepo, mb)
	//reading message from rabbit
	go func() {
		schedulerSvc.ReadMessages()
	}()

	insSvc := services.NewInspectorService(vidRepo, schedulerSvc)
	insHandler := inspectorApi.NewInspectorHandler(insSvc)

	r := inspectorApi.Routes(insHandler)
	fmt.Println("listening to :8000")
	err = http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println(err.Error())
	}
}
