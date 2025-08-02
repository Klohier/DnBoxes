package user

import (
	"dango/internal/auth/token"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService *UserService
	logger      *slog.Logger
}

type UserResponse struct {
	UserID   int    `json:"userID"`
	Username string `json:"username" validate:"required"`
}

func NewUserResponse(user *User) *UserResponse {
	return &UserResponse{
		UserID:   user.UserID,
		Username: user.Username,
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

	ctx := c.Request().Context()
	user, err := h.userService.CreateUser(ctx, username, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could Not Create User: "+err.Error())
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to Retrieve User: "+err.Error())
	}

	UserResponse := NewUserResponse(user)

	return c.JSON(http.StatusOK, UserResponse)

}

func (h *UserHandler) GetMe(c echo.Context) error {
	ctx := c.Request().Context()

	cookie, err := c.Cookie("DnB-Session")
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Failed to Retrieve Cookie: "+err.Error())
	}

	tokenString := cookie.Value

	userID, err := token.VerifyToken(tokenString)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed to Verify Token: "+err.Error())
	}

	user, err := h.userService.FindByID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Failed to Retrieve User: "+err.Error())
	}

	UserResponse := NewUserResponse(user)

	return c.JSON(http.StatusOK, UserResponse)

}

func (h *UserHandler) GetAllUsers(c echo.Context) error {
	ctx := c.Request().Context()

	users, err := h.userService.userRepo.FindAll(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to Retrieve Users: "+err.Error())
	}

	UserResponses := NewUserResponses(users)
	return c.JSON(http.StatusOK, UserResponses)
}

func (h UserHandler) HandleNewUser(c echo.Context) error {
	return nil
}
