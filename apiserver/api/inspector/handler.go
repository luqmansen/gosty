package inspector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/luqmansen/gosty/apiserver/services/inspector"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
)

type VideoInspectorHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	UploadVideo(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	inspectorService inspector.VideoInspectorService
}

func NewInspectorHandler(inspectorSvc inspector.VideoInspectorService) VideoInspectorHandler {
	return &handler{inspectorSvc}

}

func (h handler) Get(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}

func (h handler) UploadVideo(w http.ResponseWriter, r *http.Request) {
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

	f, err := ioutil.TempFile("./tmp/", fmt.Sprintf("%s-*%s", uuid.NewString(), ext[0]))
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
	// inspect
	vid := h.inspectorService.Inspect(f.Name())
	//fmt.Println(vid)
	// save to db
	// pub

	resp, err := json.Marshal(vid)
	if err != nil {
		// handle error
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)

}
