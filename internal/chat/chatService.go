package chat

import (
	"context"
	"dango/internal/events"
	"log/slog"

	// "dango/internal/websocket"

	"errors"

	// "log/slog"

	"github.com/redis/go-redis/v9"
)

type ChatService struct {
	chatRepo ChatRepository
	redisClient *redis.Client
}

func NewChatService(chatRepo ChatRepository,  RedisClient *redis.Client) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		redisClient: RedisClient,
	}
}

func (s *ChatService) SaveMessage(ctx context.Context, msg events.Message, channel string) error {
	if err := s.chatRepo.SaveMessage(ctx, msg.UserID, msg.Message, msg.TimeStamp, msg.SessionID); err != nil {
		 slog.Error("failed to send message:", "error", err)
	}

	if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, msg); err != nil {
		 slog.Error("failed to publish message:", "error", err)
	}
	
	return nil
}

func (s *ChatService) GetAllGameMessage(ctx context.Context, gameId int) ([]Message, error) {
	msg, err := s.chatRepo.GetGameMessage(ctx, gameId)
	if err != nil {
		return nil, errors.New("failed to get message from game" + err.Error())
	}
	return msg, err
}

func (s *ChatService) GetAllMessageFromSession(ctx context.Context, sessionID int) ([]Message, error) {
	msg, err := s.chatRepo.GetAllMessageFromSession(ctx, sessionID)
	if err != nil {
		return nil, errors.New("failed to get message from game" + err.Error())
	}
	return msg, err
}
