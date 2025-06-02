package game

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)
type GameHandler struct {

	gameService *GameService
	logger *slog.Logger
}

type GameResponse struct {
	UserID int `json:"userID"`
    // Username string `json:"username" validate:"required"`
}

type GameRequest struct {
	GameId int `json:"gameId"`
}




func NewGameHandler(userService *GameService) *GameHandler{
	return &GameHandler{gameService: userService,
	logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)),}
}

//CreateGame 
func(h *GameHandler) CreateGame(c echo.Context) error {

	var req struct {
		Player1        int `json:"player1"`
		Player2        int `json:"player2"`
		BoardSize int `json:"board_size"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	game, err := h.gameService.CreateGame(c.Request().Context(), req.Player1, req.Player2, req.BoardSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create game" + err.Error()})
	}

	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) GetGameState(c echo.Context) error {
    gameId := c.Param("gameId")
    if gameId == "" {
        return c.JSON(http.StatusBadRequest, "error: gameId is required")
    }

    gameIDInt, err := strconv.Atoi(gameId)
    if err != nil {
        return c.JSON(http.StatusBadRequest, "error: invalid gameId")
    }

    gameState, err := h.gameService.GetGameState(c.Request().Context(), gameIDInt)
    if err != nil {
        h.logger.Error("failed to get game state", "error", err)
        return c.JSON(http.StatusInternalServerError, "error: failed to fetch game state")
    }

    return c.JSON(http.StatusOK, gameState)
}


// func (h *GameHandler) GetGrids(c echo.Context) error {


// gameId := c.Param("gameId")

// ctx := c.Request().Context()


// // Validate gameId
// if gameId == "" {
// 	return c.JSON(http.StatusBadRequest,  "error: gameId is required")
// }

// gameIDInt, err := strconv.Atoi(gameId)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, "error: invalid gameId")
// 	}

// boxes, err := h.gameService.GetGrids(ctx, gameIDInt)
// if err != nil {
// 	slog.Error("Error retrieving grids:", err.Error())
// 	return c.JSON(http.StatusInternalServerError, "error could not retrieve grids")
// }

// // Return the grids as a JSON response
// return c.JSON(http.StatusOK, boxes)

// }

func (h *GameHandler) MakeMove(c echo.Context) error {
	var req struct {
		PlayerId int    `json:"playerId"`
		Row      int    `json:"row"`
		Col      int    `json:"col"`
		Edge     string `json:"edge"`
	}


	gameIdStr := c.Param("gameId")
	if gameIdStr == "" {
		return c.JSON(http.StatusBadRequest, "error: gameId is required")
	}

	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "error invalid gameId")
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "error: Invalid request body")
	}

	type MakeMoveResponse struct {
		GameState    *GameState `json:"gameState"`
		
	}
	// Call the service to make the move
	gameState, err := h.gameService.MakeMove(c.Request().Context(), gameId, req.PlayerId, req.Row, req.Col, req.Edge)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error: Failed to make move: " + err.Error())
	}

	response := MakeMoveResponse{
	GameState:  &gameState,
	
}

	// Return a success response
	return c.JSON(http.StatusOK, response)
}