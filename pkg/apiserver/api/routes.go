package api

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	return r
}

func AddWorkerRoutes(r *chi.Mux, h WorkerHandler) *chi.Mux {

	r.Get("/worker", h.GetWorkerInfo)
	r.Post("/worker", h.Post)

	return r
}

func AddVideoRoutes(r *chi.Mux, h VideoHandler) *chi.Mux {

	r.Get("/playlist", h.GetPlaylist)
	r.Post("/video/upload", h.UploadHandler)

	return r
}
