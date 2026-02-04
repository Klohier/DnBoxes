package chat

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/redis/go-redis/v9"
)

const maxMessageLength = 500

type ChatService struct {
	chatRepo    ChatRepository
	redisClient *redis.Client
}

func NewChatService(chatRepo ChatRepository, RedisClient *redis.Client) *ChatService {
	return &ChatService{
		chatRepo:    chatRepo,
		redisClient: RedisClient,
	}
}

type chatSavePayload struct {
	UserID    int    `json:"userID"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// StartPersistenceWorker subscribes to the "chat:save" EventBus topic
// and persists incoming chat messages to the database.
func (s *ChatService) StartPersistenceWorker(eventBus events.EventBus) {
	err := eventBus.Subscribe(context.Background(), "chat:save", func(e events.Event) {
		var payload chatSavePayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			slog.Error("chat persistence: failed to unmarshal payload", "error", err)
			return
		}

		// Defense-in-depth: sanitize even though the WebSocket layer already does this
		payload.Message = strings.TrimSpace(payload.Message)
		payload.Message = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return -1
			}
			return r
		}, payload.Message)
		if payload.Message == "" || len([]rune(payload.Message)) > maxMessageLength {
			return
		}
		payload.Message = html.EscapeString(payload.Message)

		sentAt, err := time.Parse(time.RFC3339, payload.Timestamp)
		if err != nil {
			sentAt = time.Now()
		}

		ctx := context.Background()
		topic := e.Topic

		if topic == "chat:global" {
			if err := s.chatRepo.SaveGlobalMessage(ctx, payload.UserID, payload.Message, sentAt); err != nil {
				slog.Error("chat persistence: failed to save global message", "error", err)
			}
		} else if strings.HasPrefix(topic, "game:") {
			gameIDStr := strings.TrimPrefix(topic, "game:")
			gameID, err := strconv.Atoi(gameIDStr)
			if err != nil {
				slog.Error("chat persistence: invalid game ID in topic", "topic", topic, "error", err)
				return
			}
			if err := s.chatRepo.SaveGameMessage(ctx, payload.UserID, payload.Message, sentAt, gameID); err != nil {
				slog.Error("chat persistence: failed to save game message", "error", err, "gameID", gameID)
			}
		}
	})

	if err != nil {
		slog.Error("chat persistence: failed to subscribe to chat:save", "error", err)
	} else {
		slog.Info("Chat persistence worker started, subscribed to chat:save")
	}
}

func (s *ChatService) SaveGlobalMessage(ctx context.Context, userID int, message string, sentAt time.Time) error {
	if err := s.chatRepo.SaveGlobalMessage(ctx, userID, message, sentAt); err != nil {
		return fmt.Errorf("failed to save global message: %w", err)
	}
	return nil
}

func (s *ChatService) SaveGameMessage(ctx context.Context, userID int, message string, sentAt time.Time, gameID int) error {
	if err := s.chatRepo.SaveGameMessage(ctx, userID, message, sentAt, gameID); err != nil {
		return fmt.Errorf("failed to save game message: %w", err)
	}
	return nil
}

func (s *ChatService) GetGlobalMessages(ctx context.Context) ([]Message, error) {
	msg, err := s.chatRepo.GetGlobalMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global messages: %w", err)
	}
	return msg, nil
}

func (s *ChatService) GetAllGameMessage(ctx context.Context, gameId int) ([]Message, error) {
	msg, err := s.chatRepo.GetGameMessage(ctx, gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to get game messages: %w", err)
	}
	return msg, nil
}
