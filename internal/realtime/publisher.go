package realtime

import (
	"context"
	"task_api/internal/events"
)

type Publisher struct {
	hub *Hub
}

func NewPublisher(hub *Hub) *Publisher {
	return &Publisher{hub: hub}
}

func (p *Publisher) Publish(ctx context.Context, event events.Event) error {
	p.hub.Broadcast(event)
	return nil
}