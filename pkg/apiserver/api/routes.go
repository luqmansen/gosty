package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"time"
)

func newRouter() *chi.Mux {
	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"X-Requested-With"},
		AllowCredentials:   false,
		MaxAge:             300, // Maximum value not ignored by any of major browsers
		OptionsPassthrough: false,
	}))
	return r
}

func (server *Server) AddEventStreamRoute() {
	if server.sseServer != nil {
		server.sseServer.EventTTL = 1 * time.Second
		server.router.Get("/events", server.sseServer.HTTPHandler)
	}
}

func (server *Server) AddWorkerRoutes(h WorkerHandler) {
	server.router.Get("/worker", h.GetWorkerInfo)
	server.router.Post("/worker", h.Post)
}

func (server *Server) AddVideoRoutes(h VideoHandler) {
	server.router.Get("/playlist", h.GetPlaylist)
	server.router.Post("/video/upload", h.UploadHandler)
}

func (server *Server) AddSchedulerRoutes(h SchedulerHandler) {
	server.router.Get("/progress", h.GetAllTaskProgress)
}
