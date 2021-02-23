package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type TaskStatus int

const (
	TaskStatusDone TaskStatus = iota
	TaskStatusOnprogress
	TaskStatusFailed
)

type (
	Task struct {
		Id bson.ObjectId `bson:"_id,omitempty" json:"id"`

		// Task kind, either split, transcode, or merge task
		Kind          string `json:"kind"`
		TaskSplit     SplitTask
		TaskTranscode TranscodeTask
		TaskMerge     MergeTask

		Status       TaskStatus `json:"status"`
		Worker       string     `gorm:"size:255" json:"worker"`
		CompletedAt  time.Time  `json:"completed_at"`
		TaskDuration time.Duration
	}

	SplitTask struct {
		Video Video
		// Split to X chunk
		TargetChunk    int `json:"target_chunk"`
		DurationPerVid int `json:"duration_per_vid"`
		DurationLeft   int `json:"duration_left"`
	}

	MergeTask struct {
		ListVideo []Video `gorm:"foreignKey:ID"`
	}

	TranscodeTask struct {
		Video          Video
		TargetRes      string `gorm:"size:255;" json:"target_res"`
		TargetBitrate  string `gorm:"size:255;" json:"target_bitrate"`
		TargetEncoding string `json:"target_encoding"`
	}
)
