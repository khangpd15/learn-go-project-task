package queue

import (
	"context"
)

type Queue interface {
	Enqueue(ctx context.Context, queueName string, payload []byte) error
	Dequeue(ctx context.Context, queueName string) ([]byte, error)
}