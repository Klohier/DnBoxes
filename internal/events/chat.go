package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Message struct {
	UserID    int       `json:"userID"`
	Username  string    `json:"username"`
	SessionID int       `json:"session_id"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`
}

type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

func PublishEvent(ctx context.Context, rdb *redis.Client, topic, eventType string, payload any) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal event payload:", "error", err)
	}

	event := Event{
		Type:    eventType,
		Payload: rawPayload,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		slog.Error("failed to marshal event:", "error", err)
	}

	slog.Info("Publishing event", "topic", topic, "type", eventType, "message", string(eventBytes))
	return rdb.Publish(ctx, topic, eventBytes).Err()
}
