package services

import (
	"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"math"
	"time"
)

type schedulerServices struct {
	repo repositories.TaskRepository
}

func NewSchedulerService(repo repositories.TaskRepository) SchedulerService {
	return &schedulerServices{repo: repo}
}

func (s schedulerServices) CreateSplitTask(video *model.Video) error {
	//split by size in Byte
	var sizePerVid int
	var sizeLeft int
	var minSize = 1024 << 10 // 10 MB

	// if video size less than min file size, forward to transcode task
	if video.Size < minSize {
		err := s.CreateTranscodeTask(video)
		if err != nil {
			return err
		}
		return nil
	} else {
		//split per 10 MB files
		sizePerVid = minSize
		sizeLeft = video.Size % minSize
	}

	task := model.Task{
		Kind: model.TaskSplit,
		TaskSplit: model.SplitTask{
			Video:       *video,
			TargetChunk: int(math.Ceil(float64(video.Size) / float64(minSize))),
			SizePerVid:  sizePerVid,
			SizeLeft:    sizeLeft,
		},
		Status:       model.TaskQueued,
		CompletedAt:  time.Time{},
		TaskDuration: 0,
	}
	//save to db
	err := s.repo.Add(&task)
	if err != nil {
		panic(err)
	}
	//publish
	return nil
}

func (s schedulerServices) CreateTranscodeTask(video *model.Video) error {
	panic("implement me")
}

func (s schedulerServices) UpdateTask(taskId uint) error {
	panic("implement me")
}

func (s schedulerServices) DeleteTask(taskId uint) error {
	panic("implement me")
}
