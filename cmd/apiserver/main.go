package main

import (
	"fmt"
	inspectorApi "github.com/luqmansen/gosty/apiserver/api/inspector"
	"github.com/luqmansen/gosty/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/apiserver/services"

	"log"
	"net/http"
)

func main() {
	mongoUri := "mongodb://username:password@localhost:27017/gosty?authSource=admin"
	vidRepo, err := mongo.NewVideoRepository(mongoUri, "gosty", 2)
	if err != nil {
		panic(err)
	}
	taskRepo, err := mongo.NewTaskRepository(mongoUri, "gosty", 2)
	if err != nil {
		panic(err)
	}
	schedulerSvc := services.NewSchedulerService(taskRepo)
	insSvc := services.NewInspectorService(vidRepo, schedulerSvc)
	insHandler := inspectorApi.NewInspectorHandler(insSvc)

	r := inspectorApi.Routes(insHandler)
	fmt.Println("listening to :8000")
	err = http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println(err.Error())
	}
}
