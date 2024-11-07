package auth

import (
	// "errors"
	"log/slog"
	"net/http"

	// "net/http"
	"os"

	"github.com/labstack/echo/v4"
)
type LoginHandler struct{

	loginService *LoginService
	logger *slog.Logger
}


func NewLoginHandler(loginService *LoginService) *LoginHandler{
	return &LoginHandler{loginService: loginService,
	logger:  slog.New(slog.NewJSONHandler(os.Stdout, nil)),}
}


func (h *LoginHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
ctx := c.Request().Context()

user, err := h.loginService.Login(ctx, username, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Could Not Login User: " + err.Error())
	}

	
	// TODO Assin a generated token to user and send back to client

	return c.JSON(http.StatusOK, user)


}