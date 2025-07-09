package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PgUserRepository struct {
	db *pgxpool.Pool
}

func NewPgUserRepository(db *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{
		db: db,
	}
}

func (repo *PgUserRepository) FindAll(ctx context.Context) ([]User, error) {
	var users []User

	query := `SELECT user_id, username FROM users`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserID, &user.Username); err != nil {
			return nil, err
		}
		users = append(users, user)
		if err := rows.Err(); err != nil {
			return nil, err
		}

	}
	return users, nil
}

func (repo *PgUserRepository) FindByID(ctx context.Context, id int) (*User, error) {
	var user User
	query := `SELECT user_id, username FROM users WHERE user_id = $1`
	err := repo.db.QueryRow(ctx, query, id).Scan(&user.UserID, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user %d : %w", id, err)
	}
	return &user, nil
}

func (repo *PgUserRepository) Create(ctx context.Context, username string, password string) (*User, error) {

	var user User
	query := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING user_id, username`
	err := repo.db.QueryRow(ctx, query, username, password).Scan(&user.UserID, &user.Username)
	if err != nil {
		return nil, errors.New("Failed to Create User" + err.Error())
	}

	return &user, nil

}

func (repo *PgUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	query := `SELECT username, password, user_id FROM users WHERE username = $1`
	err := repo.db.QueryRow(ctx, query, username).Scan(&user.Username, &user.Password, &user.UserID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, errors.New("failed to check username uniqueness")
	}

	return &user, nil

}

func (repo *PgUserRepository) UserExists(ctx context.Context, username string) (bool, error) {
	user, err := repo.FindByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

func (repo *PgUserRepository) UpdateGameID(ctx context.Context, userID int, gameID *int) (*User, error) {
	query := `UPDATE users SET game_id = $1 WHERE user_id = $2`
	_, err := repo.db.Exec(ctx, query, gameID, userID)
	if err != nil {
		return nil, errors.New("failed to update user")
	}

	user, err := repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to get updated user after updating gameID: " + err.Error())
	}

	return user, nil

}
