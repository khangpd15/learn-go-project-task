package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"task_api/internal/jobs"
	"task_api/internal/queue"
	"time"
)

type NotificationWorker struct {
	queue queue.Queue
}

func NewNotificationWorker(q queue.Queue) *NotificationWorker {
	return &NotificationWorker{queue: q}
}

func (w *NotificationWorker) sendNotification(job jobs.NotificationJob) error {
	if job.AssigneeID == 4 {
		return errors.New("fake notification error")
	}

	log.Printf("[WORKER] sending notification to user=%d message=%s\n", job.AssigneeID, job.Message)
	return nil
}

func (w *NotificationWorker) handleJob(ctx context.Context, job jobs.NotificationJob) {
	err := w.sendNotification(job)
	if err == nil {
		log.Printf("[WORKER] notification sent successfully: task=%d assignee=%d\n", job.TaskID, job.AssigneeID)
		return
	}
	job.RetryCount++
	if job.RetryCount <= jobs.MaxNotificationRetries {
		log.Printf("[WORKER] notification failed, retry=%d/%d, task=%d, error=%v\n",
			job.RetryCount,
			jobs.MaxNotificationRetries,
			job.TaskID,
			err,
		)
		delay := retryDelay(job.RetryCount)

		log.Printf("[WORKER] scheduling retry in %s for task=%d\n", delay, job.TaskID)

		time.Sleep(delay)
		payload, marshalErr := json.Marshal(job)
		if marshalErr != nil {
			log.Printf("Error marshaling notification job for retry: %v", marshalErr)
			return
		}
		if enqueueErr := w.queue.Enqueue(ctx, jobs.NotificationQueueName, payload); enqueueErr != nil {
			log.Printf("Error enqueuing notification job for retry: %v", enqueueErr)
		}
		return
	}
	log.Printf("[WORKER] notification failed after max retries: task=%d assignee=%d\n", job.TaskID, job.AssigneeID)

}
func (w *NotificationWorker) Start(ctx context.Context) {
	log.Println("Notification Worker started")
	for {
		payload, err := w.queue.Dequeue(ctx, jobs.NotificationQueueName)
		if err != nil {
			log.Printf("Error dequeuing notification job: %v", err)
			continue
		}
		var jobData jobs.NotificationJob
		if err := json.Unmarshal([]byte(payload), &jobData); err != nil {
			log.Printf("Error unmarshaling notification job: %v", err)
			continue
		}
		w.handleJob(ctx, jobData)
	}
}
func retryDelay(retryCount int) time.Duration {
	switch retryCount {
	case 1:
		return 10 * time.Second
	case 2:
		return 1 * time.Minute
	default:
		return 5 * time.Minute
	}
}
