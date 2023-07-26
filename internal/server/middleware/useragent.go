package middleware

import (
	"net/http"

	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
	"go.uber.org/zap"
)

// UserAgent ensures that the request has a valid user agent.
func UserAgent(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() == "" {
			perror.JSON(w, logger, perror.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "User agent is missing. Please provide a valid user agent.",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}
