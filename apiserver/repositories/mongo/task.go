package mongo

import (
	"github.com/luqmansen/gosty/apiserver/model"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/pkg/errors"
	"time"
)

type taskRepository struct {
	db mongoRepository
}

func NewTaskRepository(uri, db string, mongoTimeout int) (repositories.TaskRepository, error) {
	repo := &taskRepository{
		db: mongoRepository{
			timeout:  time.Duration(mongoTimeout) * time.Second,
			database: db,
		},
	}
	client, e := newMongoClient(uri, mongoTimeout)
	if e != nil {
		return nil, errors.Wrap(e, "repository.NewNewsRepository")
	}
	repo.db.client = client
	return repo, nil
}

func (t taskRepository) Get(taskId uint) model.Task {
	panic("implement me")
}

func (t taskRepository) Add(task *model.Task) error {
	return nil
}

func (t taskRepository) Update(taskId uint) error {
	panic("implement me")
}

func (t taskRepository) Delete(taskId uint) error {
	panic("implement me")
}
