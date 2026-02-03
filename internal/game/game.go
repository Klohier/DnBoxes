package game

import (
	"dango/internal/events"
	"fmt"
	"math/rand"
	"strconv"
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
	Row        int  `json:"row"`
	Col        int  `json:"col"`
	TopEdge    bool `json:"top_edge"`
	RightEdge  bool `json:"right_edge"`
	BottomEdge bool `json:"bottom_edge"`
	LeftEdge   bool `json:"left_edge"`
	OwnerTurn  *int `json:"owner_turn"` // turn_order of owner
}

// Game is the aggregate root for a Dots and Boxes game.
// State is projected from domain events via applyEvent.
type Game struct {
	events.Aggregate

	GameID      *int       `json:"game_id"`
	GameName    *string    `json:"game_name"`
	BoardSize   int        `json:"board_size"`
	CurrentTurn int        `json:"current_turn"` // turn_order
	WinnerID    *int       `json:"winner_id"`    // user_id of winner
	CreatedAt   time.Time  `json:"created_at"`
	EndedAt     *time.Time `json:"ended_at"`
	Players     []Player   `json:"players"`
	Grid        [][]Box    `json:"grid"`
}

// Move represents a player's move
type Move struct {
	TurnOrder int // Which player (by turn order)
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

// NewGame creates a new game by raising a GameCreated domain event.
func NewGame(gameID *int, boardSize int, players []Player) *Game {
	game := &Game{}
	id := 0
	if gameID != nil {
		id = *gameID
	}
	game.raise(EventTypeGameCreated, GameCreatedPayload{
		GameID:    id,
		BoardSize: boardSize,
		Players:   players,
	})
	return game
}

// LoadFromEvents reconstructs a Game aggregate by replaying historical events.
func LoadFromEvents(domainEvents []events.DomainEvent) *Game {
	game := &Game{}
	for _, e := range domainEvents {
		game.applyEvent(e)
	}
	return game
}

// ApplyMove validates and applies a player's move, raising domain events.
func (g *Game) ApplyMove(move Move) (MoveResult, error) {
	var result MoveResult

	if err := g.validateMove(move); err != nil {
		return result, err
	}

	// Raise edge placement event
	g.raise(EventTypeMoveApplied, MoveAppliedPayload{
		TurnOrder: move.TurnOrder,
		Row:       move.Row,
		Col:       move.Col,
		Edge:      string(move.Edge),
	})

	// Check for completed boxes after the edge was applied
	completedBoxes := g.findNewlyCompletedBoxes(move)
	result.CompletedBoxes = completedBoxes

	for _, box := range completedBoxes {
		g.raise(EventTypeBoxCompleted, BoxCompletedPayload{
			Row:       box.Row,
			Col:       box.Col,
			OwnerTurn: move.TurnOrder,
		})
	}

	// Determine next turn
	if len(completedBoxes) == 0 {
		nextTurn := (g.CurrentTurn + 1) % len(g.Players)
		g.raise(EventTypeTurnPassed, TurnPassedPayload{
			NextTurn: nextTurn,
		})
	}
	result.NextTurn = g.CurrentTurn

	// Check if game is over
	if g.isGameOver() {
		winnerID := g.determineWinner()
		g.raise(EventTypeGameEnded, GameEndedPayload{
			WinnerID: winnerID,
		})
		result.GameOver = true
		result.WinnerID = winnerID
	}

	return result, nil
}

// Forfeit ends the game with a forfeit, raising a GameForfeited event.
func (g *Game) Forfeit(playerID int) {
	var winnerID *int
	maxScore := -1
	for _, p := range g.Players {
		if p.UserID != nil && *p.UserID == playerID {
			continue
		}
		if p.Score > maxScore {
			maxScore = p.Score
			winnerID = p.UserID
		}
	}

	g.raise(EventTypeGameForfeited, GameForfeitedPayload{
		ForfeitedBy: playerID,
		WinnerID:    winnerID,
	})
}

// raise creates a domain event, applies it to state, and records it as uncommitted.
func (g *Game) raise(eventType string, payload any) {
	aggregateID := ""
	if g.GameID != nil {
		aggregateID = strconv.Itoa(*g.GameID)
	}

	event := events.DomainEvent{
		Type:        eventType,
		OccurredAt:  time.Now(),
		AggregateID: aggregateID,
		Version:     g.Version() + 1,
		Payload:     payload,
	}

	g.applyEvent(event)
	g.RecordEvent(event)
}

// applyEvent projects a domain event onto the aggregate state.
// Used for both new events (via raise) and replay (via LoadFromEvents).
func (g *Game) applyEvent(event events.DomainEvent) {
	switch event.Type {
	case EventTypeGameCreated:
		p, ok := event.Payload.(GameCreatedPayload)
		if !ok {
			return
		}
		g.GameID = &p.GameID
		g.BoardSize = p.BoardSize
		g.CurrentTurn = 0
		g.CreatedAt = event.OccurredAt
		g.Players = make([]Player, len(p.Players))
		copy(g.Players, p.Players)
		g.Grid = makeEmptyGrid(p.BoardSize)

	case EventTypeMoveApplied:
		p, ok := event.Payload.(MoveAppliedPayload)
		if !ok {
			return
		}
		g.setEdge(p.Row, p.Col, EdgeType(p.Edge))

	case EventTypeBoxCompleted:
		p, ok := event.Payload.(BoxCompletedPayload)
		if !ok {
			return
		}
		g.Grid[p.Row][p.Col].OwnerTurn = &p.OwnerTurn
		g.Players[p.OwnerTurn].Score++

	case EventTypeTurnPassed:
		p, ok := event.Payload.(TurnPassedPayload)
		if !ok {
			return
		}
		g.CurrentTurn = p.NextTurn

	case EventTypeGameEnded:
		p, ok := event.Payload.(GameEndedPayload)
		if !ok {
			return
		}
		g.WinnerID = p.WinnerID
		now := event.OccurredAt
		g.EndedAt = &now

	case EventTypeGameForfeited:
		p, ok := event.Payload.(GameForfeitedPayload)
		if !ok {
			return
		}
		g.WinnerID = p.WinnerID
		now := event.OccurredAt
		g.EndedAt = &now
	}

	g.IncrementVersion()
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
	return g.EndedAt != nil
}

// GetCurrentPlayer returns the player whose turn it is
func (g *Game) GetCurrentPlayer() *Player {
	if g.CurrentTurn < 0 || g.CurrentTurn >= len(g.Players) {
		return nil
	}
	return &g.Players[g.CurrentTurn]
}

// --- Private helper methods ---

func (g *Game) validateMove(move Move) error {
	if move.TurnOrder != g.CurrentTurn {
		return fmt.Errorf("not player %d's turn (current turn: %d)", move.TurnOrder, g.CurrentTurn)
	}

	if move.Row < 0 || move.Row >= g.BoardSize || move.Col < 0 || move.Col >= g.BoardSize {
		return fmt.Errorf("coordinates out of bounds: (%d, %d)", move.Row, move.Col)
	}

	if !g.isEdgeAvailable(move.Row, move.Col, move.Edge) {
		return fmt.Errorf("edge %s at (%d, %d) already taken", move.Edge, move.Row, move.Col)
	}

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

// findNewlyCompletedBoxes checks which boxes became complete after an edge was placed.
// This reads from the already-mutated grid (edge was set by MoveApplied event).
func (g *Game) findNewlyCompletedBoxes(move Move) []Box {
	var completed []Box

	if box := g.findCompletableBox(move.Row, move.Col); box != nil {
		completed = append(completed, *box)
	}

	adjRow, adjCol := g.getAdjacentBox(move.Row, move.Col, move.Edge)
	if adjRow >= 0 {
		if box := g.findCompletableBox(adjRow, adjCol); box != nil {
			completed = append(completed, *box)
		}
	}

	return completed
}

// findCompletableBox returns the box if all 4 edges are set and it has no owner.
func (g *Game) findCompletableBox(row, col int) *Box {
	if row < 0 || row >= g.BoardSize || col < 0 || col >= g.BoardSize {
		return nil
	}

	box := &g.Grid[row][col]
	if box.OwnerTurn != nil || !g.isBoxComplete(box) {
		return nil
	}

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

func (g *Game) isGameOver() bool {
	for row := 0; row < g.BoardSize; row++ {
		for col := 0; col < g.BoardSize; col++ {
			if g.Grid[row][col].OwnerTurn == nil {
				return false
			}
		}
	}
	return true
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
