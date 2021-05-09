package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/services"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	apiserverTempDir = "./tmp/"
)

type VideoHandler interface {
	//GetPlaylist fetch all video with finished dask task
	GetPlaylist(w http.ResponseWriter, r *http.Request)
	UploadHandler(w http.ResponseWriter, r *http.Request)
}

type video struct {
	videoService services.VideoService
	config       *config.Configuration
}

func NewVideoHandler(cfg *config.Configuration, videoService services.VideoService) VideoHandler {
	return &video{
		videoService: videoService,
		config:       cfg,
	}
}

func (h video) GetPlaylist(w http.ResponseWriter, r *http.Request) {

	vid := h.videoService.GetAll()
	if len(vid) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}

	resp, err := json.Marshal(vid)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp)
	if err != nil {
		log.Error(err)
	}
	return
}

func (h video) UploadHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "file field is expected", http.StatusBadRequest)
		return
	}

	buf := bufio.NewReader(p)

	sniff, _ := buf.Peek(512)
	if !filetype.IsVideo(sniff) {
		http.Error(w, "video file expected", http.StatusBadRequest)
		return
	}
	if _, err := os.Stat(apiserverTempDir); os.IsNotExist(err) {
		err = os.Mkdir(apiserverTempDir, 0700)
		if err != nil {
			log.Error(err)
		}
	}

	ext := mimetype.Detect(sniff).Extension()
	fileName := fmt.Sprintf("%s-*%s", uuid.NewString(), ext)
	f, err := ioutil.TempFile(apiserverTempDir, fileName)
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
	err = util.Upload(url, values)
	if err != nil {
		log.Errorf("Error uploading files %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//If we put inspect before file upload, in case of pod is down, after video
	//inspection, the error will propagate to other service since task is
	//already created, but file hasn't uploaded
	vid := h.videoService.Inspect(f.Name())

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

	if err = os.Remove(f.Name()); err != nil {
		//just log the error, don't care if file isn't removed
		//handle the disk full error on later
		log.Error(err)
	}
	return

}
