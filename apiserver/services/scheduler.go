package services

import (
	"github.com/luqmansen/gosty/apiserver/models"
)

type SchedulerService interface {
	CreateSplitTask(video *models.Video) error
	CreateTranscodeTask(video *models.Video) error
	CreateMergeTask(video *models.Video) error
	ReadMessages()
	//UpdateTask(task *models.Task) error
	DeleteTask(taskId string) error
}
