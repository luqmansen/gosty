package inspector

import (
	"github.com/go-chi/chi"
)

func Routes(h VideoInspectorHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/inspector", h.Get)
	r.Post("/inspector/upload", h.UploadVideo)

	return r
}
