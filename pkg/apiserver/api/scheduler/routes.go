package scheduler

import (
	"github.com/go-chi/chi"
)

func Routes(h Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/scheduler", h.Get)
	r.Post("/scheduler", h.Post)

	return r
}
