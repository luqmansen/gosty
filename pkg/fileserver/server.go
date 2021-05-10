package fileserver

import (
	"fmt"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type server struct {
	router  *chi.Mux
	address string
}

func NewServer(address string, router *chi.Mux) *server {
	return &server{
		router:  router,
		address: address,
	}
}

func (server *server) Serve() {
	log.Infof("File running on http://%s", server.address)
	err := http.ListenAndServe(fmt.Sprintf("%s", server.address), server.router)
	if err != nil {
		log.Error(err)
	}
}
