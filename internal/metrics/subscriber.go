package metrics

import (
	"context"
	"dango/internal/events"
	"log/slog"
)

type MetricsCollector struct {
    metrics *Metrics
    bus     events.EventBus
}

func NewMetricsCollector(ctx context.Context, bus events.EventBus) *MetricsCollector {
    m := &MetricsCollector{
        metrics: New(),
        bus:     bus,
    }
    
    m.subscribeToEvents(ctx)
    
    return m
}

func (m *MetricsCollector) subscribeToEvents(ctx context.Context) {  

	  if err := m.bus.Subscribe(ctx, "connections", m.handleConnectionEvent); err != nil {
        slog.Error("failed to subscribe to connections", "error", err)
    }

	if err := m.bus.Subscribe(ctx, "global:games", m.handleGameEvent); err != nil {
        slog.Error("failed to subscribe to global:games", "error", err)
    }

	if err := m.bus.Subscribe(ctx, "game:state_updated", m.handleGameStateUpdate); err != nil {
        slog.Error("failed to subscribe to game:state_updated", "error", err)
    }

	  if err := m.bus.Subscribe(ctx, "global:lobbies", m.handleLobbyEvent); err != nil {
        slog.Error("failed to subscribe to global:lobbies", "error", err)
    }
}

func (m *MetricsCollector) handleGameStateUpdate(event events.Event) {
    slog.Info("METRICS: Game state update received", 
        "topic", event.Topic,
        "type", event.Type)
    
    m.metrics.IncrementMessages()
}


func (m *MetricsCollector) GetMetrics() *Metrics {
    return m.metrics
}

func (m *MetricsCollector) handleGameEvent(event events.Event) {
    switch event.Type {
    case "game_created":
        m.metrics.IncrementGames()
        slog.Info("METRICS: Game created", "active_games", m.metrics.activeGames.Load())
    case "game_completed":
        m.metrics.DecrementGames()
        slog.Info("METRICS: Game completed", "active_games", m.metrics.activeGames.Load())
    case "game_updated":
        m.metrics.IncrementMessages()
        slog.Debug("METRICS: Game updated")
    default:
        slog.Debug("METRICS: Unknown game event", "type", event.Type)
    }
}

func (m *MetricsCollector) handleLobbyEvent(event events.Event) {
    
    switch event.Type {
    case "lobby_created":
        m.handleLobbyCreated()
    case "lobby_updated":
        m.handleLobbyUpdated()
    case "lobby_deleted":
        m.handleLobbyDeleted()
    default:
        slog.Debug("METRICS: Unknown lobby event", "type", event.Type)
    }
}

func (m *MetricsCollector) handleLobbyCreated() {
    m.metrics.IncrementLobbies()
    slog.Info("METRICS: Lobby created", "active_lobbies", m.metrics.activeLobbies.Load())
}

func (m *MetricsCollector) handleLobbyUpdated() {
    m.metrics.IncrementMessages()
    slog.Debug("METRICS: Lobby updated")
}

func (m *MetricsCollector) handleLobbyDeleted() {
    m.metrics.DecrementLobbies()
    slog.Info("METRICS: Lobby deleted", "active_lobbies", m.metrics.activeLobbies.Load())
}

func (m *MetricsCollector) handleConnectionEvent(event events.Event) {
    switch event.Type {
    case "user_connected":
        m.metrics.IncrementConnections()
        slog.Info("METRICS: User connected", "active_connections", m.metrics.activeConnections.Load())
    case "user_disconnected":
        m.metrics.DecrementConnections()
        slog.Info("METRICS: User disconnected", "active_connections", m.metrics.activeConnections.Load())
    default:
        slog.Debug("METRICS: Unknown connection event", "type", event.Type)
    }
}