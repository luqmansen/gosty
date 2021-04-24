package api

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Server struct {
	router    *chi.Mux
	sseServer *sse.Server
	host      string
	port      string
}

func NewServer(port, host string) *Server {
	r := newRouter()

	return &Server{
		router: r,
		host:   host,
		port:   port,
	}
}

func (server *Server) AddEventStreamServer(s *sse.Server) {
	server.sseServer = s
}

func (server *Server) Serve() {

	log.Infof("apiserver running on pod %server, listening to %s:%s server",
		os.Getenv("HOSTNAME"), server.host, server.port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", server.host, server.port), server.router)
	if err != nil {
		log.Println(err.Error())
	}
}
