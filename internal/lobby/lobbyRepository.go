package lobby

import (
	"context"
	"dango/internal/user"
)

type LobbyRepository interface {
	FindAll(ctx context.Context) ([]Lobby, error)
	FindByID(ctx context.Context, LobbyId int) (*Lobby, error)
	Create(ctx context.Context, HostId int) (*Lobby, error)
	Delete(ctx context.Context, LobbyId int) error
	JoinLobby(ctx context.Context, lobbyID int, userID int) error
	LeaveLobby(ctx context.Context, lobbyID int, userID int) error
	GetParticipants(ctx context.Context, lobbyID int) ([]user.User, error)
}
