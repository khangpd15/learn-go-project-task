package entities

import "time"

type Notification struct {
	ID        int    `json:"id"`
	TaskID    int    `json:"task_id,omitempty"`
	ProjectID int    `json:"project_id,omitempty"`
	SenderID  int    `json:"sender_id"`
	ReceiverID *int    `json:"receiver_id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	IsRead    bool   `json:"is_read"`
	ReadAt    *int64 `json:"read_at,omitempty"`
	Message   string `json:"message"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}