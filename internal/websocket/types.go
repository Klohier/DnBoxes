package websocket

import (
	"dango/internal/chat"
	"dango/internal/game"
	"encoding/json"
	"time"
)

type HandlerDeps struct {
	ChatService *chat.ChatService
	GameService *game.GameService
}

type BroadcastEvent struct {
    Topic string
    Event Event
}

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
	EventJoinPage      = "page:join"
	EventLeavePage     = "page:leave"
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
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}
