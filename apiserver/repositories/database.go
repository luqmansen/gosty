package repositories

import "github.com/luqmansen/gosty/apiserver/model"

type VideoRepository interface {
	Get(videoId uint) model.Video
	Add(video *model.Video)
	Update(videoId uint) error
	Delete(videoId uint) error
}

type TaskRepository interface {
	Get(taskId uint) model.Task
	Add(task *model.Task) error
	Update(taskId uint) error
	Delete(taskId uint) error
}

type WorkerRepository interface {
	Get(workerId uint) model.Worker
	Add(worker *model.Worker) error
	Delete(workerId uint) error
}
