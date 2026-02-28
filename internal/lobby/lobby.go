package lobby

import (
	"errors"
	"time"
)

var (
	ErrLobbyFull      = errors.New("lobby is full")
	ErrAlreadyInLobby = errors.New("user already in lobby")
	ErrNotInLobby     = errors.New("user not in lobby")
	ErrLobbyNotFound  = errors.New("lobby not found")
)

type DomainEvent interface{}

// PlayerLeftEvent is emitted when a player leaves the lobby.
type PlayerLeftEvent struct {
	UserID int64
}

// PlayerReadyEvent is emitted when a player sets ready state.
type PlayerReadyEvent struct {
	UserID int64
	Ready  bool
}

type Lobby struct {
	LobbyID     string        `json:"lobby_id"`
	HostID      int64         `json:"host_id"`
	BoardSize   int           `json:"board_size"`
	Name        string        `json:"name"`
	PlayerLimit int           `json:"player_limit"`
	IsPrivate   bool          `json:"is_private"`
	CreatedAt   time.Time     `json:"created_at"`
	Players     []LobbyPlayer `json:"players"`
	events      []DomainEvent
}

type LobbyPlayer struct {
	IsReady  bool   `json:"is_ready"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

func (l *Lobby) Events() []DomainEvent {
	return l.events
}

func (l *Lobby) ClearEvents() {
	l.events = nil
}

func (l *Lobby) AddPlayer(userID int64, username string) error {
	if len(l.Players) >= l.PlayerLimit {
		return ErrLobbyFull
	}
	for _, p := range l.Players {
		if p.UserID == userID {
			return ErrAlreadyInLobby
		}
	}
	l.Players = append(l.Players, LobbyPlayer{UserID: userID, IsReady: false, Username: username})
	l.events = append(l.events, PlayerJoinedEvent{LobbyID: l.LobbyID, UserID: userID}) // emit event

	return nil
}

func (l *Lobby) CanJoin(userID int64) error {
	if len(l.Players) >= l.PlayerLimit {
		return ErrLobbyFull
	}

	for _, p := range l.Players {
		if p.UserID == userID {
			return ErrAlreadyInLobby
		}
	}

	return nil
}

func (l *Lobby) RemovePlayer(userID int64) error {
	for i, p := range l.Players {
		if p.UserID == userID {
			l.Players = append(l.Players[:i], l.Players[i+1:]...)
			l.events = append(l.events, PlayerLeftEvent{UserID: userID})
			return nil
		}
	}
	return ErrNotInLobby
}

func (l *Lobby) SetReady(userID int64, ready bool) error {
	for i, p := range l.Players {
		if p.UserID == userID {
			l.Players[i].IsReady = ready
			l.events = append(l.events, PlayerReadyEvent{UserID: userID, Ready: ready}) // emit event

			return nil
		}
	}
	return ErrNotInLobby
}

func (l *Lobby) IsEmpty() bool {
	return len(l.Players) == 0
}
