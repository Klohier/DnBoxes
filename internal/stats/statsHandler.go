package stats

import (
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type StatsHandler struct {
	service *StatsService
}

func NewStatsHandler(service *StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetMyStats(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok || !userToken.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
	}
	userID := int(userIDFloat)

	stats, err := h.service.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get stats: "+err.Error())
	}

	return c.JSON(http.StatusOK, stats)
}

func (h *StatsHandler) GetUserStats(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
	}

	stats, err := h.service.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get stats: "+err.Error())
	}

	return c.JSON(http.StatusOK, stats)
}

func (h *StatsHandler) GetLeaderboard(c echo.Context) error {
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	leaderboard, err := h.service.GetLeaderboard(c.Request().Context(), limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get leaderboard: "+err.Error())
	}

	return c.JSON(http.StatusOK, leaderboard)
}
