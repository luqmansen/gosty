package services

import (
	"github.com/luqmansen/gosty/apiserver/model"
)

type SchedulerService interface {
	CreateSplitTask(task *model.Task) error
	//CreateCombineTask
	//CreateTranscodeTask
	UpdateTask(taskId uint) error
	DeleteTask(taskId uint) error
}
