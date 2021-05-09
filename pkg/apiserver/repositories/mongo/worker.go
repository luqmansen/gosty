package mongo

import (
	"context"
	"github.com/luqmansen/gosty/pkg/apiserver/config"
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type workerRepository struct {
	db mongoRepository
}

const (
	workerCollectionName = "worker"
)

func NewWorkerRepository(db config.Database, client *mongo.Client) repositories.WorkerRepository {
	workerRepo := &workerRepository{
		db: mongoRepository{
			client:   client,
			timeout:  time.Duration(db.Timeout) * time.Second,
			database: db.Name,
		},
	}
	return workerRepo
}

func (w workerRepository) GetAll(limit int64) (result []*models.Worker, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), w.db.timeout)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetLimit(limit)
	filter := bson.M{}

	coll := w.db.client.Database(w.db.database).Collection(workerCollectionName)
	cur, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, errors.New("repositories.Worker.GetAll :" + err.Error())
	}

	for cur.Next(ctx) {

		var elem models.Worker
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		result = append(result, &elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	err = cur.Close(ctx)
	if err != nil {
		return nil, errors.New("repositories.Worker.GetAll :" + err.Error())
	}
	return result, nil
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

func (w workerRepository) Delete(podName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.db.timeout)
	defer cancel()
	c := w.db.client.Database(w.db.database).Collection(workerCollectionName)
	filter := bson.M{"workerpodname": podName}
	if _, e := c.DeleteOne(ctx, filter); e != nil {
		return errors.Wrap(e, "repositories.Worker.Delete")
	}
	return nil
}
