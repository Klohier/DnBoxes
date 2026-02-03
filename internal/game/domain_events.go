package game

// Domain event type constants for the Game aggregate.
const (
	EventTypeGameCreated   = "game.created"
	EventTypeMoveApplied   = "game.move_applied"
	EventTypeBoxCompleted  = "game.box_completed"
	EventTypeTurnPassed    = "game.turn_passed"
	EventTypeGameEnded     = "game.ended"
	EventTypeGameForfeited = "game.forfeited"
)

// GameCreatedPayload is emitted when a new game is created.
type GameCreatedPayload struct {
	GameID    int      `json:"game_id"`
	BoardSize int      `json:"board_size"`
	Players   []Player `json:"players"`
}

// MoveAppliedPayload is emitted when a player places an edge.
type MoveAppliedPayload struct {
	TurnOrder int    `json:"turn_order"`
	Row       int    `json:"row"`
	Col       int    `json:"col"`
	Edge      string `json:"edge"`
}

// BoxCompletedPayload is emitted when a box is claimed by a player.
type BoxCompletedPayload struct {
	Row       int `json:"row"`
	Col       int `json:"col"`
	OwnerTurn int `json:"owner_turn"`
}

// TurnPassedPayload is emitted when the turn passes to another player.
type TurnPassedPayload struct {
	NextTurn int `json:"next_turn"`
}

// GameEndedPayload is emitted when the game ends naturally.
type GameEndedPayload struct {
	WinnerID *int `json:"winner_id"`
}

// GameForfeitedPayload is emitted when a player forfeits.
type GameForfeitedPayload struct {
	ForfeitedBy int  `json:"forfeited_by"`
	WinnerID    *int `json:"winner_id"`
}
