package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"dango/internal/lobby"

	"github.com/redis/go-redis/v9"
)

type RedisLobbyRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisLobbyRepository(client *redis.Client) *RedisLobbyRepository {
	return &RedisLobbyRepository{
		client: client,
		ttl:    24 * time.Hour, 
	}
}

// -------------------- Lobby --------------------

func (r *RedisLobbyRepository) CreateLobby(ctx context.Context, l lobby.Lobby) error {
	key := fmt.Sprintf("lobby:%s", l.LobbyID)
	data, err := json.Marshal(l)
	if err != nil {
		return err
	}
	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return err
	}

	// Initialize players list
	playersKey := fmt.Sprintf("lobby:%s:players", l.LobbyID)
	initialPlayers := []lobby.LobbyPlayer{{UserID: l.HostID, IsReady: false}}
	playerData, _ := json.Marshal(initialPlayers)
	return r.client.Set(ctx, playersKey, playerData, r.ttl).Err()
}

func (r *RedisLobbyRepository) GetLobby(ctx context.Context, lobbyID string) (*lobby.Lobby, error) {
	key := fmt.Sprintf("lobby:%s", lobbyID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var l lobby.Lobby
	if err := json.Unmarshal([]byte(val), &l); err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *RedisLobbyRepository) DeleteLobby(ctx context.Context, lobbyID string) error {
	r.client.Del(ctx, fmt.Sprintf("lobby:%s", lobbyID))
	r.client.Del(ctx, fmt.Sprintf("lobby:%s:players", lobbyID))
	return nil
}

// -------------------- Players --------------------

func (r *RedisLobbyRepository) GetPlayers(ctx context.Context, lobbyID string) ([]lobby.LobbyPlayer, error) {
	key := fmt.Sprintf("lobby:%s:players", lobbyID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []lobby.LobbyPlayer{}, nil
		}
		return nil, err
	}
	var players []lobby.LobbyPlayer
	if err := json.Unmarshal([]byte(val), &players); err != nil {
		return nil, err
	}
	return players, nil
}

func (r *RedisLobbyRepository) AddPlayer(ctx context.Context, lobbyID string, userID int64) error {
	players, _ := r.GetPlayers(ctx, lobbyID)
	l, _ := r.GetLobby(ctx, lobbyID)
	if len(players) >= l.PlayerLimit {
		return lobby.ErrLobbyFull
	}
	players = append(players, lobby.LobbyPlayer{UserID: userID, IsReady: false})
	return r.savePlayers(ctx, lobbyID, players)
}

func (r *RedisLobbyRepository) RemovePlayer(ctx context.Context, lobbyID string, userID int64) error {
	players, _ := r.GetPlayers(ctx, lobbyID)
	newPlayers := []lobby.LobbyPlayer{}
	for _, p := range players {
		if p.UserID != userID {
			newPlayers = append(newPlayers, p)
		}
	}
	return r.savePlayers(ctx, lobbyID, newPlayers)
}

func (r *RedisLobbyRepository) SetPlayerReady(ctx context.Context, lobbyID string, userID int64, ready bool) error {
	players, _ := r.GetPlayers(ctx, lobbyID)
	for i := range players {
		if players[i].UserID == userID {
			players[i].IsReady = ready
			break
		}
	}
	return r.savePlayers(ctx, lobbyID, players)
}

// -------------------- Helpers --------------------

func (r *RedisLobbyRepository) savePlayers(ctx context.Context, lobbyID string, players []lobby.LobbyPlayer) error {
	key := fmt.Sprintf("lobby:%s:players", lobbyID)
	data, err := json.Marshal(players)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

// GetAllLobbies retrieves all lobbies stored in Redis.
func (r *RedisLobbyRepository) GetAllLobbies(ctx context.Context) ([]lobby.Lobby, error) {
	var lobbies []lobby.Lobby

	iter := r.client.Scan(ctx, 0, "lobby:*", 0).Iterator()
	seen := map[string]struct{}{}

	for iter.Next(ctx) {
		key := iter.Val()
		// skip keys for players list
		if keyHasSuffix(key, ":players") {
			continue
		}

		// avoid duplicates
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return nil, err
		}

		var l lobby.Lobby
		if err := json.Unmarshal([]byte(val), &l); err != nil {
			return nil, err
		}

		lobbies = append(lobbies, l)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return lobbies, nil
}

// keyHasSuffix checks if a key has the specified suffix.
func keyHasSuffix(key, suffix string) bool {
	if len(key) < len(suffix) {
		return false
	}
	return key[len(key)-len(suffix):] == suffix
}