package lobby

import "time"

type Lobby struct {
	LobbyID     string `json:"lobby_id"`
	HostID      int64 `json:"host_id"`
	Name        string `json:"name"`
	PlayerLimit int `json:"player_limit"`
	IsPrivate   bool `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	// SessionID   int64
	Players     []*LobbyPlayer
}

type LobbyPlayer struct {
	IsReady bool `json:"is_ready"`
	UserID  int64 `json:"userID"`
}
