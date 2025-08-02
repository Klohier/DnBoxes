package auth

import (
	// "errors"
	// "dango/internal/auth"
	// "log"
	"dango/internal/auth/token"
	"log/slog"
	"net/http"
	"time"

	// "net/http"
	"os"

	"github.com/labstack/echo/v4"
)

type LoginHandler struct {
	loginService *LoginService
	logger       *slog.Logger
}

func NewLoginHandler(loginService *LoginService) *LoginHandler {
	return &LoginHandler{loginService: loginService,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil))}
}

func (h *LoginHandler) Login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	ctx := c.Request().Context()

	user, err := h.loginService.Login(ctx, username, password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Could Not Login User: "+err.Error())
	}

	//SETS COOKIE
	cookie := new(http.Cookie)
	cookie.Name = "DnB-Session"

	cookie.HttpOnly = true
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"

	sessionToken, nil := token.GenerateToken(user.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate session token: ")
	}

	cookie.Value = sessionToken
	c.SetCookie(cookie)

	h.logger.Info(sessionToken)
	// csrf := c.Request().Header.Get("X-CSRF-TOKEN")
	// if csrf != user.CSRF || "" {
	// 	return echo.NewHTTPError(http.StatusUnauthorized, "Could Not Login User: " + err.Error())
	// }

	return c.JSON(http.StatusOK, user)

}
func (h *LoginHandler) Logout(c echo.Context) error {
	// Clear the HttpOnly cookie by setting MaxAge to -1
	cookie := new(http.Cookie)
	cookie.Name = "DnB-Session"
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.MaxAge = -1 // Expire immediately

	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}
