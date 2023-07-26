// Package middleware provides simple middlewares for the Privatar service.
package middleware

import "net/http"

// Chain wraps a given http.Handler with middlewares.
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}
