package api

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type SchedulerHandler interface {
	GetTaskUpdate(w http.ResponseWriter, r *http.Request)
	GetAllTaskProgress(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
}
type scheduler struct {
	schedulerSvc services.Scheduler
}

func NewSchedulerHandler(schedulerSvc services.Scheduler) SchedulerHandler {
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

// GetTaskUpdate will update query last update from task repository
func (s scheduler) GetTaskUpdate(w http.ResponseWriter, r *http.Request) {
	// add this for testing purposes
	os.Exit(1)
}

func (s scheduler) Post(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
