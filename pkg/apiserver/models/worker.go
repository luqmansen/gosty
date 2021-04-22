package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type WorkerStatus int

const (
	WorkerStatusIdle WorkerStatus = iota
	WorkerStatusWorking
	WorkerStatusTerminated
)

type Worker struct {
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	WorkerPodName string             `json:"worker_pod_name"`
	Status        WorkerStatus       `json:"status"`
	WorkingOn     string             `json:"working_on"` // string of task id that worker working on
	UpdatedAt     time.Time          `json:"updated_at"`
}

func (w *Worker) TableName() string {
	return "worker"
}
