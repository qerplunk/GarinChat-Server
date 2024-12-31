package middleware

import (
	"fmt"
	"net/http"
	"qerplunk/garin-chat/envconfig"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Creates a middleware stack out of Middlewares located in this file.
// Useful for reusing middleware stacks.
//
// Example:
// stack := middleware.CreateStack(middleware.OriginCheck(), middleware.FooBarCheck())
func CreateStack(middlewares ...Middleware) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for _, middleware := range middlewares {
			next = middleware(next)
		}
		return next
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
