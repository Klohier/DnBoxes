package auth

import (
	"encoding/base64"
	"fmt"
	"net"
)

func GenerateToken(ip, timestamp, userAgent string, userId int) (string, error) {

	parsedIP := net.ParseIP(ip).To16()

	rawToken := fmt.Sprintf("%d|%s|%s|%s|%s", userId, ip, timestamp, userAgent, parsedIP)

	token := base64.StdEncoding.EncodeToString([]byte(rawToken))

	return token, nil
}
