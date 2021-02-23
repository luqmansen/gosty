package mongo

import (
	"context"
	"github.com/luqmansen/gosty/apiserver/models"
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

func (r taskRepository) Get(taskId uint) models.Task {
	panic("implement me")
}

func (r taskRepository) Add(task *models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	c := r.db.client.Database(r.db.database).Collection(task.TableName())
	if _, e := c.InsertOne(ctx, task); e != nil {
		return errors.Wrap(e, "repository.Task.Add")
	}
	return nil
}

func (r taskRepository) Update(taskId uint) error {
	panic("implement me")
}

func (r taskRepository) Delete(taskId uint) error {
	panic("implement me")
}
