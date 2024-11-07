package chat

import (
	// "dango/internal/user"
	"time"
)

type Message struct {
	UserID int `json:"userID"`
	Username string `json:"username"`
	GameID *int
	Message string `json:"message"`
	TimeStamp time.Time
}

