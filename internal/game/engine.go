package game

import (
	"fmt"
	"math/rand"
)
type Engine struct {
	Players     []Player
	BoardSize   int
	Grid        [][]Box
	CurrentTurn int
	Scores      map[int]int
	WinnerID    *int
}

type Move struct {
	UserID int
	Row    int
	Col    int
	Edge   string
}

type MoveResult struct {
	Move         Move
	BoxCompleted bool
	ClaimedBoxes []Box
	NewScores    map[int]int
	NextTurn     int
	WinnerID     *int
}

func NewEngine(gameState *GameState) *Engine {
	engine := &Engine{
		Players:     gameState.Game.Players,
		BoardSize:   gameState.Game.BoardSize,
		CurrentTurn: *gameState.Game.CurrentTurn,
		Scores:      make(map[int]int),
		WinnerID:    gameState.Game.WinnerId,
	}

	for _, p := range engine.Players {
		engine.Scores[p.UserID] = p.Score
	}

	// Initialize the 2D Grid slice
	engine.Grid = make([][]Box, engine.BoardSize)
	for i := range engine.Grid {
		engine.Grid[i] = make([]Box, engine.BoardSize)
	}

	// Convert 1D Grids slice into 2D Grid

	for _, box := range gameState.Grids {
		row := box.Row
		col := box.Col
		if row >= 0 && row < engine.BoardSize && col >= 0 && col < engine.BoardSize {
			engine.Grid[row][col] = box
		}
	}

	return engine
}

func (e *Engine) ApplyMove(move Move) (MoveResult, error) {

	var result MoveResult
	result.Move = move

	if e.CurrentTurn < 0 || e.CurrentTurn >= len(e.Players) {
		return result, fmt.Errorf("invalid current turn index: %d", e.CurrentTurn)
	}
	if e.Players[e.CurrentTurn].UserID != move.UserID {
		return result, fmt.Errorf("it's not player %d's turn", move.UserID)
	}

	if err := e.SetEdge(move.Row, move.Col, move.Edge); err != nil {
		return result, err
	}

	// Validate coordinates
	if move.Row < 0 || move.Row >= e.BoardSize || move.Col < 0 || move.Col >= e.BoardSize {
		return result, fmt.Errorf("move out of bounds")
	}

	// Check if this box or adjacent boxes are completed

	boxCompleted, claimedBoxes := e.CheckAdjacentBoxes(move)
	result.BoxCompleted = boxCompleted
	result.ClaimedBoxes = claimedBoxes

	newScores := make(map[int]int, len(e.Scores))
	for k, v := range e.Scores {
		newScores[k] = v
	}
	result.NewScores = newScores

	// Update turn if no box completed
	if !boxCompleted {
		e.CurrentTurn = (e.CurrentTurn + 1) % len(e.Players)
	}

	result.NextTurn = e.CurrentTurn

	if e.isGameOver() {
		winnerID := e.determineWinner()
		e.WinnerID = winnerID
		result.WinnerID = winnerID
	}

	return result, nil

}

func (e *Engine) SetEdge(row, col int, edge string) error {
	taken, err := e.edgeTaken(row, col, edge)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("%s already selected", edge)
	}

	box := &e.Grid[row][col]

	switch edge {
	case "top_edge":
		box.TopEdge = true
		if row > 0 {
			e.Grid[row-1][col].BottomEdge = true
		}
	case "right_edge":
		box.RightEdge = true
		if col < e.BoardSize-1 {
			e.Grid[row][col+1].LeftEdge = true
		}
	case "bottom_edge":
		box.BottomEdge = true
		if row < e.BoardSize-1 {
			e.Grid[row+1][col].TopEdge = true
		}
	case "left_edge":
		box.LeftEdge = true
		if col > 0 {
			e.Grid[row][col-1].RightEdge = true
		}

	}

	return nil
}

func (e *Engine) CheckAdjacentBoxes(move Move) (boxCompleted bool, claimedBoxes []Box) {
	check := func(row, col int) {
		completed, claim := e.CheckAndScoreBox(row, col, move.UserID)
		if completed && claim != nil {
			claimedBoxes = append(claimedBoxes, *claim)
			boxCompleted = true
		}
	}

	check(move.Row, move.Col)

	switch move.Edge {
	case "top_edge":
		check(move.Row-1, move.Col)
	case "right_edge":
		check(move.Row, move.Col+1)
	case "bottom_edge":
		check(move.Row+1, move.Col)
	case "left_edge":
		check(move.Row, move.Col-1)
	}

	return
}

func (e *Engine) CheckAndScoreBox(row, col int, userID int) (bool, *Box) {
	// Check boundaries
	if row < 0 || row >= e.BoardSize || col < 0 || col >= e.BoardSize {
		return false, nil
	}

	box := &e.Grid[row][col]

	// Check if box is already owned or not completed
	if box.Completed_by != nil || !box.TopEdge || !box.RightEdge || !box.BottomEdge || !box.LeftEdge {
		return false, nil
	}

	// Assign ownership and increment score
	box.Completed_by = &userID
	completed := true
	box.Completed = &completed
	e.Scores[userID]++

	return true, box
}

func (e *Engine) GenerateBotMove(botId int) *Move {
	completionMoves := []Move{}
	safeMoves := []Move{}
	riskyMoves := []Move{}

	// Categorize moves
	for row := 0; row < e.BoardSize; row++ {
		for col := 0; col < e.BoardSize; col++ {
			box := &e.Grid[row][col]
			if box.Completed_by != nil {
				continue
			}

			edges := []string{"top_edge", "right_edge", "bottom_edge", "left_edge"}
			for _, edge := range edges {
				if !e.isEdgeAvailable(row, col, edge) {
					continue
				}

				claimed := e.countClaimedEdges(box)
				move := Move{UserID: botId, Row: row, Col: col, Edge: edge}

				switch claimed {
				case 3:
					// Completing a box → always good
					completionMoves = append(completionMoves, move)
				case 0, 1:
					// Safe → doesn't hand over a box
					safeMoves = append(safeMoves, move)
				case 2:
					// Risky → sets up a chain
					riskyMoves = append(riskyMoves, move)
				}
			}
		}
	}

	// Priority order:
	// 1. Complete a box if possible
	if len(completionMoves) > 0 {
		return &completionMoves[rand.Intn(len(completionMoves))] // could randomize
	}

	// 2. Take a safe move
	if len(safeMoves) > 0 {
		return &safeMoves[rand.Intn(len(safeMoves))]
	}

	// 3. If nothing else, take a risky move 
	if len(riskyMoves) > 0 {
		return &riskyMoves[rand.Intn(len(riskyMoves))]
	}

	return nil
}

// Count edges claimed for a box
func (e *Engine) countClaimedEdges(box *Box) int {
	count := 0
	if box.TopEdge {
		count++
	}
	if box.RightEdge {
		count++
	}
	if box.BottomEdge {
		count++
	}
	if box.LeftEdge {
		count++
	}
	return count
}




func (e *Engine) isGameOver() bool {
	for row := 0; row < e.BoardSize; row++ {
		for col := 0; col < e.BoardSize; col++ {
			if e.Grid[row][col].Completed_by == nil {
				return false
			}
		}
	}
	return true
}

func (e *Engine) determineWinner() *int {
	var maxScore int
	var winnerID *int
	var tie bool

	for playerID, score := range e.Scores {
		if winnerID == nil || score > maxScore {
			maxScore = score
			id := playerID
			winnerID = &id
			tie = false
		} else if score == maxScore {
			tie = true
		}
	}

	if tie {
		return nil
	}

	return winnerID
}

func (e *Engine) edgeTaken(row, col int, edge string) (bool, error) {
	if row < 0 || row >= e.BoardSize || col < 0 || col >= e.BoardSize {
		return true, fmt.Errorf("coordinates out of bounds")
	}

	box := e.Grid[row][col]

	switch edge {
	case "top_edge":
		return box.TopEdge, nil
	case "right_edge":
		return box.RightEdge, nil
	case "bottom_edge":
		return box.BottomEdge, nil
	case "left_edge":
		return box.LeftEdge, nil
	default:
		return true, fmt.Errorf("invalid edge: %s", edge)
	}
}

func (e *Engine) isEdgeAvailable(row, col int, edge string) bool {
	taken, err := e.edgeTaken(row, col, edge)
	return err == nil && !taken
}
