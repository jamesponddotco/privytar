package middleware

import (
	"log/slog"
	"net/http"

	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
)

// UserAgent ensures that the request has a valid user agent.
func UserAgent(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() == "" {
			perror.JSON(r.Context(), w, logger, perror.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "User agent is missing. Please provide a valid user agent.",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}
