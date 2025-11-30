package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrLobbyFull = errors.New("lobby is full")
)

type LobbyService struct {
	rdb *redis.Client
}

func NewLobbyService(rdb *redis.Client) *LobbyService {
	return &LobbyService{rdb: rdb}
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

	data, err := json.Marshal(lobby)
	if err != nil {
		return nil, err
	}

	// Store lobby JSON
	if err := s.rdb.Set(ctx, "lobby:"+lobbyID, data, 30*time.Minute).Err(); err != nil {
		return nil, err
	}

	// Add lobby to set
	if err := s.rdb.SAdd(ctx, "lobbies", lobbyID).Err(); err != nil {
		return nil, err
	}

	// Add host to players set & state hash atomically
	pipe := s.rdb.TxPipeline()
	pipe.SAdd(ctx, "lobby:"+lobbyID+":players", hostID)
	pipe.HSet(ctx, "lobby:"+lobbyID+":state", hostID, "not_ready")
	pipe.Expire(ctx, "lobby:"+lobbyID+":players", 30*time.Minute)
	pipe.Expire(ctx, "lobby:"+lobbyID+":state", 30*time.Minute)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	// Publish event
	s.publishLobbyEvent(ctx, lobbyID, "player_joined", hostID)

	return &lobby, nil
}

// JoinLobby adds a user, respecting the player limit
func (s *LobbyService) JoinLobby(ctx context.Context, lobbyID string, userID int64) error {
	keyPlayers := "lobby:" + lobbyID + ":players"
	keyState := "lobby:" + lobbyID + ":state"

	for {
		// WATCH the players set
		err := s.rdb.Watch(ctx, func(tx *redis.Tx) error {
			count, err := tx.SCard(ctx, keyPlayers).Result()
			if err != nil {
				return err
			}

			lobbyData, err := tx.Get(ctx, "lobby:"+lobbyID).Result()
			if err != nil {
				return err
			}

			var lobby Lobby
			if err := json.Unmarshal([]byte(lobbyData), &lobby); err != nil {
				return err
			}

			if count >= int64(lobby.PlayerLimit) {
				return ErrLobbyFull
			}

			// Add player atomically
			pipe := tx.TxPipeline()
			pipe.SAdd(ctx, keyPlayers, userID)
			pipe.HSet(ctx, keyState, userID, "not_ready")
			pipe.Expire(ctx, keyPlayers, 30*time.Minute)
			pipe.Expire(ctx, keyState, 30*time.Minute)
			_, err = pipe.Exec(ctx)
			return err
		}, keyPlayers)

		if err == redis.TxFailedErr {
			// Retry on race condition
			continue
		}
		if err != nil {
			return err
		}

		// Publish event
		s.publishLobbyEvent(ctx, lobbyID, "player_joined", userID)
		return nil
	}
}

func (s *LobbyService) LeaveLobby(ctx context.Context, lobbyID string, userID int64) error {
	keyPlayers := "lobby:" + lobbyID + ":players"
	keyState := "lobby:" + lobbyID + ":state"

	pipe := s.rdb.TxPipeline()
	pipe.SRem(ctx, keyPlayers, userID)
	pipe.HDel(ctx, keyState, strconv.FormatInt(userID, 10))
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// Check if lobby empty and delete
	count, err := s.rdb.SCard(ctx, keyPlayers).Result()
	if err != nil {
		return err
	}
	if count == 0 {
		pipe := s.rdb.TxPipeline()
		pipe.Del(ctx, "lobby:"+lobbyID, keyPlayers, keyState)
		pipe.SRem(ctx, "lobbies", lobbyID)
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}

	// Publish event
	s.publishLobbyEvent(ctx, lobbyID, "player_left", userID)

	return nil
}

func (s *LobbyService) GetLobbyPlayers(ctx context.Context, lobbyID string) ([]LobbyPlayer, error) {
	keyPlayers := "lobby:" + lobbyID + ":players"
	keyState := "lobby:" + lobbyID + ":state"

	userIDs, err := s.rdb.SMembers(ctx, keyPlayers).Result()
	if err != nil {
		return nil, err
	}

	states, err := s.rdb.HGetAll(ctx, keyState).Result()
	if err != nil {
		return nil, err
	}

	players := []LobbyPlayer{}
	for _, id := range userIDs {
		uid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		players = append(players, LobbyPlayer{
			UserID:  uid,
			IsReady: states[id] == "ready",
		})
	}

	return players, nil
}

func (s *LobbyService) SetPlayerReady(ctx context.Context, lobbyID string, userID int64, ready bool) error {
	state := "not_ready"
	if ready {
		state = "ready"
	}
	keyState := "lobby:" + lobbyID + ":state"
	err := s.rdb.HSet(ctx, keyState, strconv.FormatInt(userID, 10), state).Err()
	if err == nil {
		s.publishLobbyEvent(ctx, lobbyID, "player_ready", userID)
	}
	return err
}

// publishLobbyEvent publishes events to Redis for WebSocket fan-out
func (s *LobbyService) publishLobbyEvent(ctx context.Context, lobbyID string, event string, userID int64) {
	payload := map[string]interface{}{
		"event":  event,
		"userID": userID,
		"time":   time.Now().Unix(),
	}
	data, _ := json.Marshal(payload)
	s.rdb.Publish(ctx, "lobby:"+lobbyID+":events", data)
}
