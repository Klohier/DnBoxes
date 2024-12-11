package game

import (
	"context"
	"errors"
	"fmt"
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

func (s *GameService) CreateGame(ctx context.Context, p1 int, p2 int, boardSize int) (*Game, error){

	//Checks for if boardSize is more than 5 less than 10

	if boardSize <= 5 || boardSize >= 10 {
		return nil, errors.New("invalid board size: must be greater than 5 and less than 10, selected: " + strconv.Itoa(boardSize))
	}

	game, err := s.gameRepo.Create(ctx, p1, p2, boardSize)
	if err != nil {
		return nil, err
	}

	return game, nil

}

func (s *GameService) GetGrids(ctx context.Context, gameId int) ([]Box, error){

	boxes, err := s.gameRepo.GetGrids(ctx, gameId)

	if err != nil {
		return nil, err
	}

	return boxes, nil

}



func (s *GameService) MakeMove(ctx context.Context,gameId int, playerId int, row int , col int, edge string) ([]Box, error){

	//Checks if edge is a valid option
	if edge != "top_edge" && edge != "right_edge" && edge != "left_edge" && edge != "bottom_edge" {
        return nil, errors.New("invalid edge : " + edge)
    }

	//Gets Game from ID
	game, err := s.gameRepo.FindByID(ctx, gameId)
	if err != nil {
		return nil, errors.New("failed to find game: " + err.Error())
	}

	//Checks if player is part of game
	if playerId != game.Player1 && playerId != game.Player2 {
		return nil, errors.New("player is not part of this game")
	}

	//Check if its player's turn
	if game.CurrentTurn != playerId {
        return nil, errors.New("it's not player " + strconv.Itoa(playerId) + " 's turn" )
    }

	//Checks if Selected Edge is a Valid Move (Not already chosen)
	isEdgeSelected, err := s.gameRepo.IsEdgeSelected(ctx, gameId, row, col, edge)
    if err != nil {
        return nil, errors.New("failed to check if edge is already selected: " + err.Error())
    }

    if isEdgeSelected {
        return nil, fmt.Errorf("invalid move: edge %s for box at row %d, col %d is already selected", edge, row, col)
    }

	//Updates Grid
	grid, err := s.gameRepo.UpdateGrid(ctx, gameId , row , col , edge  )
	if err != nil {
		return nil, errors.New("failed to update grid edge: " + err.Error())
	}


	return grid, nil

}

