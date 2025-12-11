package websocket

type BroadcastEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
	Topic   string `json:"topic"`
}

type Player struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}
