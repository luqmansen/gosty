package util

import (
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func GenerateID() string {
	id := uuid.New()
	return strings.Replace(id.String(), "-", "", -1)
}

func GetVersionEndpoint(router *chi.Mux, gitCommit string) {
	router.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		version := struct {
			Version string
		}{
			Version: gitCommit,
		}
		resp, err := json.Marshal(version)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			log.Error(err)
		}
	})
}
