package lobby_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"dango/internal/lobby"

	"github.com/redis/go-redis/v9"
)

func newTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   15,
	})
}

func TestLobbyLifecycleWithEvents(t *testing.T) {
	rdb := newTestRedis()
	ctx := context.Background()

	// Clean database before test
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatal("failed to flush DB:", err)
	}
	t.Log("🧹 Redis test DB flushed")

	s := lobby.NewLobbyService(rdb)

	//  Create lobby
	l, err := s.CreateLobby(ctx, 1, "Hello", 4, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("🎮 Created Lobby: %+v\n", l)

	// Subscribe to lobby events after creation
	pubsub := rdb.Subscribe(ctx, "lobby:"+l.LobbyID+":events")
	defer pubsub.Close()

	// Use a small timeout to prevent hanging
	receiveEvent := func() map[string]interface{} {
		ch := pubsub.Channel()
		select {
		case msg := <-ch:
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
				t.Fatal("failed to unmarshal pubsub message:", err)
			}
			t.Logf(" Received event: %+v", payload)
			return payload
		case <-time.After(500 * time.Millisecond):
			t.Fatal("timeout waiting for pubsub event")
		}
		return nil
	}

	// Join players
	if err := s.JoinLobby(ctx, l.LobbyID, 2); err != nil {
		t.Fatal(err)
	}
	payload := receiveEvent()
	if payload["event"] != "player_joined" || int64(payload["userID"].(float64)) != 2 {
		t.Fatal("unexpected event for player 2 join")
	}

	if err := s.JoinLobby(ctx, l.LobbyID, 3); err != nil {
		t.Fatal(err)
	}
	payload = receiveEvent()
	if payload["event"] != "player_joined" || int64(payload["userID"].(float64)) != 3 {
		t.Fatal("unexpected event for player 3 join")
	}

	//Ready state
	if err := s.SetPlayerReady(ctx, l.LobbyID, 2, true); err != nil {
		t.Fatal(err)
	}
	payload = receiveEvent()
	if payload["event"] != "player_ready" || int64(payload["userID"].(float64)) != 2 {
		t.Fatal("unexpected event for player 2 ready")
	}

	// Validate lobby players
	players, err := s.GetLobbyPlayers(ctx, l.LobbyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(players))
	}

	// Leave lobby
	if err := s.LeaveLobby(ctx, l.LobbyID, 2); err != nil {
		t.Fatal(err)
	}
	payload = receiveEvent()
	if payload["event"] != "player_left" || int64(payload["userID"].(float64)) != 2 {
		t.Fatal("unexpected event for player 2 leave")
	}

	// Delete when empty
	if err := s.LeaveLobby(ctx, l.LobbyID, 1); err != nil {
		t.Fatal(err)
	}
	receiveEvent() // event for player 1 leaving
	if err := s.LeaveLobby(ctx, l.LobbyID, 3); err != nil {
		t.Fatal(err)
	}
	receiveEvent() // event for player 3 leaving

	// Give Redis a moment to propagate deletion
	time.Sleep(50 * time.Millisecond)
	_, err = rdb.Get(ctx, "lobby:"+l.LobbyID).Result()
	if err == nil || err != redis.Nil {
		t.Fatal("lobby should have been deleted")
	}

	t.Log("🏁 Lobby lifecycle and events verified successfully")
}

func TestLobbyPlayerLimit(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   15,
	})
	ctx := context.Background()
	defer rdb.FlushDB(ctx) // Clean DB after test

	s := lobby.NewLobbyService(rdb)

	// Create a lobby with limit 2
	l, err := s.CreateLobby(ctx, 1, "LimitTest", 2, false)
	if err != nil {
		t.Fatal(err)
	}

	// First player is host (already added)
	// Add one more player (should succeed)
	if err := s.JoinLobby(ctx, l.LobbyID, 2); err != nil {
		t.Fatal("expected join to succeed:", err)
	}

	// Add third player (should fail)
	err = s.JoinLobby(ctx, l.LobbyID, 3)
	if err != lobby.ErrLobbyFull {
		t.Fatalf("expected ErrLobbyFull, got: %v", err)
	}

	// Validate lobby has exactly 2 players
	players, err := s.GetLobbyPlayers(ctx, l.LobbyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(players) != 2 {
		t.Fatalf("expected 2 players in lobby, got %d", len(players))
	}

	t.Log("Player limit enforced correctly")
}