package game

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// BotService handles bot-specific game logic
type BotService struct {
	mu       sync.RWMutex
	botGames map[int]*Game  // Changed from *GameState to *Game
	nextID   int
	bus      events.EventBus
}

// NewBotService creates a new BotService
func NewBotService(bus events.EventBus) *BotService {
	return &BotService{
		botGames: make(map[int]*Game),
		nextID:   10000,
			bus:      bus,
	}
}

// CreateBotGame creates a new in-memory bot game
func (b *BotService) CreateBotGame(playerIDs []int, numBots int, boardSize int, usernames map[int]string) (*Game, error) {
	if boardSize <= 4 || boardSize >= 20 {
		return nil, fmt.Errorf("invalid board size: must be >4 and <20")
	}

	if len(playerIDs) == 0 {
		return nil, fmt.Errorf("must provide at least one human player")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Create players (human + bots)
	players := make([]Player, 0, len(playerIDs)+numBots)

	// Add human players
	for _, id := range playerIDs {
		username := usernames[id]
		if username == "" {
			username = fmt.Sprintf("Player%d", id)
		}
		players = append(players, Player{
			UserID:      &id,
			Username:    username,
			TurnOrder:   len(players),
			IsAnonymous: false,
			Score:       0,
		})
	}

	// Add bot players
	for i := 0; i < numBots; i++ {
		botID := -(i + 1) // negative IDs for bots: -1, -2, etc.
		players = append(players, Player{
			UserID:      &botID,
			Username:    fmt.Sprintf("Bot%d", i+1),
			TurnOrder:   len(players),
			IsAnonymous: false,
			Score:       0,
		})
	}

	// Generate bot game ID
	gameID := b.nextID
	b.nextID++

	// Create game using domain model
	game := NewGame(&gameID, boardSize, players)

	// Store in memory
	b.botGames[gameID] = game

	slog.Info("Created bot game",
		"gameID", gameID,
		"players", len(players),
		"boardSize", boardSize,
	)

	return game, nil
}

// PlayBotTurn executes bot moves until turn ends or game is over
func (b *BotService) PlayBotTurn(ctx context.Context, gameID int) error {
	for {
		game, err := b.GetBotGameState(gameID)
		if err != nil {
			return err
		}

		// Check if game is over
		if game.IsGameOver() {
			slog.Info("Bot game finished", "gameID", gameID, "winner", game.WinnerID)
			break
		}

		// Get current player
		currentPlayer := game.GetCurrentPlayer()
		if currentPlayer == nil {
			return fmt.Errorf("no current player found")
		}

		// Check if current player is a bot
		if currentPlayer.UserID == nil || *currentPlayer.UserID >= 0 {
			// Not a bot's turn
			break
		}

		// Generate bot move
		move := game.GenerateBotMove(currentPlayer.TurnOrder)
		if move == nil {
			slog.Warn("Bot has no valid moves, skipping", 
				"gameID", gameID, 
				"botID", *currentPlayer.UserID)
			
			game.CurrentTurn = (game.CurrentTurn + 1) % len(game.Players)
			continue
		}

		// Apply bot move
		result, err := game.ApplyMove(*move)
		if err != nil {
				slog.Error("Bot move failed", 
				"gameID", gameID, 
				"botID", *currentPlayer.UserID,
				"error", err)
		}

		slog.Info("Bot made move",
			"gameID", gameID,
			"botID", *currentPlayer.UserID,
			"row", move.Row,
			"col", move.Col,
			"edge", move.Edge,
			"boxesCompleted", len(result.CompletedBoxes))

		// If bot didn't complete a box, turn passes to next player
		// if len(result.CompletedBoxes) == 0 {
		// 	slog.Info("Bot didn't complete box, turn ending", "gameID", gameID)
		// 	break
		// }

		b.publishGameState(ctx, gameID, game)


			if len(result.CompletedBoxes) > 0 {
			slog.Info("Bot completed boxes, checking if it gets another turn", 
				"gameID", gameID,
				"boxCount", len(result.CompletedBoxes))
				time.Sleep(600 * time.Millisecond)
		}

		slog.Info("Bot completed boxes, taking another turn", "gameID", gameID, "boxCount", len(result.CompletedBoxes))
	}

	// Publish final state
	game, err := b.GetBotGameState(gameID)
	if err != nil {
		return err
	}
	
	b.publishGameState(ctx, gameID, game)


	if game.IsGameOver() {
		b.publishGameCompleted(ctx, gameID, game.WinnerID)
		slog.Info("Bot game completed", "gameID", gameID, "winner", game.WinnerID)
	}


	return nil
}

func (b *BotService) publishGameState(ctx context.Context, gameID int, game *Game) {
	topic := fmt.Sprintf("game:%d", gameID)
	
	payloadBytes, err := json.Marshal(game)
	if err != nil {
		slog.Error("Failed to marshal game state", "gameID", gameID, "error", err)
		return
	}

	b.bus.Publish(ctx, topic, events.Event{
		Topic:   topic,
		Type:    "game:state",
		Payload: payloadBytes,
	})
}

func (b *BotService) publishGameCompleted(ctx context.Context, gameID int, winnerID *int) {
	payloadBytes, err := json.Marshal(map[string]interface{}{
		"game_id":   gameID,
		"winner_id": winnerID,
	})
	if err != nil {
		slog.Error("Failed to marshal game completed event", "gameID", gameID, "error", err)
		return
	}

	b.bus.Publish(ctx, "global:games", events.Event{
		Topic:   "global:games",
		Type:    "game_completed",
		Payload: payloadBytes,
	})
}

// GetBotGameState returns the in-memory bot game state
func (b *BotService) GetBotGameState(gameID int) (*Game, error) {
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
	slog.Info("Deleted bot game", "gameID", gameID)
}

// IsBotGame checks if a game ID is a bot game
func (b *BotService) IsBotGame(gameID int) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.botGames[gameID]
	return ok
}