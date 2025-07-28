package game

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
)

type GameService struct {
	gameRepo GameRepository
}

func NewGameService(gameRepo GameRepository) *GameService {
	return &GameService{
		gameRepo: gameRepo,
	}
}

func (s *GameService) GetGameState(ctx context.Context, gameId int) (*GameState, error) {
	game, err := s.gameRepo.FindByID(ctx, gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch game: %w", err)
	}

	grids, err := s.gameRepo.GetGrids(ctx, gameId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch grids: %w", err)
	}

	return &GameState{
		Game:  game,
		Grids: grids,
	}, nil
}

func (s *GameService) CreateGame(ctx context.Context, playerIDs []int, boardSize int, sessionId int) (*Game, error) {

	//Checks for if boardSize is more than 5 less than 10

	if boardSize <= 4 || boardSize >= 11 {
		return nil, errors.New("invalid board size: must be greater than 5 and less than 10, selected: " + strconv.Itoa(boardSize))
	}

	if len(playerIDs) < 2 {
		return nil, errors.New("a game requires at least 2 players")
	}

	// Ensure no duplicate player IDs
	playerIDSet := make(map[int]struct{})
	for _, id := range playerIDs {
		if _, exists := playerIDSet[id]; exists {
			return nil, fmt.Errorf("duplicate player ID detected: %d", id)
		}
		playerIDSet[id] = struct{}{}
	}

	game, err := s.gameRepo.Create(ctx, playerIDs, boardSize, sessionId)
	if err != nil {
		return nil, err
	}

	slog.Info("New Game Created")

	return game, nil

}

func (s *GameService) MakeMove(ctx context.Context, gameId int, playerId int, row int, col int, edge string) (*GameState, error) {

// 	//Checks if edge is a valid option
// 	if edge != "top_edge" && edge != "right_edge" && edge != "left_edge" && edge != "bottom_edge" {
// 		return nil, errors.New("invalid edge : " + edge)
// 	}

	move := Move{
		Row: row,
		Col: col,
		Edge: edge,
		UserID: playerId,
	}

	gameState, err := s.GetGameState(ctx, gameId)
	if err != nil {
		return nil, errors.New("failed to find game: " + err.Error())
	}

	
	engine := NewEngine(gameState)
    result, err := engine.ApplyMove(move)
    if err != nil {
        return nil, err
    }

	for r := 0; r < engine.BoardSize; r++ {
		for c := 0; c < engine.BoardSize; c++ {
			box := engine.Grid[r][c]
			// Update only edges that are true in this box
			if box.TopEdge {
				if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "top_edge"); err != nil {
					return nil, err
				}
			}
			if box.RightEdge {
				if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "right_edge"); err != nil {
					return nil, err
				}
			}
			if box.BottomEdge {
				if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "bottom_edge"); err != nil {
					return nil, err
				}
			}
			if box.LeftEdge {
				if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "left_edge"); err != nil {
					return nil, err
				}
			}

			
		}
	}


	for _, box := range result.ClaimedBoxes {
		if err := s.gameRepo.IncrementPlayerScore(ctx, gameId, playerId); err != nil {
			slog.Error("failed to increment player score", "error", err)
		}

		if err := s.gameRepo.SetBoxCompleted(ctx, box.GameId, box.Row, box.Col, playerId); err != nil {
        slog.Error("failed to set box completed", "error", err)
        return nil, err
    }
	}
   

	if err := s.gameRepo.UpdateTurn(ctx, gameId, result.NextTurn); err != nil {
		return nil, fmt.Errorf("failed to update turn: %w", err)
	}

	


	if result.WinnerID != nil {
	if err := s.gameRepo.SetWinner(ctx, gameId, result.WinnerID); err != nil {
		slog.Error("failed to persist winner", "error", err)
		return nil, err
	}

}

	// Reload updated game and grids after persistence

	updatedGameState, err := s.GetGameState(ctx, gameId)
	if err != nil {
		return nil, errors.New("failed to get grids after move: " + err.Error())
	}

	return updatedGameState, nil

}

// CheckSetCompletion Checks if board is completed and Sets to true, true means box has been completed
func (s *GameService) CheckSetCompletion(ctx context.Context, gameId, row, col int, playerId int) (bool, error) {
	box, err := s.gameRepo.GetBoxByRowCol(ctx, gameId, row, col)
	if err != nil {
		return false, errors.New("failed to fetch box by row and col: " + err.Error())
	}

	// Check if all edges are selected
	if box.TopEdge && box.LeftEdge && box.RightEdge && box.BottomEdge {
		// Set the box as completed
		err := s.gameRepo.SetBoxCompleted(ctx, gameId, row, col, playerId)
		if err != nil {
			return false, errors.New("failed to set box as completed: " + err.Error())
		}
		return true, nil
	}
	return false, nil
}

// SetWinner sets winner of game duh
func (s *GameService) SetWinner(ctx context.Context, gameId int, winnerId *int) error {
	err := s.gameRepo.SetWinner(ctx, gameId, winnerId)
	if err != nil {
		return fmt.Errorf("failed to set winner for game %d: %v", gameId, err)
	}
	return nil

}
