package infra_test

import (
	"context"
	"testing"
	"time"

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

func TestRedisLobbyRepository_Unit(t *testing.T) {
	client := setupRedis()
	repo := infra.NewRedisLobbyRepository(client)
	ctx := context.Background()

	// Clean up DB before and after
	client.FlushDB(ctx)
	defer client.FlushDB(ctx)

	// --- Create Lobby ---
	l := lobby.Lobby{
		LobbyID:     "test123",
		HostID:      1,
		Name:        "TestLobby",
		PlayerLimit: 2,
		IsPrivate:   false,
		CreatedAt:   time.Now(),
	}

	t.Logf("Creating lobby: %+v", l)
	if err := repo.CreateLobby(ctx, l); err != nil {
		t.Fatal(err)
	}

	// --- Get Players (should include host) ---
	players, err := repo.GetPlayers(ctx, l.LobbyID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Players after creation: %+v", players)
	if len(players) != 1 || players[0].UserID != 1 {
		t.Fatalf("expected host in players list, got %v", players)
	}

	// --- Add Second Player ---
	t.Logf("Adding player 2")
	if err := repo.AddPlayer(ctx, l.LobbyID, 2); err != nil {
		t.Fatal(err)
	}
	players, _ = repo.GetPlayers(ctx, l.LobbyID)
	t.Logf("Players after adding player 2: %+v", players)
	if len(players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(players))
	}

	// --- Set Player Ready ---
	t.Logf("Setting player 2 ready")
	if err := repo.SetPlayerReady(ctx, l.LobbyID, 2, true); err != nil {
		t.Fatal(err)
	}
	players, _ = repo.GetPlayers(ctx, l.LobbyID)
	t.Logf("Players after setting ready: %+v", players)
	foundReady := false
	for _, p := range players {
		if p.UserID == 2 && p.IsReady {
			foundReady = true
		}
	}
	if !foundReady {
		t.Fatalf("player 2 should be ready")
	}

	// --- Remove Player ---
	t.Logf("Removing player 2")
	if err := repo.RemovePlayer(ctx, l.LobbyID, 2); err != nil {
		t.Fatal(err)
	}
	players, _ = repo.GetPlayers(ctx, l.LobbyID)
	t.Logf("Players after removing player 2: %+v", players)
	if len(players) != 1 || players[0].UserID != 1 {
		t.Fatalf("expected only host remaining, got %v", players)
	}

	// --- Remove Host ---
	t.Logf("Removing host")
	if err := repo.RemovePlayer(ctx, l.LobbyID, 1); err != nil {
		t.Fatal(err)
	}
	players, _ = repo.GetPlayers(ctx, l.LobbyID)
	t.Logf("Players after removing host: %+v", players)
	if len(players) != 0 {
		t.Fatalf("expected no players remaining, got %d", len(players))
	}

	// --- Check Lobby Exists ---
	lobbyAfterRemoval, _ := repo.GetLobby(ctx, l.LobbyID)
	t.Logf("Lobby after removing all players: %+v", lobbyAfterRemoval)
	// Note: The repository does not auto-delete the lobby
	if lobbyAfterRemoval == nil {
		t.Logf("Lobby record still exists? nil")
	}
}
