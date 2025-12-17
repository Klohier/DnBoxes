package game

import (
	"context"
)

type GameRepository interface {
	FindAll(ctx context.Context) ([]Game, error)
	FindByID(ctx context.Context, id int) (*Game, error)
	Create(ctx context.Context, players []Player, boardSize int) (*Game, error)
	Update(ctx context.Context, game *Game) error
	FindAllFromUser(ctx context.Context, userId int) ([]Game, error)
}