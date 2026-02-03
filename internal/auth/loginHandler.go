package auth

import (
	"dango/internal/auth/token"
	"dango/internal/user"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

type LoginHandler struct {
	loginService *LoginService
	userService  *user.UserService
	logger       *slog.Logger
}

func NewLoginHandler(loginService *LoginService, userService *user.UserService) *LoginHandler {
	return &LoginHandler{
		loginService: loginService,
		userService:  userService,
		logger:       slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
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
	cookie.SameSite = http.SameSiteNoneMode
	cookie.Secure = true
	sessionToken, nil := token.GenerateToken(user.UserID, user.Username)
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
func (h *LoginHandler) GuestLogin(c echo.Context) error {
	ctx := c.Request().Context()

	guest, err := h.userService.CreateGuestUser(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not create guest: "+err.Error())
	}

	sessionToken, err := token.GenerateToken(guest.UserID, guest.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate session token")
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

	h.logger.Info("Guest user created", "userID", guest.UserID, "username", guest.Username)

	return c.JSON(http.StatusOK, guest)
}

func (h *LoginHandler) Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "DnB-Session"
	cookie.Value = ""
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.MaxAge = -1 // Expire immediately

	c.SetCookie(cookie)

	return c.NoContent(http.StatusOK)
}
