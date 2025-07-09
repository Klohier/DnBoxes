package game

import (
	"context"
)

type GameRepository interface {
	FindAll(ctx context.Context) ([]Game, error)
	FindByID(ctx context.Context, id int) (*Game, error)
	Create(ctx context.Context, playerIds []int, board_size int, sessionId int) (*Game, error)
	GetGrids(ctx context.Context, gameId int) ([]Box, error)
	UpdateGrid(ctx context.Context, gameId int, row int, col int, edge string) ([]Box, error)
	IsEdgeSelected(ctx context.Context, gameId int, row int, col int, edge string) (bool, error)
	SetBoxCompleted(ctx context.Context, gameId int, row int, col int, playerId int) error
	GetBoxByRowCol(ctx context.Context, gameId int, row int, col int) (*Box, error)
	UpdateTurn(ctx context.Context, gameId int, userID int) error
	SetWinner(ctx context.Context, gameId int, winnerId *int) error
	FindAllFromUser(ctx context.Context, userId int) ([]Game, error)
	IncrementPlayerScore(ctx context.Context, gameId int, userId int) error
	GetPlayerScores(ctx context.Context, gameId int) (map[int]int, error)
}
