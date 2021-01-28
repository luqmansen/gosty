package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

type TaskStatus int

const (
	TaskStatusDone TaskStatus = iota
	TaskStatusOnprogress
	TaskStatusFailed
)

type WorkerStatus int

const (
	WorkerStatusIdle WorkerStatus = iota
	WorkerStatusWorking
	WorkerStatusTerminated
)

type (
	Video struct {
		gorm.Model
		FileName   string `gorm:"size:255;not null;unique" json:"file_name"`
		FileSize   int    `json:"file_size"`
		Bitrate    int    `json:"bitrate"`
		Resolution string `gorm:"size:255;" json:"resolution"`
	}

	Task struct {
		gorm.Model
		Video          Video
		TargetRes      string     `gorm:"size:255;" json:"target_res"`
		TargetBitrate  string     `gorm:"size:255;" json:"target_bitrate""`
		TargetEncoding string     `json:"target_encoding"`
		Status         TaskStatus `json:"status"`
		Worker         string     `gorm:"size:255" json:"worker"`
		CompletedAt    time.Time  `json:"completed_at"`
		TaskDuration   time.Duration
	}

	Worker struct {
		gorm.Model
		ContainerId  string
		Status       WorkerStatus
		LastAssigned time.Time
	}
)
