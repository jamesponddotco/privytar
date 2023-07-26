package middleware

import (
	"net/http"

	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
	"go.uber.org/zap"
)

// PanicRecovery tries to recover from panics and returns a 500 error if there
// was one.
func PanicRecovery(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered", zap.Any("error", err))

				perror.JSON(w, logger, perror.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "Internal server error. Please try again later.",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
