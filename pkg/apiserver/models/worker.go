package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type WorkerStatus int

const (
	WorkerStatusReady WorkerStatus = iota
	WorkerStatusWorking
	WorkerStatusTerminated
	WorkerStatusUnreachable
)

type Worker struct {
	Id            primitive.ObjectID `bson:"id,omitempty" json:"id"`
	WorkerPodName string             `json:"worker_pod_name"`
	IpAddress     string             `json:"ip_address"`
	Status        WorkerStatus       `json:"status"`
	WorkingOn     string             `json:"working_on"` // string of task id that worker working on
	UpdatedAt     time.Time          `json:"updated_at"`
}

func (w *Worker) TableName() string {
	return "worker"
}
