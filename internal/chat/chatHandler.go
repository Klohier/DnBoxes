package chat

import (
	"errors"
	"fmt"

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
func (h *ChatHandler) GetAllMessage(c echo.Context) error {
	// Call GetAllMessage from ChatService
	messages, err := h.chatService.GetAllMessage(c.Request().Context())
	if err != nil {
		// If there's an error, return a server error response
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get messages: %v", err),
		})
	}

	// Return the list of messages in the response
	return c.JSON(http.StatusOK, messages)
}

func (h *ChatHandler) GetAllGameMessage(c echo.Context) error {

	gameId, err := strconv.Atoi(c.Param("gameId"))
	if err != nil {
		// Handle the error (e.g., return a bad request response)
		return c.JSON(http.StatusBadRequest, errors.New("invalid game id"))
}

	// Call GetAllMessage from ChatService
	messages, err := h.chatService.GetAllGameMessage(c.Request().Context(), gameId)
	if err != nil {
		// If there's an error, return a server error response
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get messages: %v", err),
		})
	}

	// Return the list of messages in the response
	return c.JSON(http.StatusOK, messages)
}
