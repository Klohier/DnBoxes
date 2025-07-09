package session

import (
	"fmt"
	"time"
)

type ConnectionStatus string

const (
	Connected    ConnectionStatus = "connected"
	Disconnected ConnectionStatus = "disconnected"
	Quit         ConnectionStatus = "quit"
)

func parseConnectionStatus(s string) (ConnectionStatus, error) {
	switch s {
	case string(Connected):
		return Connected, nil
	case string(Disconnected):
		return Disconnected, nil
	case string(Quit):
		return Quit, nil
	default:
		return "", fmt.Errorf("invalid connection status: %s", s)
	}
}

type Session struct {
	SessionID int       `json:"session_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UserCount int       `json:"user_count"`
}

type SessionUser struct {
	UserID           int              `json:"user_id"`
	Username         string           `json:"username"`
	ConnectionStatus ConnectionStatus `json:"connection_status"`
	JoinedAt         time.Time        `json:"joined_at"`
}

type FullSession struct {
	Session
	Users []SessionUser `json:"users"`
}
