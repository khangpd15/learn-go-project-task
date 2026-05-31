package jobs

const NotificationQueueName = "jobs:notification"
const MaxNotificationRetries = 3

type NotificationJob struct {
	NotificationID int `json:"notification_id"`
	RetryCount int `json:"retry_count"`
}