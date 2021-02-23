package mongo

import (
	"context"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	"github.com/pkg/errors"
	"time"
)

type videoRepository struct {
	db mongoRepository
}

func NewVideoRepository(uri, db string, timeout int) (repositories.VideoRepository, error) {
	vidRepo := &videoRepository{
		db: mongoRepository{
			timeout:  time.Duration(timeout) * time.Second,
			database: db,
		},
	}
	client, e := newMongoClient(uri, timeout)
	if e != nil {
		return nil, errors.Wrap(e, "repository.NewVideoRepository")
	}
	vidRepo.db.client = client
	return vidRepo, nil
}

func (r videoRepository) Get(videoId uint) models.Video {
	panic("implement me")
}

func (r videoRepository) Add(video *models.Video) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.db.timeout)
	defer cancel()
	c := r.db.client.Database(r.db.database).Collection(video.TableName())
	if _, e := c.InsertOne(ctx, video); e != nil {
		return errors.Wrap(e, "repository.Video.Add")
	}
	return nil
}

func (r videoRepository) Update(videoId uint) error {
	panic("implement me")
}

func (r videoRepository) Delete(videoId uint) error {
	panic("implement me")
}
