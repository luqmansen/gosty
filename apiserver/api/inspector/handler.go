package inspector

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"github.com/luqmansen/gosty/apiserver/services"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
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
	values := map[string]io.Reader{"video": f}
	actualFileName := strings.Split(f.Name(), "/")[1] // remove /tmp/ on filepath
	err = Upload("http://localhost:8001/upload?filename="+actualFileName, values)
	if err != nil {
		log.Fatal(err)
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

func Upload(url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	err = w.Close()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Check the response
	if res.StatusCode >= 400 {
		err = fmt.Errorf("bad status: %s . res.Status")
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatal(string(b))
		}
	}
	log.Debug("Upload file success")
	return
}
