package inspector

import (
	"github.com/go-chi/chi"
)

func Routes(h Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/worker", h.Get)
	r.Post("/worker", h.Post)

	return r
}
