package services

import (
	"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
)

type schedulerServices struct {
	repo repositories.TaskRepository
}

func NewSchedulerService(repo repositories.TaskRepository) SchedulerService {
	return &schedulerServices{repo: repo}
}

func (s schedulerServices) UpdateTask(taskId uint) error {
	panic("implement me")
}

func (s schedulerServices) DeleteTask(taskId uint) error {
	panic("implement me")
}

func (s schedulerServices) CreateSplitTask(task *model.Task) error {
	return nil
}
