package auth

import (
	// "errors"
	// "dango/internal/auth"
	// "log"
	"log/slog"
	"net/http"
	"time"

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


	//SETS COOKIE
	cookie := new(http.Cookie)
	cookie.Name = "DnB-Session"

	// cookie.HttpOnly = true
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.Path = "/"


	sessionToken, nil := GenerateToken(c.RealIP(), time.Now().UTC().String(), c.Request().UserAgent(), user.UserID) 
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate session token: ")
	}
	
	cookie.Value = sessionToken
	c.SetCookie(cookie)
	// TODO Assin a generated token to user and send back to client

	// session, err := c.Cookie("DnB-Session")
	// if err != nil || session.Value == "" || session.Value != sessionToken  {
	// 	return echo.NewHTTPError(http.StatusUnauthorized, "Could Not Login User: " + err.Error())	
		
	// }
	h.logger.Info(sessionToken)
	// csrf := c.Request().Header.Get("X-CSRF-TOKEN")
	// if csrf != user.CSRF || "" {
	// 	return echo.NewHTTPError(http.StatusUnauthorized, "Could Not Login User: " + err.Error())
	// }
	
	return c.JSON(http.StatusOK, user)


}