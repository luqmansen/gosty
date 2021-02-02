package services

import "github.com/luqmansen/gosty/apiserver/model"

type SchedulerService interface {
	CreateTask(task *model.Task) error
	UpdateTask(taskId uint) error
	DeleteTask(taskId uint) error
}
