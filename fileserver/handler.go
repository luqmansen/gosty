package fileserver

import (
	"bufio"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//
//type FileServerHandler interface {
//	HandleUpload() func(writer http.ResponseWriter, request *http.Request)
//	HandleFileServer(pathToServe string) func(writer http.ResponseWriter, request *http.Request)
//}
//
//type fileServer struct {
//	router *chi.Mux
//}
//
//func NewFileServerHandler(r *chi.Mux) FileServerHandler {
//	return fileServer{
//		router: r,
//	}
//}
func Index() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("OK"))
	}
}

func HandleFileServer(pathToServe string) func(writer http.ResponseWriter, request *http.Request) {

	if _, err := os.Stat(pathToServe); os.IsNotExist(err) {
		err = os.Mkdir(pathToServe, 0700)
		if err != nil {
			log.Error(err)
		}
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		ctx := chi.RouteContext(request.Context())
		pathPrefix := strings.TrimSuffix(ctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(http.Dir(pathToServe)))
		fs.ServeHTTP(writer, request)
	}
}

func HandleUpload() func(writer http.ResponseWriter, request *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if p.FormName() != "file" {
			log.Println("file field is expected")
			http.Error(w, "file field is expected", http.StatusBadRequest)
			return
		}

		params, _ := url.ParseQuery(r.URL.RawQuery)
		fileName := params.Get("filename")

		f, err := os.Create("./storage/" + fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf := bufio.NewReader(p)
		var maxSize int64 = 32 << 20
		lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))

		written, err := f.ReadFrom(lmt)
		log.Debugf("Written %s byte", written)
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
}
