package postgre

import (
	"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"gorm.io/gorm"
)

type videoRepository struct {
	db *gorm.DB
}

func NewVideoRepository(dsn string) repositories.VideoRepository {
	db := newPostgreClient(dsn)
	return &videoRepository{
		db: db,
	}
}

func (v videoRepository) Get(videoId uint) model.Video {
	panic("implement me")
}

func (v videoRepository) Add(video *model.Video) {
	v.db.Create(video)
}

func (v videoRepository) Update(videoId uint) error {
	panic("implement me")
}

func (v videoRepository) Delete(videoId uint) error {
	panic("implement me")
}
