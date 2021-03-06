package main

import (
	"fmt"
	inspectorApi "github.com/luqmansen/gosty/apiserver/api/inspector"
	"github.com/luqmansen/gosty/apiserver/pkg"
	"github.com/luqmansen/gosty/apiserver/repositories/mongo"
	"github.com/luqmansen/gosty/apiserver/repositories/rabbitmq"
	"github.com/luqmansen/gosty/apiserver/services"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

func main() {
	//init viper configuration
	pkg.InitConfig()
	config := pkg.GetConfig()

	vidRepo, err := mongo.NewVideoRepository(config.Database)
	if err != nil {
		log.Fatalf(err.Error())
	}
	taskRepo, err := mongo.NewTaskRepository(config.Database)
	if err != nil {
		log.Fatalf(err.Error())
	}

	mb := rabbitmq.NewRabbitMQRepo(viper.GetString("mb"))

	schedulerSvc := services.NewSchedulerService(taskRepo, vidRepo, mb)
	//reading message from rabbit
	go func() {
		schedulerSvc.ReadMessages()
	}()

	insSvc := services.NewInspectorService(vidRepo, schedulerSvc)
	insHandler := inspectorApi.NewInspectorHandler(insSvc)

	r := inspectorApi.Routes(insHandler)
	log.Infof("listening to :%s:%s", config.Server.Host, config.Server.Port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port), r)
	if err != nil {
		log.Println(err.Error())
	}
}
