package events

import (
	"encoding/json"
	"time"
)

// type Message struct {
// 	UserID    int       `json:"userID"`
// 	Username  string    `json:"username"`
// 	SessionID int       `json:"session_id"`
// 	Message   string    `json:"message"`
// 	TimeStamp time.Time `json:"timestamp"`
// }

// Integration event structure
type Event struct {
	// Type is the message type sent
	Topic string `json:"topic"`
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

type DomainEvent struct {
	Entity
	Type string `json:"type"`
	OccurredAt time.Time `json:"occurred_at"`
	AggregateID string 
	Payload any `json:"payload"`


}

type IDer interface {
	ID() string
}

type DDDEvent interface {
	IDer
	EventType() string
	Payload() json.RawMessage
	Metadata() any
	OccurredAt() time.Time
}


func (DomainEvent *DomainEvent) EventType() string {
	return DomainEvent.Type
}

func (DomainEvent *DomainEvent) EventPayload() any {
	return DomainEvent.Payload
}

func (DomainEvent *DomainEvent) EventOccurredAt() time.Time {
	return DomainEvent.OccurredAt
}
