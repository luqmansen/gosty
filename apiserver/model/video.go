package model

import (
	"gopkg.in/mgo.v2/bson"
)

type Video struct {
	//gorm.Model, changed to mongodb
	Id       bson.ObjectId `bson:"_id,omitempty" json:"id"`
	FileName string        `gorm:"size:255;not null;unique" json:"file_name"`
	// File size in kb
	Size     int     `json:"size"`
	Bitrate  int     `json:"bitrate"`
	Duration float32 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
}

func (m *Video) TableName() string {
	return "video"
}
