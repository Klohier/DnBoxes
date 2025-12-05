package lobby_test

import (
	"context"
	"fmt"
	"testing"

	"dango/internal/events"
	"dango/internal/infra"
	"dango/internal/lobby"

	"github.com/redis/go-redis/v9"
)


func setupRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379", 
		DB:   1,                 
	})
}

func TestLobbyServiceWithRedisEventBus(t *testing.T) {
	ctx := context.Background()
	redisClient := setupRedis()
	redisClient.FlushDB(ctx)
	defer redisClient.FlushDB(ctx)

	repo := infra.NewRedisLobbyRepository(redisClient)
	eventBus := infra.NewRedisEventBus(redisClient)

	service := lobby.NewLobbyService(repo, eventBus)

	// --- Create lobby ---
	lobbyObj, err := service.CreateLobby(ctx, 1, "TestLobby", 2, false)
	if err != nil {
		t.Fatal(err)
	}
	logRedisLobbyState(t, redisClient, lobbyObj.LobbyID)

	topic := fmt.Sprintf("lobby:%s", lobbyObj.LobbyID)
err = eventBus.Subscribe(ctx, topic, func(e events.Event) {
	t.Logf("Received event: %s %s", e.Type, string(e.Payload))

})

if err != nil {
	t.Fatalf("subscribe failed: %v", err)
}

	// --- Join player 2 ---
	if err := service.JoinLobby(ctx, lobbyObj.LobbyID, 2); err != nil {
		t.Fatal(err)
	}
	logRedisLobbyState(t, redisClient, lobbyObj.LobbyID)

	// --- Set player 2 ready ---
	if err := service.SetPlayerReady(ctx, lobbyObj.LobbyID, 2, true); err != nil {
		t.Fatal(err)
	}
	logRedisLobbyState(t, redisClient, lobbyObj.LobbyID)


	// --- Player 2 leaves (should trigger "player_left") ---
	if err := service.LeaveLobby(ctx, lobbyObj.LobbyID, 2); err != nil {
		t.Fatal(err)
	}
	logRedisLobbyState(t, redisClient, lobbyObj.LobbyID)

	// --- Host leaves (should trigger "lobby_deleted") ---
	if err := service.LeaveLobby(ctx, lobbyObj.LobbyID, 1); err != nil {
		t.Fatal(err)
	}
	logRedisLobbyState(t, redisClient, lobbyObj.LobbyID)

}


func logRedisLobbyState(t *testing.T, client *redis.Client, lobbyID string) {
	ctx := context.Background()

	// Get lobby
	lobbyKey := fmt.Sprintf("lobby:%s", lobbyID)
	lobbyData, err := client.Get(ctx, lobbyKey).Result()
	if err != nil && err != redis.Nil {
		t.Errorf("failed to get lobby: %v", err)
	} else if err == redis.Nil {
		t.Logf("Lobby %s does not exist in Redis", lobbyID)
	} else {
		t.Logf("Lobby data in Redis: %s", lobbyData)
	}
}
