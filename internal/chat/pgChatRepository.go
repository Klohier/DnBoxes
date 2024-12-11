package chat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PgChatRepository struct {
    db *pgxpool.Pool
}

func NewPgChatRepository(db *pgxpool.Pool) *PgChatRepository{
	return &PgChatRepository{
		db: db,
	}
}

func (repo *PgChatRepository) SaveMessage(ctx context.Context, userID int, message string, time time.Time, gameID *int) (error){
	
	// var msg Message
	query :=`INSERT INTO chats (user_id, message, timestamp, game_id) VALUES ($1, $2, $3, $4)`
	_, err := repo.db.Exec(ctx, query, userID, message, time, gameID)
    if err != nil {
		fmt.Printf("Executing query: %s\n", query)
fmt.Printf("Values: userID=%d, message=%s, timestamp=%s, gameID=%v\n", userID, message, time, gameID)
        return  errors.New("Failed to Save Message" + err.Error())
    }

    return nil

}


func (repo *PgChatRepository) GetAllMessage(ctx context.Context) ([]Message, error){
	var messages []Message
	


	query := `SELECT message, timestamp FROM chats WHERE game_id IS NULL`


	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.Message, &message.TimeStamp); err != nil {
			return nil, err
		}
		messages = append(messages, message)
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	return messages, nil
}

func (repo *PgChatRepository) GetGameMessage(ctx context.Context, gameID *int) ([]Message, error){
	var messages []Message
	
   
	query := `SELECT message, timestamp FROM chats WHERE game_id = $1`

	
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.Message, &message.TimeStamp); err != nil {
			return nil, err
		}
		messages = append(messages, message)
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	return messages, nil
}

