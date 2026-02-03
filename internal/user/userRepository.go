package user

import (
	"context"
)

type UserRepository interface {
	FindAll(ctx context.Context) ([]User, error)
	FindByID(ctx context.Context, id int) (*User, error)
	Create(ctx context.Context, username string, password string) (*User, error)
	CreateGuest(ctx context.Context, username string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	UserExists(ctx context.Context, username string) (bool, error)
	UpdateGameID(ctx context.Context, userID int, gameID *int) (*User, error)
}
