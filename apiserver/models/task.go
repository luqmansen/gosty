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
	TaskNew TaskKind = iota // new task, not from other task
	TaskSplit
	TaskMerge
	TaskTranscode
	TaskDash
)

type TaskTranscodeType int

const (
	TranscodeVideo TaskTranscodeType = iota
	TranscodeAudio
)

type (
	Task struct {
		Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
		OriginVideo *Video             `json:"origin_video"`

		Kind          TaskKind       `json:"kind"`
		TaskSplit     *SplitTask     `json:"task_split"`
		TaskTranscode *TranscodeTask `json:"task_transcode"`
		TaskMerge     *MergeTask     `json:"task_merge"`
		TaskDash      *DashTask      `json:"task_dash"`
		PrevTask      TaskKind       `json:"prev_task"`

		Status TaskStatus `json:"status"`
		Worker string     `gorm:"size:255" json:"worker"`

		TaskSubmitted time.Time     `json:"task_submitted"` // time (approx <1ms) when task submitted to queue
		TaskCompleted time.Time     `json:"task_completed"` // time when task finished by worker
		TaskDuration  time.Duration `json:"task_duration"`
	}

	SplitTask struct {
		Video *Video
		// Split to X chunk
		TargetChunk  int      `json:"target_chunk"`
		SizePerVid   int64    `json:"size_per_vid"`
		SizeLeft     int64    `json:"size_left"`
		SplitedVideo []*Video `json:"splited_video"`
	}

	MergeTask struct {
		ListVideo []*Video `json:"list_video"`
	}

	TranscodeTask struct {
		TranscodeType   TaskTranscodeType `json:"transcode_type"`
		Video           *Video            `json:"video"`
		TargetRes       string            `gorm:"size:255;" json:"target_res"`
		TargetBitrate   int               `gorm:"size:255;" json:"target_bitrate"`
		TargetEncoding  string            `json:"target_encoding"`
		TargetReprCount int               `json:"target_representation_count"` //Number of target representation
		ResultVideo     *Video            `json:"result_video"`
		ResultAudio     *Audio            `json:"result_audio"`
	}

	DashTask struct {
		ListVideo  []*Video
		ListAudio  []*Audio
		ResultDash []string
	}

	Audio struct {
		FileName string
		Bitrate  int
	}
)

func (t *Task) TableName() string {
	return "task"
}
