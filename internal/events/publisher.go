package events

import (
	"context"
	"encoding/json"
	"log"
)

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type NoopPublisher struct{}

func (p NoopPublisher) Publish(ctx context.Context, event Event) error {
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		log.Println("[EVENT PUBLISH ERROR]", err)
		return err
	}

	log.Println("[EVENT PUBLISHED]", string(data))
	log.Println("[PUBLISHER] send event to hub:", event.Type, "userID:", event.UserID)
	return nil
}