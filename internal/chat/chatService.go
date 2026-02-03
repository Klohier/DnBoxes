package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

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
