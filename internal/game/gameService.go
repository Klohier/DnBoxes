package game

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
)

type GameService struct {
	gameRepo    GameRepository
	// mu          sync.RWMutex
	// botGames    map[int]*GameState
	bus        events.EventBus
	botService  *BotService
	// nextID      int
}

func NewGameService(gameRepo GameRepository, bus events.EventBus, botService *BotService,) *GameService {
	return &GameService{
		gameRepo:    gameRepo,
		botService:  botService,
		// botGames:    make(map[int]*GameState),
		// nextID:      10000,
		bus:        bus,
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

func (s *GameService) CreateGame(ctx context.Context, playerIDs []int, boardSize int) (*Game, error) {

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

	game, err := s.gameRepo.Create(ctx, playerIDs, boardSize)
	if err != nil {
		return nil, err
	}

	slog.Info("New Game Created", "gameID", *game.GameId)

	payloadBytes, err := json.Marshal(game)
	if err != nil {
		slog.Error("Failed to marshal game", "error", err)
		return game, nil
	}

	s.bus.Publish(ctx, "global:games", events.Event{
		Topic:   "global:games",
		Type:    "game_created",
		Payload: payloadBytes,
	})

	return game, nil

}

func (s *GameService) MakeMove(ctx context.Context, gameID int, playerID int, row int, col int, edge string) (*GameState, error) {
	// Validate edge
	if err := validateEdge(edge); err != nil {
		return nil, err
	}

	// Check if bot game
	// if state, err := s.botService.GetBotGameState(gameID); err == nil {
	// 	return s.botMakeMove(ctx, state, playerID, row, col, edge)
	// }

	// Normal DB-backed game
	return s.dbMakeMove(ctx, gameID, playerID, row, col, edge)
}

// func (s *GameService) botMakeMove(ctx context.Context, state *GameState, playerID, row, col int, edge string) (*GameState, error) {
// 	game := NewGame(state)

// 	// Apply the human move
// 	move := Move{Row: row, Col: col, Edge: edge, UserID: playerID}
// 	result, err := game.ApplyMove(move)
// 	if err != nil {
// 		return nil, err
// 	}

	

// 	// Update bot game state
// 	state.Game.CurrentTurn = &result.NextTurn
// 	state.Game.WinnerId = result.WinnerID
// 	state.Game.Players = game.Players
// 	state.Grids = flattenGrid(game.Grid)

// 	// Publish human move
// 	// channel := fmt.Sprintf("game:%d", *state.Game.GameId)
// 	// if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, state); err != nil {
// 	// 	slog.Error("failed to publish human move", "error", err)
// 	// }
// 	s.publishGameUpdate(ctx, *state.Game.GameId, state)

// 	// If it's now a bot turn, let the bot play
// 	currentPlayer := state.Game.Players[*state.Game.CurrentTurn]
// 	if isBot(currentPlayer.UserID) {
// 		if err := s.botService.PlayBotTurn(ctx, *state.Game.GameId, currentPlayer.UserID, func(gs *GameState, mv *Move) (*GameState, error) {
// 			game := NewGame(gs)
// 			res, err := game.ApplyMove(*mv)
// 			if err != nil {
// 				return nil, err
// 			}

// 			// Update state
// 			gs.Game.CurrentTurn = &res.NextTurn
// 			gs.Game.WinnerId = res.WinnerID
// 			gs.Game.Players = game.Players
// 			gs.Grids = flattenGrid(game.Grid)

// 			// Publish bot move
// 			s.publishGameUpdate(ctx, *gs.Game.GameId, gs)
// 			// if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, gs); err != nil {
// 			// 	slog.Error("failed to publish bot move", "error", err)
// 			// }

// 			return gs, nil
// 		}); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return state, nil
// }

// ---------------------- DB Move Logic ----------------------

func (s *GameService) dbMakeMove(ctx context.Context, gameID, playerID, row, col int, edge string) (*GameState, error) {
	// channel := fmt.Sprintf("game:%d", gameID)
	gameState, err := s.GetGameState(ctx, gameID)
	if err != nil {
		return nil, err
	}

	game := NewGame(gameState)
	move := Move{Row: row, Col: col, Edge: edge, UserID: playerID}
	result, err := game.ApplyMove(move)
	if err != nil {
		return nil, err
	}

	if err := s.persistMove(ctx, gameID, playerID, game, result); err != nil {
		return nil, err
	}

	updatedState, err := s.GetGameState(ctx, gameID)
	if err != nil {
		return nil, err
	}

	s.publishGameUpdate(ctx, gameID, updatedState)
	// if err := events.PublishEvent(ctx, s.redisClient, channel, events.EventMessage, gameState); err != nil {
	// 	slog.Error("failed to publish move", "error", err)
	// }

	return updatedState, nil
}

func (s *GameService) publishGameUpdate(ctx context.Context, gameID int, gameState *GameState) {
	topic := fmt.Sprintf("game:%d", gameID)
	
	payloadBytes, err := json.Marshal(gameState)
	if err != nil {
		slog.Error("Failed to marshal game state", "gameID", gameID, "error", err)
		return
	}

	event := events.Event{
		Topic:   topic,
		Type:    "game:state",
		Payload: payloadBytes,
	}

	s.bus.Publish(ctx, topic, event)

	s.bus.Publish(ctx, "game:state_updated", event)
}


// SetWinner sets winner of game duh
func (s *GameService) SetWinner(ctx context.Context, gameId int, winnerId *int) error {
	err := s.gameRepo.SetWinner(ctx, gameId, winnerId)
	if err != nil {
		return fmt.Errorf("failed to set winner for game %d: %v", gameId, err)
	}

		slog.Info("Game completed", "gameID", gameId, "winnerID", winnerId)

		payload := map[string]interface{}{
		"game_id":   gameId,
		"winner_id": winnerId,
	}
	payloadBytes, _ := json.Marshal(payload)
	
	s.bus.Publish(ctx, "global:games", events.Event{
		Topic:   "global:games",
		Type:    "game_completed",
		Payload: payloadBytes,
	})
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

func (s *GameService) persistMove(ctx context.Context, gameId, playerId int, game *Game, result MoveResult) error {
    for r := 0; r < game.BoardSize; r++ {
        for c := 0; c < game.BoardSize; c++ {
            box := game.Grid[r][c]
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

	if err := s.gameRepo.UpdateTurn(ctx, gameId, result.NextTurn); err != nil {
        return err
    }

	if result.WinnerID != nil {
        slog.Info("Game finished", "gameId", gameId, "winnerId", *result.WinnerID)
        if err := s.SetWinner(ctx, gameId, result.WinnerID); err != nil {
            return err
        }
    }

    return nil
}
