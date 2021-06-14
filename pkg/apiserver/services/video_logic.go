package services

import (
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"strconv"
	"strings"
	"time"
)

type videoServices struct {
	vidRepo      repositories.VideoRepository
	schedulerSvc Scheduler
	cache        *cache.Cache
}

const (
	KeyGetAllVideo = "KeyGetAllVideo"
)

func NewVideoService(vidRepo repositories.VideoRepository, schedulerSvc Scheduler, cache *cache.Cache) VideoService {
	return &videoServices{vidRepo, schedulerSvc, cache}
}

func (v videoServices) Inspect(file string) models.Video {

	wd, _ := os.Getwd()
	result, err := fluentffmpeg.Probe(wd + "/" + file)
	if err != nil {
		log.Fatal(err)
	}

	format := result["format"].(map[string]interface{})
	duration, err := strconv.ParseFloat(format["duration"].(string), 32)
	if err != nil {
		log.Error(err)
	}
	size, err := strconv.ParseInt(format["size"].(string), 10, 32)
	if err != nil {
		log.Error(err)
	}

	bitrate, err := strconv.ParseInt(format["bit_rate"].(string), 10, 32)
	if err != nil {
		log.Error(err)
	}

	streams := result["streams"].([]interface{})[0].(map[string]interface{})

	//file is full path of the file, we only need the actual name
	fileName := strings.Split(file, "/")

	video := models.Video{
		Id:       primitive.NewObjectID(),
		FileName: fileName[len(fileName)-1],
		Size:     size,
		Bitrate:  int(bitrate), //streams["bit_rate"].(int),
		Duration: float32(duration),
		Width:    int(streams["coded_width"].(float64)),
		Height:   int(streams["coded_height"].(float64)),
	}

	err = v.vidRepo.Add(&video)
	if err != nil {
		log.Error(err)
	}

	err = v.schedulerSvc.CreateSplitTask(&video)
	if err != nil {
		log.Error(err)
	}

	return video
}

func (v *videoServices) GetAll() (vids []*models.Video, err error) {
	res, found := v.cache.Get(KeyGetAllVideo)
	if found {
		return res.([]*models.Video), nil
	} else {
		vids, err = v.vidRepo.GetAvailable(100)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		v.cache.Set(KeyGetAllVideo, vids, 30*time.Second)

		return vids, nil
	}
}
