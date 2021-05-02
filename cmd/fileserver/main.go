package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/luqmansen/gosty/pkg/fileserver"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
)

func main() {
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

	workDir, _ := os.Getwd()
	pathToServe := workDir + "/storage"

	fsPath := func(path string) string {
		if strings.ContainsAny(path, "{}*") {
			panic("fileServer does not permit any URL parameters.")
		}

		if path != "/" && path[len(path)-1] != '/' {
			r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
			path += "/"
		}
		path += "*"

		return path
	}("/files")

	r.Get("/", fileserver.Index())
	r.Get(fsPath, fileserver.HandleFileServer(pathToServe))
	r.Post("/upload", fileserver.HandleUpload())
	r.Get("/drop", fileserver.DropAll(pathToServe))

	port := util.GetEnv("PORT", "8001")
	host := util.GetEnv("FS_HOST", "0.0.0.0")
	log.Infof("File running on port http://%s:%s", host, port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), r)
	if err != nil {
		log.Error(err)
	}
}
