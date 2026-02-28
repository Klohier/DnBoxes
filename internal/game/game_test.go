package game

import (
	"fmt"
	"testing"
)

func mockGame(boardSize int) *Game {
	userID1 := 1
	userID2 := 2

	players := []Player{
		{UserID: &userID1, Username: "Alice", TurnOrder: 0, Score: 0},
		{UserID: &userID2, Username: "Bob", TurnOrder: 1, Score: 0},
	}

	game := NewGame(nil, boardSize, players)
	return game
}

func TestApplyMoveCompletesBox(t *testing.T) {
	game := mockGame(1)
	fmt.Printf("Players in Game: %+v\n", game.Players)
	fmt.Printf("CurrentTurn: %d\n", game.CurrentTurn)

	moves := []Move{
		{TurnOrder: 0, Row: 0, Col: 0, Edge: TopEdge},
		{TurnOrder: 1, Row: 0, Col: 0, Edge: RightEdge},
		{TurnOrder: 0, Row: 0, Col: 0, Edge: BottomEdge},
		{TurnOrder: 1, Row: 0, Col: 0, Edge: LeftEdge},
	}

	for _, m := range moves {
		_, err := game.ApplyMove(m)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		printGridState(game)
	}

	// Check that the box is marked as completed
	if game.Grid[0][0].OwnerTurn == nil {
		t.Errorf("expected box to be completed")
	}

	// Check score is 1 for whoever completed it
	ownerTurn := *game.Grid[0][0].OwnerTurn
	if game.Players[ownerTurn].Score != 1 {
		t.Errorf("expected score 1 for player at turn %d, got %d", ownerTurn, game.Players[ownerTurn].Score)
	}
}

func TestGameOverAndWinner2(t *testing.T) {
	game := mockGame(1)

	moves := []Move{
		{TurnOrder: 0, Row: 0, Col: 0, Edge: TopEdge},
		{TurnOrder: 1, Row: 0, Col: 0, Edge: RightEdge},
		{TurnOrder: 0, Row: 0, Col: 0, Edge: BottomEdge},
		{TurnOrder: 1, Row: 0, Col: 0, Edge: LeftEdge},
	}

	for _, m := range moves {
		game.ApplyMove(m)
		printGridState(game)
	}

	if !game.IsGameOver() {
		t.Errorf("expected game to be over")
	}

	winner := game.WinnerID
	if winner == nil {
		t.Errorf("expected a winner, got tie or nil")
	}
}

func TestGameOverAndWinner(t *testing.T) {
	game := mockGame(3)

	turnOrders := []int{0, 1}
	turnIndex := 0

	boardSize := game.BoardSize

	seenEdges := make(map[string]bool)

	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			boxEdges := []struct {
				Edge EdgeType
				Key  string
			}{
				{TopEdge, fmt.Sprintf("top-%d-%d", row, col)},
				{RightEdge, fmt.Sprintf("right-%d-%d", row, col)},
				{BottomEdge, fmt.Sprintf("bottom-%d-%d", row, col)},
				{LeftEdge, fmt.Sprintf("left-%d-%d", row, col)},
			}

			for _, e := range boxEdges {
				if seenEdges[e.Key] {
					continue
				}
				seenEdges[e.Key] = true

				switch e.Edge {
				case TopEdge:
					if row > 0 {
						seenEdges[fmt.Sprintf("bottom-%d-%d", row-1, col)] = true
					}
				case BottomEdge:
					if row < boardSize-1 {
						seenEdges[fmt.Sprintf("top-%d-%d", row+1, col)] = true
					}
				case LeftEdge:
					if col > 0 {
						seenEdges[fmt.Sprintf("right-%d-%d", row, col-1)] = true
					}
				case RightEdge:
					if col < boardSize-1 {
						seenEdges[fmt.Sprintf("left-%d-%d", row, col+1)] = true
					}
				}

				move := Move{
					TurnOrder: turnOrders[turnIndex],
					Row:       row,
					Col:       col,
					Edge:      e.Edge,
				}

				currentPlayer := game.Players[move.TurnOrder]
				userIDStr := "anonymous"
				if currentPlayer.UserID != nil {
					userIDStr = fmt.Sprintf("%d", *currentPlayer.UserID)
				}
				fmt.Printf("Player %s's turn: placing %s at (%d, %d)\n", userIDStr, move.Edge, move.Row, move.Col)

				result, err := game.ApplyMove(move)
				if err != nil {
					t.Fatalf("error applying move %v: %v", move, err)
				}

				printGridState(game)

				if len(result.CompletedBoxes) == 0 {
					turnIndex = (turnIndex + 1) % len(turnOrders)
				} else {
					fmt.Printf("Boxes claimed: %+v\n", result.CompletedBoxes)
					fmt.Printf("Player score: %d\n", game.Players[move.TurnOrder].Score)
				}
			}
		}
	}

	if !game.IsGameOver() {
		t.Errorf("expected game to be over")
	}

	winner := game.WinnerID
	if winner == nil {
		t.Errorf("expected a winner, got tie or nil")
	} else {
		// Find winner player
		for _, p := range game.Players {
			if p.UserID != nil && *p.UserID == *winner {
				t.Logf("Winner is player %d with score %d", *winner, p.Score)
				break
			}
		}
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
			if box.OwnerTurn != nil {
				// Show turn order instead of user_id
				fmt.Printf(" %d ", *box.OwnerTurn)
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

	// Print current scores
	fmt.Print("Scores: ")
	for _, p := range game.Players {
		if p.UserID != nil {
			fmt.Printf("Player %d: %d  ", *p.UserID, p.Score)
		} else {
			fmt.Printf("Anonymous: %d  ", p.Score)
		}
	}
	fmt.Println()
	fmt.Println()
}

func TestCompletingMultipleBoxesInOneMove(t *testing.T) {
	game := mockGame(2) // Small 2x2 board

	// Pre-fill edges so that two boxes will be completed by the last move
	moves := []Move{
		// Top-left box (0,0)
		{TurnOrder: 0, Row: 0, Col: 0, Edge: TopEdge},
		{TurnOrder: 1, Row: 0, Col: 0, Edge: LeftEdge},
		{TurnOrder: 0, Row: 0, Col: 0, Edge: BottomEdge},

		// Top-right box (0,1)
		{TurnOrder: 1, Row: 0, Col: 1, Edge: TopEdge},
		{TurnOrder: 0, Row: 0, Col: 1, Edge: RightEdge},
		{TurnOrder: 1, Row: 0, Col: 1, Edge: BottomEdge},

		// Now both boxes are missing only their shared edge
		{TurnOrder: 0, Row: 0, Col: 0, Edge: RightEdge}, // This should complete both boxes
	}

	for _, move := range moves {
		result, err := game.ApplyMove(move)
		if err != nil {
			t.Fatalf("error applying move %v: %v", move, err)
		}
		printGridState(game)
		if len(result.CompletedBoxes) > 0 {
			fmt.Printf("Boxes claimed: %+v\n", result.CompletedBoxes)
		}
	}

	// Verify both boxes were completed
	completedBoxes := 0
	for row := 0; row < 2; row++ {
		for col := 0; col < 2; col++ {
			if game.Grid[row][col].OwnerTurn != nil {
				completedBoxes++
			}
		}
	}

	if completedBoxes != 2 {
		t.Errorf("expected 2 boxes to be completed, got %d", completedBoxes)
	}

	expectedScore := 2
	if game.Players[0].Score != expectedScore {
		t.Errorf("expected player at turn 0 to have score %d, got %d", expectedScore, game.Players[0].Score)
	}
}

func mockGameWithPlayers(boardSize int, numPlayers int) *Game {
	players := []Player{}

	for i := 0; i < numPlayers; i++ {
		userID := i + 1
		players = append(players, Player{
			UserID:    &userID,
			Username:  fmt.Sprintf("Player%d", i+1),
			TurnOrder: i,
			Score:     0,
		})
	}

	game := NewGame(nil, boardSize, players)
	return game
}

func TestThreePlayerGame(t *testing.T) {
	game := mockGameWithPlayers(3, 3)

	if len(game.Players) != 3 {
		t.Fatalf("expected 3 players, got %d", len(game.Players))
	}

	turnOrders := []int{0, 1, 2}
	turnIndex := 0

	boardSize := game.BoardSize
	seenEdges := make(map[string]bool)

	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			boxEdges := []struct {
				Edge EdgeType
				Key  string
			}{
				{TopEdge, fmt.Sprintf("top-%d-%d", row, col)},
				{RightEdge, fmt.Sprintf("right-%d-%d", row, col)},
				{BottomEdge, fmt.Sprintf("bottom-%d-%d", row, col)},
				{LeftEdge, fmt.Sprintf("left-%d-%d", row, col)},
			}

			for _, e := range boxEdges {
				if seenEdges[e.Key] {
					continue
				}
				seenEdges[e.Key] = true

				switch e.Edge {
				case TopEdge:
					if row > 0 {
						seenEdges[fmt.Sprintf("bottom-%d-%d", row-1, col)] = true
					}
				case BottomEdge:
					if row < boardSize-1 {
						seenEdges[fmt.Sprintf("top-%d-%d", row+1, col)] = true
					}
				case LeftEdge:
					if col > 0 {
						seenEdges[fmt.Sprintf("right-%d-%d", row, col-1)] = true
					}
				case RightEdge:
					if col < boardSize-1 {
						seenEdges[fmt.Sprintf("left-%d-%d", row, col+1)] = true
					}
				}

				move := Move{
					TurnOrder: turnOrders[turnIndex],
					Row:       row,
					Col:       col,
					Edge:      e.Edge,
				}

				currentPlayer := game.Players[move.TurnOrder]
				fmt.Printf("Player %d's turn: placing %s at (%d, %d)\n", *currentPlayer.UserID, move.Edge, move.Row, move.Col)

				result, err := game.ApplyMove(move)
				if err != nil {
					t.Fatalf("error applying move %v: %v", move, err)
				}

				if len(result.CompletedBoxes) == 0 {
					turnIndex = (turnIndex + 1) % len(turnOrders)
				} else {
					fmt.Printf("Player %d claimed %d box(es)\n", *currentPlayer.UserID, len(result.CompletedBoxes))
				}
			}
		}
	}

	printGridState(game)

	if !game.IsGameOver() {
		t.Errorf("expected game to be over")
	}

	// Verify all 3 players have scores tracked
	totalScore := 0
	for _, p := range game.Players {
		totalScore += p.Score
		t.Logf("Player %d final score: %d", *p.UserID, p.Score)
	}

	expectedTotalBoxes := boardSize * boardSize
	if totalScore != expectedTotalBoxes {
		t.Errorf("expected total score to equal total boxes (%d), got %d", expectedTotalBoxes, totalScore)
	}

	winner := game.WinnerID
	if winner == nil {
		t.Log("Game ended in a tie")
	} else {
		t.Logf("Winner is player %d", *winner)
	}
}

func TestFourPlayerGame(t *testing.T) {
	game := mockGameWithPlayers(21, 4)

	if len(game.Players) != 4 {
		t.Fatalf("expected 4 players, got %d", len(game.Players))
	}

	turnOrders := []int{0, 1, 2, 3}
	turnIndex := 0

	boardSize := game.BoardSize
	seenEdges := make(map[string]bool)

	moveCount := 0

	for row := 0; row < boardSize; row++ {
		for col := 0; col < boardSize; col++ {
			boxEdges := []struct {
				Edge EdgeType
				Key  string
			}{
				{TopEdge, fmt.Sprintf("top-%d-%d", row, col)},
				{RightEdge, fmt.Sprintf("right-%d-%d", row, col)},
				{BottomEdge, fmt.Sprintf("bottom-%d-%d", row, col)},
				{LeftEdge, fmt.Sprintf("left-%d-%d", row, col)},
			}

			for _, e := range boxEdges {
				if seenEdges[e.Key] {
					continue
				}
				seenEdges[e.Key] = true

				switch e.Edge {
				case TopEdge:
					if row > 0 {
						seenEdges[fmt.Sprintf("bottom-%d-%d", row-1, col)] = true
					}
				case BottomEdge:
					if row < boardSize-1 {
						seenEdges[fmt.Sprintf("top-%d-%d", row+1, col)] = true
					}
				case LeftEdge:
					if col > 0 {
						seenEdges[fmt.Sprintf("right-%d-%d", row, col-1)] = true
					}
				case RightEdge:
					if col < boardSize-1 {
						seenEdges[fmt.Sprintf("left-%d-%d", row, col+1)] = true
					}
				}

				move := Move{
					TurnOrder: turnOrders[turnIndex],
					Row:       row,
					Col:       col,
					Edge:      e.Edge,
				}

				result, err := game.ApplyMove(move)
				if err != nil {
					t.Fatalf("error applying move %v: %v", move, err)
				}

				moveCount++
				if len(result.CompletedBoxes) == 0 {
					turnIndex = (turnIndex + 1) % len(turnOrders)
				}
			}
		}
	}

	printGridState(game)

	if !game.IsGameOver() {
		t.Errorf("expected game to be over")
	}

	// Verify all 4 players have scores tracked
	totalScore := 0
	for i, p := range game.Players {
		totalScore += p.Score
		t.Logf("Player %d (turn order %d) final score: %d", *p.UserID, i, p.Score)
	}

	expectedTotalBoxes := boardSize * boardSize
	if totalScore != expectedTotalBoxes {
		t.Errorf("expected total score to equal total boxes (%d), got %d", expectedTotalBoxes, totalScore)
	}

	winner := game.WinnerID
	if winner == nil {
		t.Log("Game ended in a tie")
	} else {
		t.Logf("Winner is player %d", *winner)
	}

	t.Logf("Total moves in game: %d", moveCount)
}

func TestTurnRotationWithMultiplePlayers(t *testing.T) {
	game := mockGameWithPlayers(2, 4) // 4 players on small board

	// Test that turns rotate correctly through all 4 players
	moves := []Move{
		{TurnOrder: 0, Row: 0, Col: 0, Edge: TopEdge},
		{TurnOrder: 1, Row: 0, Col: 0, Edge: RightEdge},
		{TurnOrder: 2, Row: 0, Col: 0, Edge: BottomEdge},
		{TurnOrder: 3, Row: 0, Col: 0, Edge: LeftEdge}, // Completes box
	}

	for i, move := range moves {
		if game.CurrentTurn != move.TurnOrder {
			t.Errorf("move %d: expected current turn to be %d, got %d", i, move.TurnOrder, game.CurrentTurn)
		}

		result, err := game.ApplyMove(move)
		if err != nil {
			t.Fatalf("error applying move %v: %v", move, err)
		}

		// Box completed on last move, so turn should stay with player 3
		if i == 3 && len(result.CompletedBoxes) > 0 {
			if game.CurrentTurn != 3 {
				t.Errorf("expected turn to stay with player 3 after completing box, got %d", game.CurrentTurn)
			}
		}
	}

	printGridState(game)

	// Verify box was claimed by player 3 (turn order 3)
	if game.Grid[0][0].OwnerTurn == nil {
		t.Errorf("expected box to be completed")
	} else if *game.Grid[0][0].OwnerTurn != 3 {
		t.Errorf("expected box to be owned by turn order 3, got %d", *game.Grid[0][0].OwnerTurn)
	}

	// Verify player 3 got the point
	if game.Players[3].Score != 1 {
		t.Errorf("expected player at turn order 3 to have score 1, got %d", game.Players[3].Score)
	}
}
