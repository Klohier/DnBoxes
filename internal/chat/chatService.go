package chat

import (
	"context"
	"errors"
)

type ChatService struct {
	chatRepo ChatRepository
}

func NewChatService(chatRepo ChatRepository) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
	}
}

func (s *ChatService) SaveMessage(ctx context.Context, msg Message) error {
	err := s.chatRepo.SaveMessage(ctx, msg.UserID, msg.Message, msg.TimeStamp, msg.SessionID)
	if err != nil {
		return errors.New("Failed to send message" + err.Error())
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
