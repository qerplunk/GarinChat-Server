package auth

import (
	"log"
	"qerplunk/garin-chat/envconfig"

	"github.com/golang-jwt/jwt/v5"
)

// Returns true if the JWT token is valid
// The JWT decode secret should be located in .env under JWT_DECODE_SECRET
func JWTTokenValid(token string) bool {
	if token == "" {
		log.Println("No authentication token provided")
		return false
	}

	jwtDecodeSecret := envconfig.EnvConfig.JwtSecret

	tok, jwtError := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtDecodeSecret), nil
	})

	if jwtError != nil {
		log.Println("Error trying to parse JWT:", jwtError)
		return false
	}

	if !tok.Valid {
		log.Println("Invalid token")
		return false
	}

	return true
}
