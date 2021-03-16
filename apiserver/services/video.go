package services

import (
	"github.com/luqmansen/gosty/apiserver/models"
)

type VideoService interface {
	Inspect(filePath string) models.Video
	GetAll() []*models.Video
}
