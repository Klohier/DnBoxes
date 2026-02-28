package lobby

type PlayerJoinedEvent struct {
	LobbyID string `json:"lobbyID"`
	UserID  int64  `json:"userID"`
	Time    int64  `json:"time"`
}

type LobbyEventPayload struct {
	LobbyID string `json:"lobbyID"`
	Event   string `json:"event"`
	UserID  *int64 `json:"user_id,omitempty"`
	Time    int64  `json:"time"`
}
