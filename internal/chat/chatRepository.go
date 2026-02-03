package chat

import (
	"context"
	"time"
)

type ChatRepository interface {
	SaveGlobalMessage(ctx context.Context, userID int, message string, sentAt time.Time) error
	SaveGameMessage(ctx context.Context, userID int, message string, sentAt time.Time, gameID int) error
	GetGlobalMessages(ctx context.Context) ([]Message, error)
	GetGameMessage(ctx context.Context, gameID int) ([]Message, error)
}
