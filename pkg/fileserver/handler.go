package fileserver

import (
	"bufio"
	"container/list"
	"fmt"
	"github.com/go-chi/chi"
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

	// Sync related
	InitialSync()
	SyncHook(writer http.ResponseWriter, request *http.Request)
	ExecuteSynchronization()
}

type fileServer struct {
	pathToServe string
	//This is dns of other statefulsets dns
	selfHost           string
	peerFileServerHost []string
	syncMapFileLists   *sync.Map
	syncQueue          *list.List
}

func NewFileServerHandler(pathToServe string, peerFsHost []string, host string) Handler {

	if _, err := os.Stat(pathToServe); os.IsNotExist(err) {
		log.Infof("folder %s doesn't exists, creating...", pathToServe)
		err = os.Mkdir(pathToServe, 0700)
		if err != nil {
			log.Error(err)
		}
	}

	return &fileServer{
		pathToServe:        pathToServe,
		peerFileServerHost: peerFsHost,
		selfHost:           host,
		syncMapFileLists:   &sync.Map{},
		syncQueue:          list.New(),
	}
}

func (h *fileServer) Index(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("OK"))
}

func (h *fileServer) HandleFileServer(writer http.ResponseWriter, request *http.Request) {
	ctx := chi.RouteContext(request.Context())
	pathPrefix := strings.TrimSuffix(ctx.RoutePattern(), "/*")
	fs := http.StripPrefix(pathPrefix, http.FileServer(http.Dir(h.pathToServe)))
	fs.ServeHTTP(writer, request)
}

func (h *fileServer) HandleUpload(w http.ResponseWriter, r *http.Request) {
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

	f, err := os.Create(h.pathToServe + "/" + fileName)
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

	h.triggerSync()
}

func (h *fileServer) DropAll(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(h.pathToServe)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			w.Write([]byte(fmt.Sprintf("Removing %s\n", filename)))
			err := os.Remove(h.pathToServe + "/" + filename)
			if err != nil {
				log.Errorf("error removing %s: %s", filename, err)
				w.Write([]byte(fmt.Sprintf("error removing %s: %s\n", filename, err)))
			}
		}(f.Name())
	}
	wg.Wait()
	w.WriteHeader(http.StatusNoContent)
}
