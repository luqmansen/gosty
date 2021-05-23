package worker

import (
	"bytes"
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *Svc) ProcessTaskSplit(task *models.Task) error {
	start := time.Now()

	filePath := fmt.Sprintf("%s/%s", workdir, task.TaskSplit.Video.FileName)
	log.Debug(filePath)
	url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), task.TaskSplit.Video.FileName)
	err := util.Download(filePath, url)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task %s,  id: %s", models.TASK_NAME_ENUM[task.Kind], task.Id.Hex())

	cmd := exec.Command(
		"bash", wd+"/script/split.sh", fmt.Sprintf("%s/%s", workdir, task.TaskSplit.Video.FileName),
		strconv.FormatInt(task.TaskSplit.SizePerVid, 10), "-c copy")
	log.Debug(cmd.String())

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Dir = wd

	err = cmd.Run()
	if err != nil {
		log.Error(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	log.Infof("Processing done %s: ", out.String())

	files, err := ioutil.ReadDir(workdir)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Debugf("Reading directory %s: ", workdir)

	var tempFiles []string
	for _, f := range files {
		//only append related files of current task to be sent to fileserver,
		//prevent possibility of undeleted files from previous task
		splitFileName := strings.Join(strings.Split(strings.Split(f.Name(), ".")[0], "-")[:2], "-")
		originalFileName := strings.Split(task.TaskSplit.Video.FileName, ".")[0]
		if splitFileName == originalFileName {
			tempFiles = append(tempFiles, f.Name())
		}

	}
	if len(tempFiles) == 0 {
		return errors.New("worker.ProcessTaskSplit: no files found, something wrong")
	}

	// I prefer sorted slice, easier to check every video segment order
	sort.Strings(tempFiles)

	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	for _, file := range tempFiles {
		if file == task.TaskSplit.Video.FileName {
			continue
		} // skip original file

		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()

			filePath := fmt.Sprintf("%s/%s", workdir, fileName)
			fileReader, err := os.Open(filePath)
			if err != nil {
				log.Error(err)
				errCh <- err
				return
			}

			values := map[string]io.Reader{"file": fileReader}
			url := fmt.Sprintf("%s/upload?filename=%s", s.config.FileServer.GetFileServerUri(), fileName)
			if err = util.Upload(url, values); err != nil {
				log.Error(err)
				errCh <- err
				return
			}
			if err = os.Remove(filePath); err != nil {
				log.Error(err)
				errCh <- err
				return
			}
		}(file)
	}

	wg.Add(1)
	go func(w *sync.WaitGroup) {
		log.Debugf("removing %s", filePath)
		err = os.Remove(filePath)
		if err != nil {
			log.Error(err)
			errCh <- err
		}
		w.Done()
	}(&wg)

	var videoList []*models.Video
	for _, fileName := range tempFiles {
		if fileName == task.TaskSplit.Video.FileName {
			continue
		} // skip original file

		// TODO: refactor splited video result data structure
		// this part is super redundant, as every single property
		// is same as parent video (except duration maybe)
		videoList = append(videoList, &models.Video{
			FileName: fileName,
			Size:     task.OriginVideo.Size,
			Bitrate:  task.OriginVideo.Bitrate,
			Duration: task.OriginVideo.Duration,
			Width:    task.OriginVideo.Width,
			Height:   task.OriginVideo.Height,
		})
	}

	select {
	case err = <-errCh:
		return err
	default:
		wg.Wait()
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskSplit.SplitedVideo = videoList
		return nil
	}
}
