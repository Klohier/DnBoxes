package websocket

import (
	"context"
	"dango/internal/chat"
	"encoding/json"
	"fmt"
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
	// EventSendMessage is the event name for new chat messages sent
	EventSendMessage = "send_message"
	EventNewMessage = "new_message"
	EventSendInvite = "send_invite"
	EventRecieveInvite = "receive_invite"
	//
	EventAcceptInvite = "accept_invite"
	EventDeclineInvite = "decline_invite"
	EventQuitGame = "quit_game"
	EventGameCreated = "game_created"
	EventGetGrids =  "get_grids"
	EventNewGrids = "new_grids"
	EventMakeMove = "make_move"
	//
	
	//
	EventGetPlayers = "get_players"
	EventNewPlayers = "new_players"
)

type NewMessageStruct struct {
	UserID    int        `json:"userID"`
	Username  string     `json:"username"`
	GameID    *int       `json:"gameID"`
	Message   string     `json:"message"`
	TimeStamp time.Time  `json:"timeStamp"`
}

type QuitGameEvent struct {
    PlayerID int    `json:"playerID"` // The ID of the player quitting the game
    GameID   int    `json:"gameID"`   // The ID of the game being quit
}


// SendInviteEvent is the payload for the send_invite event
type SendInviteEvent struct {
	SenderID   int    `json:"senderID"`   // The ID of the player sending the invite
	SenderName string `json:"senderName"` // The username of the player sending the invite
	ReceiverID int    `json:"receiverID"` // The ID of the player receiving the invite
	ReceiverName string `json:"receiverName"` // The username of the player receiving the invite
	BoardSize   int       `json:"board_size"`
	Timestamp  time.Time `json:"timestamp"`  // Timestamp for the invite
}

// ReceiveInviteEvent is the payload for the receive_invite event
type ReceiveInviteEvent struct {
	PlayerID   int    `json:"playerID"`   // ID of the player receiving the invite
	InviterID  int    `json:"inviterID"`  // ID of the player sending the invite
	Username   string `json:"username"`   // Username of the player sending the invite
	BoardSize   int       `json:"board_size"`
	Timestamp  time.Time `json:"timestamp"`  // Timestamp of when the invite was sent
}

// AcceptInviteEvent is the payload for the accept_invite event
type AcceptInviteEvent struct {
	PlayerID  int    `json:"playerID"`  // The ID of the player accepting the invite
	SenderId int `json:"senderID"`
	BoardSize   int       `json:"board_size"`	
}

// DeclineInviteEvent is the payload for the decline_invite event
type DeclineInviteEvent struct {
	PlayerID  int    `json:"playerID"`  // The ID of the player declining the invite
	Declined  bool   `json:"declined"`  // Whether the invite was declined or not
}

// SendMessageEvent is the payload sent in the
// send_message event
type SendMessageEvent struct {
	UserID   int    `json:"userID"`   
    Username string `json:"username"`   
    GameID   *int   `json:"gameID"`    
    Message  string `json:"message"`    
    Timestamp time.Time `json:"timestamp"` 
}



// NewMessageEvent is returned when responding to send_message
type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

// SendBoxEvent is the payload sent in the send_box_update event
type SendBoxEvent struct {
	BoxID         int  `json:"boxId"`
	GameID        int  `json:"gameId"`
	TopEdge       bool `json:"topEdge"`
	LeftEdge      bool `json:"leftEdge"`
	RightEdge     bool `json:"rightEdge"`
	BottomEdge    bool `json:"bottomEdge"`
	Row           int  `json:"row"`
	Col           int  `json:"col"`
	Completed     *bool `json:"completed,omitempty"`
	CompletedBy   *int  `json:"completedBy,omitempty"`
}


// NewBoxEvent is the payload returned when responding to send_box_update
type NewBoxEvent struct {
	SendBoxEvent
	Updated time.Time `json:"updated"`
}



type Player struct {
	UserID   int    `json:"userID"`
	Username string `json:"username"`
}



// GetPlayersHandler will send a list of all currently connected players
func GetPlayersHandler(event Event, c *Connection) error {
	// Gather the list of all connected players
	var players []Player

	
	for client := range c.manager.connections {
		// Collect player info from the connected clients


		if client.gameID == nil {
		players = append(players, Player{
			UserID:   client.userID,   // Assuming the client has UserID field
			Username: client.username, // Assuming the client has Username field
		})
	}
	}

	// Marshal the player list into the response payload
	responsePayload, err := json.Marshal(players)
	if err != nil {
		return fmt.Errorf("failed to marshal players response: %v", err)
	}
	fmt.Println("Sending player list:", players)

	// Send the response to the requesting client
	responseEvent := Event{
		Type:    EventGetPlayers, // You can reuse this constant or use a new one like "players_response"
		Payload: responsePayload,
	}

	for client := range c.manager.connections {
		client.egress <- responseEvent 
	}
	// c.egress <- responseEvent

	return nil
}

// SendMessageHandler will send out a message to all other participants in the chat
func SendMessageHandler(event Event, c *Connection) error {
	// Marshal Payload into wanted format
	var chatevent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatevent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	// Prepare an Outgoing Message to others
	var broadMessage NewMessageEvent
	

	broadMessage.Sent = time.Now().UTC() // Set the Sent time to now
    broadMessage.Message = chatevent.Message
    broadMessage.Username = chatevent.Username
    broadMessage.UserID = chatevent.UserID
    broadMessage.GameID = chatevent.GameID
    broadMessage.Timestamp = chatevent.Timestamp

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}

	// Place payload into an Event
	var outgoingEvent Event
	
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage
	// Broadcast to all other Clients


	// Broadcast to all other clients based on gameID logic
	for client := range c.manager.connections {
		// Global chat: send to clients with gameID == nil
		if chatevent.GameID == nil && client.gameID == nil {
			client.egress <- outgoingEvent
		}

		 if chatevent.GameID != nil && client.gameID != nil && *client.gameID == *chatevent.GameID {
			// Game-specific chat: send to clients with the same gameID
			client.egress <- outgoingEvent
		}
	}

	

	msg := chat.Message{
		UserID:    chatevent.UserID,
		Username:  chatevent.Username,
		GameID:    chatevent.GameID,
		Message:   chatevent.Message,
		TimeStamp: chatevent.Timestamp,
	}

	

	ctx := context.Background()
	err = c.manager.chatService.SaveMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to save message: %v", err)
	}

	return nil

}

// GetGridsHandler will send a list of grids
func GetGridsHandler(event Event, c *Connection) error {
	// Parse the incoming payload
	type GetGridsPayload struct {
		GameID int `json:"gameID"`
	}

	var payload GetGridsPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("invalid payload for get_grids: %v", err)
	}

	// Use GameService to retrieve grids
	ctx := context.Background()
	grids, err := c.manager.gameService.GetGrids(ctx, payload.GameID)
	if err != nil {
		return fmt.Errorf("failed to get grids for game ID %d: %v", payload.GameID, err)
	}

	// Marshal the response payload
	responsePayload, err := json.Marshal(grids)
	if err != nil {
		return fmt.Errorf("failed to marshal grids response: %v", err)
	}

	// Send the response to the client
	responseEvent := Event{
		Type:    EventNewGrids,
		Payload: responsePayload,
	}

	//TODO Send to clients with same gameid
	c.egress <- responseEvent

	return nil
}

// SendInviteHandler will handle sending a game invite
func SendInviteHandler(event Event, c *Connection) error {
	// Unmarshal the payload into SendInviteEvent
	var inviteEvent SendInviteEvent
	if err := json.Unmarshal(event.Payload, &inviteEvent); err != nil {
		return fmt.Errorf("bad payload in send_invite request: %v", err)
	}

	if inviteEvent.BoardSize < 5 || inviteEvent.BoardSize > 10 {
		return fmt.Errorf("invalid board_size %d, it must be between 1 and 10", inviteEvent.BoardSize)
	}


	//OutGoing Message

	var receiveEvent ReceiveInviteEvent

	receiveEvent.InviterID = inviteEvent.SenderID
	receiveEvent.PlayerID = inviteEvent.ReceiverID
	receiveEvent.Username = inviteEvent.SenderName
	receiveEvent.BoardSize = inviteEvent.BoardSize
	receiveEvent.Timestamp = inviteEvent.Timestamp

	// Prepare the outgoing invite event for the receiver
	var outgoingEvent Event
	inviteEvent.Timestamp = time.Now().UTC()// Set the timestamp
	receiveEventData, err := json.Marshal(receiveEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal send_invite event: %v", err)
	}

	outgoingEvent.Payload = receiveEventData
	outgoingEvent.Type = EventRecieveInvite

	// Find the receiver's connection
	client :=findConnectionByUserID(c.manager, inviteEvent.ReceiverID)

	client.egress <- outgoingEvent


	return nil
}


// AcceptInviteHandler handles accepting a game invite and creates a game with both players.
func AcceptInviteHandler(event Event, c *Connection) error {
    // Parse the incoming payload
    var acceptInviteEvent AcceptInviteEvent
    if err := json.Unmarshal(event.Payload, &acceptInviteEvent); err != nil {
        return fmt.Errorf("bad payload in accept_invite request: %v", err)
    }

    // Validate the payload
    if acceptInviteEvent.PlayerID == 0 {
        return fmt.Errorf("playerID is required")
    }

    // Find the sender's connection
    inviterConnection := findConnectionByUserID(c.manager, acceptInviteEvent.SenderId)
    if inviterConnection == nil {
        return fmt.Errorf("inviter with userID %d not found", acceptInviteEvent.SenderId)
    }

    // Retrieve the sender and receiver information from the connection
    senderID := acceptInviteEvent.SenderId
    inviteeID := c.userID // Assuming the current connection represents the invitee
	boardSize := acceptInviteEvent.BoardSize

    // Create a new game with both players
    ctx := context.Background()
    game, err := c.manager.gameService.CreateGame(ctx, senderID, inviteeID, boardSize)
    if err != nil {
        return fmt.Errorf("failed to create game: %v", err)
    }

	_, err = c.manager.userService.UpdateGameID(ctx, senderID, game.GameId)
    if err != nil {
        return fmt.Errorf("failed to update sender gameID in database: %v", err)
    }

    _, err = c.manager.userService.UpdateGameID(ctx, inviteeID, game.GameId)
    if err != nil {
        return fmt.Errorf("failed to update invitee gameID in database: %v", err)
    }

	c.gameID = game.GameId
    inviterConnection.gameID = game.GameId

    // Notify both players of the accepted invite and game creation
    gameCreatedEvent := Event{
        Type: EventGameCreated,
    }

    gameCreatedPayload := struct {
        GameID    *int    `json:"gameID"`
        SenderID  int    `json:"senderID"`
        InviteeID int    `json:"inviteeID"`
        Timestamp string `json:"timestamp"`
    }{
        GameID:    game.GameId,
        SenderID:  senderID,
        InviteeID: inviteeID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }

    gameCreatedData, err := json.Marshal(gameCreatedPayload)
    if err != nil {
        return fmt.Errorf("failed to marshal game created event: %v", err)
    }

    gameCreatedEvent.Payload = gameCreatedData

    // Send the event to both the sender and the invitee
    c.egress <- gameCreatedEvent
    inviterConnection.egress <- gameCreatedEvent

	BroadcastPlayerList(c.manager)

    return nil
}

func MakeMoveHandler(event Event, c *Connection) error {
		// Parse the incoming payload
		type MakeMovePayload struct {
			GameID   int    `json:"gameID"`
			PlayerID int    `json:"playerID"`
			Row      int    `json:"row"`
			Col      int    `json:"col"`
			Edge     string `json:"edge"`
		}
	
		var payload MakeMovePayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			

			

			return fmt.Errorf("invalid payload for make_move: %v", err)
		}
	
		// Use GameService to retrieve grids
		ctx := context.Background()
		grids, err := c.manager.gameService.MakeMove(ctx, payload.GameID, payload.PlayerID, payload.Row, payload.Col, payload.Edge)
		if err != nil {

			invalidMoveEvent := Event{
				Type:    "invalid_move",
				
			}

			c.egress <- invalidMoveEvent

			return fmt.Errorf("failed to make moves for game ID %d: %v", payload.GameID, err)
		}
	
		// Marshal the response payload
		responsePayload, err := json.Marshal(grids)
		if err != nil {
			return fmt.Errorf("failed to marshal grids response: %v", err)
		}
	
		// Send the response to the client
		responseEvent := Event{
			Type:    EventNewGrids,
			Payload: responsePayload,
		}
	
		for client := range c.manager.connections {
			if client.gameID != nil && *client.gameID == payload.GameID {
				client.egress <- responseEvent
			}
		}

		for client := range c.manager.connections {
			if client.gameID != nil && *client.gameID == payload.GameID && client.userID != payload.PlayerID {
				yourTurnEvent := Event{
					Type:    "your_turn",
					
				}
				client.egress <- yourTurnEvent
			}
		}
	
		return nil
}

func QuitGameHandler(event Event, c *Connection) error {
    // Parse the incoming payload
    var quitGameEvent QuitGameEvent
    if err := json.Unmarshal(event.Payload, &quitGameEvent); err != nil {
        return fmt.Errorf("invalid payload for quit_game: %v", err)
    }


    // Ensure the quitting player matches the current connection
	quitEvent := Event{
        Type:    EventQuitGame,
        Payload: event.Payload, // Sending the same payload as received
    } 

    // Set the player's gameID to nil to indicate they have quit
    c.gameID = nil

	c.manager.userService.UpdateGameID(context.Background(), quitGameEvent.PlayerID, nil)

	for client := range c.manager.connections {
        if client.gameID != nil && *client.gameID == quitGameEvent.GameID && client.userID != quitGameEvent.PlayerID {
            client.egress <- quitEvent // Send to other players with the same gameID
			client.gameID = nil
			client.manager.userService.UpdateGameID(context.Background(), client.userID, nil)
			client.manager.gameService.SetWinner(context.Background(), quitGameEvent.GameID, client.userID)
        }
    }

    return nil
}
