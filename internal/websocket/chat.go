package websocket

import (
	"context"
	"dango/internal/events"
	"encoding/json"
	"log/slog"
)

// SendMessageHandler will send out a message to all other participants in the chat
func MessageHandler(event Event, c *Connection, deps *HandlerDeps) error {

	var Message events.Message
	ctx := context.Background()

	if err := json.Unmarshal(event.Payload, &Message); err != nil {
		slog.Error("bad payload in request: ", "error", err)
	}

	channel := "lobby"

	err := deps.ChatService.SaveMessage(ctx, Message, channel)
	if err != nil {
		slog.Error("failed to save message: ", "error", err.Error())
	}

	// topic := fmt.Sprintf("chat:%d", Message.SessionID)
    c.manager.Broadcast(channel, EventMessage, Message)

	return nil

}
