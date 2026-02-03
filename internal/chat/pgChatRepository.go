package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgChatRepository struct {
	db *pgxpool.Pool
}

func NewPgChatRepository(db *pgxpool.Pool) *PgChatRepository {
	return &PgChatRepository{
		db: db,
	}
}

func (repo *PgChatRepository) SaveGlobalMessage(ctx context.Context, userID int, message string, sentAt time.Time) error {
	query := `INSERT INTO chats (user_id, message, sent_at, lobby_id) VALUES ($1, $2, $3, 'global')`
	_, err := repo.db.Exec(ctx, query, userID, message, sentAt)
	if err != nil {
		return fmt.Errorf("failed to save global message: %w", err)
	}
	return nil
}

func (repo *PgChatRepository) SaveGameMessage(ctx context.Context, userID int, message string, sentAt time.Time, gameID int) error {
	query := `INSERT INTO chats (user_id, message, sent_at, game_id) VALUES ($1, $2, $3, $4)`
	_, err := repo.db.Exec(ctx, query, userID, message, sentAt, gameID)
	if err != nil {
		return fmt.Errorf("failed to save game message: %w", err)
	}
	return nil
}

func (repo *PgChatRepository) GetGlobalMessages(ctx context.Context) ([]Message, error) {
	var messages []Message

	query := `
		SELECT c.message, c.sent_at, u.username, u.user_id
		FROM chats c
		LEFT JOIN users u ON c.user_id = u.user_id
		WHERE c.lobby_id = 'global'
		AND c.sent_at >= NOW() - INTERVAL '30 minutes'
		ORDER BY c.sent_at ASC
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.Message, &message.TimeStamp, &message.Username, &message.UserID); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (repo *PgChatRepository) GetGameMessage(ctx context.Context, gameID int) ([]Message, error) {
	var messages []Message

	query := `
		SELECT c.message, c.sent_at, u.username, u.user_id
		FROM chats c
		LEFT JOIN users u ON c.user_id = u.user_id
		WHERE c.game_id = $1
		AND c.sent_at >= NOW() - INTERVAL '30 minutes'
		ORDER BY c.sent_at ASC
	`
	rows, err := repo.db.Query(ctx, query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.Message, &message.TimeStamp, &message.Username, &message.UserID); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
