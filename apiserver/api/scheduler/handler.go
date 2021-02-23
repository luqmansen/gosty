package inspector

import (
	"github.com/luqmansen/gosty/apiserver/services"
	"net/http"
)

type SchedulerHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	inspectorService services.VideoInspectorService
}

func NewInspectorHandler(inspectorSvc services.VideoInspectorService) SchedulerHandler {
	return &handler{inspectorSvc}

}

func (h handler) Get(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (h handler) Post(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
