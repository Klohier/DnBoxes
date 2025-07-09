package session

import (
	"context"
)

type SessionRepository interface {
	FindAll(ctx context.Context) ([]Session, error)
	FindByID(ctx context.Context, SessionId int) (*FullSession, error)
	Create(ctx context.Context) (*Session, error)
	AddUserToSession(ctx context.Context, sessionID int, userID int) error
	RemoveUserFromSession(ctx context.Context, sessionID int, userID int) error
	DeleteSession(ctx context.Context, sessionID int) error
	FindSessionByUserID(ctx context.Context, userID int) (*int, error)
	SetUserConnectionStatus(ctx context.Context, sessionID, userID int, status string) error
}
