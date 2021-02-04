package inspector

import (
	"bytes"
	"github.com/luqmansen/gosty/apiserver/model"
)

type VideoInspectorService interface {
	Inspect(file *bytes.Buffer) model.Video
}
