package game

import "time"

type Player struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	TurnOrder   int    `json:"turn_order"`
	IsSpectator bool   `json:"is_spectator"`
	IsConnected bool   `json:"is_connected"`
	Score       int    `json:"score"`
}

type Game struct {
	GameId      *int      `json:"game_id"`
	SessionId   int       `json:"session_id"`
	GameName    *string   `json:"game_name"`
	Players     []Player  `json:"players"`
	BoardSize   int       `json:"board_size"`
	WinnerId    *int      `json:"winner"`
	CreatedAt   time.Time `json:"created_at"`
	CurrentTurn *int      `json:"current_turn"`
}

type Box struct {
	BoxId        int   `json:"box_id"`
	GameId       int   `json:"game_id"`
	TopEdge      bool  `json:"top_edge"`
	LeftEdge     bool  `json:"left_edge"`
	RightEdge    bool  `json:"right_edge"`
	BottomEdge   bool  `json:"bottom_edge"`
	Row          int   `json:"row"`
	Col          int   `json:"col"`
	Completed    *bool `json:"completed"`
	Completed_by *int  `json:"completed_by"`
}

type GameState struct {
	Game  *Game `json:"game"`
	Grids []Box `json:"grids"`
}
