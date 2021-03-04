package inspector

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/luqmansen/gosty/apiserver/pkg"
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
}

func NewInspectorHandler(inspectorSvc services.VideoInspectorService) VideoInspectorHandler {
	return &handler{inspectorSvc}

}

func (h handler) Get(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
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

	//hash := md5.New()
	//n, err := io.Copy(hash, buf)
	//log.Debugf("Byte written for hash: ", n)
	//if err != nil {
	//	log.Fatal(err.Error())
	//}

	fileName := fmt.Sprintf("%s-*%s", uuid.NewString(), ext[0])
	f, err := ioutil.TempFile("./tmp/", fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var maxSize int64 = 32 << 20
	lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511))

	n, err := io.Copy(f, lmt)
	log.Debugf("Byte written for file: ", n)
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		log.Errorf("Error seeking file")
	}

	//upload to file server
	values := map[string]io.Reader{"file": f}
	actualFileName := strings.Split(f.Name(), "/")[1] // remove /tmp/ on filepath
	err = pkg.Upload("http://localhost:8001/upload?filename="+actualFileName, values)
	if err != nil {
		log.Error(err)
		return
	}

	//inspect
	vid := h.inspectorService.Inspect(f.Name())

	resp, err := json.Marshal(vid)
	if err != nil {
		log.Fatal(err)
	}

	//defer func() {
	//	if err := f.Close(); err != nil {
	//		log.Error(err)
	//	}
	//}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
	return

}
