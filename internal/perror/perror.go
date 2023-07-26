// Package perror provides custom error types for the Privytar service.
package perror

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// ErrorResponse is the response returned by the API when an error occurs.
type ErrorResponse struct {
	// Message is a human-readable message describing the error.
	Message string `json:"message"`

	// Code is a machine-readable code describing the error.
	Code uint `json:"code"`
}

// JSON sends an ErrorResponse to the HTTP response writer as JSON.
func JSON(w http.ResponseWriter, logger *zap.Logger, response ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.Code))

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode error response", zap.String("error", err.Error()))
	}
}
