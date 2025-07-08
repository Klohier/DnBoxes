package lobby

import "time"


type Lobby struct {
	LobbyID int64
	HostID int64
	Name string
	PlayerLimit int
	IsPrivate bool
	CreatedAt time.Time
	SessionID int64
	Players []*LobbyPlayer

}

type LobbyPlayer struct {
    LobbyID  int64
    IsReady  bool
    UserID   int64
}

