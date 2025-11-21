package game

import (
	"context"
	"dango/internal/events"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type GameService struct {
	gameRepo    GameRepository
	// mu          sync.RWMutex
	// botGames    map[int]*GameState
	redisClient *redis.Client
	botService  *BotService
	// nextID      int
}

func NewGameService(gameRepo GameRepository, redisClient *redis.Client, botService *BotService) *GameService {
	return &GameService{
		gameRepo:    gameRepo,
		botService:  botService,
		// botGames:    make(map[int]*GameState),
		// nextID:      10000,
		redisClient: redisClient,
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

// func (s *GameService) MakeMove(ctx context.Context, gameId int, playerId int, row int, col int, edge string) (*GameState, error) {

// 	// var gameState *GameState
// 	// var err error
// 	channel := fmt.Sprintf("game:%d", gameId)

// 	//Checks if edge is a valid option
// 	if err := validateEdge(edge); err != nil {
//         return nil, err
//     }

// 	gameState, err := s.GetGameState(ctx, gameId)
//     if err != nil {
//         return nil, fmt.Errorf("failed to fetch game state: %w", err)
//     }

// 	move := Move{
// 		Row:    row,
// 		Col:    col,
// 		Edge:   edge,
// 		UserID: playerId,
// 	}




// 	engine := NewEngine(gameState)
// 	result, err := engine.ApplyMove(move)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Persist the move (scores, grids, turn)
//     if err := s.persistMove(ctx, gameId, playerId, engine, result); err != nil {
//         return nil, err
//     }

//     // Publish updated game state to subscribers
//     if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, gameState); err != nil {
//         slog.Error("failed to publish message:", "error", err)
//     }

//     // Return updated game state
//     updatedGameState, err := s.GetGameState(ctx, gameId)
//     if err != nil {
//         return nil, fmt.Errorf("failed to reload game state: %w", err)
//     }

// 	// For Bot Game
// 	// if botGame, ok := s.botGames[gameId]; ok {
// 	// 	botGame.Game.CurrentTurn = &engine.CurrentTurn
// 	// 	botGame.Game.WinnerId = engine.WinnerID
// 	// 	botGame.Game.Players = engine.Players
// 	// 	botGame.Game.BoardSize = engine.BoardSize

// 	// 	for i := range botGame.Game.Players {
// 	// 		playerID := botGame.Game.Players[i].UserID
// 	// 		if score, ok := engine.Scores[playerID]; ok {
// 	// 			botGame.Game.Players[i].Score = score
// 	// 		} else {
// 	// 			botGame.Game.Players[i].Score = 0
// 	// 		}
// 	// 	}

// 	// 	// Flatten the grid for engine
// 	// 	botGame.Grids = flattenGrid(engine.Grid)
// 	// }

// 	// if !s.isBotGame(gameState.Game) {
// 	// 	if err := s.persistMove(ctx, gameId, playerId, engine, result); err != nil {
//     //         return nil, err
//     //     }
// 	// }

// 	// nextTurn := result.NextTurn

// 	// nextPlayerID := gameState.Game.Players[nextTurn].UserID

// 	// if isBot(nextPlayerID) {
// 	// 	slog.Info("Triggering bot move")

// 	// 	//TODO: Causes race condition
// 	// 	// go func() {
// 	// 	if _, err := s.makeBotMoves(ctx, gameId, nextPlayerID); err != nil {
// 	// 		slog.Error("Bot move error", "err", err)
// 	// 	}
// 	// 	// }()
// 	// }

// 	// if result.WinnerID != nil {
// 	// 	if !s.isBotGame(gameState.Game) {
// 	// 		if err := s.gameRepo.SetWinner(ctx, gameId, result.WinnerID); err != nil {
// 	// 			slog.Error("failed to persist winner", "error", err)
// 	// 			return nil, err
// 	// 		}
// 	// 	} else {
// 	// 		slog.Info("Bot game finished", "gameId", gameId, "winnerId", *result.WinnerID)
// 	// 	}
// 	// 	if s.isBotGame(gameState.Game) {
// 	// 		s.mu.Lock()
// 	// 		delete(s.botGames, gameId)
// 	// 		s.mu.Unlock()
// 	// 		slog.Info("Deleted finished bot game from memory", "gameId", gameId)
// 	// 	}

// 	// }

// 	// // Reload updated game and grids after persistence

// 	// if !s.isBotGame(gameState.Game) {
// 	// 	updatedGameState, err := s.GetGameState(ctx, gameId)
// 	// 	if err != nil {
// 	// 		return nil, errors.New("failed to get grids after move: " + err.Error())
// 	// 	}

// 	// 	if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, gameState); err != nil {
// 	// 		slog.Error("failed to publish message:", "error", err)
// 	// 	}
// 	// 	return updatedGameState, nil

// 	// }

// 	// if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, gameState); err != nil {
// 	// 	slog.Error("failed to publish message:", "error", err)
// 	// }
// 	// return gameState, nil
//  return updatedGameState, nil
// }

func (s *GameService) MakeMove(ctx context.Context, gameID int, playerID int, row int, col int, edge string) (*GameState, error) {
	// Validate edge
	if err := validateEdge(edge); err != nil {
		return nil, err
	}

	// Check if bot game
	if state, err := s.botService.GetBotGameState(gameID); err == nil {
		return s.botMakeMove(ctx, state, playerID, row, col, edge)
	}

	// Normal DB-backed game
	return s.dbMakeMove(ctx, gameID, playerID, row, col, edge)
}

func (s *GameService) botMakeMove(ctx context.Context, state *GameState, playerID, row, col int, edge string) (*GameState, error) {
	engine := NewEngine(state)

	// Apply the human move
	move := Move{Row: row, Col: col, Edge: edge, UserID: playerID}
	result, err := engine.ApplyMove(move)
	if err != nil {
		return nil, err
	}

	

	// Update bot game state
	state.Game.CurrentTurn = &result.NextTurn
	state.Game.WinnerId = result.WinnerID
	state.Game.Players = engine.Players
	state.Grids = flattenGrid(engine.Grid)

	// Publish human move
	channel := fmt.Sprintf("game:%d", *state.Game.GameId)
	if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, state); err != nil {
		slog.Error("failed to publish human move", "error", err)
	}

	// If it's now a bot turn, let the bot play
	currentPlayer := state.Game.Players[*state.Game.CurrentTurn]
	if isBot(currentPlayer.UserID) {
		if err := s.botService.PlayBotTurn(ctx, *state.Game.GameId, currentPlayer.UserID, func(gs *GameState, mv *Move) (*GameState, error) {
			engine := NewEngine(gs)
			res, err := engine.ApplyMove(*mv)
			if err != nil {
				return nil, err
			}

			// Update state
			gs.Game.CurrentTurn = &res.NextTurn
			gs.Game.WinnerId = res.WinnerID
			gs.Game.Players = engine.Players
			gs.Grids = flattenGrid(engine.Grid)

			// Publish bot move
			if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, gs); err != nil {
				slog.Error("failed to publish bot move", "error", err)
			}

			return gs, nil
		}); err != nil {
			return nil, err
		}
	}

	return state, nil
}

// ---------------------- DB Move Logic ----------------------

func (s *GameService) dbMakeMove(ctx context.Context, gameID, playerID, row, col int, edge string) (*GameState, error) {
	channel := fmt.Sprintf("game:%d", gameID)
	gameState, err := s.GetGameState(ctx, gameID)
	if err != nil {
		return nil, err
	}

	engine := NewEngine(gameState)
	move := Move{Row: row, Col: col, Edge: edge, UserID: playerID}
	result, err := engine.ApplyMove(move)
	if err != nil {
		return nil, err
	}

	if err := s.persistMove(ctx, gameID, playerID, engine, result); err != nil {
		return nil, err
	}

	if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, gameState); err != nil {
		slog.Error("failed to publish move", "error", err)
	}

	return s.GetGameState(ctx, gameID)
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


func validateEdge(edge string) error {
    switch edge {
    case "top_edge", "right_edge", "bottom_edge", "left_edge":
        return nil
    default:
        return fmt.Errorf("invalid edge: %s", edge)
    }
}

func (s *GameService) persistMove(ctx context.Context, gameId, playerId int, engine *Engine, result MoveResult) error {
    for r := 0; r < engine.BoardSize; r++ {
        for c := 0; c < engine.BoardSize; c++ {
            box := engine.Grid[r][c]
            if box.TopEdge {
                if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "top_edge"); err != nil {
                    return err
                }
            }
            if box.RightEdge {
                if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "right_edge"); err != nil {
                    return err
                }
            }
            if box.BottomEdge {
                if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "bottom_edge"); err != nil {
                    return err
                }
            }
            if box.LeftEdge {
                if _, err := s.gameRepo.UpdateGrid(ctx, gameId, r, c, "left_edge"); err != nil {
                    return err
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
            return err
        }
    }

    return s.gameRepo.UpdateTurn(ctx, gameId, result.NextTurn)
}
