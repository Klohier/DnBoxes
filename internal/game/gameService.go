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

func (s *GameService) MakeMove(ctx context.Context, gameId int, playerId int, row int, col int, edge string) (GameState, error) {

	//Checks if edge is a valid option
	if edge != "top_edge" && edge != "right_edge" && edge != "left_edge" && edge != "bottom_edge" {
		return GameState{}, errors.New("invalid edge : " + edge)
	}

	//Gets Game from ID
	game, err := s.gameRepo.FindByID(ctx, gameId)
	if err != nil {
		return GameState{}, errors.New("failed to find game: " + err.Error())
	}

	// Checks if player is part of the game
	playerTurnOrder := -1
	found := false
	for _, p := range game.Players {
		if p.UserID == playerId {
			playerTurnOrder = p.TurnOrder
			found = true
			break
		}
	}

	if !found {
		return GameState{}, errors.New("player is not part of this game")
	}

	//Check if its player's turn
	if *game.CurrentTurn != playerTurnOrder {
		return GameState{}, errors.New("it's not player " + strconv.Itoa(playerId) + " 's turn")
	}

	//Checks if Selected Edge is a Valid Move (Not already chosen)
	isEdgeSelected, err := s.gameRepo.IsEdgeSelected(ctx, gameId, row, col, edge)
	if err != nil {
		return GameState{}, errors.New("failed to check if edge is already selected: " + err.Error())
	}

	if isEdgeSelected {
		return GameState{}, fmt.Errorf("invalid move: edge %s for box at row %d, col %d is already selected", edge, row, col)
	}

	//Updates Grid
	_, err = s.gameRepo.UpdateGrid(ctx, gameId, row, col, edge)
	if err != nil {
		return GameState{}, errors.New("failed to update grid edge: " + err.Error())
	}

	boxCompleted := false
	// Check if the box at the current position was completed
	if completed, err := s.CheckSetCompletion(ctx, gameId, row, col, playerId); err == nil && completed {
		boxCompleted = true
		if err := s.gameRepo.IncrementPlayerScore(ctx, gameId, playerId); err != nil {
			slog.Error("failed to increment player score", "error", err)
		}
	}

	//Checks if other boxes also get completed
	switch edge {
	case "top_edge":
		if row > 0 {
			completed, err := s.CheckSetCompletion(ctx, gameId, row-1, col, playerId)
			if err == nil && completed {
				boxCompleted = true
				if err := s.gameRepo.IncrementPlayerScore(ctx, gameId, playerId); err != nil {
					slog.Error("failed to increment player score", "error", err)
				}
			}
		}
	case "left_edge":
		if col > 0 {
			completed, err := s.CheckSetCompletion(ctx, gameId, row, col-1, playerId)
			if err == nil && completed {
				boxCompleted = true
				if err := s.gameRepo.IncrementPlayerScore(ctx, gameId, playerId); err != nil {
					slog.Error("failed to increment player score", "error", err)
				}
			}
		}
	case "right_edge":
		if col < game.BoardSize-1 {
			completed, err := s.CheckSetCompletion(ctx, gameId, row, col+1, playerId)
			if err == nil && completed {
				boxCompleted = true
				if err := s.gameRepo.IncrementPlayerScore(ctx, gameId, playerId); err != nil {
					slog.Error("failed to increment player score", "error", err)
				}
			}
		}
	case "bottom_edge":
		if row < game.BoardSize-1 {
			completed, err := s.CheckSetCompletion(ctx, gameId, row+1, col, playerId)
			if err == nil && completed {
				boxCompleted = true
				if err := s.gameRepo.IncrementPlayerScore(ctx, gameId, playerId); err != nil {
					slog.Error("failed to increment player score", "error", err)
				}
			}
		}
	}

	var nextPlayerId int
	if !boxCompleted {
		slog.Info("box is not completed, updating turn")
		nextTurnOrder := (*game.CurrentTurn + 1) % len(game.Players)
		if err := s.gameRepo.UpdateTurn(ctx, gameId, nextTurnOrder); err != nil {
			return GameState{}, fmt.Errorf("failed to update turn: %w", err)
		}

		slog.Info("Updating turn to player:", "nextPlayerId", nextPlayerId)

	}

	// Check if there are any more moves left
	boxes, err := s.gameRepo.GetGrids(ctx, gameId)
	if err != nil {
		return GameState{}, errors.New("failed to get grids:" + err.Error())
	}

	allCompleted := true
	//Checks if any box has an edge left as false
	for _, box := range boxes {
		if !box.TopEdge || !box.LeftEdge || !box.RightEdge || !box.BottomEdge {
			allCompleted = false
			break
		}
	}

	if allCompleted {
		playerScores, err := s.gameRepo.GetPlayerScores(ctx, gameId)
		if err != nil {
			return GameState{}, fmt.Errorf("failed to get player scores: %w", err)
		}

		var winnerId *int
		highestScore := -1
		tie := false

		for userId, score := range playerScores {
			if score > highestScore {
				highestScore = score
				winnerId = &userId
				tie = false
			} else if score == highestScore {
				tie = true
			}
		}

		// If there's a tie, set winnerId to nil
		if tie {
			winnerId = nil
		}

		if err := s.SetWinner(ctx, gameId, winnerId); err != nil {
			return GameState{}, fmt.Errorf("failed to set winner: %w", err)
		}
	}

	// Reload updated game state
	updatedGame, err := s.gameRepo.FindByID(ctx, gameId)
	if err != nil {
		return GameState{}, errors.New("failed to reload game after move: " + err.Error())
	}

	//Grab Most Latest Board to Send
	updatedGrids, err := s.gameRepo.GetGrids(ctx, gameId)
	if err != nil {
		return GameState{}, errors.New("failed to get grids after move: " + err.Error())
	}

	return GameState{
		Game:  updatedGame,
		Grids: updatedGrids,
	}, nil

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
