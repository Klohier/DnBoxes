package game

import (
	"context"
	"dango/internal/events"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// UserFinder looks up a username by user ID.
type UserFinder interface {
	FindUsername(ctx context.Context, userID int) (string, error)
}

type GameHandler struct {
	gameService  *GameService
	timerService *GameTimerService
	wsSubscriber events.WebSocketSubscriber
	userFinder   UserFinder
	logger       *slog.Logger
}

func NewGameHandler(gameService *GameService, timerService *GameTimerService, wsSubscriber events.WebSocketSubscriber, userFinder UserFinder) *GameHandler {
	return &GameHandler{
		gameService:  gameService,
		timerService: timerService,
		wsSubscriber: wsSubscriber,
		userFinder:   userFinder,
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

	// Look up actual usernames for all players
	usernames := make(map[int]string)
	for _, id := range req.PlayerIDs {
		username, err := h.userFinder.FindUsername(c.Request().Context(), id)
		if err != nil {
			slog.Warn("Failed to look up username", "userID", id, "error", err)
		} else {
			usernames[id] = username
		}
	}

	game, err := h.gameService.CreateGame(c.Request().Context(), req.PlayerIDs, req.BoardSize, usernames)
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

	// Look up actual username for the human player
	usernames := make(map[int]string)
	username, err := h.userFinder.FindUsername(c.Request().Context(), req.HumanPlayerID)
	if err != nil {
		slog.Warn("Failed to look up username for bot game", "userID", req.HumanPlayerID, "error", err)
	} else {
		usernames[req.HumanPlayerID] = username
	}

	game, err := h.gameService.botService.CreateBotGame(playerIDs, req.NumBots, req.BoardSize, usernames)
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

func (h *GameHandler) ForfeitGame(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid game ID"})
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
	}
	playerID := int(userIDFloat)

	game, err := h.gameService.ForfeitGame(c.Request().Context(), gameID, playerID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, game)
}

func (h *GameHandler) GetGameHistory(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
	}
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok || !userToken.Valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
	}
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token subject"})
	}
	userID := int(userIDFloat)

	history, err := h.gameService.GetUserGameHistory(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load game history"})
	}

	if history == nil {
		history = []GameHistoryEntry{}
	}

	return c.JSON(http.StatusOK, history)
}

func (h *GameHandler) GetGameEvents(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid game ID"})
	}

	domainEvents, err := h.gameService.GetGameEvents(c.Request().Context(), gameID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to load game events"})
	}

	if domainEvents == nil {
		domainEvents = []events.DomainEvent{}
	}

	return c.JSON(http.StatusOK, domainEvents)
}

func (h *GameHandler) GetTimerState(c echo.Context) error {
	gameID, err := strconv.Atoi(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid game ID"})
	}

	if h.timerService == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Timer not available"})
	}

	state := h.timerService.GetTimerState(gameID)
	if state == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No active timer for this game"})
	}

	return c.JSON(http.StatusOK, state)
}