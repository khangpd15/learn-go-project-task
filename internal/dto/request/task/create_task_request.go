package task

type CreateTaskRequest struct {
	ProjectID   int    `json:"project_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}