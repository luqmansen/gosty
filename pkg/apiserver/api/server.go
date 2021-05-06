package api

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/gorilla/handlers"
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

func NewServer(port, host string, router *chi.Mux) *Server {
	return &Server{
		router: router,
		host:   host,
		port:   port,
	}
}

func (server *Server) GetRouter() *chi.Mux {
	return server.router
}

func (server *Server) Serve() {

	log.Infof("apiserver running on pod %server, listening to %s:%s server",
		os.Getenv("HOSTNAME"), server.host, server.port)

	loggedRouter := handlers.LoggingHandler(os.Stdout, server.router)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", server.host, server.port), loggedRouter)
	if err != nil {
		log.Println(err.Error())
	}
}
