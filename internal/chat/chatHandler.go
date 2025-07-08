package chat

import (
	"errors"

	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

//
func NewChatHandler(chatService *ChatService) *ChatHandler{
	return &ChatHandler{
	chatService: chatService,	
	logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)),}
}



// GetAllMessagesHandler handles the request to get all messages
func (h *ChatHandler) GetAllMessageFromSession(c echo.Context) error {
	sessionID, err := strconv.Atoi(c.QueryParam("sessionID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid session id"))
}

	messages, err := h.chatService.GetAllMessageFromSession(c.Request().Context(), sessionID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.New("Failed to get messages: " + err.Error()),
		)
		 
	}

	return c.JSON(http.StatusOK, messages)
}

func (h *ChatHandler) GetAllGameMessage(c echo.Context) error {

	gameId, err := strconv.Atoi(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid game id"))
}

	messages, err := h.chatService.GetAllGameMessage(c.Request().Context(), gameId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.New("Failed to get messages: "  + err.Error()),
		)
	}

	return c.JSON(http.StatusOK, messages)
}
