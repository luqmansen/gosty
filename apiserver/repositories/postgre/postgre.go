package postgre

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type postgresRepository struct {
	connUrl string
	timeout time.Duration
}

func newPostgreClient(dsn string) *gorm.DB {
	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
	}

	//err = db.Debug().Migrator().DropTable(&model.Video{}, &model.Task{}, &model.Worker{})
	//if err != nil {
	//	panic(err)
	//}
	//
	//
	//err = db.Debug().AutoMigrate(&model.Video{}, &model.Task{}, &model.Worker{})
	//if err != nil {
	//	panic(err)
	//}

	return db
}
