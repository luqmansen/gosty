package services

import (
	"github.com/luqmansen/gosty/apiserver/models"
)

type VideoInspectorService interface {
	Inspect(filePath string) models.Video
}
