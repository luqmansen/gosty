package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type TaskStatus int

const (
	TaskQueued TaskStatus = iota
	TaskStatusDone
	TaskStatusOnprogress
	TaskStatusFailed
)

type TaskKind int

const (
	TaskSplit TaskKind = iota
	TaskMerge
	TaskTranscode
)

type (
	Task struct {
		Id primitive.ObjectID `bson:"_id,omitempty" json:"id"`

		// Task kind, either split, transcode, or merge task
		Kind          TaskKind `json:"kind"`
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
		TargetChunk int     `json:"target_chunk"`
		SizePerVid  int     `json:"size_per_vid"`
		SizeLeft    int     `json:"size_left"`
		VideoList   []Video `json:"video_list"`
	}

	MergeTask struct {
		ListVideo []Video `gorm:"foreignKey:ID"`
	}

	TranscodeTask struct {
		Video          Video
		TargetRes      string  `gorm:"size:255;" json:"target_res"`
		TargetBitrate  int     `gorm:"size:255;" json:"target_bitrate"`
		TargetEncoding string  `json:"target_encoding"`
		TranscodeVideo []Video `json:"transcode_video"`
	}
)

func (t *Task) TableName() string {
	return "task"
}
