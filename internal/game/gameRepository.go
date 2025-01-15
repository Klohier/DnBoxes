package game

import (
	"context"
)

type GameRepository interface {
	FindAll(ctx context.Context) ([]Game, error)
	FindByID(ctx context.Context, id int) (*Game, error)
	Create(ctx context.Context, player1 int, player2 int, board_size int) (*Game, error)
	GetGrids(ctx context.Context, gameId int) ([]Box, error)
	UpdateGrid(ctx context.Context, gameId int, row int, col int, edge string) ([]Box, error)
	IsEdgeSelected(ctx context.Context, gameId int, row int, col int, edge string) (bool, error)
	SetBoxCompleted(ctx context.Context, gameId int, row int, col int, playerId int) error
	GetBoxByRowCol(ctx context.Context, gameId int, row int, col int) (*Box, error)
	UpdateTurn(ctx context.Context, gameId int, userID int) error
	SetWinner(ctx context.Context, gameId int, winnerId *int) error
}