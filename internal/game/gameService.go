package game

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

type GameService struct {
	gameRepo GameRepository
	mu       sync.RWMutex
	botGames map[int]*GameState 
	nextID   int           
}

func NewGameService(gameRepo GameRepository) *GameService {
	return &GameService{
		gameRepo: gameRepo,
		botGames: make(map[int]*GameState),
		nextID:   10000,
	}
}

func (s *GameService) GetBotGameState(gameID int) (*GameState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	game, ok := s.botGames[gameID]
	if !ok {
        fmt.Printf("Bot game %d not found\n", gameID)
        return nil, fmt.Errorf("bot game %d not found", gameID)
    }
	return game, nil
}

func (s *GameService) CreateBotGameInMemory(ctx context.Context, playerIDs []int, boardSize int, sessionId int) (*GameState, error) {
	if boardSize <= 4 || boardSize >= 11 {
		return nil, errors.New("invalid board size: must be >4 and <11")
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create players
	players := []Player{
		{
			UserID:      playerIDs[0],
			Username:    "Human",
			TurnOrder:   0,
			IsSpectator: false,
			IsConnected: true,
			Score:       0,
		},
		{
			UserID:      -1,
			Username:    "Bot",
			TurnOrder:   1,
			IsSpectator: false,
			IsConnected: true,
			Score:       0,
		},
	}

	s.nextID++
	gameID := s.nextID
	now := time.Now()

	// Construct Game
	turnIndex := 0
	game := &Game{
		GameId:      &gameID,
		SessionId:   sessionId,
		GameName:    nil,
		Players:     players,
		BoardSize:   boardSize,
		WinnerId:    nil,
		CreatedAt:   now,
		CurrentTurn: &turnIndex, 
	}

	// Create empty boxes
	grids := []Box{}
	boxID := 1
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			grids = append(grids, Box{
				BoxId:      boxID,
				GameId:     gameID,
				TopEdge:    false,
				LeftEdge:   false,
				RightEdge:  false,
				BottomEdge: false,
				Row:        row,
				Col:        col,
				Completed:  nil,
				Completed_by: nil,
			})
			boxID++
		}
	}

	state := &GameState{
		Game:  game,
		Grids: grids,
	}

	s.botGames[gameID] = state

	slog.Info("Created in-memory bot game",
		"gameID", gameID,
		"sessionID", sessionId,
		"boardSize", boardSize,
		"players", players,
		"boxesCount", len(grids),
		"state", state,
	)


	

	return state, nil
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

var gameState *GameState
var err error

	//Checks if edge is a valid option
	if edge != "top_edge" && edge != "right_edge" && edge != "left_edge" && edge != "bottom_edge" {
		return nil, errors.New("invalid edge : " + edge)
	}

	move := Move{
		Row: row,
		Col: col,
		Edge: edge,
		UserID: playerId,
	}

		

// s.mu.Lock() 
    // defer s.mu.Unlock()
s.mu.RLock()
	if inMemGame, exists := s.botGames[gameId]; exists {
		slog.Info("Detected bot game: using in-memory state")
		gameState = inMemGame
	} else {
		s.mu.RUnlock()
		gameState, err = s.GetGameState(ctx, gameId)
		if err != nil {
			return nil, errors.New("failed to find game: " + err.Error())
		}
		// s.mu.Lock()
	}



	
	engine := NewEngine(gameState)
    result, err := engine.ApplyMove(move)
    if err != nil {
        return nil, err
    }

	// For Bot Game
	if botGame, ok := s.botGames[gameId]; ok {
    botGame.Game.CurrentTurn = &engine.CurrentTurn
    botGame.Game.WinnerId = engine.WinnerID
    botGame.Game.Players = engine.Players
    botGame.Game.BoardSize = engine.BoardSize

   
    for i := range botGame.Game.Players {
        playerID := botGame.Game.Players[i].UserID
        if score, ok := engine.Scores[playerID]; ok {
            botGame.Game.Players[i].Score = score
        } else {
            botGame.Game.Players[i].Score = 0
        }
    }

    // Flatten the grid for engine
    botGame.Grids = flattenGrid(engine.Grid)
}

	


	

	if !s.isBotGame(gameState.Game) {

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

	gameState.Game.CurrentTurn = &result.NextTurn


}

	nextTurn := result.NextTurn

	nextPlayerID := gameState.Game.Players[nextTurn].UserID

if isBot(nextPlayerID) {
	slog.Info("Triggering bot move")
	
	//TODO: Causes race condition
	go func() {
            if _, err := s.makeBotMoves(ctx, gameId, nextPlayerID); err != nil {
                slog.Error("Bot move error", "err", err)
            }
        }()
}


	if result.WinnerID != nil {
	if err := s.gameRepo.SetWinner(ctx, gameId, result.WinnerID); err != nil {
		slog.Error("failed to persist winner", "error", err)
		return nil, err
	}

}

	// Reload updated game and grids after persistence

	if !s.isBotGame(gameState.Game) {
    updatedGameState, err := s.GetGameState(ctx, gameId)
    if err != nil {
        return nil, errors.New("failed to get grids after move: " + err.Error())
    }
    return updatedGameState, nil
}
	return gameState, nil

}



func (s *GameService) isBotGame(game *Game) bool {
    for _, playerId := range game.Players {
        if isBot(playerId.UserID) {
            return true
        }
    }
    return false
}

func (s *GameService) makeBotMoves(ctx context.Context, gameId int, botId int) (*GameState, error) {
	for {
		gameState, err := s.GetBotGameState(gameId)
		if err != nil {
			return nil, fmt.Errorf("failed to get game state for bot: %w", err)
		}

		engine := NewEngine(gameState)
		move := engine.GenerateBotMove(botId)
		if move == nil {
			slog.Warn("Bot has no valid moves")
			break
		}

		slog.Info("Bot is making move", "move", move)

		_, err = s.MakeMove(ctx, gameId, botId, move.Row, move.Col, move.Edge)
		if err != nil {
			slog.Error("Bot move failed", "error", err)
			return nil, err
		}

		time.Sleep(10 * time.Millisecond)

		// Stop if bot no longer gets a second move (i.e., didn't close a box)
		if gameState.Game.CurrentTurn == nil || *gameState.Game.CurrentTurn != botId {
			break
		}
	}

	return s.GetBotGameState(gameId)
}

// SetWinner sets winner of game duh
func (s *GameService) SetWinner(ctx context.Context, gameId int, winnerId *int) error {
	err := s.gameRepo.SetWinner(ctx, gameId, winnerId)
	if err != nil {
		return fmt.Errorf("failed to set winner for game %d: %v", gameId, err)
	}
	return nil

}


// TODO: Look to change this

func isBot(playerID int) bool {
	return playerID == -1 
}

func flattenGrid(grid [][]Box) []Box {
    var boxes []Box
    for _, row := range grid {
        boxes = append(boxes, row...)
    }
    return boxes

}
