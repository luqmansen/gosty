package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/luqmansen/gosty/pkg/apiserver/services"
	log "github.com/sirupsen/logrus"
)

type WorkerHandler interface {
	GetWorkerInfo(w http.ResponseWriter, r *http.Request)
	ScaleHandler(w http.ResponseWriter, r *http.Request)
}

type worker struct {
	workerSvc services.WorkerService
}

func NewWorkerHandler(workerSvc services.WorkerService) WorkerHandler {
	return &worker{workerSvc}

}

func (wrk worker) GetWorkerInfo(w http.ResponseWriter, r *http.Request) {
	workerList, err := wrk.workerSvc.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(workerList) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(workerList)
	if err != nil {
		log.Errorf("Failed to marshal worker list: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
	}
}

func (wrk worker) ScaleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.PostFormValue("replicanum")
	if query == "" {
		errResp := map[string]interface{}{"error": "Replica num can't be empty"}
		resp, _ := json.Marshal(errResp)
		w.Write(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	replicaNum, err := strconv.Atoi(query)
	if err != nil {
		errResp := map[string]interface{}{"error": "replica number should be integer"}
		resp, _ := json.Marshal(errResp)
		w.Write(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sc, err := wrk.workerSvc.Scale(int32(replicaNum))
	if err != nil {
		//todo: create models for response
		errResp := map[string]interface{}{"error": err}
		resp, _ := json.Marshal(errResp)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
	}

	resp, _ := json.Marshal(sc)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
