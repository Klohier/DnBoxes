package game

import (
	"dango/internal/events"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type GameHandler struct {
	gameService  *GameService
	wsSubscriber events.WebSocketSubscriber
	logger       *slog.Logger
}

func NewGameHandler(gameService *GameService, wsSubscriber events.WebSocketSubscriber) *GameHandler {
	return &GameHandler{
		gameService:  gameService,
		wsSubscriber: wsSubscriber,
		logger:       slog.Default(),
	}
}

func (h *GameHandler) CreateGame(c echo.Context) error {
	var req struct {
		PlayerIDs []int `json:"player_ids"`
		BoardSize int   `json:"board_size"`
	}
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	game, err := h.gameService.CreateGame(c.Request().Context(), req.PlayerIDs, req.BoardSize)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	
	// Subscribe players to game room
	if game.GameID != nil {
		topic := fmt.Sprintf("game:%d", *game.GameID)
		for _, playerID := range req.PlayerIDs {
			h.wsSubscriber.SubscribeUser(playerID, topic)
		}
	}
	
	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) GetGameState(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid game ID"})
	}
	
	// Subscribe user to game
	if userToken := c.Get("user"); userToken != nil {
		if claims := userToken.(*jwt.Token).Claims.(jwt.MapClaims); claims != nil {
			if userIDFloat, ok := claims["sub"].(float64); ok {
				userID := int(userIDFloat)
				topic := fmt.Sprintf("game:%d", gameID)
				h.wsSubscriber.SubscribeUser(userID, topic)
			}
		}
	}
	
	game, err := h.gameService.GetGame(c.Request().Context(), gameID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Game not found"})
	}
	
	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) MakeMove(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid game ID"})
	}
	
	var req struct {
		PlayerID int    `json:"playerId"`
		Row      int    `json:"row"`
		Col      int    `json:"col"`
		Edge     string `json:"edge"`
	}
	
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	
	game, err := h.gameService.MakeMove(c.Request().Context(), gameID, req.PlayerID, req.Row, req.Col, req.Edge)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) CreateBotGame(c echo.Context) error {
	var req struct {
		HumanPlayerID int `json:"human_player_id"`
		BoardSize     int `json:"board_size"`
		NumBots       int `json:"num_bots"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	slog.Info("CreateBotGame called", "req", req)

	if req.HumanPlayerID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "human_player_id is required"})
	}
	if req.BoardSize <= 4 || req.BoardSize >= 20 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "board_size must be >4 and <20"})
	}
	if req.NumBots <= 0 || req.NumBots > 3 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "num_bots must be 1-3"})
	}

	playerIDs := []int{req.HumanPlayerID}

	game, err := h.gameService.botService.CreateBotGame(playerIDs, req.NumBots, req.BoardSize)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create bot game: " + err.Error()})
	}

	if game.GameID != nil {
		topic := fmt.Sprintf("game:%d", *game.GameID)
		h.wsSubscriber.SubscribeUser(req.HumanPlayerID, topic)
		slog.Info("Subscribed player to bot game room", "playerID", req.HumanPlayerID, "topic", topic)
	}

	return c.JSON(http.StatusOK, game)
}