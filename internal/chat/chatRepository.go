package chat

import (
	"context"
	"time"
)

type ChatRepository interface {
	SaveMessage(ctx context.Context, userID int, message string, time time.Time, sessionID int)  error
	GetAllMessageFromSession(ctx context.Context, sessionID int) ([]Message, error)
	GetGameMessage(ctx context.Context, gameID int) ([]Message, error)
}