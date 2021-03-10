package inspector

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Routes(h VideoInspectorHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/inspector", h.Get)
	r.Post("/inspector/upload", h.UploadHandler)

	return r
}
