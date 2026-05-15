package entities

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Assignee    string `json:"assignee"`
}

func NewTask(title, description, status, assignee string) Task {
	return Task{
		Title:       title,
		Description: description,
		Status:      status,
		Assignee:    assignee,
	}
}