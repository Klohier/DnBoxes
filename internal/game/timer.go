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

const (
	DefaultPlayerTime     = 5 * time.Minute
	DisconnectGracePeriod = 30 * time.Second
	TimerTickInterval     = 1 * time.Second
)

// TimerState represents the timer data sent to clients
type TimerState struct {
	GameID     int               `json:"game_id"`
	Players    []PlayerTimerInfo `json:"players"`
	ActiveTurn int               `json:"active_turn"`
}

// PlayerTimerInfo represents timer info for a single player
type PlayerTimerInfo struct {
	TurnOrder    int   `json:"turn_order"`
	UserID       int   `json:"user_id"`
	RemainingMs  int64 `json:"remaining_ms"`
	Disconnected bool  `json:"disconnected"`
}

// GameTimer tracks timer state for a single game
type GameTimer struct {
	mu               sync.Mutex
	gameID           int
	playerTimes      map[int]time.Duration // turnOrder -> remaining time (checkpoint)
	playerUserIDs    map[int]int           // turnOrder -> userID
	userIDToTurn     map[int]int           // userID -> turnOrder
	activeTurn       int
	turnStartedAt   time.Time
	disconnected     map[int]bool        // userID -> is disconnected
	disconnectTimers map[int]*time.Timer // userID -> 30s grace timer
	stopped          bool
	stopChan         chan struct{}
}

// GameTimerService manages timers for all active games
type GameTimerService struct {
	mu        sync.RWMutex
	timers    map[int]*GameTimer
	bus       events.EventBus
	forfeitFn func(ctx context.Context, gameID, playerID int) error
}

// NewGameTimerService creates a new timer service
func NewGameTimerService(bus events.EventBus) *GameTimerService {
	return &GameTimerService{
		timers: make(map[int]*GameTimer),
		bus:    bus,
	}
}

// SetForfeitFunc sets the function called when a player times out or disconnects
func (s *GameTimerService) SetForfeitFunc(fn func(ctx context.Context, gameID, playerID int) error) {
	s.forfeitFn = fn
}

// StartTimer initializes and starts a timer for a game
func (s *GameTimerService) StartTimer(gameID int, players []Player, startingTurn int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't start duplicate timers
	if _, exists := s.timers[gameID]; exists {
		return
	}

	timer := &GameTimer{
		gameID:           gameID,
		playerTimes:      make(map[int]time.Duration),
		playerUserIDs:    make(map[int]int),
		userIDToTurn:     make(map[int]int),
		activeTurn:       startingTurn,
		turnStartedAt:   time.Now(),
		disconnected:     make(map[int]bool),
		disconnectTimers: make(map[int]*time.Timer),
		stopChan:         make(chan struct{}),
	}

	for _, p := range players {
		timer.playerTimes[p.TurnOrder] = DefaultPlayerTime
		if p.UserID != nil {
			timer.playerUserIDs[p.TurnOrder] = *p.UserID
			timer.userIDToTurn[*p.UserID] = p.TurnOrder
		}
	}

	s.timers[gameID] = timer

	go timer.run(s)

	slog.Info("Game timer started", "gameID", gameID, "players", len(players))
}

// SwitchTurn is called when a move changes the active turn
func (s *GameTimerService) SwitchTurn(gameID, newTurn int) {
	s.mu.RLock()
	timer, exists := s.timers[gameID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	timer.mu.Lock()
	defer timer.mu.Unlock()

	if timer.stopped || timer.activeTurn == newTurn {
		return
	}

	// Deduct elapsed time from the old active player
	elapsed := time.Since(timer.turnStartedAt)
	timer.playerTimes[timer.activeTurn] -= elapsed

	// Switch to new turn
	timer.activeTurn = newTurn
	timer.turnStartedAt = time.Now()
}

// StopTimer stops and removes a game timer
func (s *GameTimerService) StopTimer(gameID int) {
	s.mu.Lock()
	timer, exists := s.timers[gameID]
	if !exists {
		s.mu.Unlock()
		return
	}
	delete(s.timers, gameID)
	s.mu.Unlock()

	timer.mu.Lock()
	defer timer.mu.Unlock()

	if !timer.stopped {
		timer.stopped = true
		close(timer.stopChan)
	}

	// Cancel all disconnect timers
	for _, t := range timer.disconnectTimers {
		t.Stop()
	}

	slog.Info("Game timer stopped", "gameID", gameID)
}

// HandleDisconnect starts a 30s grace period for a disconnected player
func (s *GameTimerService) HandleDisconnect(userID int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, timer := range s.timers {
		timer.mu.Lock()
		if _, inGame := timer.userIDToTurn[userID]; inGame {
			timer.disconnected[userID] = true

			// Cancel any existing disconnect timer
			if existing, ok := timer.disconnectTimers[userID]; ok {
				existing.Stop()
			}

			// Start 30s grace period
			gameID := timer.gameID
			timer.disconnectTimers[userID] = time.AfterFunc(DisconnectGracePeriod, func() {
				s.handleDisconnectTimeout(gameID, userID)
			})

			slog.Info("Player disconnected, starting 30s grace period",
				"userID", userID, "gameID", timer.gameID)
		}
		timer.mu.Unlock()
	}
}

// HandleReconnect cancels the disconnect grace period
func (s *GameTimerService) HandleReconnect(userID int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, timer := range s.timers {
		timer.mu.Lock()
		if _, inGame := timer.userIDToTurn[userID]; inGame {
			timer.disconnected[userID] = false

			if t, ok := timer.disconnectTimers[userID]; ok {
				t.Stop()
				delete(timer.disconnectTimers, userID)
			}

			slog.Info("Player reconnected, grace period cancelled",
				"userID", userID, "gameID", timer.gameID)
		}
		timer.mu.Unlock()
	}
}

// GetTimerState returns the current timer state for a game
func (s *GameTimerService) GetTimerState(gameID int) *TimerState {
	s.mu.RLock()
	timer, exists := s.timers[gameID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	return timer.buildState()
}

// HasTimer checks if a game has an active timer
func (s *GameTimerService) HasTimer(gameID int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.timers[gameID]
	return exists
}

// handleDisconnectTimeout is called when 30s grace period expires
func (s *GameTimerService) handleDisconnectTimeout(gameID, userID int) {
	slog.Info("Disconnect grace period expired, forfeiting",
		"gameID", gameID, "userID", userID)

	if s.forfeitFn != nil {
		if err := s.forfeitFn(context.Background(), gameID, userID); err != nil {
			slog.Error("Failed to forfeit on disconnect timeout",
				"gameID", gameID, "userID", userID, "error", err)
		}
	}

	s.StopTimer(gameID)
}

// handleTimeout is called when a player's clock reaches zero
func (s *GameTimerService) handleTimeout(gameID, userID int) {
	slog.Info("Player timed out", "gameID", gameID, "userID", userID)

	if s.forfeitFn != nil {
		if err := s.forfeitFn(context.Background(), gameID, userID); err != nil {
			slog.Error("Failed to forfeit on timeout",
				"gameID", gameID, "userID", userID, "error", err)
		}
	}

	s.StopTimer(gameID)
}

// run is the main timer loop, ticking every second
func (t *GameTimer) run(service *GameTimerService) {
	ticker := time.NewTicker(TimerTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if t.tick(service) {
				return
			}
		case <-t.stopChan:
			return
		}
	}
}

// tick checks time and publishes state. Returns true if timer expired.
func (t *GameTimer) tick(service *GameTimerService) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.stopped {
		return true
	}

	// Calculate remaining time for active player
	elapsed := time.Since(t.turnStartedAt)
	remaining := t.playerTimes[t.activeTurn] - elapsed

	if remaining <= 0 {
		t.stopped = true
		userID := t.playerUserIDs[t.activeTurn]
		go service.handleTimeout(t.gameID, userID)
		return true
	}

	// Publish timer state
	state := t.buildStateLocked()
	service.publishTimerState(state)

	return false
}

// buildState creates a TimerState snapshot (acquires lock)
func (t *GameTimer) buildState() *TimerState {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.buildStateLocked()
}

// buildStateLocked creates a TimerState snapshot (caller must hold lock)
func (t *GameTimer) buildStateLocked() *TimerState {
	now := time.Now()
	state := &TimerState{
		GameID:     t.gameID,
		ActiveTurn: t.activeTurn,
		Players:    make([]PlayerTimerInfo, 0, len(t.playerTimes)),
	}

	for turnOrder, remaining := range t.playerTimes {
		effectiveRemaining := remaining
		if turnOrder == t.activeTurn {
			effectiveRemaining -= now.Sub(t.turnStartedAt)
			if effectiveRemaining < 0 {
				effectiveRemaining = 0
			}
		}

		userID := t.playerUserIDs[turnOrder]
		state.Players = append(state.Players, PlayerTimerInfo{
			TurnOrder:    turnOrder,
			UserID:       userID,
			RemainingMs:  effectiveRemaining.Milliseconds(),
			Disconnected: t.disconnected[userID],
		})
	}

	return state
}

// publishTimerState publishes timer state to the game room
func (s *GameTimerService) publishTimerState(state *TimerState) {
	topic := fmt.Sprintf("game:%d", state.GameID)

	payload, err := json.Marshal(state)
	if err != nil {
		slog.Error("Failed to marshal timer state", "error", err)
		return
	}

	s.bus.Publish(context.Background(), topic, events.Event{
		Topic:   topic,
		Type:    events.EventGameTimer,
		Payload: payload,
	})
}

// ListenForConnectionEvents subscribes to connection events for disconnect/reconnect handling
func (s *GameTimerService) ListenForConnectionEvents() {
	err := s.bus.Subscribe(context.Background(), "connections", func(e events.Event) {
		var payload struct {
			UserID int `json:"user_id"`
		}
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			slog.Error("Failed to unmarshal connection event", "error", err)
			return
		}

		switch e.Type {
		case "user_disconnected":
			s.HandleDisconnect(payload.UserID)
		case "user_connected":
			s.HandleReconnect(payload.UserID)
		}
	})

	if err != nil {
		slog.Error("Failed to subscribe to connection events", "error", err)
	}
}
