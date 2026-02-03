package game

import (
	"context"
	"fmt"
	"math/rand"
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

func (repo *PgGameRepository) FindByID(ctx context.Context, gameID int) (*Game, error) {
	batch := &pgx.Batch{}
	
	// Query 1: Game info
	batch.Queue(`
		SELECT game_id, name, board_size, current_turn, winner_id, created_at, ended_at
		FROM games WHERE game_id = $1
	`, gameID)
	
	// Query 2: Players with scores
	batch.Queue(`
		SELECT gd.user_id, u.username, gd.turn_order, gd.score
		FROM game_details gd
		JOIN users u ON u.user_id = gd.user_id
		WHERE gd.game_id = $1
		ORDER BY gd.turn_order
	`, gameID)
	
	// Query 3: Boxes
	batch.Queue(`
		SELECT grid_row, grid_col, top_edge, right_edge, bottom_edge, left_edge, completed_by
		FROM grids
		WHERE game_id = $1
		ORDER BY grid_row, grid_col
	`, gameID)
	
	br := repo.db.SendBatch(ctx, batch)
	defer br.Close()
	
	var game Game
	
	// Scan game
	err := br.QueryRow().Scan(
		&game.GameID, &game.GameName, &game.BoardSize, &game.CurrentTurn,
		&game.WinnerID, &game.CreatedAt, &game.EndedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("game %d not found", gameID)
		}
		return nil, fmt.Errorf("failed to get game: %w", err)
	}
	
	// Scan players with scores
	playerRows, err := br.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	
	for playerRows.Next() {
		var p Player
		err := playerRows.Scan(&p.UserID, &p.Username, &p.TurnOrder, &p.Score)
		if err != nil {
			playerRows.Close()
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		// Assume registered users (not anonymous) since we're joining with users table
		p.IsAnonymous = false
		game.Players = append(game.Players, p)
	}
	playerRows.Close()
	
	// Scan boxes and build grid
	game.Grid = make([][]Box, game.BoardSize)
	for i := range game.Grid {
		game.Grid[i] = make([]Box, game.BoardSize)
	}
	
	boxRows, err := br.Query()
	if err != nil {
		return nil, fmt.Errorf("failed to query boxes: %w", err)
	}
	defer boxRows.Close()
	
	for boxRows.Next() {
		var row, col int
		var topEdge, rightEdge, bottomEdge, leftEdge bool
		var completedBy *int
		
		err := boxRows.Scan(&row, &col, &topEdge, &rightEdge, &bottomEdge, &leftEdge, &completedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan box: %w", err)
		}
		
		// Convert user_id to turn_order
		var ownerTurn *int
		if completedBy != nil {
			// Find the turn_order for this user_id
			for _, p := range game.Players {
				if p.UserID != nil && *p.UserID == *completedBy {
					ownerTurn = &p.TurnOrder
					break
				}
			}
		}
		
		game.Grid[row][col] = Box{
			Row:        row,
			Col:        col,
			TopEdge:    topEdge,
			RightEdge:  rightEdge,
			BottomEdge: bottomEdge,
			LeftEdge:   leftEdge,
			OwnerTurn:  ownerTurn,
		}
	}
	
	return &game, boxRows.Err()
}

func (repo *PgGameRepository) Create(ctx context.Context, players []Player, boardSize int) (*Game, error) {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	
	// Shuffle players for random turn order
	rand.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	
	// Assign turn orders
	for i := range players {
		players[i].TurnOrder = i
		players[i].Score = 0
	}
	
	// Insert game
	var gameID int
	err = tx.QueryRow(ctx, `
		INSERT INTO games (board_size, current_turn)
		VALUES ($1, 0)
		RETURNING game_id
	`, boardSize).Scan(&gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	
	// Insert players into game_details
	for _, player := range players {
		if player.UserID == nil {
			return nil, fmt.Errorf("cannot create game with anonymous players in database")
		}
		
		_, err = tx.Exec(ctx, `
			INSERT INTO game_details (game_id, user_id, turn_order, score)
			VALUES ($1, $2, $3, $4)
		`, gameID, *player.UserID, player.TurnOrder, player.Score)
		if err != nil {
			return nil, fmt.Errorf("failed to insert player: %w", err)
		}
	}
	
	// Insert empty boxes into grids
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			_, err = tx.Exec(ctx, `
				INSERT INTO grids (game_id, grid_row, grid_col)
				VALUES ($1, $2, $3)
			`, gameID, row, col)
			if err != nil {
				return nil, fmt.Errorf("failed to create box: %w", err)
			}
		}
	}
	
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	// Return the created game
	return repo.FindByID(ctx, gameID)
}

func (repo *PgGameRepository) Update(ctx context.Context, game *Game) error {
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	
	// Update game metadata
	_, err = tx.Exec(ctx, `
		UPDATE games 
		SET current_turn = $2, winner_id = $3, ended_at = $4
		WHERE game_id = $1
	`, game.GameID, game.CurrentTurn, game.WinnerID, game.EndedAt)
	if err != nil {
		return fmt.Errorf("failed to update game: %w", err)
	}
	
	// Update player scores in game_details
	for _, player := range game.Players {
		if player.UserID == nil {
			continue // Skip anonymous players
		}
		
		_, err = tx.Exec(ctx, `
			UPDATE game_details
			SET score = $3
			WHERE game_id = $1 AND turn_order = $2
		`, game.GameID, player.TurnOrder, player.Score)
		if err != nil {
			return fmt.Errorf("failed to update player score: %w", err)
		}
	}
	
	// Update all boxes in grids
	for i := 0; i < len(game.Grid); i++ {
		for j := 0; j < len(game.Grid[i]); j++ {
			box := &game.Grid[i][j]
			
			// Convert turn_order back to user_id for completed_by
			var completedBy *int
			if box.OwnerTurn != nil {
				// Find the user_id for this turn_order
				for _, p := range game.Players {
					if p.TurnOrder == *box.OwnerTurn && p.UserID != nil {
						completedBy = p.UserID
						break
					}
				}
			}
			
			_, err = tx.Exec(ctx, `
				UPDATE grids
				SET top_edge = $3, right_edge = $4, bottom_edge = $5, 
				    left_edge = $6, completed = $7, completed_by = $8
				WHERE game_id = $1 AND grid_row = $2 AND grid_col = $9
			`, game.GameID, box.Row, box.TopEdge, box.RightEdge, box.BottomEdge,
				box.LeftEdge, completedBy != nil, completedBy, box.Col)
			if err != nil {
				return fmt.Errorf("failed to update box at (%d,%d): %w", i, j, err)
			}
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