// Package perror provides custom error types for the Privytar service.
package perror

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse is the response returned by the API when an error occurs.
type ErrorResponse struct {
	// Message is a human-readable message describing the error.
	Message string `json:"message"`

	// Code is a machine-readable code describing the error.
	Code uint `json:"code"`
}

// JSON sends an ErrorResponse to the HTTP response writer as JSON.
func JSON(ctx context.Context, w http.ResponseWriter, logger *slog.Logger, response ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.Code))

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.LogAttrs(
			ctx,
			slog.LevelError,
			"failed to encode error response",
			slog.String("error", err.Error()),
		)
	}
}
