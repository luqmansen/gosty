package fileserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const (
	StoragePath = "./storage/"
)

type Handler interface {
	Index(writer http.ResponseWriter, request *http.Request)
	HandleUpload(writer http.ResponseWriter, request *http.Request)
	HandleFileServer(writer http.ResponseWriter, request *http.Request)
	DropAll(writer http.ResponseWriter, request *http.Request)
	GetAll(writer http.ResponseWriter, request *http.Request)

	PeerDiscovery()
	StartSync()
}

type fileServer struct {
	pathToServe string
	//This is dns of other statefulsets dns
	selfHost           string
	peerFileServerHost []string
}

func NewFileServerHandler(pathToServe string, peerFsHost []string, host string) Handler {

	if _, err := os.Stat(pathToServe); os.IsNotExist(err) {
		log.Infof("folder %s doesn't exists, creating...", pathToServe)
		err = os.Mkdir(pathToServe, 7777)
		if err != nil {
			log.Error(errors.Wrap(err, "Error initiating file server handler"))
		}
	}

	return &fileServer{
		pathToServe:        pathToServe,
		peerFileServerHost: peerFsHost,
		selfHost:           host,
	}
}

func (fs *fileServer) Index(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("OK"))
}

func (fs *fileServer) GetAll(writer http.ResponseWriter, request *http.Request) {
	files, err := ioutil.ReadDir(fs.pathToServe)
	if err != nil {
		log.Error(errors.Wrap(err, "Error ioutil.ReadDir"))
	}
	var payload []string
	for _, f := range files {
		payload = append(payload, f.Name())
	}
	b, err := json.Marshal(payload)
	if err != nil {
		log.Error(errors.Wrap(err, "Error json.Marshal"))
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
	}
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(b)
}

func (fs *fileServer) HandleFileServer(writer http.ResponseWriter, request *http.Request) {
	ctx := chi.RouteContext(request.Context())
	pathPrefix := strings.TrimSuffix(ctx.RoutePattern(), "/*")
	f := http.StripPrefix(pathPrefix, http.FileServer(http.Dir(fs.pathToServe)))
	f.ServeHTTP(writer, request)
}

func (fs *fileServer) HandleUpload(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		log.Println(errors.Wrap(err, "Error MultipartReader"))
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

	f, err := os.Create(fs.pathToServe + "/" + fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := bufio.NewReader(p)
	var maxSize int64 = 32 << 20
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))

	written, err := f.ReadFrom(lmt)
	log.Debugf("Written %v byte", written)
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

func (fs *fileServer) DropAll(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(fs.pathToServe)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			w.Write([]byte(fmt.Sprintf("Removing %s\n", filename)))
			err := os.Remove(fs.pathToServe + "/" + filename)
			if err != nil {
				log.Errorf("error removing %s: %s", filename, err)
				w.Write([]byte(fmt.Sprintf("error removing %s: %s\n", filename, err)))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}(f.Name())
	}
	wg.Wait()
	w.WriteHeader(http.StatusNoContent)
}
