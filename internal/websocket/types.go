package websocket

import (
	"dango/internal/chat"
	"dango/internal/game"
)

type HandlerDeps struct {
	ChatService *chat.ChatService
	GameService *game.GameService

}

