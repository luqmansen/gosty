package services

import (
	"encoding/json"
	"fmt"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/luqmansen/gosty/pkg/apiserver/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MessageBrokerQueueTaskNew          = "task_new"
	MessageBrokerQueueTaskFinished     = "task_finished"
	MessageBrokerQueueTaskUpdateStatus = "task_update_status"
)

type schedulerServices struct {
	taskRepo  repositories.TaskRepository
	videoRepo repositories.VideoRepository
	mb        repositories.MessageBrokerRepository
}

func NewSchedulerService(
	taskRepo repositories.TaskRepository,
	videoRepo repositories.VideoRepository,
	mb repositories.MessageBrokerRepository,
) SchedulerService {
	return &schedulerServices{
		taskRepo:  taskRepo,
		videoRepo: videoRepo,
		mb:        mb,
	}
}

func (s schedulerServices) GetAllTaskProgress() (result []*models.TaskProgressResponse) {
	//for every task from db, group them if they are from the same video
	allTask, err := s.taskRepo.GetAll(100)
	if err != nil {
		log.Error(err)
	}
	if len(allTask) == 0 {
		return nil
	}

	tempTask := make(map[string][]*models.Task)
	tempOriginVideo := make(map[string]*models.Video)
	for _, task := range allTask {
		idVideo := task.OriginVideo.FileName
		if val, ok := tempTask[idVideo]; ok {
			val = append(val, task)
			tempTask[idVideo] = val
		} else {
			tempTask[idVideo] = []*models.Task{task}
		}

		if _, ok := tempOriginVideo[idVideo]; !ok {
			tempOriginVideo[idVideo] = task.OriginVideo
		}

	}
	//sorting map by keys
	keys := make([]string, 0, len(tempTask))
	for k := range tempTask {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		originVideo := tempOriginVideo[k]
		result = append(result, &models.TaskProgressResponse{
			OriginVideo: originVideo,
			TaskList:    tempTask[k],
			TotalDuration: func() (t time.Duration) {
				for _, v := range tempTask[k] {
					t += v.TaskDuration
				}
				return
			}(),
		})
	}

	return
}

func (s schedulerServices) ReadMessages() {
	log.Debugf("Starting read message from %s", MessageBrokerQueueTaskFinished)
	forever := make(chan bool, 1)

	finishedTask := make(chan interface{})
	go s.mb.ReadMessage(finishedTask, MessageBrokerQueueTaskFinished)
	go s.scheduleTaskFromQueue(finishedTask)

	updateTaskStatusQueue := make(chan interface{})
	go s.mb.ReadMessage(updateTaskStatusQueue, MessageBrokerQueueTaskUpdateStatus)
	go s.updateTaskStatus(updateTaskStatusQueue)

	<-forever
}

func (s schedulerServices) updateTaskStatus(updateTaskStatusQueue chan interface{}) {
	for w := range updateTaskStatusQueue {
		msg := w.(amqp.Delivery)
		var task models.Task
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			log.Error(err)
		}

		if err := s.taskRepo.Update(&task); err != nil {
			log.Error(err)
		} else {
			if err = msg.Ack(false); err != nil {
				log.Error(err)
			}
		}

	}
}

func (s schedulerServices) scheduleTaskFromQueue(finishedTask chan interface{}) {
	//TODO: Refactor this repetitive message ack
	for t := range finishedTask {
		msg := t.(amqp.Delivery)
		var task models.Task
		err := json.Unmarshal(msg.Body, &task)
		if err != nil {
			log.Error(err)
		}

		log.Debugf("Updating finished task %s", task.Id.String())
		log.Debugf("Updating filename %s", task.OriginVideo.FileName)
		log.Debugf("Updating task kind %d", task.Kind)

		err = s.taskRepo.Update(&task)
		if err != nil {
			log.Error(err)
		}

		switch taskKind := task.Kind; taskKind {
		case models.TaskSplit:
			//save each splitted video into its own record
			if err := s.videoRepo.AddMany(task.TaskSplit.SplitedVideo); err != nil {
				log.Error(err)
				break
			}

			if err := s.createTranscodeTaskFromSplitTask(&task); err != nil {
				log.Error(err)
				break
			}

			if err = msg.Ack(false); err != nil {
				log.Error(err)
				break
			}

		case models.TaskTranscode:
			//TODO : Find one and update
			toUpdate, err := s.videoRepo.GetOneByName(task.TaskTranscode.Video.FileName)
			if err != nil {
				log.Error(err)
				break
			}

			if task.TaskTranscode.TranscodeType == models.TranscodeAudio {
				toUpdate.Audio = task.TaskTranscode.ResultAudio

			} else {
				toUpdate.Video = append(toUpdate.Video, task.TaskTranscode.ResultVideo)
			}
			if err = s.videoRepo.Update(toUpdate); err != nil {
				log.Error(err)
				break
			}

			//	check if previously is split task, then merge
			if task.PrevTask == models.TaskSplit {
				if err = s.CreateMergeTask(&task); err != nil {
					log.Error(errors.Wrap(err, "failed to create merge task"))
					break
				}
			} else {
				// must be a video with small size (prevTask == MessageBrokerQueueTaskNew)
				if err = s.CreateDashTask(&task); err != nil {
					log.Error(errors.Wrap(err, "failed to create dash task"))
					break
				}
			}

			if err = msg.Ack(false); err != nil {
				log.Error(err)
				break
			}
		case models.TaskMerge:
			//update video with result of merged video that has been merged
			//and of course also transcoded
			toUpdate, err := s.videoRepo.GetOneByName(task.OriginVideo.FileName)
			if err != nil {
				log.Error(errors.Wrap(err, "services.Scheduler.ReadTask.TaskMerge"))
				break
			}
			toUpdate.Video = append(toUpdate.Video, task.TaskMerge.Result)
			if err = s.videoRepo.Update(toUpdate); err != nil {
				log.Error(err)
				break
			}

			if err = s.CreateDashTask(&task); err != nil {
				log.Error(errors.Wrap(err, "failed to create dash task"))
				break
			}

			if err = msg.Ack(false); err != nil {
				log.Error(err)
				break
			}

		case models.TaskDash:
			file := strings.Split(task.TaskDash.ListVideo[0].FileName, "_")
			toUpdate, err := s.videoRepo.GetOneByName(fmt.Sprintf("%s.mp4", file[0]))
			if err != nil {
				log.Error(err)
			}

			toUpdate.DashFile = task.TaskDash.ResultDash
			if err = s.videoRepo.Update(toUpdate); err != nil {
				log.Error(err)
			}

			if err = msg.Ack(false); err != nil {
				log.Error(err)
			}

		}

	}
}

func (s schedulerServices) CreateSplitTask(video *models.Video) error {
	//split by size in Byte
	var sizePerVid int64
	var sizeLeft int64
	var minSize int64 = 10240 << 10 // 10 MB (Skip this until merge task is done)

	// if video size less than min file size, forward to transcode task
	if video.Size < minSize {
		//Todo: make this task definition not redundant.
		//Task is re-defined on CreateTranscodeTask, but
		//this is a current workaround for preserve origin
		//video field, later pls redesign the data models
		task := &models.Task{
			OriginVideo:   video,
			TaskTranscode: &models.TranscodeTask{Video: video},
		}
		err := s.CreateTranscodeTask(task)
		if err != nil {
			return err
		}

		//if err := s.createTranscodeAudioTask(video); err != nil {
		//	log.Error(err)
		//	return err
		//}
		return nil
	} else {
		//split per 10 MB files
		sizePerVid = minSize
		sizeLeft = video.Size % minSize
	}

	//if err := s.createTranscodeAudioTask(video); err != nil {
	//	log.Error(err)
	//	return err
	//}

	task := models.Task{
		OriginVideo: video,
		Kind:        models.TaskSplit,
		TaskSplit: &models.SplitTask{
			Video:       video,
			TargetChunk: int(math.Ceil(float64(video.Size) / float64(minSize))),
			SizePerVid:  sizePerVid,
			SizeLeft:    sizeLeft,
		},
		PrevTask:      models.TaskNew,
		Status:        models.TaskQueued,
		TaskSubmitted: time.Now(),
	}

	err := s.taskRepo.Add(&task)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = s.mb.Publish(task, MessageBrokerQueueTaskNew)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

// Transcode video input to all representation, each of video represent as task
// this function might be transcoding a part video (video with _00X name)
func (s schedulerServices) CreateTranscodeTask(task *models.Task) error {
	// check if previously from task split
	// file name on task split contain random-name_001
	videoToTranscode := task.TaskTranscode.Video
	fileName := strings.Split(videoToTranscode.FileName, "_")
	var prevTask models.TaskKind
	if len(fileName) == 1 {
		prevTask = models.TaskNew
	} else {
		prevTask = models.TaskSplit
	}

	//list taken from youtube available videoToTranscode resolution
	//br = audio bitrate, currently ignored
	availRes := []map[string]interface{}{
		{"res": "256x144", "br": 80_000},
		//{"res": "426x240", "br": 300_000},
		//{"res": "640x360", "br": 400_000},
		//{"res": "854x480", "br": 500_000},
		//{"res": "1280x720", "br": 1_500_000},
		//{"res": "1920x1080", "br": 3_000_000},
		//{"res": "2560x1440", "br": 6_000_000},
		//{"res": "3840x2160", "br": 13_000_000},
		//{"res": "7680x4320", "br": 20_000_000},
	}
	var target []map[string]interface{}
	for _, v := range availRes {
		//compare original videoToTranscode height with available resolution
		//only transcode to below or same resolution of original videoToTranscode
		//Take idx 1 of slice [width, height]
		h, _ := strconv.Atoi(strings.Split(v["res"].(string), "x")[1])
		if videoToTranscode.Height >= h {
			target = append(target, v)
		}
	}
	if len(target) == 0 {
		return errors.New("No target available")
	}

	var taskList []*models.Task
	for _, t := range target {
		taskList = append(taskList, &models.Task{
			OriginVideo: task.OriginVideo,
			Kind:        models.TaskTranscode,
			TaskTranscode: &models.TranscodeTask{
				TranscodeType:   models.TranscodeVideo,
				Video:           videoToTranscode,
				TargetRes:       t["res"].(string),
				TargetBitrate:   t["br"].(int),
				TargetEncoding:  "",
				TargetReprCount: len(target),
			},
			PrevTask:      prevTask,
			Status:        models.TaskQueued,
			TaskSubmitted: time.Now(),
		})
	}

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	for _, task := range taskList {
		fmt.Println("add transcode task")
		wg.Add(1)
		go func(t *models.Task) {
			defer wg.Done()

			//todo: add retry mechanism
			err := s.taskRepo.Add(t)
			if err != nil {
				log.Error(err)
				errChan <- err
				return
			}
			err = s.mb.Publish(t, MessageBrokerQueueTaskNew)
			if err != nil {
				log.Error(err)
				errChan <- err
				return
			}

		}(task)

	}
	select {
	case err := <-errChan:
		return err
	default:
		wg.Wait()
		return nil
	}
}

//For transcode audio
//Currently audio transcoding is disabled, original audio is embedded on audio,
//see note on task_transcode line ~60
func (s schedulerServices) createTranscodeAudioTask(video *models.Video) error {
	task := models.Task{
		Kind: models.TaskTranscode,
		TaskTranscode: &models.TranscodeTask{
			TranscodeType: models.TranscodeAudio,
			Video:         video,
		},
		Status:        models.TaskQueued,
		TaskSubmitted: time.Now(),
	}
	err := s.taskRepo.Add(&task)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = s.mb.Publish(task, MessageBrokerQueueTaskNew)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (s schedulerServices) createTranscodeTaskFromSplitTask(task *models.Task) error {
	var wg sync.WaitGroup
	errCh := make(chan error)
	for _, vid := range task.TaskSplit.SplitedVideo {
		wg.Add(1)
		func(v *models.Video) {
			defer wg.Done()
			task := &models.Task{
				OriginVideo:   task.OriginVideo,
				TaskTranscode: &models.TranscodeTask{Video: v},
			}
			err := s.CreateTranscodeTask(task)
			if err != nil {
				log.Error(err)
				errCh <- err
			}
		}(vid)
	}
	select {
	case err := <-errCh:
		return err
	default:
		wg.Wait()
		return nil
	}
}

// Merge task previously must be a split task, so the task parameter is models.Task
// with TaskTranscode struct filled. This function will be invoked everytime a
// Transcode task is finished.
//
// This function will only merge chunk of video with specific resolution
// one at a time. eg: 001_240p + 002_240p + 003_240p
func (s schedulerServices) CreateMergeTask(task *models.Task) error {
	if task.PrevTask != models.TaskSplit {
		return errors.New("previous task isn't split task")
	}
	if task.TaskTranscode == nil {
		return errors.New("TaskTranscode is nil")
	}
	// get all task's video with specific resolution that will be merged
	fileName := task.OriginVideo.FileName
	targetRes := task.TaskTranscode.TargetRes
	taskTranscodeList, err := s.taskRepo.GetTranscodeTasksByVideoNameAndResolution(fileName, targetRes)
	if err != nil {
		log.Error(err)
		return err
	}
	// get the split task to get the number of splited video
	splitTask, err := s.taskRepo.GetOneByVideoNameAndKind(fileName, models.TaskSplit)
	if err != nil {
		util.DebugStruct(*task)
		return err
	}
	// check if all all chunk of video is already transcoded
	if len(taskTranscodeList) != len(splitTask.TaskSplit.SplitedVideo) {
		log.Debugf("all transcode task for resolution %s haven't finished, need %s, get %s",
			targetRes, len(taskTranscodeList), len(splitTask.TaskSplit.SplitedVideo))
		return nil
	}

	var toMerge []*models.Video
	for _, t := range taskTranscodeList {
		toMerge = append(toMerge, t.TaskTranscode.ResultVideo)
	}
	if len(toMerge) == 0 {
		return errors.New("nothing to merge")
	}

	mergeTask := &models.Task{
		OriginVideo:   task.OriginVideo,
		Kind:          models.TaskMerge,
		TaskMerge:     &models.MergeTask{ListVideo: toMerge},
		PrevTask:      models.TaskTranscode,
		Status:        models.TaskQueued,
		TaskSubmitted: time.Now(),
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := s.taskRepo.Add(mergeTask)
		if err != nil {
			log.Error(err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		err = s.mb.Publish(mergeTask, MessageBrokerQueueTaskNew)
		if err != nil {
			log.Error(err)
			return
		}
	}()
	wg.Wait()

	return nil
}

func (s schedulerServices) CreateDashTask(task *models.Task) error {
	//Dash task will only be called after task transcode or task merge,
	video, err := s.videoRepo.GetOneByName(task.OriginVideo.FileName)
	if err != nil {
		log.Error(errors.Wrap(err, "services.Scheduler.CreateDashTask"))
		return err
	}

	//Disable this, see not on worker/task_transcode ~60
	//check if audio already transcoded
	//if video.Audio == nil{
	//	log.Debug("Audio still empty")
	//	return nil
	//}
	var prevTask models.TaskKind
	var reprCount int
	if task.Kind == models.TaskTranscode {
		reprCount = task.TaskTranscode.TargetReprCount
		prevTask = models.TaskTranscode
	} else {
		// must be previously from task merge
		prevTask = models.TaskMerge
		t, err := s.taskRepo.GetOneByVideoNameAndKind(task.OriginVideo.FileName, models.TaskTranscode)
		if err != nil {
			log.Errorf("scheduler.CreateDashTask: %s", err)
			return err
		}
		reprCount = t.TaskTranscode.TargetReprCount

	}
	if len(video.Video) != reprCount {
		log.Debug("Video transcoding haven't finished")
		return nil
	}
	if len(video.Video) == 0 {
		log.Errorf("scheduler.CreateDashTask: nothing to create dash")
		return nil
	}

	taskDash := &models.Task{
		OriginVideo: task.OriginVideo,
		Kind:        models.TaskDash,
		TaskDash: &models.DashTask{
			ListVideo: video.Video,
			ListAudio: []*models.Audio{video.Audio},
		},
		PrevTask:      prevTask,
		Status:        models.TaskQueued,
		TaskSubmitted: time.Now(),
	}

	err = s.taskRepo.Add(taskDash)
	if err != nil {
		log.Error(errors.Wrap(err, "services.Scheduler.CreateDashTask"))
		return err
	}

	err = s.mb.Publish(taskDash, MessageBrokerQueueTaskNew)
	if err != nil {
		log.Error(errors.Wrap(err, "services.Scheduler.CreateDashTask"))
		return err
	}

	return nil
}

func (s schedulerServices) DeleteTask(taskId string) error {
	panic("implement me")
}