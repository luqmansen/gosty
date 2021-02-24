package services

import (
	"encoding/json"
	"fmt"
	"github.com/luqmansen/gosty/apiserver/models"
	"github.com/luqmansen/gosty/apiserver/repositories"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type videoInspectorServices struct {
	vidRepo      repositories.VideoRepository
	schedulerSvc SchedulerService
}

func NewInspectorService(vidRepo repositories.VideoRepository, schedulerSvc SchedulerService) VideoInspectorService {
	return &videoInspectorServices{vidRepo, schedulerSvc}
}

func (v videoInspectorServices) Inspect(file string) models.Video {

	//only get video stream (v:0 means video stream idx 0)
	wd, _ := os.Getwd()
	cmd := exec.Command("/usr/bin/ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", "-select_streams", "v:0", wd + "/" + file)
	//cmd.Dir = wd
	//stdout, err := cmd.Output()
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(fmt.Sprint(err) + ": " + string(stdout))
	}

	// json to map string
	var result map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &result)
	if err != nil {
		log.Error(err)
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
		FileName: fileName[len(fileName)-1],
		Size:     int(size),
		//mkv doesn't contains metadata for bitrate, and we don't really need it right now
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
