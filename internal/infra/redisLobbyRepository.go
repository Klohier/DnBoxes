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

func lobbyKey(id string) string {
	return fmt.Sprintf("lobby:%s", id)
}

// -------------------- Lobby --------------------

func (r *RedisLobbyRepository) CreateLobby(ctx context.Context, l *lobby.Lobby) error {
	data, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, lobbyKey(l.LobbyID), data, r.ttl).Err()
}

func (r *RedisLobbyRepository) GetLobby(ctx context.Context, lobbyID string) (*lobby.Lobby, error) {
	val, err := r.client.Get(ctx, lobbyKey(lobbyID)).Result()
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

func (r *RedisLobbyRepository) Save(ctx context.Context, l *lobby.Lobby) error {
	data, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, lobbyKey(l.LobbyID), data, r.ttl).Err()
}

func (r *RedisLobbyRepository) DeleteLobby(ctx context.Context, lobbyID string) error {
	return r.client.Del(ctx, lobbyKey(lobbyID)).Err()
}

// -------------------- Get All --------------------

func (r *RedisLobbyRepository) GetAllLobbies(ctx context.Context) ([]*lobby.Lobby, error) {
	var lobbies []*lobby.Lobby

	iter := r.client.Scan(ctx, 0, "lobby:*", 0).Iterator()
	for iter.Next(ctx) {
		val, err := r.client.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}

		var l lobby.Lobby
		if err := json.Unmarshal([]byte(val), &l); err != nil {
			continue
		}

		lobbies = append(lobbies, &l)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return lobbies, nil
}
