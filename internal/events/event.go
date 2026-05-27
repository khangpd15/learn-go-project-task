package events



type Event struct {
	Type   string `json:"type"`
	UserID int   `json:"user_id"`
	ProjectID int `json:"project_id"`
	TaskID int `json:"task_id"`
	Data   interface{} `json:"data"`
}

const (
	EventTaskCreated = "task.created"
)

