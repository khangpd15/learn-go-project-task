package task

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Assignee    *int `json:"assignee"`
}