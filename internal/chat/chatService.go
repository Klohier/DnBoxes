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

func (s *ChatService) SaveMessage(ctx context.Context, msg Message) error {
	err := s.chatRepo.SaveMessage(ctx, msg.UserID, msg.Message, msg.TimeStamp, msg.GameID)
	if err != nil {
		return errors.New("Failed to send message" + err.Error())
	}

	return nil
}

func (s *ChatService) GetAllGameMessage(ctx context.Context, gameId int) ([]Message , error) {
	msg, err := s.chatRepo.GetGameMessage(ctx, gameId)
	if err != nil {
		return nil, errors.New("failed to get message from game" + err.Error())
	}
	return msg, err
}
 

func (s *ChatService) GetAllMessage(ctx context.Context) ([]Message , error) {
	msg, err := s.chatRepo.GetAllMessage(ctx)
	if err != nil {
		return nil, errors.New("failed to get message from game" + err.Error())
	}
	return msg, err
}

