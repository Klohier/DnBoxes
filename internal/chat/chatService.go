package chat

import (
	"context"
	"errors"
)
type ChatService struct {
	chatRepo ChatRepository

}

func NewChatService(chatRepo ChatRepository) *ChatService{
	return &ChatService{
		chatRepo: chatRepo,
	}
}

func (s *ChatService) SendMessage(ctx context.Context, msg Message) error {
	err := s.chatRepo.SaveMessage(ctx, msg.UserID, msg.Message, msg.TimeStamp, msg.GameID)
	if err != nil {
		return errors.New("Failed to send message" + err.Error())
	}

	return nil
}
