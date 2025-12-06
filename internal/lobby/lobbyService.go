package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dango/internal/events"

	"github.com/google/uuid"
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
	lobby := &Lobby{
		LobbyID:     lobbyID,
		HostID:      hostID,
		Name:        name,
		PlayerLimit: limit,
		IsPrivate:   isPrivate,
		CreatedAt:   time.Now(),
	}

	// Add host as first player
	   if err := lobby.AddPlayer(hostID); err != nil {
        return nil, err
    }

	if err := s.lobbyRepo.CreateLobby(ctx, lobby); err != nil {
		return nil, err
	}

	s.publishLobbyEvent(ctx, "global:lobbies", "lobby_created", lobby)


	return lobby, nil
}

func (s *LobbyService) JoinLobby(ctx context.Context, lobbyID string, userID int64) error {

	lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)
	if err != nil {
		return err
	}
	if lobby == nil {
        return ErrLobbyNotFound
    }

	if err := lobby.AddPlayer(userID); err != nil {
        return err
    }

	 if err := s.lobbyRepo.Save(ctx, lobby); err != nil {
        return err
    }

// Publish domain events
	for _, e := range lobby.Events() {
	s.publishLobbyEvent(ctx, "lobby:"+lobbyID, fmt.Sprintf("%T", e), e)
	}
	lobby.ClearEvents()

	return nil
}

func (s *LobbyService) LeaveLobby(ctx context.Context, lobbyID string, userID int64) error {


	    lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)
    if err != nil {
        return err
    }
    if lobby == nil {
        return ErrLobbyNotFound
    }

    lobby.RemovePlayer(userID)

    if lobby.IsEmpty() {
        if err := s.lobbyRepo.DeleteLobby(ctx, lobbyID); err != nil {
            return err
        }
	s.publishLobbyEvent(ctx, "global:lobbies", "lobby_deleted", nil)
        return nil
    }

    if err := s.lobbyRepo.Save(ctx, lobby); err != nil {
        return err
    }


	for _, e := range lobby.Events() {
		s.publishLobbyEvent(ctx, "lobby:"+lobbyID, fmt.Sprintf("%T", e), e)
	}
	lobby.ClearEvents() 

	return nil
}

func (s *LobbyService) SetPlayerReady(ctx context.Context, lobbyID string, userID int64, ready bool) error {

	 lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)
    if err != nil {
        return err
    }
    if lobby == nil {
        return ErrLobbyNotFound
    }

    lobby.SetReady(userID, ready)

    if err := s.lobbyRepo.Save(ctx, lobby); err != nil {
        return err
    }

	for _, e := range lobby.Events() {
		s.publishLobbyEvent(ctx, "lobby:"+lobbyID, fmt.Sprintf("%T", e), e)
	}
	lobby.ClearEvents() 

	return nil
}

func (s *LobbyService) GetLobbyPlayers(ctx context.Context, lobbyID string) ([]LobbyPlayer, error) {
	lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)
    if err != nil {
        return nil, err
    }
    if lobby == nil {
        return nil, ErrLobbyNotFound
    }
    return lobby.Players, nil
}

func (s *LobbyService) GetAllLobbies(ctx context.Context) ([]LobbyResponse, error) {
    lobbies, err := s.lobbyRepo.GetAllLobbies(ctx)
    if err != nil {
        return nil, err
    }

    response := make([]LobbyResponse, len(lobbies))
    for i, l := range lobbies {
        response[i] = LobbyResponse{
            LobbyID:     l.LobbyID,
            Name:        l.Name,
            HostID:      int(l.HostID),
            PlayerLimit: l.PlayerLimit,
            IsPrivate:   l.IsPrivate,
            CreatedAt:   l.CreatedAt.Format(time.RFC3339),
        }
    }

    return response, nil
}




func (s *LobbyService) publishLobbyEvent(ctx context.Context, channel string, eventType string, payload any) {
    data, err := json.Marshal(payload)
    if err != nil {
        fmt.Println("failed to marshal event payload:", err)
        return
    }

    e := events.Event{
        Type:    eventType,
        Payload: data,
    }

    if err := s.bus.Publish(ctx, channel, e); err != nil {
        fmt.Println("failed to publish event:", err)
    } else {
        fmt.Printf("Published event to channel %s: %s\n", channel, string(data))
    }
}