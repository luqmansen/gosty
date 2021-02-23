package main

import (
	"bufio"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"strings"
)

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
		http.Error(w, "video field is expected", http.StatusBadRequest)
		return
	}

	buf := bufio.NewReader(p)
	sniff, _ := buf.Peek(512)
	if !filetype.IsVideo(sniff) {
		http.Error(w, "video file expected", http.StatusBadRequest)
		return
	}

	ext, err := mime.ExtensionsByType(p.Header.Get("Content-Type"))
	if len(ext) == 0 {
		contentType := http.DetectContentType(sniff)
		ext, err = mime.ExtensionsByType(contentType)
	}

	if err != nil {
		log.Println(err.Error())
	}

	//TODO: hash to MD5, skip if already exists
	f, err := ioutil.TempFile("./fileserver/storage", fmt.Sprintf("%s-*%s", uuid.NewString(), ext[0]))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maxSize int64 = 32 << 20
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))

	_, err = io.Copy(f, lmt)
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
