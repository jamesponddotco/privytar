package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
)

// AcceptRequests is a middleware that only allows GET, HEAD, and OPTIONS requests.
func AcceptRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			perror.JSON(r.Context(), w, logger, perror.ErrorResponse{
				Code:    http.StatusMethodNotAllowed,
				Message: fmt.Sprintf("Method %s not allowed. Must be GET, HEAD, or OPTIONS.", r.Method),
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}
