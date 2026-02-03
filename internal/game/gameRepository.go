package game

import (
	"context"
	"dango/internal/events"
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
	Create(ctx context.Context, game *Game) error
	FindAllFromUser(ctx context.Context, userId int) ([]Game, error)
	FindUserGameHistory(ctx context.Context, userID int) ([]GameHistoryEntry, error)

	// Event store: append new domain events for a game
	AppendEvents(ctx context.Context, gameID int, domainEvents []events.DomainEvent) error
	// Event store: load all domain events for a game (replay)
	LoadEvents(ctx context.Context, gameID int) ([]events.DomainEvent, error)
	// Projection: update games/game_details tables from current aggregate state
	UpdateProjection(ctx context.Context, game *Game) error
}
