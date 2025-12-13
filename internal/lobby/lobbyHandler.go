package lobby

import (
	"dango/internal/events"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)



type LobbyHandler struct {
	lobbyService *LobbyService
	wsSubscriber events.WebSocketSubscriber
	logger       *slog.Logger
}

type CreateLobbyRequest struct {
	Name        string `json:"name"`
	PlayerLimit int    `json:"player_limit"`
	IsPrivate   bool   `json:"is_private"`
	BoardSize	 int    `json:"board_size"`
}

type LobbyResponse struct {
	LobbyID     string `json:"lobby_id"`
	Name        string `json:"name"`
	BoardSize	 int    `json:"board_size"`
	HostID      int    `json:"host_id"`
	PlayerLimit int    `json:"player_limit"`
	IsPrivate   bool   `json:"is_private"`
	CreatedAt   string `json:"created_at"`
	Players     []LobbyPlayer `json:"players"`
}


// CreateLobby creates a new lobby for authenticated users
func NewLobbyHandler(lobbyService *LobbyService, wsSubscriber events.WebSocketSubscriber) *LobbyHandler {
	return &LobbyHandler{
		wsSubscriber: wsSubscriber,
		lobbyService: lobbyService,
	}
}

func (h *LobbyHandler) CreateLobby(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateLobbyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Validate board size
	if req.BoardSize < 5 || req.BoardSize > 10 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "board_size must be between 5 and 10"})
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


	username, ok := claims["username"].(string)
if !ok {
    return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token: missing username"})
}

	// Save to persistence
	lobby, err := h.lobbyService.CreateLobby(ctx, userID, username, req.Name, req.PlayerLimit, req.IsPrivate, req.BoardSize)
if err != nil {
    return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create lobby"})
}

// Convert to LobbyResponse for consistent frontend data
//TODO: fix int casting
	resp := LobbyResponse{
		LobbyID:     lobby.LobbyID,
		BoardSize: lobby.BoardSize,
		Name:        lobby.Name,
		HostID:      int(lobby.HostID),
		PlayerLimit: lobby.PlayerLimit,
		IsPrivate:   lobby.IsPrivate,
		CreatedAt:   lobby.CreatedAt.Format(time.RFC3339),
		Players:     lobby.Players,
	}



h.wsSubscriber.SubscribeUser(int(userID), fmt.Sprintf("lobby:%s", lobby.LobbyID))

return c.JSON(http.StatusOK, resp)

}

func (h *LobbyHandler) GetAllLobbies(c echo.Context) error {
    ctx := c.Request().Context()

    lobbies, err := h.lobbyService.GetAllLobbies(ctx)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to fetch lobbies",
        })
    }

    // Return list of lobbies
    responses := make([]LobbyResponse, len(lobbies))
    for i, l := range lobbies {
        responses[i] = LobbyResponse{
            LobbyID:     l.LobbyID,
			BoardSize: l.BoardSize,
            Name:        l.Name,
            HostID:      int(l.HostID),
            PlayerLimit: l.PlayerLimit,
            IsPrivate:   l.IsPrivate,
            CreatedAt:   l.CreatedAt.Format(time.RFC3339),
            Players:     l.Players,
        }
    }

    return c.JSON(http.StatusOK, responses)
}

func (h *LobbyHandler) JoinLobby(c echo.Context) error {
ctx := c.Request().Context()

lobbyID := c.Param("lobbyId")
if lobbyID == "" {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing lobby ID"})
}

// Extract user from JWT
userToken := c.Get("user").(*jwt.Token)
claims := userToken.Claims.(jwt.MapClaims)
userIDFloat, ok := claims["sub"].(float64)
if !ok {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
}
userID := int64(userIDFloat)

username, ok := claims["username"].(string)
if !ok {
    return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token: missing username"})
}

// Join the lobby
if err := h.lobbyService.JoinLobby(ctx, lobbyID, userID, username); err != nil {
	if err == ErrLobbyNotFound {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "lobby not found"})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to join lobby"})
}

//fetch the updated lobby to return
lobby, err := h.lobbyService.lobbyRepo.GetLobby(ctx, lobbyID)
if err != nil || lobby == nil {
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch updated lobby"})
}

//join room in websocket
h.wsSubscriber.SubscribeUser(int(userID), fmt.Sprintf("lobby:%s", lobbyID))

resp := LobbyResponse{
	LobbyID:     lobby.LobbyID,
	BoardSize: lobby.BoardSize,
	Name:        lobby.Name,
	HostID:      int(lobby.HostID),
	PlayerLimit: lobby.PlayerLimit,
	IsPrivate:   lobby.IsPrivate,
	CreatedAt:   lobby.CreatedAt.Format(time.RFC3339),
	Players:     lobby.Players,
}

return c.JSON(http.StatusOK, resp)

}

func (h *LobbyHandler) GetLobby(c echo.Context) error {
	ctx := c.Request().Context()

	lobbyID := c.Param("lobbyId")
	if lobbyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing lobby ID"})
	}

	lobby, err := h.lobbyService.GetLobby(ctx, lobbyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch lobby"})
	}
	if lobby == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "lobby not found"})
	}
	userToken := c.Get("user").(*jwt.Token)
claims := userToken.Claims.(jwt.MapClaims)
userIDFloat, ok := claims["sub"].(float64)
if !ok {
	return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
}
userID := int64(userIDFloat)

//join room in websocket
h.wsSubscriber.SubscribeUser(int(userID), fmt.Sprintf("lobby:%s", lobbyID))

	

	resp := LobbyResponse{
		LobbyID:     lobby.LobbyID,
		BoardSize: lobby.BoardSize,
		Name:        lobby.Name,
		HostID:      int(lobby.HostID),
		PlayerLimit: lobby.PlayerLimit,
		IsPrivate:   lobby.IsPrivate,
		CreatedAt:   lobby.CreatedAt.Format(time.RFC3339),
		Players:     lobby.Players,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *LobbyHandler) ToggleReady(c echo.Context) error {
	ctx := c.Request().Context()

	lobbyID := c.Param("lobbyId")
	if lobbyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing lobby ID"})
	}

	// Extract user from JWT
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(jwt.MapClaims)
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
	}
	userID := int64(userIDFloat)

	// Get current lobby to check player's ready status
	lobby, err := h.lobbyService.GetLobby(ctx, lobbyID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch lobby"})
	}
	if lobby == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "lobby not found"})
	}

	// Find the player and their current ready status
	var currentReadyStatus bool
	playerFound := false
	for _, player := range lobby.Players {
		if player.UserID == userID {
			currentReadyStatus = player.IsReady
			playerFound = true
			break
		}
	}

	if !playerFound {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "player not in lobby"})
	}

	// Toggle the ready status
	newReadyStatus := !currentReadyStatus
	if err := h.lobbyService.SetPlayerReady(ctx, lobbyID, userID, newReadyStatus); err != nil {
		if err == ErrLobbyNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "lobby not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to toggle ready status"})
	}

	// Fetch the updated lobby to return
	updatedLobby, err := h.lobbyService.GetLobby(ctx, lobbyID)
	if err != nil || updatedLobby == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch updated lobby"})
	}

	resp := LobbyResponse{
		LobbyID:     updatedLobby.LobbyID,
		BoardSize: updatedLobby.BoardSize,
		Name:        updatedLobby.Name,
		HostID:      int(updatedLobby.HostID),
		PlayerLimit: updatedLobby.PlayerLimit,
		IsPrivate:   updatedLobby.IsPrivate,
		CreatedAt:   updatedLobby.CreatedAt.Format(time.RFC3339),
		Players:     updatedLobby.Players,
	}

	return c.JSON(http.StatusOK, resp)
}
