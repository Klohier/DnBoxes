package chat

import (
	// "dango/internal/user"
	"log/slog"
	"time"
)

//Message
type Message struct {
	UserID int `json:"userID"`
	Username string `json:"username"`
	GameID *int
	Message string `json:"message"`
	TimeStamp time.Time
}


type ChatHandler struct {
    chatService *ChatService
	logger *slog.Logger
}
