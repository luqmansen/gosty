package worker

import (
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

	originalFilePath := fmt.Sprintf("%s/%s", workdir, task.TaskSplit.Video.FileName)
	url := fmt.Sprintf("%s/files/%s", s.config.FileServer.GetFileServerUri(), task.TaskSplit.Video.FileName)
	err := util.Download(originalFilePath, url)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Processing task %s,  id: %s", models.TASK_NAME_ENUM[task.Kind], task.Id.Hex())

	cmd := exec.Command(
		"bash", wd+"/script/split.sh", fmt.Sprintf("%s/%s", workdir, task.TaskSplit.Video.FileName),
		strconv.FormatInt(task.TaskSplit.SizePerVid, 10), "-c copy")
	log.Debug(cmd.String())

	err = util.CommandExecLogger(cmd)
	if err != nil {
		log.Error(err)
		return err
	}

	files, err := ioutil.ReadDir(workdir)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Debugf("Reading directory %s: ", workdir)

	var tempFiles []string
	origFileNameWithoutExt := strings.Split(task.TaskSplit.Video.FileName, ".")[0]
	for _, f := range files {
		//only append related files of current task to be sent to fileserver,
		//prevent possibility of undeleted files from previous task
		//Below function is to remove filename(-1.ext) <- parentheses part and left the original filename
		splitFileName := strings.Join(strings.Split(strings.Split(f.Name(), ".")[0], "-")[:2], "-")
		if splitFileName == origFileNameWithoutExt {
			if f.Name() != task.OriginVideo.FileName { // yet we don't need the original file to be uploaded again
				tempFiles = append(tempFiles, f.Name())
			}
		}
	}

	if len(tempFiles) == 0 {
		return errors.New("worker.ProcessTaskSplit: no files found, something wrong")
	}

	// I prefer sorted slice, easier to check every video segment order
	sort.Strings(tempFiles)
	fmt.Println(tempFiles)

	errCh := make(chan error, 1)
	waitCh := make(chan struct{}, 1)
	videoResults := make([]*models.Video, len(tempFiles))
	wg := sync.WaitGroup{}

	go func() {
		defer close(waitCh)
		defer wg.Wait()

		for idx, file := range tempFiles {
			if file == task.TaskSplit.Video.FileName {
				continue
			} // skip original file, in case previous checking is failed

			wg.Add(1)

			go func(idx int, fileName string) {
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
				// Append successfully uploaded file to video result list
				// TODO: refactor splited video result data structure
				// this part is super redundant, as every single property
				// is same as parent video (except duration & size)
				videoResults[idx] = &models.Video{
					FileName: fileName,
					Size:     task.OriginVideo.Size,
					Bitrate:  task.OriginVideo.Bitrate,
					Duration: task.OriginVideo.Duration,
					Width:    task.OriginVideo.Width,
					Height:   task.OriginVideo.Height,
				}

				if err = os.Remove(filePath); err != nil {
					log.Errorf("Failed to remove file %s: %s", filePath, err)
					return
				}

			}(idx, file)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Debugf("removing %s", originalFilePath)
		err = os.Remove(originalFilePath)
		if err != nil {
			log.Errorf("Failed to remove original file %s: %s", originalFilePath, err)
		}

	}()

	wg.Wait()

	select {
	case err = <-errCh:
		task.Status = models.TaskStatusFailed
		return err
	case <-waitCh:
		task.TaskDuration = time.Since(start)
		task.TaskCompleted = time.Now()
		task.Status = models.TaskStatusDone
		task.TaskSplit.SplitedVideo = videoResults
		return nil
	}
}
