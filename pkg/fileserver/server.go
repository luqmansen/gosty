package fileserver

import (
	"fmt"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type server struct {
	router *chi.Mux
	host   string
	port   string
}

func NewServer(port, host string, router *chi.Mux) *server {
	return &server{
		router: router,
		host:   host,
		port:   port,
	}
}

func (server *server) Serve() {
	log.Infof("File running on port http://%s:%s", server.host, server.port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", server.host, server.port), server.router)
	if err != nil {
		log.Error(err)
	}
}
