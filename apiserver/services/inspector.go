package services

import (
	"bytes"
	"github.com/luqmansen/gosty/apiserver/model"
)

type InspectorService interface {
	Inspect(file *bytes.Buffer) model.Video
}
