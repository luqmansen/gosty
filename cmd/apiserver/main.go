package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/luqmansen/gosty/apiserver/model"
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
	s.DB.Debug().AutoMigrate(&model.Video{}, &model.Task{}, &model.Worker{})
}

var server = Server{}

func main() {
	db := fmt.Sprintf("postgres")
	dsn := fmt.Sprintf("host=localhost user=postgres password=password dbname=gosty port=5432 sslmode=disable")
	server.InitDB(db, dsn)
}
