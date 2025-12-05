package lobby

import "context"

type LobbyRepository interface {
	CreateLobby(ctx context.Context, lobby *Lobby) error
	DeleteLobby(ctx context.Context, lobbyID string) error
	GetLobby(ctx context.Context, lobbyID string) (*Lobby, error)
	GetAllLobbies(ctx context.Context) ([]*Lobby, error)
	Save(ctx context.Context, lobby *Lobby) error
}
