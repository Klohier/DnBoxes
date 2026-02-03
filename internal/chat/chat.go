package chat

import (
	"log/slog"
	"time"
)

type Message struct {
	UserID    int       `json:"userID"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timestamp"`
}

type ChatHandler struct {
	chatService *ChatService
	logger      *slog.Logger
}
