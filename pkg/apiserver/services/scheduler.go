package services

import (
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/r3labs/sse/v2"
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
	sse       *sse.Server
}

const (
	TaskHTTPEventStream = "task"
)

type SchedulerService interface {
	GetAllTaskProgress() []*models.TaskProgressResponse
	CreateSplitTask(video *models.Video) error
	CreateTranscodeTask(task *models.Task) error
	CreateDashTask(task *models.Task) error
	CreateMergeTask(task *models.Task) error
	ReadMessages()
	//UpdateTask(task *models.Task) error
	DeleteTask(taskId string) error
}
