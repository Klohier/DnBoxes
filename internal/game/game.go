package game




type Game struct {

	GameId *int
	Player1 int
	Player2 int
	BoardSize int
	WinnerId *int
	CurrentTurn int
}


type Box struct {
	BoxId int
	GameId int
	TopEdge bool `json:"top_edge"`
	LeftEdge bool `json:"left_edge"`
	RightEdge bool `json:"right_edge"`
	BottomEdge bool `json:"bottom_edge"`
	Row int
	Col int
	Completed *bool `json:"completed"`
	Completed_by *int`json:"completed_by"` 
	
}

type GameState struct {
    Game  *Game  `json:"game"`
    Grids []Box  `json:"grids"`
}