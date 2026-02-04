package game

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"
)

type GameService struct {
	gameRepo     GameRepository
	bus          events.EventBus
	botService   *BotService
	timerService *GameTimerService
}

func NewGameService(gameRepo GameRepository, bus events.EventBus, botService *BotService, timerService *GameTimerService) *GameService {
	return &GameService{
		gameRepo:     gameRepo,
		bus:          bus,
		botService:   botService,
		timerService: timerService,
	}
}

func (s *GameService) GetGame(ctx context.Context, gameID int) (*Game, error) {
	// Check if it's a bot game first
	if s.botService.IsBotGame(gameID) {
		return s.botService.GetBotGameState(gameID)
	}

	// Otherwise load from database (event replay or legacy fallback)
	return s.gameRepo.FindByID(ctx, gameID)
}

func (s *GameService) GetUserGameHistory(ctx context.Context, userID int) ([]GameHistoryEntry, error) {
	return s.gameRepo.FindUserGameHistory(ctx, userID)
}

// GetGameEvents returns all domain events for a game (for replay).
func (s *GameService) GetGameEvents(ctx context.Context, gameID int) ([]events.DomainEvent, error) {
	return s.gameRepo.LoadEvents(ctx, gameID)
}

func (s *GameService) CreateGame(ctx context.Context, playerIDs []int, boardSize int, usernames map[int]string) (*Game, error) {
	if boardSize <= 4 || boardSize >= 11 {
		return nil, errors.New("invalid board size: must be > 4 and < 11, got: " + strconv.Itoa(boardSize))
	}

	if len(playerIDs) < 2 || len(playerIDs) > 4 {
		return nil, errors.New("game requires 2-4 players")
	}

	seen := make(map[int]bool)
	for _, id := range playerIDs {
		if seen[id] {
			return nil, fmt.Errorf("duplicate player ID: %d", id)
		}
		seen[id] = true
	}

	players := make([]Player, len(playerIDs))
	for i, id := range playerIDs {
		username := usernames[id]
		if username == "" {
			username = fmt.Sprintf("Player%d", id)
		}
		players[i] = Player{
			UserID:   &id,
			Username: username,
		}
	}

	// Shuffle players for random turn order before raising the event
	rand.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	for i := range players {
		players[i].TurnOrder = i
		players[i].Score = 0
	}

	// Create aggregate via domain event (GameCreated)
	tempID := 0
	game := NewGame(&tempID, boardSize, players)

	// Persist: projection + events
	if err := s.gameRepo.Create(ctx, game); err != nil {
		return nil, err
	}

	slog.Info("Game created", "gameID", *game.GameID, "boardSize", boardSize, "players", len(players))

	s.publishIntegrationEvent(ctx, "global:games", "game_created", game)

	if s.timerService != nil {
		s.timerService.StartTimer(*game.GameID, game.Players, game.CurrentTurn)
	}

	return game, nil
}

func (s *GameService) MakeMove(ctx context.Context, gameID, playerID, row, col int, edge string) (*Game, error) {
	isBotGame := s.botService.IsBotGame(gameID)

	var game *Game
	var err error

	if isBotGame {
		game, err = s.botService.GetBotGameState(gameID)
	} else {
		game, err = s.gameRepo.FindByID(ctx, gameID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load game: %w", err)
	}

	// Find player's turn order
	var turnOrder int
	found := false
	for _, p := range game.Players {
		if p.UserID != nil && *p.UserID == playerID {
			turnOrder = p.TurnOrder
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("player %d not in game", playerID)
	}

	edgeType, err := parseEdge(edge)
	if err != nil {
		return nil, err
	}

	previousTurn := game.CurrentTurn
	move := Move{
		TurnOrder: turnOrder,
		Row:       row,
		Col:       col,
		Edge:      edgeType,
	}

	// ApplyMove now raises domain events internally
	result, err := game.ApplyMove(move)
	if err != nil {
		return nil, fmt.Errorf("invalid move: %w", err)
	}

	// Persist events and update projection (only for DB games)
	if !isBotGame {
		uncommitted := game.UncommittedEvents()
		if err := s.gameRepo.AppendEvents(ctx, gameID, uncommitted); err != nil {
			return nil, fmt.Errorf("failed to save events: %w", err)
		}
		game.ClearEvents()

		if err := s.gameRepo.UpdateProjection(ctx, game); err != nil {
			slog.Error("Failed to update projection", "gameID", gameID, "error", err)
		}
	}

	// Switch timer if turn changed
	if s.timerService != nil && result.NextTurn != previousTurn {
		s.timerService.SwitchTurn(*game.GameID, result.NextTurn)
	}

	slog.Info("Move applied",
		"gameID", *game.GameID,
		"playerID", playerID,
		"isBotGame", isBotGame,
		"boxesCompleted", len(result.CompletedBoxes),
		"gameOver", result.GameOver)

	// Publish integration event for WebSocket
	topic := fmt.Sprintf("game:%d", *game.GameID)
	s.publishIntegrationEvent(ctx, topic, "game:state", game)

	// If bot game and it's now a bot's turn, trigger bot moves
	if isBotGame && !result.GameOver {
		currentPlayer := game.GetCurrentPlayer()
		if currentPlayer != nil && currentPlayer.UserID != nil && *currentPlayer.UserID < 0 {
			time.Sleep(800 * time.Millisecond)
			go func() {
				if err := s.botService.PlayBotTurn(context.Background(), *game.GameID); err != nil {
					slog.Error("Bot turn failed", "gameID", *game.GameID, "error", err)
				}
			}()
		}
	}

	if result.GameOver {
		if s.timerService != nil {
			s.timerService.StopTimer(*game.GameID)
		}
		s.publishIntegrationEvent(ctx, "global:games", "game_completed", map[string]interface{}{
			"game_id":   *game.GameID,
			"winner_id": game.WinnerID,
		})
	}

	return game, nil
}

func (s *GameService) ForfeitGame(ctx context.Context, gameID, playerID int) (*Game, error) {
	isBotGame := s.botService.IsBotGame(gameID)

	var game *Game
	var err error

	if isBotGame {
		game, err = s.botService.GetBotGameState(gameID)
	} else {
		game, err = s.gameRepo.FindByID(ctx, gameID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load game: %w", err)
	}

	if game.EndedAt != nil {
		return nil, fmt.Errorf("game has already ended")
	}

	// Forfeit raises a GameForfeited domain event
	game.Forfeit(playerID)

	if !isBotGame {
		uncommitted := game.UncommittedEvents()
		if err := s.gameRepo.AppendEvents(ctx, gameID, uncommitted); err != nil {
			return nil, fmt.Errorf("failed to save forfeit events: %w", err)
		}
		game.ClearEvents()

		if err := s.gameRepo.UpdateProjection(ctx, game); err != nil {
			slog.Error("Failed to update projection after forfeit", "gameID", gameID, "error", err)
		}
	}

	slog.Info("Game forfeited", "gameID", *game.GameID, "forfeitedBy", playerID, "winnerID", game.WinnerID)

	if s.timerService != nil {
		s.timerService.StopTimer(*game.GameID)
	}

	topic := fmt.Sprintf("game:%d", *game.GameID)
	s.publishIntegrationEvent(ctx, topic, "game:state", game)
	s.publishIntegrationEvent(ctx, "global:games", "game_completed", map[string]interface{}{
		"game_id":   *game.GameID,
		"winner_id": game.WinnerID,
	})

	return game, nil
}

// publishIntegrationEvent publishes a WebSocket integration event (separate from domain events).
func (s *GameService) publishIntegrationEvent(ctx context.Context, topic, eventType string, payload interface{}) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal event payload", "type", eventType, "error", err)
		return
	}

	s.bus.Publish(ctx, topic, events.Event{
		Topic:   topic,
		Type:    eventType,
		Payload: payloadBytes,
	})
}

func parseEdge(edge string) (EdgeType, error) {
	switch edge {
	case "top_edge", "top":
		return TopEdge, nil
	case "right_edge", "right":
		return RightEdge, nil
	case "bottom_edge", "bottom":
		return BottomEdge, nil
	case "left_edge", "left":
		return LeftEdge, nil
	default:
		return "", fmt.Errorf("invalid edge: %s", edge)
	}
}
