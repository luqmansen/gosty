package postgre

import (
	"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
)

type videoRepository struct {
}

func NewVideoRepository() repositories.VideoRepository {
	return &videoRepository{}
}

func (v videoRepository) Get(videoId uint) model.Video {
	panic("implement me")
}

func (v videoRepository) Add(video *model.Video) error {
	panic("implement me")
}

func (v videoRepository) Update(videoId uint) error {
	panic("implement me")
}

func (v videoRepository) Delete(videoId uint) error {
	panic("implement me")
}
