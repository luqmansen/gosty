package inspector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/luqmansen/gosty/apiserver/pkg"
	"github.com/luqmansen/gosty/apiserver/pkg/util/config"
	"github.com/luqmansen/gosty/apiserver/services"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

type VideoInspectorHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	UploadHandler(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	inspectorService services.VideoInspectorService
	config           *config.Configuration
}

func NewInspectorHandler(
	cfg *config.Configuration,
	inspectorSvc services.VideoInspectorService,
) VideoInspectorHandler {
	return &handler{
		inspectorService: inspectorSvc,
		config:           cfg,
	}

}

func (h handler) Get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}

func (h handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
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

	if p.FormName() != "file" {
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
	if err != nil {
		log.Println(err.Error())
	}
	if len(ext) == 0 {
		contentType := http.DetectContentType(sniff)
		ext, err = mime.ExtensionsByType(contentType)
		if len(ext) == 0 {
			log.Error("no content type detected, set to mp4")
			ext = append(ext, ".mp4")
		}
	}

	fileName := fmt.Sprintf("%s-*%s", uuid.NewString(), ext[0])
	f, err := ioutil.TempFile("./tmp/", fileName)
	if err != nil {
		log.Errorf("Error creating temp file %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maxSize int64 = 32 << 20
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))

	n, err := io.Copy(f, lmt)
	log.Debugf("Byte written for file %d: ", n)
	if err != nil && err != io.EOF {
		log.Errorf("Error copying file %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		log.Errorf("Error seeking file %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//upload to file server
	values := map[string]io.Reader{"file": f}
	actualFileName := strings.Split(f.Name(), "/")[1] // remove /tmp/ on filepath
	url := fmt.Sprintf("%s/upload?filename=%s", h.config.FileServer.GetFileServerUri(), actualFileName)
	err = pkg.Upload(url, values)
	if err != nil {
		log.Errorf("Error uploading files %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//if err = os.Remove(f.Name()); err != nil{
	//	//just log the error, don't care if file isn't removed
	//	//handle the disk full error on later
	//	log.Error(err)
	//}

	//If we put inspect before file upload, in case of pod is down, after video
	//inspection, the error will propagate to other service since task is
	//already created, but file hasn't uploaded
	vid := h.inspectorService.Inspect(f.Name())

	resp, err := json.Marshal(vid)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Error(err)
	}
	return

}
