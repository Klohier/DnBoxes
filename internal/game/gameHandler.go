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


// func NewGameResponse(game *Game) *GameResponse {
// 	return &GameResponse{
// 		UserID: game.player1,
		
// 	}
// }


// Creates a Slice of UserResponses Populated from a slice of Users
// func NewUserResponses(users []User) []UserResponse {
// 	var userResponses []UserResponse
// 	for _, user := range users {
// 		userResponses = append(userResponses, *NewUserResponse(&user))
// 	}
// 	return userResponses
// }

func NewGameHandler(userService *GameService) *GameHandler{
	return &GameHandler{gameService: userService,
	logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)),}
}


func(h *GameHandler) CreateGame(c echo.Context) error {

	var req struct {
		P1        int `json:"p1"`
		P2        int `json:"p2"`
		BoardSize int `json:"board_size"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Call the service to create the game
	game, err := h.gameService.CreateGame(c.Request().Context(), req.P1, req.P2, req.BoardSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create game" + err.Error()})
	}

	// Return the created game details as the response
	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) GetGrids(c echo.Context) error {
// Bind the incoming JSON to the GameRequest struct


gameId := c.Param("gameId")

ctx := c.Request().Context()


// Validate gameId
if gameId == "" {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "gameId is required"})
}

gameIDInt, err := strconv.Atoi(gameId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid gameId"})
	}

// Retrieve the grids (boxes) from the repository
boxes, err := h.gameService.GetGrids(ctx, gameIDInt)
if err != nil {
	slog.Error("Error retrieving grids:", err.Error())
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not retrieve grids"})
}

// Return the grids as a JSON response
return c.JSON(http.StatusOK, boxes)

}

func (h *GameHandler) MakeMove(c echo.Context) error {
	var req struct {
		PlayerId int    `json:"player_id"`
		Row      int    `json:"row"`
		Col      int    `json:"col"`
		Edge     string `json:"edge"`
	}


	gameIdStr := c.Param("gameId")
	if gameIdStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "gameId is required"})
	}

	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid gameId"})
	}

	// Bind request data
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}


	// Call the service to make the move
	grid, err := h.gameService.MakeMove(c.Request().Context(), gameId, req.PlayerId, req.Row, req.Col, req.Edge)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to make move: " + err.Error()})
	}

	// Return a success response
	return c.JSON(http.StatusOK, grid)
}