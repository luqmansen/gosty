package inspector

import (
	"github.com/go-chi/chi"
)

func Routes(h SchedulerHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/scheduler", h.Get)
	r.Post("/scheduler", h.Post)

	return r
}
