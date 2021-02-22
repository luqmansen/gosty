package inspector

import (
	"github.com/luqmansen/gosty/apiserver/model"
)

type VideoInspectorService interface {
	Inspect(filePath string) model.Video
}
