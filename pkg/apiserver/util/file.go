package util

import (
	"bytes"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

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
		logrus.Error(err)
		return err
	}

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		logrus.Error(err)
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	var res *http.Response
	post := func() (err error) {
		client := http.Client{}
		defer client.CloseIdleConnections()

		res, err = client.Do(req)
		if err != nil {
			logrus.Errorf("Uploading file error: %s, url: %s", err, url)
			return
		}
		if res.StatusCode != http.StatusCreated {
			logrus.Errorf("Uploading file error: %d, url: %s", res.StatusCode, url)
			return errors.New(fmt.Sprintf("Failed to uploads to %s, status code : %d", url, res.StatusCode))
		}
		return
	}

	if err := backoff.Retry(post, backoff.NewExponentialBackOff()); err != nil {
		return err
	}

	// Check the response
	if res.StatusCode >= 400 {
		err = fmt.Errorf("bad status: %s", res.Status)
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.Error(string(b))
		}

	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logrus.Error(err)
		}
	}()

	logrus.Debugf("Upload to %s success", url)
	return
}

func Download(filepath string, url string) (err error) {

	var resp *http.Response

	get := func() error {
		resp, err = http.Get(url)
		if err != nil {
			logrus.Errorf("Downloading file error: %s, url: %s", err, url)
			return err
		}
		if resp.StatusCode != http.StatusOK {
			logrus.Errorf("Downloading file error: %d, url: %s", resp.StatusCode, url)
			return errors.New(fmt.Sprintf("Failed to download from %s, status code : %d", url, resp.StatusCode))
		}
		return nil
	}

	if err := backoff.Retry(get, backoff.NewConstantBackOff(1*time.Second)); err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Error(err)
		}
	}()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer func() {
		if err := out.Close(); err != nil {
			logrus.Error(err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	return err
}
