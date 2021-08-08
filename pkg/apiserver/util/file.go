package util

import (
	"bytes"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"
)

func Upload(uri string, values map[string]io.Reader) (err error) {

	// rewrite url to upload to primary replica
	newUri, err := url.ParseRequestURI(uri)
	if err != nil {
		log.Error(err)
	}
	//newUri.Host = "gosty-fileserver-0." + newUri.Host
	newUri.Host = newUri.Host
	uri = newUri.String()

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
		log.Error(err)
		return err
	}

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", uri, &b)
	if err != nil {
		log.Error(err)
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	var res *http.Response
	client := http.Client{}
	defer client.CloseIdleConnections()

	post := func() (err error) {

		res, err = client.Do(req)
		if err != nil {
			log.Errorf("Uploading file error: %s, uri: %s", err, uri)
			return
		}
		if res.StatusCode != http.StatusCreated {
			log.Errorf("Uploading file error: %d, uri: %s", res.StatusCode, uri)
			return errors.New(fmt.Sprintf("Failed to uploads to %s, status code : %d", uri, res.StatusCode))
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				log.Error(err)
			}
		}()
		return
	}
	back := backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 20)
	if err := backoff.Retry(post, back); err != nil {
		return err
	}

	log.Debugf("Upload to %s success", uri)
	return
}

func Download(filepath string, uri string) (err error) {

	var resp *http.Response

	get := func() error {
		resp, err = http.Get(uri)
		if err != nil {
			log.Errorf("Downloading file error: %s, url: %s", err, uri)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			log.Errorf("Downloading file error: %d, url: %s", resp.StatusCode, uri)
			return errors.New(fmt.Sprintf("Failed to download from %s, status code : %d", uri, resp.StatusCode))
		}
		return nil
	}

	back := backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), 20)
	if err := backoff.Retry(get, back); err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error(err)
		}
	}()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer func() {
		if err := out.Close(); err != nil {
			log.Error(err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	return err
}
