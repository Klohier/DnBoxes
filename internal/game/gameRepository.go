package game

import (
	"context"
	"time"
)

type GameHistoryEntry struct {
	GameID    int        `json:"game_id"`
	BoardSize int        `json:"board_size"`
	WinnerID  *int       `json:"winner_id"`
	CreatedAt time.Time  `json:"created_at"`
	EndedAt   *time.Time `json:"ended_at"`
	Players   []Player   `json:"players"`
}

type GameRepository interface {
	FindAll(ctx context.Context) ([]Game, error)
	FindByID(ctx context.Context, id int) (*Game, error)
	Create(ctx context.Context, players []Player, boardSize int) (*Game, error)
	Update(ctx context.Context, game *Game) error
	FindAllFromUser(ctx context.Context, userId int) ([]Game, error)
	FindUserGameHistory(ctx context.Context, userID int) ([]GameHistoryEntry, error)
}
