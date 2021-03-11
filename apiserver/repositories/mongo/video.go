package mongo

import (
	"context"
	"github.com/luqmansen/gosty/apiserver/config"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type videoRepository struct {
	db mongoRepository
}

func NewVideoRepository(cfg config.Database) (repositories.VideoRepository, error) {
	vidRepo := &videoRepository{
		db: mongoRepository{
			timeout:  time.Duration(cfg.Timeout) * time.Second,
			database: cfg.Name,
		},
	}
	client, e := newMongoClient(cfg.GetDatabaseUri(), cfg.Timeout)
	if e != nil {
		return nil, errors.Wrap(e, "repositories.NewVideoRepository")
	}
	vidRepo.db.client = client
	return vidRepo, nil
}

func (r videoRepository) GetOneByName(key string) (*models.Video, error) {
	log.Debugf("GetOneByName :%s", key)
	filter := &bson.M{"filename": key}
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()

	result := &models.Video{}
	c := r.db.client.Database(r.db.database).Collection(result.TableName())
	err := c.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, errors.New("repositories.Video.GetOneByName :" + err.Error())
	}
	return result, nil

}

func (r videoRepository) AddMany(videoList []*models.Video) error {

	docs := make([]interface{}, len(videoList))
	for i := range videoList {
		docs[i] = videoList[i]
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	c := r.db.client.Database(r.db.database).Collection(videoList[0].TableName())
	if _, e := c.InsertMany(ctx, docs); e != nil {
		return errors.Wrap(e, "repositories.Video.AddMany")
	}
	return nil
}

func (r videoRepository) Find(key string) []*models.Video {
	panic("implement me")
}

func (r videoRepository) Get(videoId uint) models.Video {
	panic("implement me")
}

func (r videoRepository) Add(video *models.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	c := r.db.client.Database(r.db.database).Collection(video.TableName())
	if _, e := c.InsertOne(ctx, video); e != nil {
		return errors.Wrap(e, "repositories.Video.Add")
	}
	return nil
}

func (r videoRepository) Update(video *models.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	pByte, err := bson.Marshal(video)
	if err != nil {
		return err
	}

	var update bson.M
	err = bson.Unmarshal(pByte, &update)
	if err != nil {
		return err
	}
	c := r.db.client.Database(r.db.database).Collection(video.TableName())
	oId, _ := primitive.ObjectIDFromHex(video.Id.Hex())
	_, err = c.UpdateOne(
		ctx,
		bson.M{
			"_id": oId,
		},
		bson.D{{Key: "$set", Value: update}},
	)

	if err != nil {
		return errors.Wrap(err, "repositories.Video.Update")
	}
	return nil
}

func (r videoRepository) Delete(videoId uint) error {
	panic("implement me")
}
