package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Video struct {
	//gorm.Model, changed to mongodb
	Id       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FileName string             `gorm:"size:255;not null;unique" json:"file_name"`
	// File size in Byte
	Size     int64   `json:"size"`
	Bitrate  int     `json:"bitrate"`
	Duration float32 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`

	Audio *Audio   `json:"audio"`
	Video []*Video `json:"video"`

	DashFile []string
}

func (m *Video) TableName() string {
	return "video"
}
