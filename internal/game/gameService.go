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

func NewGameService(gameRepo GameRepository) *GameService{
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

func (s *GameService) CreateGame(ctx context.Context, p1 int, p2 int, boardSize int) (*Game, error){

	//Checks for if boardSize is more than 5 less than 10

	// if boardSize <= 4 || boardSize >= 11 {
	// 	return nil, errors.New("invalid board size: must be greater than 5 and less than 10, selected: " + strconv.Itoa(boardSize))
	// }

	if p1 == p2 {
		return nil, errors.New("players must be different")
	}

	game, err := s.gameRepo.Create(ctx, p1, p2, boardSize)
	if err != nil {
		return nil, err
	}

	slog.Info("New Game Created")

	return game, nil

}

// func (s *GameService) GetGrids(ctx context.Context, gameId int) ([]Box, error){

// 	boxes, err := s.gameRepo.GetGrids(ctx, gameId)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return boxes, nil

// }

// func (s *GameService) GetCurrentTurn(ctx context.Context, gameID int) (int, error) {
//     game, err := s.gameRepo.FindByID(ctx, gameID) 
//     if err != nil {
//         return 0, fmt.Errorf("failed to fetch game with ID %d: %v", gameID, err)
//     }
//     return game.CurrentTurn, nil
// }



func (s *GameService) MakeMove(ctx context.Context,gameId int, playerId int, row int , col int, edge string) (GameState, error){

	//Checks if edge is a valid option
	if edge != "top_edge" && edge != "right_edge" && edge != "left_edge" && edge != "bottom_edge" {
        return GameState{}, errors.New("invalid edge : " + edge)
    }

	//Gets Game from ID
	game, err := s.gameRepo.FindByID(ctx, gameId)
	if err != nil {
		return GameState{},  errors.New("failed to find game: " + err.Error())
	}

	//Checks if player is part of game
	if playerId != game.Player1 && playerId != game.Player2 {
		return GameState{}, errors.New("player is not part of this game")
	}

	//Check if its player's turn
	if game.CurrentTurn != playerId {
        return GameState{}, errors.New("it's not player " + strconv.Itoa(playerId) + " 's turn" )
    }

	//Checks if Selected Edge is a Valid Move (Not already chosen)
	isEdgeSelected, err := s.gameRepo.IsEdgeSelected(ctx, gameId, row, col, edge)
    if err != nil {
        return GameState{}, errors.New("failed to check if edge is already selected: " + err.Error())
    }

    if isEdgeSelected {
        return GameState{}, fmt.Errorf("invalid move: edge %s for box at row %d, col %d is already selected", edge, row, col)
    }

	boxCompleted := false

	//Updates Grid
	_, err = s.gameRepo.UpdateGrid(ctx, gameId , row , col , edge  )
	if err != nil {
		return GameState{}, errors.New("failed to update grid edge: " + err.Error())
	}

	

	// Check if the box at the current position was completed
	if completed, err := s.CheckSetCompletion(ctx, gameId, row, col, playerId); err == nil && completed {
		boxCompleted = true
	}


	//Checks if other boxes also get completed
	switch edge {
	case "top_edge":
		if row > 0 {
			completed, err := s.CheckSetCompletion(ctx, gameId, row-1, col, playerId)
			if err == nil && completed {
				boxCompleted = true
			}
		}
	case "left_edge":
		if col > 0 {
			completed, err := s.CheckSetCompletion(ctx, gameId, row, col-1, playerId)
			if err == nil && completed {
				boxCompleted = true
			}
		}
	case "right_edge":
		if col > game.BoardSize-1 {
			completed, err := s.CheckSetCompletion(ctx, gameId, row, col +1, playerId)
			if err == nil && completed {
				boxCompleted = true
			}
		}
	case "bottom_edge":
		if row > game.BoardSize-1{
			completed, err := s.CheckSetCompletion(ctx, gameId, row+1, col, playerId)
			if err == nil && completed {
				boxCompleted = true
			}
		}
	}
var nextPlayer int
	if !boxCompleted {
		slog.Info("box is not completed, updating turn")
    
    if game.CurrentTurn == playerId {
        // If it's the current player's turn, set the next player
        if game.Player1 == playerId {
            nextPlayer = game.Player2
		} else if 
         game.Player2 == playerId {
            nextPlayer = game.Player1
        } else {
            return GameState{}, errors.New("invalid player ID in game")
        }

		slog.Info("Updating turn to player:", nextPlayer)

        // Update the turn in the database
        if err := s.gameRepo.UpdateTurn(ctx, gameId, nextPlayer); err != nil {
            return GameState{}, fmt.Errorf("failed to update turn: %w", err)
        }
    }
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
		player1Score := 0
		player2Score := 0

		// Count the number of boxes each player has completed
		for _, box := range boxes {
			if *box.Completed_by == game.Player1 {
				player1Score++
			} else if *box.Completed_by == game.Player2 {
				player2Score++
			}
		}

		var winnerId *int
		if player1Score > player2Score {
			winnerId = &game.Player1
		} else if player2Score > player1Score {
			winnerId = &game.Player2
		} else {
			// It's a draw if the scores are equal
			winnerId = nil
		}

		if err := s.SetWinner(ctx, gameId, winnerId); err != nil {
			return GameState{}, errors.New("failed to set winner: " + err.Error())
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

//CheckSetCompletion Checks if board is completed and Sets to true, true means box has been completed
func (s *GameService) CheckSetCompletion(ctx context.Context, gameId, row, col int, playerId int) (bool, error) {
	box, err := s.gameRepo.GetBoxByRowCol(ctx, gameId, row, col)
	if err != nil {
		return false ,errors.New("failed to fetch box by row and col: " + err.Error())
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

//SetWinner sets winner of game duh
func (s *GameService) SetWinner(ctx context.Context, gameId int, winnerId *int) error {
	err := s.gameRepo.SetWinner(ctx, gameId, winnerId)
    if err != nil {
        return fmt.Errorf("failed to set winner for game %d: %v", gameId, err)
    }
    return nil	

}
