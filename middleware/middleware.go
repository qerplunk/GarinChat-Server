package middleware

import (
	"fmt"
	"net/http"
	"qerplunk/garin-chat/auth"
	"qerplunk/garin-chat/envconfig"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Creates a middleware stack out of Middlewares located in this file.
// Useful for reusing middleware stacks.
//
// Example:
// stack := middleware.CreateStack(middleware.JWT_Check(), middleware.OriginCheck())
func CreateStack(middlewares ...Middleware) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for _, middleware := range middlewares {
			next = middleware(next)
		}
		return next
	}
}

// Checks if the JWT decode secret provided can decode the URL query value of "token"
// The JWT decode secret should be located in .env under JWT_DECODE_SECRET
func JWTCheck() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token_str := r.URL.Query().Get("token")
			if !auth.JWTTokenValid(token_str) {
				return
			}

			next(w, r)
		})
	}
}

// Checks if the request origin is allowed
// Allowed origins should be located in .env under ALLOWED_ORIGINS as a list
func OriginCheck() Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowedOrigins := envconfig.EnvConfig.AllowedOrigins

			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					next(w, r)
					return
				}
			}

			fmt.Printf("Origin %s NOT allowed\n", origin)
			return
		})
	}
}
