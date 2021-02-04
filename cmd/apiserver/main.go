package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	inspectorApi "github.com/luqmansen/gosty/apiserver/api/inspector"
	"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/luqmansen/gosty/apiserver/repositories/postgre"
	inspectorSvc "github.com/luqmansen/gosty/apiserver/services/inspector"
	"log"
	"net/http"
)

type Server struct {
	DB *gorm.DB
}

func (s *Server) InitDB(db string, dsn string) {
	var err error
	s.DB, err = gorm.Open(db, dsn)
	if err != nil {
		fmt.Println(err.Error())
	}
	//s.DB.Debug().DropTable(&model.Video{}, &model.Task{}, &model.Worker{})
	s.DB.Debug().AutoMigrate(&model.Video{}, &model.Task{}, &model.Worker{})
	s.DB.Debug().AutoMigrate(&model.Task{}).AddForeignKey("video", "videos(id)", "CASCADE", "CASCADE")

}

//var server = Server{}

func main() {
	//db := fmt.Sprintf("postgres")
	//dsn := fmt.Sprintf("host=localhost user=postgres password=password dbname=gosty port=5432 sslmode=disable")
	//server.InitDB(db, dsn)

	vidRepo := postgre.NewVideoRepository()
	insSvc := inspectorSvc.NewInspectorService(repositories.VideoRepository(vidRepo))
	insHandler := inspectorApi.NewInspectorHandler(insSvc)

	r := inspectorApi.Routes(insHandler)
	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Println(err.Error())
	}
}
