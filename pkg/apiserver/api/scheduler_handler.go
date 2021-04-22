package api

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type SchedulerHandler interface {
	GetAllTaskProgress(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
}
type scheduler struct {
	schedulerSvc services.SchedulerService
}

func NewSchedulerHandler(schedulerSvc services.SchedulerService) SchedulerHandler {
	return &scheduler{schedulerSvc}

}

func (s scheduler) GetAllTaskProgress(w http.ResponseWriter, r *http.Request) {
	tasks := s.schedulerSvc.GetAllTaskProgress()
	if len(tasks) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}
	resp, err := json.Marshal(tasks)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		log.Error(err)
	}
	return
}

func (s scheduler) Post(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
