package token

import (
	"fmt"
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

func VerifyToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return key, nil
	})
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, fmt.Errorf("missing or invalid subject claim")
	}

	return int(userIDFloat), nil
}
