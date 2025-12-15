package game

import (
	"fmt"
	"testing"
	"time"
)

func mockGameState(boardSize int) *GameState {
	players := []Player{
		{UserID: 1, Username: "Alice", TurnOrder: 0, Score: 0},
		{UserID: 2, Username: "Bob", TurnOrder: 1, Score: 0},
		// {UserID: 3, Username: "John", TurnOrder: 2, Score: 0},
		// {UserID: 4, Username: "Tan", TurnOrder: 3, Score: 0},
	}

	currentTurn := 0

	// Create empty grid boxes
	grids := []Box{}
	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			grids = append(grids, Box{
				Row: row,
				Col: col,
			})
		}
	}

	game := &Game{
		GameId:      nil,
		GameName:    nil,
		CreatedAt:   time.Now(),
		Players:     players,
		BoardSize:   boardSize,
		CurrentTurn: currentTurn,
		WinnerId:    nil,
	}

	return &GameState{
		Game:  game,
		Grids: grids,
	}
}
func TestApplyMoveCompletesBox(t *testing.T) {
	gs := mockGameState(1)
	game := NewGame(gs)
		fmt.Printf("Players in Game: %+v\n", game.Players)
	fmt.Printf("CurrentTurn: %d\n", game.CurrentTurn)

	moves := []Move{
		{UserID: 1, Row: 0, Col: 0, Edge: "top_edge"},
		{UserID: 2, Row: 0, Col: 0, Edge: "right_edge"},
		{UserID: 1, Row: 0, Col: 0, Edge: "bottom_edge"},
		{UserID: 2, Row: 0, Col: 0, Edge: "left_edge"},
	}

	for _, m := range moves {
		_, err := game.ApplyMove(m)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		printGridState(game)
	}

	// Check that the box is marked as completed
	if game.Grid[0][0].Completed_by == nil {
		t.Errorf("expected box to be completed")
	}

	// Check score is 1 for whoever completed it
	completedBy := *game.Grid[0][0].Completed_by
	if game.Scores[completedBy] != 1 {
		t.Errorf("expected score 1 for player %d, got %d", completedBy, game.Scores[completedBy])
	}
}

func TestGameOverAndWinner2(t *testing.T) {
	gs := mockGameState(1)
	game := NewGame(gs)

	moves := []Move{
		{UserID: 1, Row: 0, Col: 0, Edge: "top_edge"},
		{UserID: 2, Row: 0, Col: 0, Edge: "right_edge"},
		{UserID: 1, Row: 0, Col: 0, Edge: "bottom_edge"},
		{UserID: 2, Row: 0, Col: 0, Edge: "left_edge"},
	}

	for _, m := range moves {
		game.ApplyMove(m)
		printGridState(game)
	}

	if !game.isGameOver() {
		t.Errorf("expected game to be over")
	}

	winner := game.determineWinner()
	if winner == nil {
		t.Errorf("expected a winner, got tie or nil")
	}
}

func TestGameOverAndWinner(t *testing.T) {
	gs := mockGameState(3)
	game := NewGame(gs)

	playerIDs := []int{1, 2}
	playerIndex := 0

	boardSize := gs.Game.BoardSize

	seenEdges := make(map[string]bool)

	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			boxEdges := []struct {
				Edge string
				Key  string
			}{
				{"top_edge", fmt.Sprintf("top-%d-%d", row, col)},
				{"right_edge", fmt.Sprintf("right-%d-%d", row, col)},
				{"bottom_edge", fmt.Sprintf("bottom-%d-%d", row, col)},
				{"left_edge", fmt.Sprintf("left-%d-%d", row, col)},
			}

			for _, e := range boxEdges {
				if seenEdges[e.Key] {
					continue
				}
				seenEdges[e.Key] = true

				switch e.Edge {
				case "top_edge":
					if row > 0 {
						seenEdges[fmt.Sprintf("bottom-%d-%d", row-1, col)] = true
					}
				case "bottom_edge":
					if row < boardSize-1 {
						seenEdges[fmt.Sprintf("top-%d-%d", row+1, col)] = true
					}
				case "left_edge":
					if col > 0 {
						seenEdges[fmt.Sprintf("right-%d-%d", row, col-1)] = true
					}
				case "right_edge":
					if col < boardSize-1 {
						seenEdges[fmt.Sprintf("left-%d-%d", row, col+1)] = true
					}
				}

				move := Move{
					UserID: playerIDs[playerIndex],
					Row:    row,
					Col:    col,
					Edge:   e.Edge,
				}

				fmt.Printf("Player %d's turn: placing %s at (%d, %d)\n", move.UserID, move.Edge, move.Row, move.Col)

				result, err := game.ApplyMove(move)
				if err != nil {
					t.Fatalf("error applying move %v: %v", move, err)
				}

				printGridState(game)

				if !result.BoxCompleted {
					playerIndex = (playerIndex + 1) % len(playerIDs)
				} else {
					fmt.Printf("Boxes claimed: %+v\n", result.ClaimedBoxes)
					fmt.Printf("Updated scores: %+v\n", result.NewScores)
				}

			}
		}
	}

	if !game.isGameOver() {
		t.Errorf("expected game to be over")
	}

	winner := game.determineWinner()
	if winner == nil {
		t.Errorf("expected a winner, got tie or nil")
	} else {
		t.Logf("Winner is player %d with score %d", *winner, game.Scores[*winner])
	}
}

func printGridState(game *Game) {
	boardSize := len(game.Grid)

	for col := 0; col < boardSize; col++ {
		if game.Grid[0][col].TopEdge {
			fmt.Print(" ---")
		} else {
			fmt.Print("    ")
		}
	}
	fmt.Println()

	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			if col == 0 {
				if game.Grid[row][col].LeftEdge {
					fmt.Print("|")
				} else {
					fmt.Print(" ")
				}
			} else {
				if game.Grid[row][col-1].RightEdge {
					fmt.Print("|")
				} else {
					fmt.Print(" ")
				}
			}

			box := game.Grid[row][col]
			if box.Completed_by != nil {
				fmt.Printf(" %d ", *box.Completed_by)
			} else {
				fmt.Print("   ")
			}
		}

		if game.Grid[row][boardSize-1].RightEdge {
			fmt.Print("|")
		}
		fmt.Println()

		for col := 0; col < boardSize; col++ {
			if game.Grid[row][col].BottomEdge {
				fmt.Print(" ---")
			} else if row < boardSize-1 && game.Grid[row+1][col].TopEdge {
				fmt.Print(" ---")
			} else {
				fmt.Print("    ")
			}
		}
		fmt.Println()
	}

	fmt.Println()
}

func TestCompletingMultipleBoxesInOneMove(t *testing.T) {
	gs := mockGameState(2) // Small 2x2 board
	game := NewGame(gs)

	// Pre-fill edges so that two boxes will be completed by the last move
	moves := []Move{
		// Top-left box (0,0)
		{UserID: 1, Row: 0, Col: 0, Edge: "top_edge"},
		{UserID: 2, Row: 0, Col: 0, Edge: "left_edge"},
		{UserID: 1, Row: 0, Col: 0, Edge: "bottom_edge"},

		// Top-right box (0,1)
		{UserID: 2, Row: 0, Col: 1, Edge: "top_edge"},
		{UserID: 1, Row: 0, Col: 1, Edge: "right_edge"},
		{UserID: 2, Row: 0, Col: 1, Edge: "bottom_edge"},

		// Now both boxes are missing only their shared edge
		{UserID: 1, Row: 0, Col: 0, Edge: "right_edge"}, // This should complete both boxes
	}

	for _, move := range moves {
		result, err := game.ApplyMove(move)
		if err != nil {
			t.Fatalf("error applying move %v: %v", move, err)
		}
		printGridState(game)
		if result.BoxCompleted {
			fmt.Printf("Boxes claimed: %+v\n", result.ClaimedBoxes)
		}
	}

	// Verify both boxes were completed
	completedBoxes := 0
	for row := 0; row < 2; row++ {
		for col := 0; col < 2; col++ {
			if game.Grid[row][col].Completed_by != nil {
				completedBoxes++
			}
		}
	}

	if completedBoxes != 2 {
		t.Errorf("expected 2 boxes to be completed, got %d", completedBoxes)
	}

	expectedScore := 2
	if game.Scores[1] != expectedScore {
		t.Errorf("expected player 1 to have score %d, got %d", expectedScore, game.Scores[1])
	}
}
