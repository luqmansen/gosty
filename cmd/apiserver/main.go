package main

import (
	"fmt"
	inspectorApi "github.com/luqmansen/gosty/apiserver/api/inspector"
	//"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/luqmansen/gosty/apiserver/repositories/postgre"
	inspectorSvc "github.com/luqmansen/gosty/apiserver/services/inspector"
	"log"
	"net/http"
)

func main() {
	dsn := fmt.Sprintf("host=localhost user=postgres password=password dbname=gosty port=5432")
	vidRepo := postgre.NewVideoRepository(dsn)
	insSvc := inspectorSvc.NewInspectorService(repositories.VideoRepository(vidRepo))
	insHandler := inspectorApi.NewInspectorHandler(insSvc)

	r := inspectorApi.Routes(insHandler)
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println(err.Error())
	}
}
