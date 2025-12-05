package lobby

import (
	"dango/internal/events"
	"net/http"

	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type LobbyHandler struct {
	lobbyService *LobbyService
	eventBus     events.EventBus
	logger       *slog.Logger
}

type CreateLobbyRequest struct {
	Name        string `json:"name"`
	PlayerLimit int    `json:"player_limit"`
	IsPrivate   bool   `json:"is_private"`
}

type LobbyResponse struct {
	LobbyID     string `json:"lobby_id"`
	Name        string `json:"name"`
	HostID      int    `json:"host_id"`
	PlayerLimit int    `json:"player_limit"`
	IsPrivate   bool   `json:"is_private"`
	CreatedAt   string `json:"created_at"`
}


// CreateLobby creates a new lobby for authenticated users
func NewLobbyHandler(eventBus events.EventBus, lobbyService *LobbyService) *LobbyHandler {
	return &LobbyHandler{
		eventBus:     eventBus,
		lobbyService: lobbyService,
	}
}

func (h *LobbyHandler) CreateLobby(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateLobbyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	 // Extract token from echo context
    userToken := c.Get("user").(*jwt.Token)
    claims := userToken.Claims.(jwt.MapClaims)

    // Extract "sub"
    userIDFloat, ok := claims["sub"].(float64)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
    }
    userID := int64(userIDFloat)

	// Save to persistence
	createdLobby, err := h.lobbyService.CreateLobby(ctx, userID, req.Name, req.PlayerLimit, req.IsPrivate)
if err != nil {
    return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create lobby"})
}

return c.JSON(http.StatusOK, createdLobby)

}

func (h *LobbyHandler) GetAllLobbies(c echo.Context) error {
    ctx := c.Request().Context()

    lobbies, err := h.lobbyService.GetAllLobbies(ctx)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to fetch lobbies",
        })
    }

    // Return array of LobbyResponse objects
    return c.JSON(http.StatusOK, lobbies)
}

