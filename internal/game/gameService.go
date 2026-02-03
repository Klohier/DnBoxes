package game

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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
	
	// Otherwise load from database
	return s.gameRepo.FindByID(ctx, gameID)
}

func (s *GameService) GetUserGameHistory(ctx context.Context, userID int) ([]GameHistoryEntry, error) {
	return s.gameRepo.FindUserGameHistory(ctx, userID)
}

func (s *GameService) GetGameMoves(ctx context.Context, gameID int) ([]GameMove, error) {
	return s.gameRepo.FindMovesByGameID(ctx, gameID)
}

func (s *GameService) CreateGame(ctx context.Context, playerIDs []int, boardSize int) (*Game, error) {
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
	
	// TODO: Fetch actual usernames from user repository/service
	players := make([]Player, len(playerIDs))
	for i, id := range playerIDs {
		players[i] = Player{
			UserID:   &id,
			Username: fmt.Sprintf("Player%d", id), // Placeholder
		}
	}
	
	game, err := s.gameRepo.Create(ctx, players, boardSize)
	if err != nil {
		return nil, err
	}
	
	slog.Info("Game created", "gameID", *game.GameID, "boardSize", boardSize, "players", len(players))

	s.publishEvent(ctx, "global:games", "game_created", game)

	// Start game timer (server-authoritative chess clock)
	if s.timerService != nil {
		s.timerService.StartTimer(*game.GameID, game.Players, game.CurrentTurn)
	}

	return game, nil
}

func (s *GameService) MakeMove(ctx context.Context, gameID, playerID, row, col int, edge string) (*Game, error) {
	// Check if it's a bot game
	isBotGame := s.botService.IsBotGame(gameID)
	
	// Load game
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
	
	// Convert edge string to EdgeType
	edgeType, err := parseEdge(edge)
	if err != nil {
		return nil, err
	}
	
	// Apply move
	previousTurn := game.CurrentTurn
	move := Move{
		TurnOrder: turnOrder,
		Row:       row,
		Col:       col,
		Edge:      edgeType,
	}

	result, err := game.ApplyMove(move)
	if err != nil {
		return nil, fmt.Errorf("invalid move: %w", err)
	}

	// Save game (only for DB games)
	if !isBotGame {
		if err := s.gameRepo.Update(ctx, game); err != nil {
			return nil, fmt.Errorf("failed to save game: %w", err)
		}
		// Record move for replay
		if err := s.gameRepo.SaveMove(ctx, gameID, GameMove{
			TurnOrder: turnOrder,
			Row:       row,
			Col:       col,
			Edge:      string(edgeType),
		}); err != nil {
			slog.Error("Failed to save move for replay", "gameID", gameID, "error", err)
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
	
	// Publish human move
	topic := fmt.Sprintf("game:%d", *game.GameID)
	s.publishEvent(ctx, topic, "game:state", game)
	
	// If bot game and it's now a bot's turn, trigger bot moves
	if isBotGame && !result.GameOver {
		currentPlayer := game.GetCurrentPlayer()
		if currentPlayer != nil && currentPlayer.UserID != nil && *currentPlayer.UserID < 0 {
			// It's a bot's turn, play bot moves asynchronously
			// BotService will handle publishing updates
			time.Sleep(800 * time.Millisecond)
			go func() {
				if err := s.botService.PlayBotTurn(context.Background(), *game.GameID); err != nil {
					slog.Error("Bot turn failed", "gameID", *game.GameID, "error", err)
				}
			}()
		}
	}
	
	if result.GameOver {
		// Stop timer when game ends
		if s.timerService != nil {
			s.timerService.StopTimer(*game.GameID)
		}
		s.publishEvent(ctx, "global:games", "game_completed", map[string]interface{}{
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

	// Determine winner: the opponent with the highest score
	var winnerID *int
	maxScore := -1
	for _, p := range game.Players {
		if p.UserID != nil && *p.UserID == playerID {
			continue
		}
		if p.Score > maxScore {
			maxScore = p.Score
			winnerID = p.UserID
		}
	}

	game.WinnerID = winnerID
	now := time.Now()
	game.EndedAt = &now

	if !isBotGame {
		if err := s.gameRepo.Update(ctx, game); err != nil {
			return nil, fmt.Errorf("failed to save game: %w", err)
		}
	}

	slog.Info("Game forfeited", "gameID", *game.GameID, "forfeitedBy", playerID, "winnerID", winnerID)

	// Stop timer on forfeit
	if s.timerService != nil {
		s.timerService.StopTimer(*game.GameID)
	}

	topic := fmt.Sprintf("game:%d", *game.GameID)
	s.publishEvent(ctx, topic, "game:state", game)
	s.publishEvent(ctx, "global:games", "game_completed", map[string]interface{}{
		"game_id":   *game.GameID,
		"winner_id": game.WinnerID,
	})

	return game, nil
}

func (s *GameService) publishEvent(ctx context.Context, topic, eventType string, payload interface{}) {
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