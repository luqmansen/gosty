package fileserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type fileInfo struct {
	Origin   string
	FileName string
}

func (h *fileServer) InitialSync() {
	files, err := ioutil.ReadDir(h.pathToServe)
	if err != nil {
		log.Error(err)
	}
	if len(files) > 0 {
		errs, _ := errgroup.WithContext(context.Background())
		// initial sync, broadcast to all peers every file
		// todo: optimize to single network call
		for _, f := range files {
			f := f
			h.syncMapFileLists.Store(f.Name(), f)
			errs.Go(
				func() error {
					err := h.peerChecker()
					if err != nil {
						log.Error(err)
						return err
					} else {
						// skip broadcast to dead peer
						return h.broadcastToAllPeers(f)
					}
				})
		}
		if err = errs.Wait(); err != nil {
			log.Error(err)
		}
	}
}

// Make sure all peer hosts alive
func (h *fileServer) peerChecker() error {
	ctx, cancelFunc := context.WithTimeout(
		context.Background(), 10*time.Second)
	defer cancelFunc()

	errs, _ := errgroup.WithContext(ctx)
	for _, peer := range h.peerFileServerHost {
		peer := peer
		errs.Go(func() error {
			err := backoff.Retry(
				func() error {
					resp, err := http.Get("http://" + peer)
					if err != nil {
						return err
					}
					if resp.StatusCode != http.StatusOK {
						return errors.New("status code != 200")
					}
					return nil
				}, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))

			return err
		})
	}
	return errs.Wait()
}

func (h *fileServer) ExecuteSynchronization() {
	for {
		for h.syncQueue.Len() > 0 {
			e := h.syncQueue.Front()
			file := e.Value.(fileInfo)
			url := fmt.Sprintf("http://%s/files/%s", file.Origin, file.FileName)
			path := h.pathToServe + "/" + file.FileName

			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := util.Download(path, url); err != nil {
					log.Errorf("Failed to download %s from %s", file.FileName, file.Origin)
					h.syncQueue.PushBack(e) //re-queue
				} else {
					h.syncMapFileLists.Store(file.FileName, file)
					h.syncQueue.Remove(e) // dequeue
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (h *fileServer) SyncHook(w http.ResponseWriter, r *http.Request) {
	var file fileInfo
	err := json.NewDecoder(r.Body).Decode(&file)
	if err != nil {
		log.Error(err)
	}
	h.syncQueue.PushBack(file)
}

func (h *fileServer) triggerSync() {
	files, err := ioutil.ReadDir(h.pathToServe)
	if err != nil {
		log.Error(err)
	}
	errs, _ := errgroup.WithContext(context.Background())
	for _, f := range files {
		f := f
		_, exists := h.syncMapFileLists.Load(f.Name())
		if !exists {
			h.syncMapFileLists.Store(f.Name(), f)
			//not exists on map, so it must be the new file
			//inform to other file server
			errs.Go(
				func() error {
					return h.broadcastToAllPeers(f)
				})

		}
	}
	if err = errs.Wait(); err != nil {
		log.Error(err)
	}
}

func (h *fileServer) broadcastToAllPeers(file fs.FileInfo) error {
	errs, _ := errgroup.WithContext(context.Background())
	for _, hosts := range h.peerFileServerHost {
		hosts := hosts
		errs.Go(
			func() error {
				payload := &fileInfo{
					Origin:   h.selfHost,
					FileName: file.Name(),
				}
				b, err := json.Marshal(payload)
				if err != nil {
					log.Error(err)
					return err
				}
				url := fmt.Sprintf("http://%s/sync", hosts)
				return backoff.Retry(
					func() error {
						resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
						if err != nil {
							log.Error(err)
						}

						if resp.StatusCode != 200 {
							log.Errorf("resp status code %d", resp.StatusCode)
						}
						return err
					}, backoff.NewExponentialBackOff())
			},
		)
	}
	return errs.Wait()
}

func (h *fileServer) uploadToAllPeers(filename string) error {
	errs, _ := errgroup.WithContext(context.Background())
	filePath := h.pathToServe + "/" + filename
	fileReader, err := os.Open(filePath)
	if err != nil {
		log.Errorf("error opening file : %s", err)
		return err
	}
	for _, hosts := range h.peerFileServerHost {
		hosts := hosts
		errs.Go(
			func() error {
				values := map[string]io.Reader{"file": fileReader}
				url := fmt.Sprintf("http://%s/upload?filename=%s", hosts, filename)
				err := util.Upload(url, values)
				if err != nil {
					return err
				}
				return nil
			})

	}
	return errs.Wait()
}
