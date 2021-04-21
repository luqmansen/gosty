package repositories

import "github.com/luqmansen/gosty/apiserver/models"

type VideoRepository interface {
	Get(videoId uint) models.Video
	GetAll(limit int64) ([]*models.Video, error)
	GetOneByName(key string) (*models.Video, error)
	Find(key string) []*models.Video
	Add(video *models.Video) error
	AddMany(videoList []*models.Video) error
	Update(video *models.Video) error
	Delete(videoId uint) error
}

type TaskRepository interface {
	Get(taskId string) models.Task
	GetOneByVideoNameAndKind(name string, kind models.TaskKind) (*models.Task, error)
	GetTranscodeTasksByVideoNameAndResolution(name, resolution string) ([]*models.Task, error)
	Add(task *models.Task) error
	Update(task *models.Task) error
	Delete(taskId string) error
}

type WorkerRepository interface {
	Get(workerId uint) models.Worker
	Add(worker *models.Worker) error
	Upsert(worker *models.Worker) error
	Delete(workerId uint) error
}
