package services

import (
	"github.com/luqmansen/gosty/apiserver/models"
)

type SchedulerService interface {
	CreateSplitTask(video *models.Video) error
	CreateTranscodeTask(video *models.Video) error
	//CreateCombineTask
	UpdateTask(taskId uint) error
	DeleteTask(taskId uint) error
}
