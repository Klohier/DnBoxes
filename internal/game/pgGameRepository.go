package game

import (
	"context"
	"errors"
	"fmt"

	// "log"
	"math/rand"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgGameRepository struct {
	db *pgxpool.Pool
}

func NewPgGameRepository(db *pgxpool.Pool) *PgGameRepository {
	return &PgGameRepository{
		db: db,
	}
}

func (repo *PgGameRepository) FindAll(ctx context.Context) ([]Game, error) {
	var games []Game

	query := `SELECT game_id FROM games`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var game Game
		if err := rows.Scan(&game.GameId); err != nil {
			return nil, err
		}
		games = append(games, game)
		if err := rows.Err(); err != nil {
			return nil, err
		}

	}
	return games, nil
}

func (repo *PgGameRepository) FindByID(ctx context.Context, gameID int) (*Game, error) {
	var game Game
	query := `SELECT game_id, name, board_size, winner_id, current_turn, created_at, session_id
			  FROM games
			  WHERE game_id = $1`
	err := repo.db.QueryRow(ctx, query, gameID).Scan(&game.GameId, &game.GameName, &game.BoardSize, &game.WinnerId, &game.CurrentTurn, &game.CreatedAt, &game.SessionId)
	if err != nil {
		return nil, fmt.Errorf("failed to find game %d : %w", gameID, err)
	}

	//get players from game
	players, err := repo.GetPlayersForGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch players for game %d: %w", gameID, err)
	}
	game.Player = players

	return &game, nil
}

func (repo *PgGameRepository) GetPlayersForGame(ctx context.Context, gameId int) ([]Player, error) {
	query := `SELECT gd.user_id, u.username, gd.turn_order, gd.score
			  FROM game_details gd
			  JOIN users u ON u.user_id = gd.user_id
			  WHERE gd.game_id = $1
			  ORDER BY gd.turn_order`

	rows, err := repo.db.Query(ctx, query, gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to query players for game %d: %w", gameId, err)
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player

		err := rows.Scan(
			&player.UserID,
			&player.Username,
			&player.TurnOrder,
			&player.Score,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player row: %w", err)
		}

		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating player rows: %w", err)
	}

	return players, nil
}

func (repo *PgGameRepository) Create(ctx context.Context, playerIds []int, boardSize int, sessionID int) (*Game, error) {
	// Start a new transaction
	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errors.New("Failed to begin transaction: " + err.Error())
	}

	//Rollsback if failure
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var game Game

	query := `INSERT INTO games (board_size, session_id) 
			   VALUES ($1, $2) 
			   RETURNING game_id, board_size, session_id`
	err = tx.QueryRow(ctx, query, boardSize, sessionID).Scan(&game.GameId, &game.BoardSize, &game.SessionId)
	if err != nil {
		return nil, errors.New("Failed to create game: " + err.Error())
	}

	//randomize turn order
	rand.Shuffle(len(playerIds), func(i, j int) {
		playerIds[i], playerIds[j] = playerIds[j], playerIds[i]
	})

	//Put players in to tables
	for turnOrder, playerId := range playerIds {
		playerQuery := `INSERT INTO game_details (game_id, user_id, turn_order, score) VALUES ($1, $2, $3, 0)`
		_, err = tx.Exec(ctx, playerQuery, game.GameId, playerId, turnOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to insert player %d into game_details: %w", playerId, err)
		}
	}

	// Set current_turn to first player's user_id
	setTurnQuery := `UPDATE games SET current_turn = $1 WHERE game_id = $2`
	_, err = tx.Exec(ctx, setTurnQuery, 0, game.GameId)
	if err != nil {
		return nil, fmt.Errorf("failed to set current turn: %w", err)
	}

	currentTurn := 0
	game.CurrentTurn = &currentTurn

	// Create the boxes associated with the game
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			boxQuery := `INSERT INTO grids (game_id, grid_row, grid_col, completed_by) 
						  VALUES ($1, $2, $3, NULL)`
			_, err = tx.Exec(ctx, boxQuery, game.GameId, i, j)
			if err != nil {
				return nil, errors.New("Failed to create box: " + err.Error())
			}
		}
	}

	// Commit the transaction if all operations are successful
	err = tx.Commit(ctx)
	if err != nil {
		return nil, errors.New("Failed to commit transaction: " + err.Error())
	}

	return &game, nil
}

func (repo *PgGameRepository) GetGrids(ctx context.Context, gameId int) ([]Box, error) {
	var boxes []Box
	query := `
	SELECT box_id, game_id, top_edge, right_edge, left_edge, bottom_edge, grid_row, grid_col, completed, completed_by 
	FROM grids WHERE game_id =$1
	ORDER BY box_id ASC
	`
	rows, err := repo.db.Query(ctx, query, gameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var box Box
		if err := rows.Scan(&box.BoxId, &box.GameId, &box.TopEdge, &box.RightEdge, &box.LeftEdge, &box.BottomEdge, &box.Row, &box.Col, &box.Completed, &box.Completed_by); err != nil {
			return nil, err
		}
		boxes = append(boxes, box)
		if err := rows.Err(); err != nil {
			return nil, err
		}

	}
	return boxes, nil
}

func (repo *PgGameRepository) UpdateGrid(ctx context.Context, gameId int, row int, col int, edge string) ([]Box, error) {

	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, errors.New("failed to begin transaction: " + err.Error())
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	game, err := repo.FindByID(ctx, gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve game %d: %w", gameId, err)
	}

	query := fmt.Sprintf(`UPDATE grids SET %s = TRUE WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3`, edge)
	//TODO think about getting results
	_, err = repo.db.Exec(ctx, query, gameId, row, col)
	if err != nil {
		return nil, fmt.Errorf("failed to update edge for box at row %d, col %d: %w", row, col, err)
	}

	//Update Adjacent Edges

	//TODO Can this be made better?
	switch edge {
	case "top_edge":
		if row > 0 {
			_, err = repo.db.Exec(ctx, `UPDATE grids SET bottom_edge = TRUE WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3`, gameId, row-1, col)
			if err != nil {
				return nil, fmt.Errorf("failed to update adjacent bottom edge for box at row %d, col %d: %w", row-1, col, err)
			}
		}

	case "right_edge":
		if col < game.BoardSize-1 {
			_, err = repo.db.Exec(ctx, `UPDATE grids SET left_edge = TRUE WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3`, gameId, row, col+1)
			if err != nil {
				return nil, fmt.Errorf("failed to update adjacent left edge for box at row %d, col %d: %w", row, col+1, err)
			}
		}

	case "left_edge":
		if col > 0 {
			_, err = repo.db.Exec(ctx, `UPDATE grids SET right_edge = TRUE WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3`, gameId, row, col-1)
			if err != nil {
				return nil, fmt.Errorf("failed to update adjacent right edge for box at row %d, col %d: %w", row, col-1, err)
			}
		}

	case "bottom_edge":
		if row < game.BoardSize-1 {
			_, err = repo.db.Exec(ctx, `UPDATE grids SET top_edge = TRUE WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3`, gameId, row+1, col)
			if err != nil {
				return nil, fmt.Errorf("failed to update adjacent top edge for box at row %d, col %d: %w", row+1, col, err)
			}
		}
	}

	// Retrieve the updated grid
	boxes, err := repo.GetGrids(ctx, gameId)
	if err != nil {
		return nil, errors.New("failed to get updated grids after updating edge: " + err.Error())
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errors.New("failed to commit transaction: " + err.Error())
	}

	return boxes, nil
}

func (repo *PgGameRepository) SetBoxCompleted(ctx context.Context, gameId int, row int, col int, playerId int) error {
	query := `
        UPDATE grids
        SET completed = true, completed_by = $4
        WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3
    `
	_, err := repo.db.Exec(ctx, query, gameId, row, col, playerId)
	return err
}

func (repo *PgGameRepository) GetBoxByRowCol(ctx context.Context, gameId int, row int, col int) (*Box, error) {

	query := `
        SELECT box_id, grid_row, grid_col, top_edge, left_edge, right_edge, bottom_edge, completed
        FROM grids
        WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3
    `

	var box Box
	err := repo.db.QueryRow(ctx, query, gameId, row, col).Scan(
		&box.BoxId,
		&box.Row,
		&box.Col,
		&box.TopEdge,
		&box.LeftEdge,
		&box.RightEdge,
		&box.BottomEdge,
		&box.Completed,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("box not found for gameId: %d, row: %d, col: %d", gameId, row, col)
		}
		return nil, errors.New("failed to get box by row and col: " + err.Error())
	}

	return &box, nil
}

func (r *PgGameRepository) UpdateTurn(ctx context.Context, gameId int, turnOrder int) error {
	query := `UPDATE games SET current_turn = $1 WHERE game_id = $2`
	_, err := r.db.Exec(ctx, query, turnOrder, gameId)
	return err
}

func (repo *PgGameRepository) IsEdgeSelected(ctx context.Context, gameId int, row int, col int, edge string) (bool, error) {
	var result bool
	query := fmt.Sprintf(`SELECT %s FROM grids WHERE game_id = $1 AND grid_row = $2 AND grid_col = $3`, edge)
	err := repo.db.QueryRow(ctx, query, gameId, row, col).Scan(&result)
	if err != nil {
		return false, fmt.Errorf("failed to query edge status for box at row %d, col %d: %w", row, col, err)
	}
	return result, nil
}

func (repo *PgGameRepository) SetWinner(ctx context.Context, gameId int, winnerId *int) error {
	query := "UPDATE games SET winner_id = $1 WHERE game_id = $2"
	_, err := repo.db.Exec(ctx, query, winnerId, gameId)
	return err
}

func (repo *PgGameRepository) FindAllFromUser(ctx context.Context, userId int) ([]Game, error) {
	var games []Game

	query := "SELECT name, game_id, board_size WHERE user_id = $1"
	rows, err := repo.db.Query(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var game Game
		if err := rows.Scan(&game.GameName, &game.GameId, &game.BoardSize); err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return games, nil
}

func (repo *PgGameRepository) IncrementPlayerScore(ctx context.Context, gameId int, userId int) error {
	query := `
		UPDATE game_details
		SET score = score + 1
		WHERE game_id = $1 AND user_id = $2
	`
	results, err := repo.db.Exec(ctx, query, gameId, userId)
	if err != nil {
		return fmt.Errorf("failed to increment score for user %d in game %d: %w", userId, gameId, err)
	}
	if results.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for user %d in game %d", userId, gameId)
	}
	return nil
}

func (repo *PgGameRepository) GetPlayerScores(ctx context.Context, gameId int) (map[int]int, error) {
	query := `SELECT user_id, score FROM game_details WHERE game_id = $1`

	rows, err := repo.db.Query(ctx, query, gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch player scores: %w", err)
	}
	defer rows.Close()

	playerScores := make(map[int]int)
	for rows.Next() {
		var userId, score int
		if err := rows.Scan(&userId, &score); err != nil {
			return nil, fmt.Errorf("failed to scan player score: %w", err)
		}
		playerScores[userId] = score
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating player scores: %w", rows.Err())
	}

	return playerScores, nil
}
