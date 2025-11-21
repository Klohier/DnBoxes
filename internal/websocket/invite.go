package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// SendInviteHandler will handle sending a game invite
func InviteHandler(event Event, c *Connection) error {
	var inviteEvent InviteEvent
	if err := json.Unmarshal(event.Payload, &inviteEvent); err != nil {
		slog.Error("bad payload in send_invite request: ", "error", err)
	}

	if inviteEvent.BoardSize < 5 || inviteEvent.BoardSize > 10 {
		slog.Error("invalid board size: must be between 5 and 10", "board_size", inviteEvent.BoardSize)
	}

	var outgoingEvent Event
	inviteEvent.Timestamp = time.Now().UTC()
	outgoingPayload, err := json.Marshal(inviteEvent)
	if err != nil {
		slog.Error("failed to marshal invite event:", "error", err)
	}

	outgoingEvent.Payload = outgoingPayload
	outgoingEvent.Type = EventSendInvite

	// Find the receiver's connection
	client := c.manager.GetConnectionByUserID(inviteEvent.ReceiverID)
	if client == nil {
		slog.Error("could not find connection for receiver ID", "reciever", inviteEvent.ReceiverID)
		return nil
	}

	slog.Info("Sending invite",
		"sender_id", inviteEvent.SenderID,
		"receiver_id", inviteEvent.ReceiverID,
	)

client.Send(outgoingEvent)
	return nil
}

// AcceptInviteHandler handles accepting a game invite and creates a game with both players.
func AcceptInviteHandler(event Event, c *Connection, deps *HandlerDeps) error {
	var acceptInviteEvent AcceptInviteEvent
	if err := json.Unmarshal(event.Payload, &acceptInviteEvent); err != nil {
		slog.Error("bad payload in accept_invite request:", "error", err)
		return nil
	}

	// Find the sender's connection
	inviterConnection := c.manager.GetConnectionByUserID(acceptInviteEvent.SenderId)
	if inviterConnection == nil {
		slog.Error("inviter with userID not found", "error", acceptInviteEvent.SenderId)
		return nil
	}
	if c == nil {
		slog.Error("current connection is nil in AcceptInviteHandler")
		return nil
	}

	// Retrieve the sender and receiver information from the connection
	senderID := acceptInviteEvent.SenderId
	inviteeID := c.userID
	boardSize := acceptInviteEvent.BoardSize

	playerIDs := []int{senderID, inviteeID}

	// Session creation
	ctx := context.Background()
	session, err := c.manager.sessionService.CreateSession(ctx)
	if err != nil {
		slog.Error("failed to create session:", "error", err)
	}

	// Create a new game with both players
	game, err := deps.GameService.CreateGame(ctx, playerIDs, boardSize, session.SessionID)
	if err != nil {
		slog.Error("failed to create game: ", "error", err)
	}

	gameTopic := fmt.Sprintf("game:%d", game.SessionId)
	c.manager.Subscribe(gameTopic, c)
	c.manager.Subscribe(gameTopic, inviterConnection)


	slog.Info("Subscribed players to game topic",
    "topic", gameTopic,
    "subscribed_connections", len(c.manager.rooms[gameTopic]),
)

	// Notify both players of the accepted invite and game creation
	gameCreatedEvent := Event{
		Type: EventGameCreated,
	}

	var payload GameCreatedPayload

	payload.GameID = game.GameId
	payload.SenderID = senderID
	payload.InviteeID = inviteeID

	gameCreatedData, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal game created event: ", "error", err)
	}

	gameCreatedEvent.Payload = gameCreatedData

	// Send the event to both the sender and the invitee
c.Send(gameCreatedEvent)	
inviterConnection.Send(gameCreatedEvent)
	return nil
}