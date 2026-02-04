package game

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgGameRepository struct {
	db *pgxpool.Pool
}

func NewPgGameRepository(db *pgxpool.Pool) *PgGameRepository {
	return &PgGameRepository{db: db}
}

func (repo *PgGameRepository) FindAll(ctx context.Context) ([]Game, error) {
	query := `
		SELECT game_id, name, board_size, current_turn, winner_id, created_at, ended_at
		FROM games
		ORDER BY created_at DESC
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		err := rows.Scan(&g.GameID, &g.GameName, &g.BoardSize, &g.CurrentTurn,
			&g.WinnerID, &g.CreatedAt, &g.EndedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, g)
	}

	return games, rows.Err()
}

// FindByID loads a game by replaying its domain events.
// Falls back to the legacy grid-based snapshot for games created before event sourcing.
func (repo *PgGameRepository) FindByID(ctx context.Context, gameID int) (*Game, error) {
	domainEvents, err := repo.LoadEvents(ctx, gameID)
	if err != nil {
		return nil, err
	}

	if len(domainEvents) > 0 {
		return LoadFromEvents(domainEvents), nil
	}

	
	return nil, err
}


// Create inserts the game projection (games + game_details) and persists the
// initial domain events from the aggregate.
func (repo *PgGameRepository) Create(ctx context.Context, game *Game) error {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert game row (projection)
	var gameID int
	err = tx.QueryRow(ctx, `
		INSERT INTO games (board_size, current_turn)
		VALUES ($1, 0)
		RETURNING game_id
	`, game.BoardSize).Scan(&gameID)
	if err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}
	game.GameID = &gameID

	// Insert players (projection)
	for _, player := range game.Players {
		if player.UserID == nil {
			return fmt.Errorf("cannot create game with anonymous players in database")
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO game_details (game_id, user_id, turn_order, score)
			VALUES ($1, $2, $3, $4)
		`, gameID, *player.UserID, player.TurnOrder, player.Score)
		if err != nil {
			return fmt.Errorf("failed to insert player: %w", err)
		}
	}

	// Fix GameCreated event payload with the real DB-assigned gameID
	for i, evt := range game.UncommittedEvents() {
		if evt.Type == EventTypeGameCreated {
			if p, ok := evt.Payload.(GameCreatedPayload); ok {
				p.GameID = gameID
				game.UncommittedEvents()[i].Payload = p
			}
		}
	}

	// Persist initial domain events (event store)
	for _, evt := range game.UncommittedEvents() {
		payloadBytes, err := json.Marshal(evt.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO game_events (game_id, event_type, version, payload, occurred_at)
			VALUES ($1, $2, $3, $4, $5)
		`, gameID, evt.Type, evt.Version, payloadBytes, evt.OccurredAt)
		if err != nil {
			return fmt.Errorf("failed to save event: %w", err)
		}
	}
	game.ClearEvents()

	return tx.Commit(ctx)
}

// AppendEvents persists new domain events to the event store.
func (repo *PgGameRepository) AppendEvents(ctx context.Context, gameID int, domainEvents []events.DomainEvent) error {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, evt := range domainEvents {
		payloadBytes, err := json.Marshal(evt.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal event payload: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO game_events (game_id, event_type, version, payload, occurred_at)
			VALUES ($1, $2, $3, $4, $5)
		`, gameID, evt.Type, evt.Version, payloadBytes, evt.OccurredAt)
		if err != nil {
			return fmt.Errorf("failed to append event: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// LoadEvents loads all domain events for a game, ordered by version.
func (repo *PgGameRepository) LoadEvents(ctx context.Context, gameID int) ([]events.DomainEvent, error) {
	rows, err := repo.db.Query(ctx, `
		SELECT event_type, version, payload, occurred_at
		FROM game_events
		WHERE game_id = $1
		ORDER BY version
	`, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var domainEvents []events.DomainEvent
	for rows.Next() {
		var eventType string
		var version int
		var payloadBytes []byte
		var occurredAt time.Time

		if err := rows.Scan(&eventType, &version, &payloadBytes, &occurredAt); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		payload, err := deserializePayload(eventType, payloadBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event payload: %w", err)
		}

		domainEvents = append(domainEvents, events.DomainEvent{
			Type:        eventType,
			Version:     version,
			OccurredAt:  occurredAt,
			AggregateID: fmt.Sprintf("%d", gameID),
			Payload:     payload,
		})
	}

	return domainEvents, rows.Err()
}

// UpdateProjection updates the games and game_details tables to reflect
// the current aggregate state. Keeps query tables in sync with events.
func (repo *PgGameRepository) UpdateProjection(ctx context.Context, game *Game) error {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE games
		SET current_turn = $2, winner_id = $3, ended_at = $4
		WHERE game_id = $1
	`, game.GameID, game.CurrentTurn, game.WinnerID, game.EndedAt)
	if err != nil {
		return fmt.Errorf("failed to update game projection: %w", err)
	}

	for _, player := range game.Players {
		if player.UserID == nil {
			continue
		}
		_, err = tx.Exec(ctx, `
			UPDATE game_details
			SET score = $3
			WHERE game_id = $1 AND turn_order = $2
		`, game.GameID, player.TurnOrder, player.Score)
		if err != nil {
			return fmt.Errorf("failed to update player score projection: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (repo *PgGameRepository) FindUserGameHistory(ctx context.Context, userID int) ([]GameHistoryEntry, error) {
	query := `
		SELECT g.game_id, g.board_size, g.winner_id, g.created_at, g.ended_at,
		       gd.user_id, u.username, gd.score, gd.turn_order
		FROM games g
		INNER JOIN game_details gd ON g.game_id = gd.game_id
		INNER JOIN users u ON u.user_id = gd.user_id
		WHERE g.game_id IN (
			SELECT game_id FROM game_details WHERE user_id = $1
		)
		AND g.ended_at IS NOT NULL
		ORDER BY g.created_at DESC, gd.turn_order
	`

	rows, err := repo.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query game history: %w", err)
	}
	defer rows.Close()

	gamesMap := make(map[int]*GameHistoryEntry)
	var order []int

	for rows.Next() {
		var gameID, boardSize int
		var winnerID *int
		var createdAt time.Time
		var endedAt *time.Time
		var playerUserID int
		var username string
		var score, turnOrder int

		err := rows.Scan(&gameID, &boardSize, &winnerID, &createdAt, &endedAt,
			&playerUserID, &username, &score, &turnOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game history row: %w", err)
		}

		entry, exists := gamesMap[gameID]
		if !exists {
			entry = &GameHistoryEntry{
				GameID:    gameID,
				BoardSize: boardSize,
				WinnerID:  winnerID,
				CreatedAt: createdAt,
				EndedAt:   endedAt,
			}
			gamesMap[gameID] = entry
			order = append(order, gameID)
		}

		uid := playerUserID
		entry.Players = append(entry.Players, Player{
			UserID:    &uid,
			Username:  username,
			Score:     score,
			TurnOrder: turnOrder,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("game history rows error: %w", err)
	}

	var result []GameHistoryEntry
	for _, id := range order {
		result = append(result, *gamesMap[id])
	}

	return result, nil
}

func (repo *PgGameRepository) FindAllFromUser(ctx context.Context, userId int) ([]Game, error) {
	query := `
		SELECT g.game_id, g.name, g.board_size, g.current_turn,
		       g.winner_id, g.created_at, g.ended_at
		FROM games g
		INNER JOIN game_details gd ON g.game_id = gd.game_id
		WHERE gd.user_id = $1
		ORDER BY g.created_at DESC
	`

	rows, err := repo.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query user games: %w", err)
	}
	defer rows.Close()

	var games []Game
	for rows.Next() {
		var g Game
		err := rows.Scan(&g.GameID, &g.GameName, &g.BoardSize, &g.CurrentTurn,
			&g.WinnerID, &g.CreatedAt, &g.EndedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, g)
	}

	return games, rows.Err()
}

func (repo *PgGameRepository) FindUsernamesByIDs(ctx context.Context, userIDs []int) (map[int]string, error) {
	usernames := make(map[int]string, len(userIDs))
	if len(userIDs) == 0 {
		return usernames, nil
	}

	rows, err := repo.db.Query(ctx, `
		SELECT user_id, username FROM users WHERE user_id = ANY($1)
	`, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query usernames: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("failed to scan username: %w", err)
		}
		usernames[id] = name
	}
	return usernames, rows.Err()
}

// deserializePayload converts raw JSON into the correct typed payload struct.
func deserializePayload(eventType string, raw []byte) (any, error) {
	switch eventType {
	case EventTypeGameCreated:
		var p GameCreatedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		return p, nil
	case EventTypeMoveApplied:
		var p MoveAppliedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		return p, nil
	case EventTypeBoxCompleted:
		var p BoxCompletedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		return p, nil
	case EventTypeTurnPassed:
		var p TurnPassedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		return p, nil
	case EventTypeGameEnded:
		var p GameEndedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		return p, nil
	case EventTypeGameForfeited:
		var p GameForfeitedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, err
		}
		return p, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}
