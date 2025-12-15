package game

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

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
	GameName    *string   `json:"game_name"`
	CreatedAt   time.Time `json:"created_at"`
	BoardSize   int       `json:"board_size"`
	Players     []Player  `json:"players"`
	WinnerId    *int      `json:"winner"`
	CurrentTurn int      `json:"current_turn"`
	Grid        [][]Box   `json:"grids"` 
	Scores      map[int]int `json:"scores"` 
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

// NewGame creates a new game instance from GameState
func NewGame(gameState *GameState) *Game {
	game := &Game{
		GameId:      gameState.Game.GameId,
		GameName:    gameState.Game.GameName,
		CreatedAt:   gameState.Game.CreatedAt,
		Players:     gameState.Game.Players,
		BoardSize:   gameState.Game.BoardSize,
		CurrentTurn: gameState.Game.CurrentTurn,
		WinnerId:    gameState.Game.WinnerId,
		Scores:      make(map[int]int),
	}

	// Initialize scores from players
	for _, p := range game.Players {
		game.Scores[p.UserID] = p.Score
	}

	// Initialize the 2D Grid slice
	game.Grid = make([][]Box, game.BoardSize)
	for i := range game.Grid {
		game.Grid[i] = make([]Box, game.BoardSize)
	}

	// Convert 1D Grids slice into 2D Grid
	for _, box := range gameState.Grids {
		row := box.Row
		col := box.Col
		if row >= 0 && row < game.BoardSize && col >= 0 && col < game.BoardSize {
			game.Grid[row][col] = box
		}
	}

	return game
}

// ApplyMove applies a player's move to the game
func (g *Game) ApplyMove(move Move) (MoveResult, error) {
		if len(g.Players) > 0 {
	slog.Debug("ApplyMove called",
		"currentTurn", g.CurrentTurn,
		"numPlayers", len(g.Players),
		"currentPlayerID", g.Players[g.CurrentTurn].UserID,
		"movePlayerID", move.UserID)
} else {
	slog.Debug("ApplyMove called",
		"currentTurn", g.CurrentTurn,
		"numPlayers", len(g.Players))
}
	var result MoveResult
	result.Move = move

	// Validate turn
	if g.CurrentTurn < 0 || g.CurrentTurn >= len(g.Players) {
		return result, fmt.Errorf("invalid current turn index: %d", g.CurrentTurn)
	}
	if g.Players[g.CurrentTurn].UserID != move.UserID {
		return result, fmt.Errorf("it's not player %d's turn", move.UserID)
	}

	// Validate coordinates
	if move.Row < 0 || move.Row >= g.BoardSize || move.Col < 0 || move.Col >= g.BoardSize {
		return result, fmt.Errorf("move out of bounds")
	}

	// Set the edge
	if err := g.setEdge(move.Row, move.Col, move.Edge); err != nil {
		return result, err
	}

	// Check if this box or adjacent boxes are completed
	boxCompleted, claimedBoxes := g.checkAdjacentBoxes(move)
	result.BoxCompleted = boxCompleted
	result.ClaimedBoxes = claimedBoxes

	// Copy current scores to result
	newScores := make(map[int]int, len(g.Scores))
	for k, v := range g.Scores {
		newScores[k] = v
	}
	result.NewScores = newScores

	// Update turn if no box completed
	if !boxCompleted {
		nextTurn := (g.CurrentTurn + 1) % len(g.Players)
		g.CurrentTurn = nextTurn
	}

	result.NextTurn = g.CurrentTurn

	// Check if game is over
	if g.isGameOver() {
		winnerID := g.determineWinner()
		g.WinnerId = winnerID
		result.WinnerID = winnerID
	}

	return result, nil
}

// GenerateBotMove generates a move for a bot player using strategy
func (g *Game) GenerateBotMove(botId int) *Move {
	completionMoves := []Move{}
	safeMoves := []Move{}
	riskyMoves := []Move{}

	// Categorize moves
	for row := 0; row < g.BoardSize; row++ {
		for col := 0; col < g.BoardSize; col++ {
			box := &g.Grid[row][col]
			if box.Completed_by != nil {
				continue
			}

			edges := []string{"top_edge", "right_edge", "bottom_edge", "left_edge"}
			for _, edge := range edges {
				if !g.isEdgeAvailable(row, col, edge) {
					continue
				}

				claimed := g.countClaimedEdges(box)
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
		return &completionMoves[rand.Intn(len(completionMoves))]
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

// IsGameOver checks if all boxes are completed
func (g *Game) IsGameOver() bool {
	return g.isGameOver()
}

// GetWinner returns the current winner (if any)
func (g *Game) GetWinner() *int {
	return g.WinnerId
}

// GetCurrentPlayer returns the player whose turn it is
func (g *Game) GetCurrentPlayer() *Player {
	if  g.CurrentTurn < 0 || g.CurrentTurn >= len(g.Players) {
		return nil
	}
	return &g.Players[g.CurrentTurn]
}

// GetScore returns a player's current score
func (g *Game) GetScore(userID int) int {
	return g.Scores[userID]
}

// Private methods

func (g *Game) setEdge(row, col int, edge string) error {
	taken, err := g.edgeTaken(row, col, edge)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("%s already selected", edge)
	}

	box := &g.Grid[row][col]

	switch edge {
	case "top_edge":
		box.TopEdge = true
		if row > 0 {
			g.Grid[row-1][col].BottomEdge = true
		}
	case "right_edge":
		box.RightEdge = true
		if col < g.BoardSize-1 {
			g.Grid[row][col+1].LeftEdge = true
		}
	case "bottom_edge":
		box.BottomEdge = true
		if row < g.BoardSize-1 {
			g.Grid[row+1][col].TopEdge = true
		}
	case "left_edge":
		box.LeftEdge = true
		if col > 0 {
			g.Grid[row][col-1].RightEdge = true
		}
	}

	return nil
}

func (g *Game) checkAdjacentBoxes(move Move) (boxCompleted bool, claimedBoxes []Box) {
	check := func(row, col int) {
		completed, claim := g.checkAndScoreBox(row, col, move.UserID)
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

func (g *Game) checkAndScoreBox(row, col int, userID int) (bool, *Box) {
	// Check boundaries
	if row < 0 || row >= g.BoardSize || col < 0 || col >= g.BoardSize {
		return false, nil
	}

	box := &g.Grid[row][col]

	// Check if box is already owned or not completed
	if box.Completed_by != nil || !box.TopEdge || !box.RightEdge || !box.BottomEdge || !box.LeftEdge {
		return false, nil
	}

	// Assign ownership and increment score
	box.Completed_by = &userID
	completed := true
	box.Completed = &completed
	g.Scores[userID]++

	return true, box
}

func (g *Game) countClaimedEdges(box *Box) int {
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

func (g *Game) isGameOver() bool {
	for row := 0; row < g.BoardSize; row++ {
		for col := 0; col < g.BoardSize; col++ {
			if g.Grid[row][col].Completed_by == nil {
				return false
			}
		}
	}
	return true
}

func (g *Game) determineWinner() *int {
	var maxScore int
	var winnerID *int
	var tie bool

	for playerID, score := range g.Scores {
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

func (g *Game) edgeTaken(row, col int, edge string) (bool, error) {
	if row < 0 || row >= g.BoardSize || col < 0 || col >= g.BoardSize {
		return true, fmt.Errorf("coordinates out of bounds")
	}

	box := g.Grid[row][col]

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

func (g *Game) isEdgeAvailable(row, col int, edge string) bool {
	taken, err := g.edgeTaken(row, col, edge)
	return err == nil && !taken
}
