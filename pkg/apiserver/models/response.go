package models

import "time"

type TaskProgressResponse struct {
	OriginVideo   *Video        `json:"origin_video"`
	TotalDuration time.Duration `json:"total_duration"`
	TaskList      []*Task       `json:"task_list"`
}
