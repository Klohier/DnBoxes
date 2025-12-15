package game

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// BotService handles bot-specific game logic
type BotService struct {
	mu       sync.RWMutex
	botGames map[int]*GameState
}

// NewBotService creates a new BotService
func NewBotService() *BotService {
	return &BotService{
		botGames: make(map[int]*GameState),
	}
}

// CreateBotGame creates a new in-memory bot game
func (b *BotService) CreateBotGame(playerIDs []int, numBots int, boardSize int) (*GameState, error) {
	if boardSize <= 4 || boardSize >= 20 {
		return nil, fmt.Errorf("invalid board size: must be >4 and <20")
	}

	if len(playerIDs) == 0 {
		return nil, fmt.Errorf("must provide at least one human player")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Create human players first
	players := make([]Player, len(playerIDs))
	for i, id := range playerIDs {
		players[i] = Player{
			UserID:      id,
			Username:    fmt.Sprintf("Human%d", i+1),
			TurnOrder:   i,
			IsSpectator: false,
			IsConnected: true,
			Score:       0,
		}
	}

	// Add bot players
	for i := 1; i < numBots; i++ {
		botID := -i // negative IDs for bots: -1, -2, etc.
		players = append(players, Player{
			UserID:      botID,
			Username:    fmt.Sprintf("Bot%d", i+1),
			TurnOrder:   len(players),
			IsSpectator: false,
			IsConnected: true,
			Score:       0,
		})
	}

	gameID := generateBotGameID()
	turnIndex := 0
	now := time.Now()

	game := &Game{
		GameId:      &gameID,
		// SessionId:   sessionID,
		Players:     players,
		BoardSize:   boardSize,
		WinnerId:    nil,
		CreatedAt:   now,
		CurrentTurn: turnIndex,
	}

	// Create empty boxes
	grids := make([]Box, 0, boardSize*boardSize)
	boxID := 1
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			grids = append(grids, Box{
				BoxId:      boxID,
				GameId:     gameID,
				Row:        row,
				Col:        col,
			})
			boxID++
		}
	}

	state := &GameState{
		Game:  game,
		Grids: grids,
	}

	b.botGames[gameID] = state

	slog.Info("Created bot game",
		"gameID", gameID,
		"players", len(players),
		"boardSize", boardSize,
	)

	return state, nil
}


// PlayBotTurn executes bot moves until turn ends
func (b *BotService) PlayBotTurn(ctx context.Context, gameID int, botID int, applyMove func(*GameState, *Move) (*GameState, error)) error {
	for {
		gameState, err := b.GetBotGameState(gameID)
		if err != nil {
			return err
		}

		game := NewGame(gameState)
		move := game.GenerateBotMove(botID)
		if move == nil {
			slog.Info("Bot has no valid moves")
			break
		}

		gameState, err = applyMove(gameState, move)
		if err != nil {
			return fmt.Errorf("bot move failed: %w", err)
		}

		currentPlayer := gameState.Game.Players[gameState.Game.CurrentTurn]
		if currentPlayer.UserID != botID {
			break
		}

		// Optional delay for better UX
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

// GetBotGameState returns the in-memory bot game state
func (b *BotService) GetBotGameState(gameID int) (*GameState, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	game, ok := b.botGames[gameID]
	if !ok {
		return nil, fmt.Errorf("bot game %d not found", gameID)
	}
	return game, nil
}

// DeleteBotGame removes a finished bot game
func (b *BotService) DeleteBotGame(gameID int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.botGames, gameID)
}

// generateBotGameID generates a simple in-memory ID
var nextBotGameID = 10000
func generateBotGameID() int {
	nextBotGameID++
	return nextBotGameID
}

