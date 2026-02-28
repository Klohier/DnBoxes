package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"dango/internal/events"

	"github.com/redis/go-redis/v9"
)

type RedisEventBus struct {
	client *redis.Client
}

// NewRedisEventBus creates a Redis-backed EventBus
func NewRedisEventBus(client *redis.Client) *RedisEventBus {
	return &RedisEventBus{client: client}
}

// Publish sends an event to a topic
func (r *RedisEventBus) Publish(ctx context.Context, topic string, event events.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if err := r.client.Publish(ctx, topic, data).Err(); err != nil {
		return fmt.Errorf("redis publish error: %w", err)
	}
	return nil
}

// Subscribe listens to a topic and calls the handler for each incoming message
func (r *RedisEventBus) Subscribe(ctx context.Context, topic string, handler func(events.Event)) error {
	pubsub := r.client.Subscribe(ctx, topic)

	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("redis subscribe error: %w", err)
	}

	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			var evt events.Event
			if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
				fmt.Printf("failed to unmarshal event: %v\n", err)
				continue
			}

			// Convert Payload to string for logging/handling
			var payloadStr string
			if err := json.Unmarshal(evt.Payload, &payloadStr); err != nil {
				// If payload is not a string, fallback to raw JSON
				payloadStr = string(evt.Payload)
			}

			fmt.Printf("handler received event: Type=%s, Payload=%s\n", evt.Type, payloadStr)
			handler(evt) // still pass full event with original Payload
		}
	}()

	return nil
}

func (r *RedisEventBus) Close() error {
	return nil
}
