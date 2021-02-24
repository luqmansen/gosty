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
	Id           primitive.ObjectID
	ContainerId  string
	Status       WorkerStatus
	LastAssigned time.Time
}
