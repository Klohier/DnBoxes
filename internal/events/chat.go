package events

import (
	"encoding/json"
	"time"
)

// Event is an integration event used for real-time broadcasting (e.g. WebSocket).
type Event struct {
	Topic   string          `json:"topic"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// DomainEvent represents an event that occurred within a domain aggregate.
// Domain events are the source of truth in an event-sourced system.
type DomainEvent struct {
	Entity
	Type        string    `json:"type"`
	OccurredAt  time.Time `json:"occurred_at"`
	AggregateID string    `json:"aggregate_id"`
	Version     int       `json:"version"`
	Payload     any       `json:"payload"`
}

func (e *DomainEvent) EventType() string {
	return e.Type
}

func (e *DomainEvent) EventPayload() any {
	return e.Payload
}

func (e *DomainEvent) EventOccurredAt() time.Time {
	return e.OccurredAt
}
