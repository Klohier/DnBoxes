package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"dango/internal/events"

	"github.com/google/uuid"
)

var (
	ErrLobbyFull = errors.New("lobby is full")
)

type LobbyService struct {
	lobbyRepo LobbyRepository
	bus 	  events.EventBus
}

func NewLobbyService(lobbyRepo LobbyRepository, bus events.EventBus) *LobbyService {
	return &LobbyService{lobbyRepo: lobbyRepo, bus: bus}
}

func (s *LobbyService) CreateLobby(ctx context.Context, hostID int64, name string, limit int, isPrivate bool) (*Lobby, error) {
	lobbyID := uuid.New().String()
	lobby := Lobby{
		LobbyID:     lobbyID,
		HostID:      hostID,
		Name:        name,
		PlayerLimit: limit,
		IsPrivate:   isPrivate,
		CreatedAt:   time.Now(),
	}

	if err := s.lobbyRepo.CreateLobby(ctx, lobby); err != nil {
		return nil, err
	}

	s.publishLobbyEvent(ctx, lobbyID, "lobby_created", nil)
	
	// Publish host joined event

	s.publishLobbyEvent(ctx, lobbyID, "player_joined", &hostID)
	return &lobby, nil
}

func (s *LobbyService) JoinLobby(ctx context.Context, lobbyID string, userID int64) error {
	players, err := s.lobbyRepo.GetPlayers(ctx, lobbyID)
	if err != nil {
		return err
	}

	lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)
	if err != nil {
		return err
	}

	if len(players) >= lobby.PlayerLimit {
		return ErrLobbyFull
	}

	if err := s.lobbyRepo.AddPlayer(ctx, lobbyID, userID); err != nil {
		return err
	}

	s.publishLobbyEvent(ctx, lobbyID, "player_joined", &userID)
	return nil
}

func (s *LobbyService) LeaveLobby(ctx context.Context, lobbyID string, userID int64) error {
	if err := s.lobbyRepo.RemovePlayer(ctx, lobbyID, userID); err != nil {
		return err
	}

	players, err := s.lobbyRepo.GetPlayers(ctx, lobbyID)
	if err != nil {
		return err
	}

	if len(players) == 0 {
		if err := s.lobbyRepo.DeleteLobby(ctx, lobbyID); err != nil {
			return err
		}
		s.publishLobbyEvent(ctx, lobbyID, "lobby_deleted", nil)
	} else {
		s.publishLobbyEvent(ctx, lobbyID, "player_left", &userID)
	}

	
	return nil
}

func (s *LobbyService) SetPlayerReady(ctx context.Context, lobbyID string, userID int64, ready bool) error {
	if err := s.lobbyRepo.SetPlayerReady(ctx, lobbyID, userID, ready); err != nil {
		return err
	}

	s.publishLobbyEvent(ctx, lobbyID, "player_ready", &userID)
	return nil
}

func (s *LobbyService) GetLobbyPlayers(ctx context.Context, lobbyID string) ([]LobbyPlayer, error) {
	return s.lobbyRepo.GetPlayers(ctx, lobbyID)
}

// ---------------------- Event Publishing ----------------------

func (s *LobbyService) publishLobbyEvent(ctx context.Context, lobbyID string, eventType string, userID *int64) {
	payload := LobbyEventPayload{
		LobbyID: lobbyID,
		Event:   eventType,
		Time:    time.Now().Unix(),
	}

	if userID != nil {
		payload.UserID = userID
	}
	// Marshal payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("failed to marshal event payload:", err)
		return
	}

	// Publish the event
	e := events.Event{
		Type:    "lobby_event",
		Payload: data,
	}

	if err := s.bus.Publish(ctx, "lobby:"+lobbyID+":events", e); err != nil {
		fmt.Println("failed to publish event:", err)
	} else {
    fmt.Printf("Published event to channel lobby:%s:events: %s\n", lobbyID, string(data))
}
}