package entities

import "time"

type Task struct {
	ID          int       `json:"id"`
	ProjectID   int       `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	AssigneeID *int      `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
}
func NewTask(
	projectID int,
	title string,
	description string,
	status string,
	assigneeID *int,
) Task {
	if status == "" {
		status = "TODO"
	}

	return Task{
		ProjectID:  projectID,
		Title:      title,
		Description: description,
		Status:     status,
		AssigneeID: assigneeID,
		CreatedAt:  time.Now(),
	}
}