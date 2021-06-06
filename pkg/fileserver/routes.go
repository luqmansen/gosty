package fileserver

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"net/http"
	"strings"
)

func NewRouter(fsHandler Handler) *chi.Mux {
	r := initRouter()

	r.Get("/", fsHandler.Index)
	r.Get(getFsPath("/files", r), fsHandler.HandleFileServer)
	r.Post("/upload", fsHandler.HandleUpload)
	r.Get("/drop", fsHandler.DropAll)
	r.Get("/all", fsHandler.GetAll)

	return r
}

func getFsPath(path string, router *chi.Mux) string {
	if strings.ContainsAny(path, "{}*") {
		panic("fileServer does not permit any URL parameters.")
	}
	if path != "/" && path[len(path)-1] != '/' {
		router.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	return path
}

func initRouter() *chi.Mux {
	config.LoadConfig(".")
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
