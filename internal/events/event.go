package events

type Event struct {
	Type      string                 `json:"type"`
	UserID    int                    `json:"user_id,omitempty"`
	UserIDs   []int                  `json:"user_ids,omitempty"`
	ProjectID int                    `json:"project_id,omitempty"`
	TaskID    int                    `json:"task_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

const (
	EventTaskCreated         = "task.created"
	EventUpdateStatus        = "task.updateStatus"
	EventTaskAssigned        = "task.assigned"
	EventTaskUnassigned      = "task.unassigned"
)
