package storage

import "github.com/luqmansen/gosty/apiserver/model"

type VideoStorage interface {
	GetVideos() []model.Video
}
