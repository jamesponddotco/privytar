package middleware

import (
	"log/slog"
	"net/http"

	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
)

// PanicRecovery tries to recover from panics and returns a 500 error if there
// was one.
func PanicRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.LogAttrs(
					r.Context(),
					slog.LevelError,
					"panic recovered",
					slog.Any("error", err),
				)

				perror.JSON(r.Context(), w, logger, perror.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error. Please try again later.",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
