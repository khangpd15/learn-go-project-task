package jobs

const NotificationQueueName = "jobs:notification"
const MaxNotificationRetries = 3

type NotificationJob struct {
	TaskID int `json:"task_id"`
	AssigneeID int `json:"assignee_id"`
	Message string `json:"message"`
	RetryCount int `json:"retry_count"`
}