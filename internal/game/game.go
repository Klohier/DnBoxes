package game

import "dango/internal/user"

type Game struct {

	gameId int
	player1 user.User
	player2 user.User
	boardSize int
	winnerId user.User
}


type Box struct {
	boxId int
	game Game
	topEdge bool
	leftEdge bool
	rightEdge bool
	bottomEdge bool
	row int
	col int
	completed bool
	completed_by user.User
}