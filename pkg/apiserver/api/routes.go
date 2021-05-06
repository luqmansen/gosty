package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/r3labs/sse/v2"
	"time"
)

func NewRouter(
	schedulerHandler SchedulerHandler,
	workerHandler WorkerHandler,
	videoHandler VideoHandler,
) *chi.Mux {
	r := initRouter()

	r.Route("/api", func(r chi.Router) {
		r.Route("/worker", func(r chi.Router) {
			r.Get("/", workerHandler.GetWorkerInfo)
			r.Post("/", workerHandler.Post)
		})

		r.Route("/video", func(r chi.Router) {
			r.Get("/playlist", videoHandler.GetPlaylist)
			r.Post("/upload", videoHandler.UploadHandler)
		})

		r.Route("/scheduler", func(r chi.Router) {
			r.Get("/progress", schedulerHandler.GetAllTaskProgress)
		})

	})
	return r
}

func initRouter() *chi.Mux {
	r := chi.NewRouter()
	//TODO: find compatible logger middleware with SSE Server
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

func (server *Server) AddEventStreamRoute(s *sse.Server) {
	server.sseServer = s
	if server.sseServer != nil {
		server.sseServer.EventTTL = 1 * time.Second
		server.router.Get("/api/events", server.sseServer.HTTPHandler)
	}
}
