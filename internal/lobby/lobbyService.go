package lobby

import (
	"context"
	"encoding/json"
	"time"

	"dango/internal/events"

	"github.com/google/uuid"
)


type LobbyService struct {
	lobbyRepo LobbyRepository
	bus 	  events.EventBus
}

type PlayerJoinedPayload struct {
    LobbyID string       `json:"lobby_id"`
    Player  LobbyPlayer `json:"player"`
}

type LobbyUpdatedPayload struct {
	LobbyID string        `json:"lobby_id"`
	Players []LobbyPlayer `json:"players"`
	Status  string        `json:"status"` 
}

func NewLobbyService(lobbyRepo LobbyRepository, bus events.EventBus) *LobbyService {
	return &LobbyService{lobbyRepo: lobbyRepo, bus: bus}
}

func (s *LobbyService) CreateLobby(ctx context.Context, hostID int64, username string, name string, limit int, isPrivate bool, boardSize int) (*Lobby, error) {
	lobbyID := uuid.New().String()
	lobby := &Lobby{
		LobbyID:     lobbyID,
		HostID:      hostID,
		Name:        name,
        BoardSize:   boardSize,
		PlayerLimit: limit,
		IsPrivate:   isPrivate,
		CreatedAt:   time.Now(),
        Players:     []LobbyPlayer{},
	}

	// Add host as first player
	   if err := lobby.AddPlayer(hostID, username); err != nil {
        return nil, err
    }

	if err := s.lobbyRepo.CreateLobby(ctx, lobby); err != nil {
		return nil, err
	}


payloadBytes, err := json.Marshal(lobby)
if err != nil {
    return nil, err
}

// When websockets  readsd from "global:lobbies", they will get this event from event bus and emit to topic subscribers
    s.bus.Publish(ctx, "global:lobbies", events.Event{Topic: "global:lobbies", Type: "lobby_created", Payload: payloadBytes})

	return lobby, nil
}

func (s *LobbyService) JoinLobby(ctx context.Context, lobbyID string, userID int64, username string) error {

	lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)
	if err != nil {
		return err
	}
	if lobby == nil {
        return ErrLobbyNotFound
    }

	if err := lobby.AddPlayer(userID, username); err != nil {
        return err
    }

	 if err := s.lobbyRepo.Save(ctx, lobby); err != nil {
        return err
    }

    payload := LobbyUpdatedPayload{
		LobbyID: lobby.LobbyID,
		Players: lobby.Players,
		Status:  "waiting",
	}

    payloadBytes, err := json.Marshal(payload)
if err != nil {
    return err
}

    s.bus.Publish(ctx, "lobby:"+lobbyID, events.Event{Topic:   "lobby:" + lobbyID, Type: "lobby_updated", Payload: payloadBytes})

    s.bus.Publish(ctx, "global:lobbies", events.Event{
    Topic:   "lobby:" + lobbyID,
    Type:    "lobby_updated",
    Payload: payloadBytes,
})

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

	if err := lobby.RemovePlayer(userID); err != nil {
		return err
	}

	if lobby.IsEmpty() {
		if err := s.lobbyRepo.DeleteLobby(ctx, lobbyID); err != nil {
			return err
		}

		deletedPayload, _ := json.Marshal(map[string]string{"lobby_id": lobbyID})
		s.bus.Publish(ctx, "global:lobbies", events.Event{Topic: "lobby:" + lobbyID, Type: "lobby_deleted", Payload: deletedPayload})

		return nil
	}

	if err := s.lobbyRepo.Save(ctx, lobby); err != nil {
		return err
	}

	payload := LobbyUpdatedPayload{
		LobbyID: lobby.LobbyID,
		Players: lobby.Players,
		Status:  "waiting",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	s.bus.Publish(ctx, "lobby:"+lobbyID, events.Event{Topic: "lobby:" + lobbyID, Type: "lobby_updated", Payload: payloadBytes})
	s.bus.Publish(ctx, "global:lobbies", events.Event{Topic: "lobby:" + lobbyID, Type: "lobby_updated", Payload: payloadBytes})

	return nil
}

func (s *LobbyService) DeleteLobby(ctx context.Context, lobbyID string) error {
	lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)
	if err != nil {
		return err
	}
	if lobby == nil {
		return ErrLobbyNotFound
	}

	if err := s.lobbyRepo.DeleteLobby(ctx, lobbyID); err != nil {
		return err
	}

	deletedPayload, _ := json.Marshal(map[string]string{"lobby_id": lobbyID})
	s.bus.Publish(ctx, "global:lobbies", events.Event{Topic: "lobby:" + lobbyID, Type: "lobby_deleted", Payload: deletedPayload})

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

    payload := LobbyUpdatedPayload{
		LobbyID: lobby.LobbyID,
		Players: lobby.Players,
		Status:  "waiting",
	}

    payloadBytes, err := json.Marshal(payload)
if err != nil {
    return err
}

    s.bus.Publish(ctx, "lobby:"+lobbyID, events.Event{ Topic:   "lobby:" + lobbyID, Type: "lobby_updated", Payload: payloadBytes})

    s.bus.Publish(ctx, "global:lobbies", events.Event{
    Topic:   "lobby:" + lobbyID,
    Type:    "lobby_updated",
    Payload: payloadBytes,
})

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

func (s *LobbyService) GetAllLobbies(ctx context.Context) ([]*Lobby, error) {
    lobbies, err := s.lobbyRepo.GetAllLobbies(ctx)
    if err != nil {
        return nil, err
    }

    return lobbies, nil
}

func (s *LobbyService) GetLobby(ctx context.Context, lobbyID string) (*Lobby, error) {
    lobby, err := s.lobbyRepo.GetLobby(ctx, lobbyID)    
    if err != nil {
        return nil, err
    }
    return lobby, nil
}