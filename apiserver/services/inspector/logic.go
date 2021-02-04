package inspector

import (
	"bytes"
	"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
)

type videoInspectorServices struct {
	InspectorRepo repositories.VideoRepository
}

func NewInspectorService(repo repositories.VideoRepository) VideoInspectorService {
	return &videoInspectorServices{repo}
}

func (v videoInspectorServices) Inspect(file *bytes.Buffer) model.Video {
	panic("implement me")
}
