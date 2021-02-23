package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type WorkerStatus int

const (
	WorkerStatusIdle WorkerStatus = iota
	WorkerStatusWorking
	WorkerStatusTerminated
)

type Worker struct {
	gorm.Model
	ContainerId  string
	Status       WorkerStatus
	LastAssigned time.Time
}
