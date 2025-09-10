package websocket

import (
	"context"
	"dango/internal/game"
	"encoding/json"
	"fmt"
	"log/slog"
)

func GameStateHandler(event Event, c *Connection, deps *HandlerDeps) error {
	var payload struct {
		GameID int `json:"gameID"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.Error("invalid payload for game_state:", "error", err)
	}

	ctx := context.Background()

	gameState, err := deps.GameService.GetGameState(ctx, payload.GameID)
	if err != nil {
		slog.Error("failed to get game state",
			"gameID", payload.GameID,
			"error", err,
		)
	}

	responsePayload, err := json.Marshal(gameState)
	if err != nil {
		slog.Error("failed to marshal game state response:", "error", err)
	}

	responseEvent := Event{
		Type:    EventGameState,
		Payload: responsePayload,
	}

	c.egress <- responseEvent

	if gameState.Game.WinnerId != nil {
		if err := broadcastWinnerEvent(c, gameState, payload.GameID); err != nil {
			slog.Error("failed to broadcast winner", "error", err)
		}

	}

	return nil
}



func MoveHandler(event Event, c *Connection, deps *HandlerDeps) error {
	type MakeMovePayload struct {
		GameID   int    `json:"gameID"`
		PlayerID int    `json:"playerID"`
		Row      int    `json:"row"`
		Col      int    `json:"col"`
		Edge     string `json:"edge"`
	}

	var payload MakeMovePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.Error("invalid payload for make_move: ", "error", err)
	}

	ctx := context.Background()

	gameState, err := deps.GameService.MakeMove(ctx, payload.GameID, payload.PlayerID, payload.Row, payload.Col, payload.Edge)
	if err != nil {

		invalidMoveEvent := Event{
			Type: "invalid_move",
		}

		c.egress <- invalidMoveEvent

		slog.Error("failed to make moves for game",
			"gameID", payload.GameID,
			"error", err,
		)
	}

	topic := fmt.Sprintf("game:%d", payload.GameID)
	c.manager.broadcast(topic, "game:state", gameState)

	return nil
}

func QuitGameHandler(event Event, c *Connection, deps *HandlerDeps) error {
	var quitGameEvent QuitGameEvent
	if err := json.Unmarshal(event.Payload, &quitGameEvent); err != nil {
		slog.Error("invalid payload for quit_game: ", "error", err)
	}

	winnerSet := false

	if winnerSet {
		gameState, err := deps.GameService.GetGameState(context.Background(), quitGameEvent.GameID)
		if err != nil {
			slog.Error("failed to get game state: ", "error", err)
		}

		if err := broadcastWinnerEvent(c, gameState, quitGameEvent.GameID); err != nil {
			slog.Error("failed to broadcast winner: ", "error", err)
		}
	}
	return nil
}
func broadcastWinnerEvent(c *Connection, gameState *game.GameState, gameID int) error {
	if gameState.Game.WinnerId == nil {
		return nil
	}

	playerMap := make(map[int]string, len(gameState.Game.Players))
	for _, player := range gameState.Game.Players {
		playerMap[player.UserID] = player.Username
	}

	return nil
}
