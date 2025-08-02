package chat

import (
	// "dango/internal/user"
	"log/slog"
	"time"
)

// Message
type Message struct {
	UserID    int       `json:"userID"`
	Username  string    `json:"username"`
	SessionID int       `json:"session_id"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`
}

type ChatHandler struct {
	chatService *ChatService
	logger      *slog.Logger
}
