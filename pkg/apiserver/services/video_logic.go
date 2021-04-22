package services

import (
	"github.com/luqmansen/gosty/pkg/apiserver/models"
	"github.com/luqmansen/gosty/pkg/apiserver/repositories"
	fluentffmpeg "github.com/modfy/fluent-ffmpeg"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"strconv"
	"strings"
)

type videoServices struct {
	vidRepo      repositories.VideoRepository
	schedulerSvc SchedulerService
}

func NewVideoService(vidRepo repositories.VideoRepository, schedulerSvc SchedulerService) VideoService {
	return &videoServices{vidRepo, schedulerSvc}
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

func (v videoServices) GetAll() (vids []*models.Video) {
	vids, err := v.vidRepo.GetAll(12)
	if err != nil {
		log.Error(err)
	}
	return
}
