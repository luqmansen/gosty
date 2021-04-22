package models

type TaskProgressResponse struct {
	OriginVideo *Video  `json:"origin_video"`
	TaskList    []*Task `json:"task_list"`
}
