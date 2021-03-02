package repositories

import "github.com/luqmansen/gosty/apiserver/models"

type VideoRepository interface {
	Get(videoId uint) models.Video
	Add(video *models.Video) error
	Update(videoId uint) error
	Delete(videoId uint) error
}

type TaskRepository interface {
	Get(taskId string) models.Task
	Add(task *models.Task) error
	Update(task *models.Task) error
	Delete(taskId string) error
}

type WorkerRepository interface {
	Get(workerId uint) models.Worker
	Add(worker *models.Worker) error
	Delete(workerId uint) error
}
