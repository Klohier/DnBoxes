package game

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PgGameRepository struct {
    db *pgxpool.Pool
}

func NewPgGameRepository(db *pgxpool.Pool) *PgGameRepository{
	return &PgGameRepository{
		db: db,
	}
}

func (repo *PgGameRepository) FindAll( ctx context.Context) ([]Game, error){
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
	
func (repo *PgGameRepository) FindByID(ctx context.Context, id int) (*Game, error){
	var game Game
	query := `SELECT player1_id, player2_id, board_size, winner_id, current_turn FROM games WHERE game_id = $1`
	err := repo.db.QueryRow(ctx, query, id).Scan(&game.Player1, &game.Player2, &game.BoardSize, &game.WinnerId, &game.CurrentTurn)
	if err !=nil {
		return nil, fmt.Errorf("failed to find game %d : %w", id, err)
	}
	return &game, nil
}
	
func (repo *PgGameRepository) Create(ctx context.Context, p1 int, p2 int, boardSize int) (*Game, error){
	
	 // Start a new transaction
	 tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	 if err != nil {
		 return nil, errors.New("Failed to begin transaction: " + err.Error())
	 }
 
	 // Ensure that the transaction gets rolled back if there's an error
	 defer func() {
		 if err != nil {
			 tx.Rollback(ctx)
		 }
	 }()
 
	 var game Game
 
	 // Insert the game
	 query := `INSERT INTO games (player1_id, player2_id, board_size, current_turn) 
			   VALUES ($1, $2, $3, $4) 
			   RETURNING game_id, player1_id, player2_id, board_size, current_turn`
	 err = tx.QueryRow(ctx, query, p1, p2, boardSize, p1).Scan(&game.GameId, &game.Player1, &game.Player2, &game.BoardSize, &game.CurrentTurn)
	 if err != nil {
		 return nil, errors.New("Failed to create game: " + err.Error())
	 }
 
	 // Create the boxes associated with the game
	 for i := 0; i < boardSize; i++ {
		 for j := 0; j < boardSize; j++ {
			 // Insert each box associated with the game
			 boxQuery := `INSERT INTO grids (game_id, row, col, completed_by) 
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
 
	 // Return the created game object
	 return &game, nil
}


func (repo *PgGameRepository) GetGrids(ctx context.Context, gameId int) ([]Box, error){
	var boxes []Box
	query := `
	SELECT box_id, game_id, top_edge, right_edge, left_edge, bottom_edge, row, col, completed, completed_by 
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
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }

	game, err := repo.FindByID(ctx, gameId)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve game %d: %w", gameId, err)
    }



	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

    query := fmt.Sprintf(`UPDATE grids SET %s = TRUE WHERE game_id = $1 AND row = $2 AND col = $3`, edge)
    _, err = repo.db.Exec(ctx, query, gameId, row, col)
    if err != nil {
        return nil, fmt.Errorf("failed to update edge for box at row %d, col %d: %w", row, col, err)
    }

	switch edge {
    case "top_edge":
        // Update the bottom edge of the box above, if it exists
        if row > 0 {
            _, err = repo.db.Exec(ctx, `UPDATE grids SET bottom_edge = TRUE WHERE game_id = $1 AND row = $2 AND col = $3`, gameId, row-1, col)
            if err != nil {
                return nil, fmt.Errorf("failed to update adjacent bottom edge for box at row %d, col %d: %w", row-1, col, err)
            }
        }

    case "right_edge":
        // Update the left edge of the box to the right, if it exists
        if col < game.BoardSize-1 { 
            _, err = repo.db.Exec(ctx, `UPDATE grids SET left_edge = TRUE WHERE game_id = $1 AND row = $2 AND col = $3`, gameId, row, col+1)
            if err != nil {
                return nil, fmt.Errorf("failed to update adjacent left edge for box at row %d, col %d: %w", row, col+1, err)
            }
        }

    case "left_edge":
        // Update the right edge of the box to the left, if it exists
        if col > 0 {
            _, err = repo.db.Exec(ctx, `UPDATE grids SET right_edge = TRUE WHERE game_id = $1 AND row = $2 AND col = $3`, gameId, row, col-1)
            if err != nil {
                return nil, fmt.Errorf("failed to update adjacent right edge for box at row %d, col %d: %w", row, col-1, err)
            }
        }

    case "bottom_edge":
        // Update the top edge of the box below, if it exists
        if row < game.BoardSize-1 { 
            _, err = repo.db.Exec(ctx, `UPDATE grids SET top_edge = TRUE WHERE game_id = $1 AND row = $2 AND col = $3`, gameId, row+1, col)
            if err != nil {
                return nil, fmt.Errorf("failed to update adjacent top edge for box at row %d, col %d: %w", row+1, col, err)
            }
        }
    }

    // Retrieve the updated grid
    boxes, err := repo.GetGrids(ctx, gameId)
    if err != nil {
        return nil, fmt.Errorf("failed to get updated grids after updating edge: %w", err)
    }

	if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    return boxes, nil
}

func (repo *PgGameRepository) SetBoxCompleted(ctx context.Context, gameId int, row int, col int, playerId int) error {
    query := `
        UPDATE grids
        SET completed = true, completed_by = $4
        WHERE game_id = $1 AND row = $2 AND col = $3
    `
    _, err := repo.db.Exec(ctx, query, gameId, row, col, playerId)
    return err
}


func (repo *PgGameRepository) GetBoxByRowCol(ctx context.Context, gameId int, row int, col int) (*Box, error) {

	query := `
        SELECT box_id, row, col, top_edge, left_edge, right_edge, bottom_edge, completed
        FROM grids
        WHERE game_id = $1 AND row = $2 AND col = $3
    `
    
    var box Box
    // Execute the query with pgx and scan the result
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
        return nil, fmt.Errorf("failed to get box by row and col: %w", err)
    }
	// log.Println(box.LeftEdge)

    return &box, nil
}

func (r *PgGameRepository) UpdateTurn(ctx context.Context, gameId int, userID int) error {
	log.Println(userID)
    query := `UPDATE games SET current_turn = $1 WHERE game_id = $2`
    _, err := r.db.Exec(ctx, query, userID, gameId)
    return err
}


func (repo *PgGameRepository) IsEdgeSelected(ctx context.Context, gameId int, row int, col int, edge string) (bool, error) {
    var result bool
    query := fmt.Sprintf(`SELECT %s FROM grids WHERE game_id = $1 AND row = $2 AND col = $3`, edge)
    err := repo.db.QueryRow(ctx, query, gameId, row, col).Scan(&result)
    if err != nil {
        return false, fmt.Errorf("failed to query edge status for box at row %d, col %d: %w", row, col, err)
    }
    return result, nil
}

func (repo *PgGameRepository) SetWinner(ctx context.Context, gameId int, winnerId int) error {
	query := "UPDATE games SET winner_id = $1 WHERE game_id = $2"
    _, err := repo.db.Exec(ctx, query, winnerId, gameId)
    return err
}