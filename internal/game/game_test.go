package game

import (
	"fmt"
	"testing"
)

func mockGameState(boardSize int) *GameState {
    players := []Player{
        {UserID: 1, Username: "Alice"},
        {UserID: 2, Username: "Bob"},
		{UserID: 3, Username: "John"},
		{UserID: 4, Username: "Tan"},

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

    return &GameState{
        Game: &Game{
            Players: players,
            BoardSize: boardSize,
            CurrentTurn: &currentTurn,
        },
        Grids: grids,
    }
}

func TestApplyMoveCompletesBox(t *testing.T) {
    gs := mockGameState(1)
    engine := NewEngine(gs)

    moves := []Move{
        {UserID: 1, Row: 0, Col: 0, Edge: "top_edge"},
        {UserID: 2, Row: 0, Col: 0, Edge: "right_edge"},
        {UserID: 1, Row: 0, Col: 0, Edge: "bottom_edge"},
        {UserID: 2, Row: 0, Col: 0, Edge: "left_edge"},
    }

    for _, m := range moves {
        _, err := engine.ApplyMove(m)
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }
		printGridState(engine)
    }

    // Check that the box is marked as completed
    if engine.Grid[0][0].Completed_by == nil {
        t.Errorf("expected box to be completed")
    }

    // Check score is 1 for whoever completed it
    completedBy := *engine.Grid[0][0].Completed_by
    if engine.Scores[completedBy] != 1 {
        t.Errorf("expected score 1 for player %d, got %d", completedBy, engine.Scores[completedBy])
    }
}

func TestGameOverAndWinner2(t *testing.T) {
    gs := mockGameState(5)
    engine := NewEngine(gs)

    moves := []Move{
        {UserID: 1, Row: 0, Col: 0, Edge: "top_edge"},
        {UserID: 2, Row: 0, Col: 0, Edge: "right_edge"},
        {UserID: 1, Row: 0, Col: 0, Edge: "bottom_edge"},
        {UserID: 2, Row: 0, Col: 0, Edge: "left_edge"},
    }

    for _, m := range moves {
        engine.ApplyMove(m)
		printGridState(engine)
    }

    if !engine.isGameOver() {
        t.Errorf("expected game to be over")
    }

    winner := engine.determineWinner()
    if winner == nil {
        t.Errorf("expected a winner, got tie or nil")
    }
}

func TestGameOverAndWinner(t *testing.T) {
    gs := mockGameState(3)
    engine := NewEngine(gs)

    playerIDs := []int{1, 2,3, 4}
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

                result, err := engine.ApplyMove(move)
                if err != nil {
                    t.Fatalf("error applying move %v: %v", move, err)
                }

                printGridState(engine)

				if !result.BoxCompleted {
                playerIndex = (playerIndex + 1) % len(playerIDs)
            } else {
                fmt.Printf("Boxes claimed: %+v\n", result.ClaimedBoxes)
                fmt.Printf("Updated scores: %+v\n", result.NewScores)
            }
            
            }
        }
    }

    if !engine.isGameOver() {
        t.Errorf("expected game to be over")
    }

    winner := engine.determineWinner()
    if winner == nil {
        t.Errorf("expected a winner, got tie or nil")
    } else {
        t.Logf("Winner is player %d with score %d", *winner, engine.Scores[*winner])
    }
}



func printGridState(engine *Engine) {
    boardSize := len(engine.Grid)

    for col := 0; col < boardSize; col++ {
        if engine.Grid[0][col].TopEdge {
            fmt.Print(" ---")
        } else {
            fmt.Print("    ")
        }
    }
    fmt.Println()

    for row := 0; row < boardSize; row++ {
        for col := 0; col < boardSize; col++ {
            if col == 0 {
                if engine.Grid[row][col].LeftEdge {
                    fmt.Print("|")
                } else {
                    fmt.Print(" ")
                }
            } else {
                if engine.Grid[row][col-1].RightEdge {
                    fmt.Print("|")
                } else {
                    fmt.Print(" ")
                }
            }

            box := engine.Grid[row][col]
            if box.Completed_by != nil {
                fmt.Printf(" %d ", *box.Completed_by)
            } else {
                fmt.Print("   ")
            }
        }

        if engine.Grid[row][boardSize-1].RightEdge {
            fmt.Print("|")
        }
        fmt.Println()

        for col := 0; col < boardSize; col++ {
            if engine.Grid[row][col].BottomEdge {
                fmt.Print(" ---")
            } else if row < boardSize-1 && engine.Grid[row+1][col].TopEdge {
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
    engine := NewEngine(gs)

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
        result, err := engine.ApplyMove(move)
        if err != nil {
            t.Fatalf("error applying move %v: %v", move, err)
        }
        printGridState(engine)
        if result.BoxCompleted {
            fmt.Printf("Boxes claimed: %+v\n", result.ClaimedBoxes)
        }
    }

    // Verify both boxes were completed
    completedBoxes := 0
    for row := 0; row < 2; row++ {
        for col := 0; col < 2; col++ {
            if engine.Grid[row][col].Completed_by != nil {
                completedBoxes++
            }
        }
    }

    if completedBoxes != 2 {
        t.Errorf("expected 2 boxes to be completed, got %d", completedBoxes)
    }

    expectedScore := 2
    if engine.Scores[1] != expectedScore {
        t.Errorf("expected player 1 to have score %d, got %d", expectedScore, engine.Scores[1])
    }
}