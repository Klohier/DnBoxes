package game

import (
	"fmt"
	"math/rand"
	"time"
)

// Player represents a player in the game
type Player struct {
	UserID      *int   `json:"user_id"`     
	Username    string `json:"username"`
	TurnOrder   int    `json:"turn_order"`
	IsAnonymous bool   `json:"is_anonymous"`
	Score       int    `json:"score"`
}

// Box represents a single cell in the grid
type Box struct {
	Row        int   `json:"row"`
	Col        int   `json:"col"`
	TopEdge    bool  `json:"top_edge"`
	RightEdge  bool  `json:"right_edge"`
	BottomEdge bool  `json:"bottom_edge"`
	LeftEdge   bool  `json:"left_edge"`
	OwnerTurn  *int  `json:"owner_turn"` // turn_order of owner,
}

// Game represents the complete game state
type Game struct {
	GameID      *int       `json:"game_id"`
	GameName    *string    `json:"game_name"`
	BoardSize   int        `json:"board_size"`
	CurrentTurn int        `json:"current_turn"` // turn_order, 
	WinnerID    *int       `json:"winner_id"`    // user_id of winner
	CreatedAt   time.Time  `json:"created_at"`
	EndedAt     *time.Time `json:"ended_at"`
	Players     []Player   `json:"players"`
	Grid        [][]Box    `json:"grid"`
}

// Move represents a player's move
type Move struct {
	TurnOrder int    // Which player (by turn order)
	Row       int
	Col       int
	Edge      EdgeType
}

// EdgeType represents the four possible edges
type EdgeType string

const (
	TopEdge    EdgeType = "top"
	RightEdge  EdgeType = "right"
	BottomEdge EdgeType = "bottom"
	LeftEdge   EdgeType = "left"
)

// MoveResult contains the outcome of applying a move
type MoveResult struct {
	CompletedBoxes []Box
	NextTurn       int
	GameOver       bool
	WinnerID       *int
}

// NewGame creates a new game with the given parameters
func NewGame(gameID *int, boardSize int, players []Player) *Game {
	game := &Game{
		GameID:      gameID,
		BoardSize:   boardSize,
		CurrentTurn: 0,
		Players:     players,
		CreatedAt:   time.Now(),
		Grid:        makeEmptyGrid(boardSize),
	}
	return game
}

// ApplyMove applies a player's move and returns the result
func (g *Game) ApplyMove(move Move) (MoveResult, error) {
	var result MoveResult
	
	// Validate move
	if err := g.validateMove(move); err != nil {
		return result, err
	}
	
	// Set the edge
	g.setEdge(move.Row, move.Col, move.Edge)
	
	// Check for completed boxes
	completedBoxes := g.checkCompletedBoxes(move)
	result.CompletedBoxes = completedBoxes
	
	// Update score if boxes were completed
	if len(completedBoxes) > 0 {
		g.Players[move.TurnOrder].Score += len(completedBoxes)
	}
	
	// Determine next turn (same player if they completed a box)
	if len(completedBoxes) == 0 {
		g.CurrentTurn = (g.CurrentTurn + 1) % len(g.Players)
	}
	result.NextTurn = g.CurrentTurn
	
	// Check if game is over
	if g.IsGameOver() {
		result.GameOver = true
		result.WinnerID = g.determineWinner()
		g.WinnerID = result.WinnerID
		now := time.Now()
		g.EndedAt = &now
	}
	
	return result, nil
}

// GenerateBotMove generates a strategic move for a bot
func (g *Game) GenerateBotMove(turnOrder int) *Move {
	var completionMoves, safeMoves, riskyMoves []Move
	
	for row := 0; row < g.BoardSize; row++ {
		for col := 0; col < g.BoardSize; col++ {
			if g.Grid[row][col].OwnerTurn != nil {
				continue
			}
			
			for _, edge := range []EdgeType{TopEdge, RightEdge, BottomEdge, LeftEdge} {
				if !g.isEdgeAvailable(row, col, edge) {
					continue
				}
				
				move := Move{TurnOrder: turnOrder, Row: row, Col: col, Edge: edge}
				edgeCount := g.countEdges(&g.Grid[row][col])
				
				switch edgeCount {
				case 3:
					completionMoves = append(completionMoves, move)
				case 0, 1:
					safeMoves = append(safeMoves, move)
				case 2:
					riskyMoves = append(riskyMoves, move)
				}
			}
		}
	}
	
	// Prioritize: completion > safe > risky
	if len(completionMoves) > 0 {
		return &completionMoves[rand.Intn(len(completionMoves))]
	}
	if len(safeMoves) > 0 {
		return &safeMoves[rand.Intn(len(safeMoves))]
	}
	if len(riskyMoves) > 0 {
		return &riskyMoves[rand.Intn(len(riskyMoves))]
	}
	
	return nil
}

// IsGameOver checks if all boxes are claimed
func (g *Game) IsGameOver() bool {
	for row := 0; row < g.BoardSize; row++ {
		for col := 0; col < g.BoardSize; col++ {
			if g.Grid[row][col].OwnerTurn == nil {
				return false
			}
		}
	}
	return true
}

// GetCurrentPlayer returns the player whose turn it is
func (g *Game) GetCurrentPlayer() *Player {
	if g.CurrentTurn < 0 || g.CurrentTurn >= len(g.Players) {
		return nil
	}
	return &g.Players[g.CurrentTurn]
}

// Private helper methods

func (g *Game) validateMove(move Move) error {
	// Validate turn
	if move.TurnOrder != g.CurrentTurn {
		return fmt.Errorf("not player %d's turn (current turn: %d)", move.TurnOrder, g.CurrentTurn)
	}
	
	// Validate coordinates
	if move.Row < 0 || move.Row >= g.BoardSize || move.Col < 0 || move.Col >= g.BoardSize {
		return fmt.Errorf("coordinates out of bounds: (%d, %d)", move.Row, move.Col)
	}
	
	// Validate edge not already taken
	if !g.isEdgeAvailable(move.Row, move.Col, move.Edge) {
		return fmt.Errorf("edge %s at (%d, %d) already taken", move.Edge, move.Row, move.Col)
	}
	
	// Validate game not over
	if g.EndedAt != nil {
		return fmt.Errorf("game has already ended")
	}
	
	return nil
}

func (g *Game) setEdge(row, col int, edge EdgeType) {
	box := &g.Grid[row][col]
	
	switch edge {
	case TopEdge:
		box.TopEdge = true
		if row > 0 {
			g.Grid[row-1][col].BottomEdge = true
		}
	case RightEdge:
		box.RightEdge = true
		if col < g.BoardSize-1 {
			g.Grid[row][col+1].LeftEdge = true
		}
	case BottomEdge:
		box.BottomEdge = true
		if row < g.BoardSize-1 {
			g.Grid[row+1][col].TopEdge = true
		}
	case LeftEdge:
		box.LeftEdge = true
		if col > 0 {
			g.Grid[row][col-1].RightEdge = true
		}
	}
}

func (g *Game) checkCompletedBoxes(move Move) []Box {
	var completed []Box
	
	// Check the box where the move was made
	if box := g.tryCompleteBox(move.Row, move.Col, move.TurnOrder); box != nil {
		completed = append(completed, *box)
	}
	
	// Check adjacent box if applicable
	adjRow, adjCol := g.getAdjacentBox(move.Row, move.Col, move.Edge)
	if adjRow >= 0 {
		if box := g.tryCompleteBox(adjRow, adjCol, move.TurnOrder); box != nil {
			completed = append(completed, *box)
		}
	}
	
	return completed
}

func (g *Game) tryCompleteBox(row, col, turnOrder int) *Box {
	if row < 0 || row >= g.BoardSize || col < 0 || col >= g.BoardSize {
		return nil
	}
	
	box := &g.Grid[row][col]
	
	// Already owned or not complete
	if box.OwnerTurn != nil || !g.isBoxComplete(box) {
		return nil
	}
	
	// Claim the box
	box.OwnerTurn = &turnOrder
	return box
}

func (g *Game) getAdjacentBox(row, col int, edge EdgeType) (int, int) {
	switch edge {
	case TopEdge:
		return row - 1, col
	case RightEdge:
		return row, col + 1
	case BottomEdge:
		return row + 1, col
	case LeftEdge:
		return row, col - 1
	}
	return -1, -1
}

func (g *Game) isEdgeAvailable(row, col int, edge EdgeType) bool {
	if row < 0 || row >= g.BoardSize || col < 0 || col >= g.BoardSize {
		return false
	}
	
	box := &g.Grid[row][col]
	
	switch edge {
	case TopEdge:
		return !box.TopEdge
	case RightEdge:
		return !box.RightEdge
	case BottomEdge:
		return !box.BottomEdge
	case LeftEdge:
		return !box.LeftEdge
	}
	
	return false
}

func (g *Game) isBoxComplete(box *Box) bool {
	return box.TopEdge && box.RightEdge && box.BottomEdge && box.LeftEdge
}

func (g *Game) countEdges(box *Box) int {
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

func (g *Game) determineWinner() *int {
	if len(g.Players) == 0 {
		return nil
	}
	
	maxScore := g.Players[0].Score
	winnerIdx := 0
	tie := false
	
	for i := 1; i < len(g.Players); i++ {
		if g.Players[i].Score > maxScore {
			maxScore = g.Players[i].Score
			winnerIdx = i
			tie = false
		} else if g.Players[i].Score == maxScore {
			tie = true
		}
	}
	
	if tie {
		return nil
	}
	
	return g.Players[winnerIdx].UserID
}

func makeEmptyGrid(size int) [][]Box {
	grid := make([][]Box, size)
	for i := range grid {
		grid[i] = make([]Box, size)
		for j := range grid[i] {
			grid[i][j] = Box{Row: i, Col: j}
		}
	}
	return grid
}