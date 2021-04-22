package mongo

import (
	"context"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const (
	taskCollectionName = "task"
)

type taskRepository struct {
	db mongoRepository
}

func NewTaskRepository(cfg config.Database) (repositories.TaskRepository, error) {
	repo := &taskRepository{
		db: mongoRepository{
			timeout:  time.Duration(cfg.Timeout) * time.Second,
			database: cfg.Name,
		},
	}
	client, e := newMongoClient(cfg.GetDatabaseUri(), cfg.Timeout)
	if e != nil {
		return nil, errors.Wrap(e, "repository.NewNewsRepository")
	}
	repo.db.client = client
	return repo, nil
}

func (r taskRepository) Get(taskId string) models.Task {
	panic("implement me")
}

func (r taskRepository) GetOneByVideoNameAndKind(name string, kind models.TaskKind) (*models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()

	filter := &bson.M{
		"originvideo.filename": name,
		"kind":                 kind,
	}

	log.Debugf("repositories.task.GetOneByNameAndKind, name :%s, kind: %d, filter: %v", name, kind, filter)

	result := &models.Task{}
	c := r.db.client.Database(r.db.database).Collection(taskCollectionName)
	err := c.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "repositories.task.GetOneByNameAndKind")
	}
	return result, nil

}
func (r taskRepository) GetTranscodeTasksByVideoNameAndResolution(name, resolution string) (result []*models.Task, err error) {
	log.Debugf("repositories.task.GetTranscodeTasksByVideoNameAndResolution :%s %s", name, resolution)
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()

	filter := &bson.M{
		"originvideo.filename":      name,
		"tasktranscode.targetres":   resolution,
		"tasktranscode.resultvideo": bson.M{"$ne": nil}, // make sure to only query finished task
	}

	coll := r.db.client.Database(r.db.database).Collection(taskCollectionName)
	cur, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "error finding document")
	}
	for cur.Next(ctx) {
		var elem models.Task
		if err := cur.Decode(&elem); err != nil {
			return nil, errors.Wrap(err, "error when decoding element")
		}
		result = append(result, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, errors.Wrap(err, "error on cursor")

	}

	err = cur.Close(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to close cursor")
	}
	return result, nil
}

func (r taskRepository) Add(task *models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	c := r.db.client.Database(r.db.database).Collection(taskCollectionName)
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
	c := r.db.client.Database(r.db.database).Collection(taskCollectionName)
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
