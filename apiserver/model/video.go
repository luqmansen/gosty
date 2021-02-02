package model

import "github.com/jinzhu/gorm"

type Video struct {
	gorm.Model
	FileName   string `gorm:"size:255;not null;unique" json:"file_name"`
	FileSize   int    `json:"file_size"`
	Bitrate    int    `json:"bitrate"`
	Resolution string `gorm:"size:255;" json:"resolution"`
}
