package chat

import (
	"context"
	"time"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, userID int, message string, time time.Time, gameID *int)  error
	GetAllMessage(ctx context.Context) ([]Message, error)
	GetGameMessage(ctx context.Context, gameID int) ([]Message, error)
}