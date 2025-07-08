package session

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

type SessionHandler struct {
	sessionService *SessionService
	logger         *slog.Logger
}

func NewSessionHandler(sessionService *SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		logger:         slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

// GET /sessions
func (h *SessionHandler) GetAllSessions(c echo.Context) error {
	sessions, err := h.sessionService.FindAllSessions(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to get sessions", "error", err)
		return c.JSON(http.StatusInternalServerError, errors.New("failed to get sessions"))
	}
	return c.JSON(http.StatusOK, sessions)
}

// GET /sessions/:id
func (h *SessionHandler) GetSessionByID(c echo.Context) error {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid session id"))
	}

	session, err := h.sessionService.FindSessionByID(c.Request().Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get session by ID", "error", err)
		return c.JSON(http.StatusInternalServerError, errors.New("failed to get session"))
	}

	return c.JSON(http.StatusOK, session)
}

// POST /sessions
func (h *SessionHandler) CreateSession(c echo.Context) error {
	var session Session
	if err := c.Bind(&session); err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid session body"))
	}

	created, err := h.sessionService.CreateSession(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to create session", "error", err)
		return c.JSON(http.StatusInternalServerError, errors.New("failed to create session"))
	}

	return c.JSON(http.StatusCreated, created)
}

// DELETE /sessions/:id
func (h *SessionHandler) DeleteSession(c echo.Context) error {
	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid session id"))
	}

	if err := h.sessionService.DeleteSession(c.Request().Context(), sessionID); err != nil {
		h.logger.Error("Failed to delete session", "error", err)
		return c.JSON(http.StatusInternalServerError, errors.New("failed to delete session"))
	}

	return c.NoContent(http.StatusNoContent)
}

// POST /sessions/:sessionId/users/:userId
func (h *SessionHandler) AddUserToSession(c echo.Context) error {
	sessionID, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid session id"))
	}

	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid user id"))
	}

	err = h.sessionService.AddUserToSession(c.Request().Context(), sessionID, userID)
	if err != nil {
		h.logger.Error("Failed to add user to session", "error", err)
		return c.JSON(http.StatusInternalServerError, errors.New("failed to add user to session"))
	}

	return c.NoContent(http.StatusNoContent)
}

// DELETE /sessions/:sessionId/users/:userId
func (h *SessionHandler) RemoveUserFromSession(c echo.Context) error {
	sessionID, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid session id"))
	}

	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid user id"))
	}

	err = h.sessionService.RemoveUserFromSession(c.Request().Context(), sessionID, userID)
	if err != nil {
		h.logger.Error("Failed to remove user from session", "error", err)
		return c.JSON(http.StatusInternalServerError, errors.New("failed to remove user from session"))
	}

	return c.NoContent(http.StatusNoContent)
}
