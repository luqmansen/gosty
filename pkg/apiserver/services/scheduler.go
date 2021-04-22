package services

import (
	"github.com/luqmansen/gosty/pkg/apiserver/models"
)

const (
	MessageBrokerQueueTaskNew          = "task_new"
	MessageBrokerQueueTaskFinished     = "task_finished"
	MessageBrokerQueueTaskUpdateStatus = "task_update_status"
)

const (
	TaskHTTPEventStream = "task"
)

type Scheduler interface {
	GetAllTaskProgress() []*models.TaskProgressResponse
	CreateSplitTask(video *models.Video) error
	CreateTranscodeTask(task *models.Task) error
	CreateDashTask(task *models.Task) error
	CreateMergeTask(task *models.Task) error
	ReadMessages()
	//UpdateTask(task *models.Task) error
	DeleteTask(taskId string) error
}
