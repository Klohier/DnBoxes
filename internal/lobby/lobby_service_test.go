package lobby_test

// import (
// 	"context"
// 	"dango/internal/events"
// 	"dango/internal/lobby"
// 	"encoding/json"
// 	"errors"
// 	"testing"
// )

// // MockEventBus captures published events for verification
// // MockEventBus captures published events for verification
// type MockEventBus struct {
//     events   []events.Event
//     handlers map[string][]func(events.Event)
// }

// func NewMockEventBus() *MockEventBus {
//     return &MockEventBus{
//         events:   []events.Event{},
//         handlers: make(map[string][]func(events.Event)),
//     }
// }

// func (m *MockEventBus) Publish(ctx context.Context, topic string, e events.Event) error {
//     m.events = append(m.events, e)
//     // Call any subscribed handlers
//     if hs, ok := m.handlers[topic]; ok {
//         for _, h := range hs {
//             h(e)
//         }
//     }
//     return nil
// }

// func (m *MockEventBus) Subscribe(ctx context.Context, topic string, handler func(events.Event)) error {
//     m.handlers[topic] = append(m.handlers[topic], handler)
//     return nil
// }

// // MockLobbyRepo simulates a simple in-memory lobby repository
// type MockLobbyRepo struct {
// 	lobbies map[string]*lobby.Lobby
// 	players map[string][]lobby.LobbyPlayer
// }

// func NewMockLobbyRepo() *MockLobbyRepo {
// 	return &MockLobbyRepo{
// 		lobbies: make(map[string]*lobby.Lobby),
// 		players: make(map[string][]lobby.LobbyPlayer),
// 	}
// }

// func (m *MockLobbyRepo) CreateLobby(ctx context.Context, l lobby.Lobby) error {
// 	m.lobbies[l.LobbyID] = &l
// 	m.players[l.LobbyID] = []lobby.LobbyPlayer{
// 		{UserID: l.HostID, IsReady: false},
// 	}
// 	return nil
// }

// func (m *MockLobbyRepo) GetLobby(ctx context.Context, lobbyID string) (*lobby.Lobby, error) {
// 	l, ok := m.lobbies[lobbyID]
// 	if !ok {
// 		return nil, nil
// 	}
// 	return l, nil
// }

// func (m *MockLobbyRepo) DeleteLobby(ctx context.Context, lobbyID string) error {
// 	delete(m.lobbies, lobbyID)
// 	delete(m.players, lobbyID)
// 	return nil
// }

// func (m *MockLobbyRepo) AddPlayer(ctx context.Context, lobbyID string, userID int64) error {
// 	players := m.players[lobbyID]
// 	// Check limit
// 	l := m.lobbies[lobbyID]
// 	if len(players) >= l.PlayerLimit {
// 		return lobby.ErrLobbyFull
// 	}
// 	m.players[lobbyID] = append(players, lobby.LobbyPlayer{UserID: userID, IsReady: false})
// 	return nil
// }

// func (m *MockLobbyRepo) RemovePlayer(ctx context.Context, lobbyID string, userID int64) error {
// 	players := m.players[lobbyID]
// 	newPlayers := []lobby.LobbyPlayer{}
// 	for _, p := range players {
// 		if p.UserID != userID {
// 			newPlayers = append(newPlayers, p)
// 		}
// 	}
// 	m.players[lobbyID] = newPlayers
// 	return nil
// }

// func (m *MockLobbyRepo) GetPlayers(ctx context.Context, lobbyID string) ([]lobby.LobbyPlayer, error) {
// 	return m.players[lobbyID], nil
// }

// func (m *MockLobbyRepo) SetPlayerReady(ctx context.Context, lobbyID string, userID int64, ready bool) error {
// 	for i := range m.players[lobbyID] {
// 		if m.players[lobbyID][i].UserID == userID {
// 			m.players[lobbyID][i].IsReady = ready
// 		}
// 	}
// 	return nil
// }

// func (m *MockLobbyRepo) GetAllLobbies(ctx context.Context) ([]lobby.Lobby, error) {
// 	lobbies := make([]lobby.Lobby, 0, len(m.lobbies))
// 	for _, l := range m.lobbies {
// 		lobbies = append(lobbies, *l)
// 	}
// 	return lobbies, nil
// }

// func TestLobbyServiceWithMocks(t *testing.T) {
// 	ctx := context.Background()
// 	repo := NewMockLobbyRepo()
// 	bus := &MockEventBus{}

// 	service := lobby.NewLobbyService(repo, bus)

// 	// Create lobby
// 	l, err := service.CreateLobby(ctx, 1, "TestLobby", 2, false)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	logLobbyPlayers(t, service, l.LobbyID)

// 	// Add player 2
// 	if err := service.JoinLobby(ctx, l.LobbyID, 2); err != nil {
// 		t.Fatal(err)
// 	}
// 	logLobbyPlayers(t, service, l.LobbyID)

// 	// Add player 3 (should hit full lobby)
// 	err = service.JoinLobby(ctx, l.LobbyID, 3)
// 	if err != nil {
// 		if errors.Is(err, lobby.ErrLobbyFull) {
// 			t.Logf("Expected error on joining full lobby: %v", err) // just log
// 		} else {
// 			t.Fatalf("Unexpected error: %v", err) // fail if any other error
// 		}
// 	}
// 	logLobbyPlayers(t, service, l.LobbyID)

// 	// Set player ready
// 	if err := service.SetPlayerReady(ctx, l.LobbyID, 2, true); err != nil {
// 		t.Fatal(err)
// 	}

// 	// Leave player 2
// 	if err := service.LeaveLobby(ctx, l.LobbyID, 2); err != nil {
// 		t.Fatal(err)
// 	}
// 	logLobbyPlayers(t, service, l.LobbyID)

// 	// Check players remaining (should be only host)
// 	players, err := service.GetLobbyPlayers(ctx, l.LobbyID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(players) != 1 || players[0].UserID != 1 {
// 		t.Fatalf("expected only host remaining, got %d", len(players))
// 	}

// 	// Leave host to delete lobby
// 	if err := service.LeaveLobby(ctx, l.LobbyID, 1); err != nil {
// 		t.Fatal(err)
// 	}

// 	logLobbyPlayers(t, service, l.LobbyID)

// // Check that players list is empty
// 	players, err = repo.GetPlayers(ctx, l.LobbyID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(players) != 0 {
// 		t.Fatalf("expected no players in deleted lobby, got %d", len(players))
// 	} else {
// 		t.Logf("All players successfully removed from lobby %s", l.LobbyID)
// 	}

// 	// Check that lobby is deleted
// 	deletedLobby, err := repo.GetLobby(ctx, l.LobbyID)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if deletedLobby != nil {
// 		t.Fatalf("expected lobby to be deleted, but it still exists")
// 	} else {
// 		t.Logf("Lobby %s successfully deleted", l.LobbyID)
// 	}

// 	// Verify events published
// 	for i, e := range bus.events {
// 		data, _ := json.Marshal(e)
// 		t.Logf("Event %d: %s", i+1, string(data))
// 	}

// 	if len(bus.events) != 5 { // player_joined x2, ready x1, leave x2 (including lobby_deleted)
// 		t.Fatalf("expected 5 events, got %d", len(bus.events))
// 	}
// }

// func logLobbyPlayers(t *testing.T, service *lobby.LobbyService, lobbyID string) {
//     players, err := service.GetLobbyPlayers(context.Background(), lobbyID)
//     if err != nil {
//         t.Fatalf("failed to get players: %v", err)
//     }
//     ids := make([]int64, 0, len(players))
//     for _, p := range players {
//         ids = append(ids, p.UserID)
//     }
//     t.Logf("Players in lobby %s: %v", lobbyID, ids)
// }
