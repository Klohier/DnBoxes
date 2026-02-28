package user

import (
	"dango/internal/auth/token"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService *UserService
	logger      *slog.Logger
}

type UserResponse struct {
	UserID   int    `json:"userID"`
	Username string `json:"username" validate:"required"`
	IsGuest  bool   `json:"isGuest"`
}

func NewUserResponse(user *User) *UserResponse {
	return &UserResponse{
		UserID:   user.UserID,
		Username: user.Username,
		IsGuest:  user.IsGuest,
	}
}

// Creates a Slice of UserResponses Populated from a slice of Users
func NewUserResponses(users []User) []UserResponse {
	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, *NewUserResponse(&user))
	}
	return userResponses
}

func NewUserHandler(userService *UserService) *UserHandler {
	return &UserHandler{userService: userService,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))}
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "username and password are required")
	}
	if len(username) > 100 || len(password) > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "input exceeds maximum length")
	}

	ctx := c.Request().Context()
	user, err := h.userService.CreateUser(ctx, username, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	h.logger.Info("New User Created",
		"uri", c.Request().RequestURI,
		"status", http.StatusCreated,
	)
	userResponse := NewUserResponse(user)

	return c.JSON(http.StatusCreated, userResponse)
}

func (h *UserHandler) FindByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("invalid User ID"))
	}

	user, err := h.userService.FindByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Failed to Retrieve User")
	}

	UserResponse := NewUserResponse(user)

	return c.JSON(http.StatusOK, UserResponse)

}

func (h *UserHandler) GetMe(c echo.Context) error {
	ctx := c.Request().Context()
	// Get the JWT token object injected by echo-jwt middleware
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthenticated")
	}

	// Extract claims
	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok || !userToken.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
	}

	// Get user ID from claims
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token subject")
	}
	userID := int(userIDFloat)

	// Fetch the user
	user, err := h.userService.FindByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "failed to retrieve user")
	}

	userResponse := NewUserResponse(user)
	return c.JSON(http.StatusOK, userResponse)

}

func (h *UserHandler) UpgradeGuest(c echo.Context) error {
	ctx := c.Request().Context()

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

	username := c.FormValue("username")
	password := c.FormValue("password")

	user, err := h.userService.UpgradeGuest(ctx, userID, username, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Issue new token with updated username
	sessionToken, err := token.GenerateToken(user.UserID, user.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate session token")
	}

	cookie := new(http.Cookie)
	cookie.Name = "DnB-Session"
	cookie.Value = sessionToken
	cookie.HttpOnly = true
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteNoneMode
	cookie.Secure = true
	c.SetCookie(cookie)

	h.logger.Info("Guest upgraded to full account", "userID", user.UserID, "username", user.Username)

	return c.JSON(http.StatusOK, NewUserResponse(user))
}

func (h *UserHandler) GetAllUsers(c echo.Context) error {
	ctx := c.Request().Context()

	users, err := h.userService.userRepo.FindAll(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Failed to Retrieve Users")
	}

	UserResponses := NewUserResponses(users)
	return c.JSON(http.StatusOK, UserResponses)
}
