package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgSessionRepository struct {
	db *pgxpool.Pool
}

func NewPgSessionRepository(db *pgxpool.Pool) *PgSessionRepository {
	return &PgSessionRepository{
		db: db,
	}
}

func (repo *PgSessionRepository) FindAll(ctx context.Context) ([]Session, error) {
	var sessions []Session

	query := `SELECT s.session_id, s.status, s.created_at, COUNT(su.user_id) as user_count
	FROM sessions s
	LEFT JOIN session_users su ON s.session_id = su.session_id
	GROUP BY s.session_id, s.status, s.created_at
	ORDER BY s.created_at DESC
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var session Session
		if err := rows.Scan(&session.SessionID, &session.Status, &session.CreatedAt, &session.UserCount); err != nil {
			return nil, err
		}

		sessions = append(sessions, session)

		if err := rows.Err(); err != nil {
			return nil, err
		}

	}
	return sessions, nil
}

func (repo *PgSessionRepository) FindByID(ctx context.Context, sessionID int) (*FullSession, error) {
	var session Session
	querySession := `SELECT session_id, status, created_at FROM sessions WHERE session_id = $1`

	err := repo.db.QueryRow(ctx, querySession, sessionID).Scan(
		&session.SessionID, &session.Status, &session.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	queryUsers := `
		SELECT u.user_id, u.username, su.connection_status, su.joined_at
		FROM session_users su
		JOIN users u ON su.user_id = u.user_id
		WHERE su.session_id = $1
	`

	rows, err := repo.db.Query(ctx, queryUsers, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var status string
	var users []SessionUser

	for rows.Next() {
		var sessionUser SessionUser

		if err := rows.Scan(&sessionUser.UserID, &sessionUser.Username, &status, &sessionUser.JoinedAt); err != nil {
			return nil, err
		}
		connectionStatus, err := parseConnectionStatus(status)
		if err != nil {
			return nil, err
		}
		sessionUser.ConnectionStatus = connectionStatus
		users = append(users, sessionUser)
	}

	return &FullSession{
		Session: session,
		Users:   users,
	}, nil
}

func (repo *PgSessionRepository) Create(ctx context.Context) (*Session, error) {
	query := `
		INSERT INTO sessions (status, created_at, session_type)
		VALUES ('active', NOW(), 'lobby')
		RETURNING session_id, status, created_at
	`

	var session Session

	err := repo.db.QueryRow(ctx, query).Scan(&session.SessionID, &session.Status, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

func (repo *PgSessionRepository) AddUserToSession(ctx context.Context, sessionID, userID int) error {
	query := `
		INSERT INTO session_users (session_id, user_id, connection_status, joined_at)
		VALUES ($1, $2, 'connected', NOW())
		ON CONFLICT (user_id) DO UPDATE 
		SET session_id = EXCLUDED.session_id, connection_status = 'connected', joined_at = NOW()
	`

	_, err := repo.db.Exec(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to add user %d to session %d: %w", userID, sessionID, err)
	}
	return nil
}

func (repo *PgSessionRepository) RemoveUserFromSession(ctx context.Context, sessionID, userID int) error {
	query := `
		DELETE FROM session_users 
		WHERE session_id = $1 AND user_id = $2
	`

	_, err := repo.db.Exec(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user %d from session %d: %w", userID, sessionID, err)
	}
	return nil
}

func (repo *PgSessionRepository) DeleteSession(ctx context.Context, sessionID int) error {
	query := `DELETE FROM sessions WHERE session_id = $1`

	result, err := repo.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session %d: %w", sessionID, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no session found with id %d", sessionID)
	}

	return nil
}

func (repo *PgSessionRepository) FindSessionByUserID(ctx context.Context, userID int) (*int, error) {
	var sessionID int

	query := `
		SELECT session_id
		FROM session_users
		WHERE user_id = $1
	`

	err := repo.db.QueryRow(ctx, query, userID).Scan(&sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying session for user %d: %w", userID, err)
	}

	return &sessionID, nil
}

func (repo *PgSessionRepository) SetUserConnectionStatus(ctx context.Context, sessionID, userID int, status string) error {
	query := `
		UPDATE session_users
		SET connection_status = $1
		WHERE session_id = $2 AND user_id = $3
	`
	_, err := repo.db.Exec(ctx, query, status, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to update connection status for user %d in session %d: %w", userID, sessionID, err)
	}
	return nil
}
