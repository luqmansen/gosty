package video

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Routes(h Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/playlist", h.GetPlaylist)
	r.Post("/video/upload", h.UploadHandler)

	return r
}
