// internal/metrics/metrics.go
package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metrics struct {
    activeConnections atomic.Int64
    activeGames       atomic.Int64
    totalMessages     atomic.Int64
    totalMoves        atomic.Int64
    
    mu              sync.RWMutex
    messagesPerSec  int64
    lastMessageTime time.Time
}

func New() *Metrics {
    m := &Metrics{
        lastMessageTime: time.Now(),
    }
    
    // Start background goroutine to calculate messages per second
    go m.calculateRate()
    
    return m
}

func (m *Metrics) IncrementConnections() {
    m.activeConnections.Add(1)
}

func (m *Metrics) DecrementConnections() {
    m.activeConnections.Add(-1)
}

func (m *Metrics) IncrementGames() {
    m.activeGames.Add(1)
}

func (m *Metrics) DecrementGames() {
    m.activeGames.Add(-1)
}

func (m *Metrics) IncrementMessages() {
    m.totalMessages.Add(1)
}

func (m *Metrics) IncrementMoves() {
    m.totalMoves.Add(1)
}

func (m *Metrics) calculateRate() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    lastCount := m.totalMessages.Load()
    
    for range ticker.C {
        currentCount := m.totalMessages.Load()
        rate := currentCount - lastCount
        
        m.mu.Lock()
        m.messagesPerSec = rate
        m.mu.Unlock()
        
        lastCount = currentCount
    }
}

func (m *Metrics) GetSnapshot() map[string]interface{} {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    return map[string]interface{}{
        "active_connections": m.activeConnections.Load(),
        "active_games":       m.activeGames.Load(),
        "total_messages":     m.totalMessages.Load(),
        "total_moves":        m.totalMoves.Load(),
        "messages_per_sec":   m.messagesPerSec,
    }
}