package token

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var key = []byte(os.Getenv("TOKEN_KEY"))

func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"iss": "DnBoxes-Auth",
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(key)
}