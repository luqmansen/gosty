package api

import (
	"encoding/json"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type WorkerHandler interface {
	GetWorkerInfo(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
}

type worker struct {
	workerSvc services.WorkerService
}

func NewWorkerHandler(workerSvc services.WorkerService) WorkerHandler {
	return &worker{workerSvc}

}

func (h worker) GetWorkerInfo(w http.ResponseWriter, r *http.Request) {
	wrk, err := h.workerSvc.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(wrk) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}

	resp, err := json.Marshal(wrk)
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

func (h worker) Post(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
