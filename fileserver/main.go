package main

import (
	"bufio"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	workDir, _ := os.Getwd()
	path := workDir + "/fileserver/storage"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0700)
		if err != nil {
			log.Error(err)
		}
	}

	r.Post("/upload", UploadVideo)

	filesDir := http.Dir(path)
	FileServer(r, "/files", filesDir)

	err := http.ListenAndServe(":8001", r)
	if err != nil {
		panic(err)
	}
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func UploadVideo(w http.ResponseWriter, r *http.Request) {
	// uncomment to give upload limit
	//r.Body = http.MaxBytesReader(w, r.Body, 32 << 20+1024)

	reader, err := r.MultipartReader()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if p.FormName() != "video" {
		log.Println("video field is expected")
		http.Error(w, "video field is expected", http.StatusBadRequest)
		return
	}
	//

	params, _ := url.ParseQuery(r.URL.RawQuery)
	fileName := params.Get("filename")
	fmt.Println(fileName)
	f, err := os.Create("./fileserver/storage/" + fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := bufio.NewReader(p)
	var maxSize int64 = 32 << 20
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))

	written, err := f.ReadFrom(lmt)
	log.Debugf("Written %s byte" , written)
	if err != nil {
		log.Error(err)
	}

	if err := f.Close(); err != nil {
		log.Println(err.Error())
	}
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("Upload file successful"))
	if err != nil {
		log.Error(err)
	}

}
