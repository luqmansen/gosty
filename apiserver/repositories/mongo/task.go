package mongo

import (
	"context"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r taskRepository) Get(taskId string) models.Task {
	panic("implement me")
}

func (r taskRepository) Add(task *models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	c := r.db.client.Database(r.db.database).Collection(task.TableName())
	res, e := c.InsertOne(ctx, task)
	if e != nil {
		return errors.Wrap(e, "repository.Task.Add")
	}
	task.Id = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r taskRepository) Update(task *models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	pByte, err := bson.Marshal(task)
	if err != nil {
		return err
	}

	var update bson.M
	err = bson.Unmarshal(pByte, &update)
	if err != nil {
		return err
	}
	c := r.db.client.Database(r.db.database).Collection(task.TableName())
	oId, _ := primitive.ObjectIDFromHex(task.Id.Hex())
	_, err = c.UpdateOne(
		ctx,
		bson.M{
			"_id": oId,
		},
		bson.D{{Key: "$set", Value: update}},
	)

	if err != nil {
		return errors.Wrap(err, "repository.Task.Add")
	}
	return nil
}

func (r taskRepository) Delete(taskId string) error {
	panic("implement me")
}
