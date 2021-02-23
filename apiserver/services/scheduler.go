package services

import (
	"github.com/luqmansen/gosty/apiserver/model"
)

type SchedulerService interface {
	CreateSplitTask(video *model.Video) error
	CreateTranscodeTask(video *model.Video) error
	//CreateCombineTask
	UpdateTask(taskId uint) error
	DeleteTask(taskId uint) error
}
