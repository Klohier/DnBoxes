package game




type Game struct {

	GameId int
	Player1 int
	Player2 int
	BoardSize int
	WinnerId *int
	CurrentTurn int
}


type Box struct {
	BoxId int
	GameId int
	TopEdge bool
	LeftEdge bool
	RightEdge bool
	BottomEdge bool
	Row int
	Col int
	Completed *bool
	Completed_by *int
}