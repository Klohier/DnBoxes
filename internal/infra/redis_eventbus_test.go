package infra_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dango/internal/events"
	"dango/internal/infra"

	"github.com/redis/go-redis/v9"
)

func TestRedisEventBus(t *testing.T) {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379" ,DB:   1,})
	bus := infra.NewRedisEventBus(rdb)

	received := make(chan events.Event, 1)

	err := bus.Subscribe(ctx, "test_topic", func(evt events.Event) {
t.Logf("handler received event: Type=%s Payload=%s", evt.Type, string(evt.Payload))
		received <- evt
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	testEvent := events.Event{
		Type: "USER_JOINED",
		Payload: json.RawMessage(`{"user_id": 124}`),
	}

	if err := bus.Publish(ctx, "test_topic", testEvent); err != nil {
		t.Fatalf("publish failed: %v", err)
	}

	select {
	case evt := <-received:
		t.Logf("test received event: Type=%s Payload=%s",  evt.Type, string(evt.Payload))
		if evt.Type != testEvent.Type {
			t.Fatalf("expected event name %s, got %s", testEvent.Type, evt.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("did not receive event")
	}
}

