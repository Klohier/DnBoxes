package websocket

import (
	"context"
	"dango/internal/chat"
	"dango/internal/game"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Event is the Messages sent over the websocket
// Used to differ between different actions
type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventHandler func(event Event, c *Connection) error

const (
	EventMessage       = "chat:new"
	EventSendInvite    = "invite:new"
	EventAcceptInvite  = "invite:accept"
	EventDeclineInvite = "invite:decline"
	EventQuitGame      = "game:quit"
	EventGameCreated   = "game:new"
	EventGameState     = "game:state"
	EventMakeMove      = "game:move"
	EventGetPlayers    = "player:get"
	EventNewPlayers    = "player:new"
)

type Message struct {
	UserID    int       `json:"userID"`
	Username  string    `json:"username"`
	SessionID int       `json:"session_id"`
	Message   string    `json:"message"`
	TimeStamp time.Time `json:"timeStamp"`
}
type QuitGameEvent struct {
	PlayerID  int `json:"playerID"` // The ID of the player quitting the game
	SessionID int `json:"session_id"`
	GameID    int `json:"gameID"` // The ID of the game being quit
}
type GetGridsPayload struct {
	GameID int `json:"gameID"`
}
type InviteEvent struct {
	SenderID     int       `json:"senderID"`     // The ID of the player sending the invite
	SenderName   string    `json:"senderName"`   // The username of the player sending the invite
	ReceiverID   int       `json:"receiverID"`   // The ID of the player receiving the invite
	ReceiverName string    `json:"receiverName"` // The username of the player receiving the invite
	BoardSize    int       `json:"board_size"`
	Timestamp    time.Time `json:"timestamp"`
}

type AcceptInviteEvent struct {
	PlayerID  int `json:"playerID"` // The ID of the player accepting the invite
	SenderId  int `json:"senderID"` //ID of who sent Invite
	BoardSize int `json:"board_size"`
}

type DeclineInviteEvent struct {
	PlayerID int `json:"playerID"` // The ID of the player declining the invite

}

// NewMessageEvent is returned when responding to send_message
type MessageEvent struct {
	UserID    int       `json:"userID"`
	Username  string    `json:"username"`
	SessionID int       `json:"session_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type GameCreatedPayload struct {
	GameID    *int `json:"gameID"`
	SenderID  int  `json:"senderID"`
	InviteeID int  `json:"inviteeID"`
}

// SendBoxEvent is the payload sent in the new_grids event
type SendBoxEvent struct {
	BoxID       int   `json:"boxId"`
	GameID      int   `json:"gameId"`
	TopEdge     bool  `json:"topEdge"`
	LeftEdge    bool  `json:"leftEdge"`
	RightEdge   bool  `json:"rightEdge"`
	BottomEdge  bool  `json:"bottomEdge"`
	Row         int   `json:"row"`
	Col         int   `json:"col"`
	Completed   *bool `json:"completed,omitempty"`
	CompletedBy *int  `json:"completedBy,omitempty"`
}

// NewBoxEvent is the payload returned when responding to new_rids
type NewBoxEvent struct {
	SendBoxEvent
	Updated time.Time `json:"updated"`
}

type Player struct {
	UserID   int    `json:"userID"`
	Username string `json:"username"`
}

// GetPlayersHandler will send a list of all currently connected players
func PlayerHandler(event Event, c *Connection) error {
	var players []Player

	sessionID := c.sessionID

	c.manager.RLock()
	conns, ok := c.manager.rooms[sessionID]
	c.manager.RUnlock()
	if !ok {
		return fmt.Errorf("session room %d does not exist", sessionID)
	}

	// Collect players in this session room
	for client := range conns {
		players = append(players, Player{
			UserID:   client.userID,
			Username: client.username,
		})
	}

	responsePayload, err := json.Marshal(players)
	if err != nil {
		return errors.New("failed to marshal players response: " + err.Error())
	}
	slog.Info("Sending player list:", slog.Any("players", players))

	responseEvent := Event{
		Type:    EventGetPlayers,
		Payload: responsePayload,
	}

	for client := range conns {
		client.egress <- responseEvent
	}

	return nil
}

// SendMessageHandler will send out a message to all other participants in the chat
func MessageHandler(event Event, c *Connection) error {
	var chatevent MessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Prepare an Outgoing Message to others
	var Message MessageEvent

	Message.Message = chatevent.Message
	Message.Username = chatevent.Username
	Message.UserID = chatevent.UserID
	Message.SessionID = chatevent.SessionID
	Message.Timestamp = chatevent.Timestamp

	data, err := json.Marshal(Message)
	if err != nil {
		return errors.New("failed to marshal broadcast message: " + err.Error())
	}

	var outgoingEvent Event

	outgoingEvent.Payload = data
	outgoingEvent.Type = EventMessage

	c.manager.RLock()
	defer c.manager.RUnlock()

	conns, ok := c.manager.rooms[chatevent.SessionID]
	if !ok {
		return fmt.Errorf("session room %d does not exist", chatevent.SessionID)
	}

	for conn := range conns {
		conn.egress <- outgoingEvent
	}

	msg := chat.Message{
		UserID:    chatevent.UserID,
		Username:  chatevent.Username,
		SessionID: chatevent.SessionID,
		Message:   chatevent.Message,
		TimeStamp: chatevent.Timestamp,
	}

	ctx := context.Background()
	err = c.manager.chatService.SaveMessage(ctx, msg)
	if err != nil {
		return errors.New("failed to save message: " + err.Error())
	}

	return nil

}

func GameStateHandler(event Event, c *Connection) error {
	var payload struct {
		GameID int `json:"gameID"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("invalid payload for game_state: %v", err)
	}

	ctx := context.Background()

	gameState, err := c.manager.gameService.GetGameState(ctx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to get game state for gameID %d: %v", payload.GameID, err)
	}

	responsePayload, err := json.Marshal(gameState)
	if err != nil {
		return fmt.Errorf("failed to marshal game state response: %v", err)
	}

	responseEvent := Event{
		Type:    EventGameState,
		Payload: responsePayload,
	}

	c.egress <- responseEvent

	if err := broadcastWinnerEvent(c, gameState, payload.GameID); err != nil {
	slog.Error("failed to broadcast winner", "error", err)
}



	return nil
}

// SendInviteHandler will handle sending a game invite
func InviteHandler(event Event, c *Connection) error {
	var inviteEvent InviteEvent
	if err := json.Unmarshal(event.Payload, &inviteEvent); err != nil {
		return errors.New("bad payload in send_invite request: " + err.Error())
	}

	if inviteEvent.BoardSize < 5 || inviteEvent.BoardSize > 10 {
		return fmt.Errorf("invalid board_size %d, it must be between 5 and 10", inviteEvent.BoardSize)
	}

	var outgoingEvent Event
	inviteEvent.Timestamp = time.Now().UTC()
	outgoingPayload, err := json.Marshal(inviteEvent)
	if err != nil {
		return errors.New("failed to marshal invite event:" + err.Error())
	}

	outgoingEvent.Payload = outgoingPayload
	outgoingEvent.Type = EventSendInvite

	// Find the receiver's connection
	client := findConnectionByUserID(c.manager, inviteEvent.ReceiverID)

	client.egress <- outgoingEvent

	return nil
}

// AcceptInviteHandler handles accepting a game invite and creates a game with both players.
func AcceptInviteHandler(event Event, c *Connection) error {
	var acceptInviteEvent AcceptInviteEvent
	if err := json.Unmarshal(event.Payload, &acceptInviteEvent); err != nil {
		return errors.New("bad payload in accept_invite request: " + err.Error())
	}

	// Find the sender's connection
	inviterConnection := findConnectionByUserID(c.manager, acceptInviteEvent.SenderId)
	if inviterConnection == nil {
		return fmt.Errorf("inviter with userID %d not found", acceptInviteEvent.SenderId)
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
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create a new game with both players
	game, err := c.manager.gameService.CreateGame(ctx, playerIDs, boardSize, session.SessionID)
	if err != nil {
		return errors.New("failed to create game: " + err.Error())
	}

	c.manager.JoinRoom(c, session.SessionID)

	inviterConnection.manager.JoinRoom(inviterConnection, session.SessionID)

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
		return errors.New("failed to marshal game created event: " + err.Error())
	}

	gameCreatedEvent.Payload = gameCreatedData

	// Send the event to both the sender and the invitee
	c.egress <- gameCreatedEvent
	inviterConnection.egress <- gameCreatedEvent

	c.manager.BroadcastPlayerListToRoom(session.SessionID)

	return nil
}

func MoveHandler(event Event, c *Connection) error {
	type MakeMovePayload struct {
		GameID   int    `json:"gameID"`
		PlayerID int    `json:"playerID"`
		Row      int    `json:"row"`
		Col      int    `json:"col"`
		Edge     string `json:"edge"`
	}

	var payload MakeMovePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return errors.New("invalid payload for make_move: " + err.Error())
	}

	ctx := context.Background()
	gameState, err := c.manager.gameService.MakeMove(ctx, payload.GameID, payload.PlayerID, payload.Row, payload.Col, payload.Edge)
	if err != nil {

		invalidMoveEvent := Event{
			Type: "invalid_move",
		}

		c.egress <- invalidMoveEvent

		return fmt.Errorf("failed to make moves for game ID %d: %v", payload.GameID, err)
	}

	// Marshal the response payload
	responsePayload, err := json.Marshal(gameState)
	if err != nil {
		return errors.New("failed to marshal grids response: " + err.Error())
	}

	// Send the response to the client
	responseEvent := Event{
		Type:    EventGameState,
		Payload: responsePayload,
	}

	for client := range c.manager.connections {
		if client.sessionID == c.sessionID {
			client.egress <- responseEvent
		}
	}
	if err := broadcastWinnerEvent(c, gameState, payload.GameID); err != nil {
	slog.Error("failed to marshal winner_set payload", "error", err)
}
if gameState.Game.WinnerId != nil {
	// Broadcast winner


	// Move all players back to the main lobby
	if err := c.manager.movePlayersToMainLobby(c.sessionID); err != nil {
		slog.Error("failed to move players to main lobby", "error", err)
	}
}


	return nil
}

func QuitGameHandler(event Event, c *Connection) error {
	var quitGameEvent QuitGameEvent
	if err := json.Unmarshal(event.Payload, &quitGameEvent); err != nil {
		return errors.New("invalid payload for quit_game: " + err.Error())
	}

	quitEvent := Event{
		Type: EventQuitGame,
	}

	mainLobbyID := 1

	if err := c.manager.sessionService.AddUserToSession(context.Background(), mainLobbyID, quitGameEvent.PlayerID); err != nil {
		return err
	}

	c.sessionID = mainLobbyID

	winnerSet := false

	// Notify other players in the same session (the old game session)
	for client := range c.manager.connections {
		if client.sessionID == quitGameEvent.SessionID {
			if client.userID != quitGameEvent.PlayerID {
				// Send quit event to other players
				client.egress <- quitEvent

				// Move them to main lobby
				if err := client.manager.sessionService.AddUserToSession(context.Background(), mainLobbyID, client.userID); err != nil {
					return err
				}
				client.sessionID = mainLobbyID

				// Set first remaining player as winner
				if !winnerSet {
					if err := client.manager.gameService.SetWinner(context.Background(), quitGameEvent.GameID, &client.userID); err != nil {
						return err
					}
					winnerSet = true
				}
			}
		}
	}

	return nil
}    
func  broadcastWinnerEvent(c *Connection, gameState *game.GameState, gameID int) error {
	if gameState.Game.WinnerId == nil {
		return nil
	}

	playerMap := make(map[int]string, len(gameState.Game.Players))
	for _, player := range gameState.Game.Players {
		playerMap[player.UserID] = player.Username
	}

	winnerPayload, err := json.Marshal(map[string]interface{}{
		"gameId":         gameID,
		"winnerId":       *gameState.Game.WinnerId,
		"winnerUsername": playerMap[*gameState.Game.WinnerId],
	})
	if err != nil {
		return fmt.Errorf("failed to marshal winner_set payload: %v", err)
	}

	winnerEvent := Event{
		Type:    "winner_set",
		Payload: winnerPayload,
	}

	for client := range c.manager.connections {
		if client.sessionID == c.sessionID {
			client.egress <- winnerEvent
		}
	}

	return nil
}