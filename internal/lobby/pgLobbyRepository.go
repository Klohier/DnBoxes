package lobby

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgLobbyRepository struct {
	db *pgxpool.Pool
}

func NewPgLobbyRepository(db *pgxpool.Pool) *PgLobbyRepository {
	return &PgLobbyRepository{
		db: db,
	}
}

func (repo *PgLobbyRepository) FindAll(ctx context.Context) ([]Lobby, error) {
	var lobbies []Lobby

	query := `SELECT lobby_id, name, host_user_id, created_at, session_id, is_private FROM lobbies`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lobby Lobby
		if err := rows.Scan(&lobby.LobbyID, &lobby.Name, &lobby.HostID, &lobby.CreatedAt, &lobby.SessionID, &lobby.IsPrivate); err != nil {
			return nil, err
		}
		lobbies = append(lobbies, lobby)
		if err := rows.Err(); err != nil {
			return nil, err
		}

	}
	return lobbies, nil
}

func (repo *PgLobbyRepository) FindByID(ctx context.Context, lobbyId int) (*Lobby, error) {
	var lobby Lobby
	query := `SELECT name, host_user_id, created_at, session_id, is_private FROM lobbies where lobby_id = $1`
	err := repo.db.QueryRow(ctx, query, lobbyId).Scan(&lobby.Name, &lobby.HostID, &lobby.CreatedAt, &lobby.SessionID, &lobby.IsPrivate)
	if err != nil {
		return nil, fmt.Errorf("failed to find game %d : %w", lobbyId, err)
	}
	return &lobby, nil
}
