package services

import (
	"github.com/luqmansen/gosty/pkg/apiserver/models"
)

type SchedulerService interface {
	CreateSplitTask(video *models.Video) error
	CreateTranscodeTask(task *models.Task) error
	CreateDashTask(task *models.Task) error
	CreateMergeTask(task *models.Task) error
	ReadMessages()
	//UpdateTask(task *models.Task) error
	DeleteTask(taskId string) error
}
