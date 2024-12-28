package auth

import (
	"fmt"
	"qerplunk/garin-chat/envconfig"

	"github.com/golang-jwt/jwt/v5"
)

// Returns true if the JWT token is valid
// The JWT decode secret should be located in .env under JWT_DECODE_SECRET
func JWTTokenValid(token string) bool {
	if token == "" {
		fmt.Println("No authentication token")
		return false
	}

	jwtDecodeSecret := envconfig.EnvConfig.JwtSecret

	tok, jwtError := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtDecodeSecret), nil
	})

	if jwtError != nil {
		fmt.Println("Error trying to parse JWT:", jwtError)
		return false
	}

	if !tok.Valid {
		fmt.Println("Invalid token")
		return false
	}

	return true
}
