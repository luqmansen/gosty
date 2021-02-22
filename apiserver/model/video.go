package model

import "github.com/jinzhu/gorm"

type Video struct {
	gorm.Model
	FileName string `gorm:"size:255;not null;unique" json:"file_name"`
	// File size in kb
	Size     int     `json:"size"`
	Bitrate  int     `json:"bitrate"`
	Duration float32 `json:"duration"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
}
