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

type Task struct {
	gorm.Model
	Video          Video      `gorm:"foreignKey:ID"`
	TargetRes      string     `gorm:"size:255;" json:"target_res"`
	TargetBitrate  string     `gorm:"size:255;" json:"target_bitrate"`
	TargetEncoding string     `json:"target_encoding"`
	Status         TaskStatus `json:"status"`
	Worker         string     `gorm:"size:255" json:"worker"`
	CompletedAt    time.Time  `json:"completed_at"`
	TaskDuration   time.Duration
}
