package lobby

import (
	"errors"
	"time"
)

var (
    ErrLobbyFull       = errors.New("lobby is full")
    ErrAlreadyInLobby  = errors.New("user already in lobby")
    ErrNotInLobby      = errors.New("user not in lobby")
	ErrLobbyNotFound = errors.New("lobby not found")
)

type Lobby struct {
	LobbyID     string `json:"lobby_id"`
	HostID      int64 `json:"host_id"`
	Name        string `json:"name"`
	PlayerLimit int `json:"player_limit"`
	IsPrivate   bool `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	Players     []LobbyPlayer
}

type LobbyPlayer struct {
	IsReady bool `json:"is_ready"`
	UserID  int64 `json:"userID"`
}

func (l *Lobby) AddPlayer(userID int64) error {
    if len(l.Players) >= l.PlayerLimit {
        return ErrLobbyFull
    }
    l.Players = append(l.Players, LobbyPlayer{UserID: userID, IsReady: false})
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
            return nil
        }
    }
    return ErrNotInLobby
}

func (l *Lobby) SetReady(userID int64, ready bool) error {
    for i, p := range l.Players {
        if p.UserID == userID {
            l.Players[i].IsReady = ready
            return nil
        }
    }
    return ErrNotInLobby
}

func (l *Lobby) IsEmpty() bool {
    return len(l.Players) == 0
}