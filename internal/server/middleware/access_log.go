package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLogResponseWriter is a small adapter for http.ResponseWriter that
// exists so we can grab the HTTP status code of a response.
type AccessLogResponseWriter struct {
	http.ResponseWriter

	// statusCode is the HTTP status code.
	statusCode int
}

// WriteHeader sets the HTTP status code.
func (w *AccessLogResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// AccessLog is a middleware that logs privacy-aware information about every
// request.
func AccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			start  = time.Now().UTC()
			writer = &AccessLogResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
		)

		next.ServeHTTP(writer, r)

		logger.LogAttrs(
			r.Context(),
			slog.LevelInfo,
			"request",
			slog.Int("status", writer.statusCode),
			slog.String("protocol", r.Proto),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", r.RequestURI),
			slog.String("duration", time.Since(start).String()),
			slog.String("user-agent", r.UserAgent()),
		)
	})
}
