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

type GameMove struct {
	MoveNumber int    `json:"move_number"`
	TurnOrder  int    `json:"turn_order"`
	Row        int    `json:"row"`
	Col        int    `json:"col"`
	Edge       string `json:"edge"`
}

type GameRepository interface {
	FindAll(ctx context.Context) ([]Game, error)
	FindByID(ctx context.Context, id int) (*Game, error)
	Create(ctx context.Context, players []Player, boardSize int) (*Game, error)
	Update(ctx context.Context, game *Game) error
	FindAllFromUser(ctx context.Context, userId int) ([]Game, error)
	FindUserGameHistory(ctx context.Context, userID int) ([]GameHistoryEntry, error)
	SaveMove(ctx context.Context, gameID int, move GameMove) error
	FindMovesByGameID(ctx context.Context, gameID int) ([]GameMove, error)
}
