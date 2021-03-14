package mongo

import (
	"context"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/pkg/util/config"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type workerRepository struct {
	db mongoRepository
}

func NewWorkerRepository(db config.Database) (repositories.WorkerRepository, error) {
	workerRepo := &workerRepository{
		db: mongoRepository{
			timeout:  time.Duration(db.Timeout) * time.Second,
			database: db.Name,
		},
	}
	client, e := newMongoClient(db.GetDatabaseUri(), db.Timeout)
	if e != nil {
		return nil, errors.Wrap(e, "repositories.NewWorkerRepository")
	}
	workerRepo.db.client = client
	return workerRepo, nil
}

func (w workerRepository) Upsert(worker *models.Worker) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.db.timeout)
	defer cancel()

	pByte, err := bson.Marshal(worker)
	if err != nil {
		return err
	}

	var workerUpdate bson.M
	err = bson.Unmarshal(pByte, &workerUpdate)
	if err != nil {
		return err
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"workerpodname": worker.WorkerPodName}
	update := bson.D{{Key: "$set", Value: workerUpdate}}

	c := w.db.client.Database(w.db.database).Collection(worker.TableName())
	if _, e := c.UpdateOne(ctx, filter, update, opts); e != nil {
		return errors.Wrap(e, "repositories.Worker.Upsert")
	}
	return nil
}

func (w workerRepository) Get(workerId uint) models.Worker {
	panic("implement me")
}

func (w workerRepository) Add(worker *models.Worker) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.db.timeout)
	defer cancel()
	c := w.db.client.Database(w.db.database).Collection(worker.TableName())
	if _, e := c.InsertOne(ctx, worker); e != nil {
		return errors.Wrap(e, "repositories.Worker.Add")
	}
	return nil
}

func (w workerRepository) Delete(workerId uint) error {
	panic("implement me")
}
