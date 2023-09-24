package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"git.sr.ht/~jamesponddotco/privytar/internal/cache"
	"git.sr.ht/~jamesponddotco/privytar/internal/fetch"
	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
	"git.sr.ht/~jamesponddotco/xstd-go/xhash/xfnv"
)

const (
	// HashSizeMD5 is the size of the MD5 hash.
	HashSizeMD5 int = 32

	// HashSizeSHA256 is the size of the SHA256 hash.
	HashSizeSHA256 int = 64
)

// AvatarHandler is the HTTP handler for the /avatar endpoint.
type AvatarHandler struct {
	fetchClient *fetch.Client
	cache       *cache.Cache
	logger      *slog.Logger
	homepage    string
}

// NewAvatarHandler returns a new AvatarHandler instance.
func NewAvatarHandler(
	homepage string,
	fetchClient *fetch.Client,
	cacheInstance *cache.Cache,
	logger *slog.Logger,
) *AvatarHandler {
	return &AvatarHandler{
		fetchClient: fetchClient,
		cache:       cacheInstance,
		logger:      logger,
		homepage:    homepage,
	}
}

// ServeHTTP handles HTTP requests for the /avatar endpoint.
func (h *AvatarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Path[len("/avatar/"):]

	if hash == "" {
		http.Redirect(w, r, h.homepage, http.StatusMovedPermanently)

		return
	}

	if (len(hash) != HashSizeMD5 && len(hash) != HashSizeSHA256) || !IsHexadecimal(hash) {
		h.logger.LogAttrs(
			r.Context(),
			slog.LevelError,
			"invalid hash format",
			slog.String("hash", hash),
		)

		perror.JSON(r.Context(), w, h.logger, perror.ErrorResponse{
			Message: "Invalid hash format",
			Code:    http.StatusBadRequest,
		})

		return
	}

	normalizedQuery, err := NormalizeQueryString(r.URL.RawQuery)
	if err != nil {
		h.logger.LogAttrs(
			r.Context(),
			slog.LevelError,
			"failed to normalize query string",
			slog.String("query", r.URL.RawQuery),
			slog.String("error", err.Error()),
		)

		perror.JSON(r.Context(), w, h.logger, perror.ErrorResponse{
			Message: "Failed to normalize query string",
			Code:    http.StatusInternalServerError,
		})
	}

	var (
		uri      = fmt.Sprintf("https://secure.gravatar.com/avatar/%s?%s", hash, normalizedQuery)
		cacheKey = xfnv.String(uri)
	)

	image, err := h.cache.Get(cacheKey)
	if err != nil {
		if !errors.Is(err, cache.ErrKeyNotFound) && !errors.Is(err, cache.ErrKeyExpired) {
			h.logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to get image from cache",
				slog.String("cacheKey", cacheKey),
				slog.String("error", err.Error()),
			)

			perror.JSON(r.Context(), w, h.logger, perror.ErrorResponse{
				Message: "Failed to get image from cache",
				Code:    http.StatusInternalServerError,
			})

			return
		}

		// Image not found in cache. Fetch from Gravatar.com.
		image, err = h.fetchClient.Remote(r.Context(), uri)
		if err != nil {
			h.logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to fetch image from Gravatar",
				slog.String("url", uri),
				slog.String("error", err.Error()),
			)

			perror.JSON(r.Context(), w, h.logger, perror.ErrorResponse{
				Message: "Failed to fetch Gravatar image",
				Code:    http.StatusBadGateway,
			})

			return
		}

		if err := h.cache.Set(cacheKey, image); err != nil {
			h.logger.LogAttrs(
				r.Context(),
				slog.LevelError,
				"failed to save image to cache",
				slog.String("cacheKey", cacheKey),
				slog.String("error", err.Error()),
			)

			perror.JSON(r.Context(), w, h.logger, perror.ErrorResponse{
				Message: "Failed to save image to cache",
				Code:    http.StatusInternalServerError,
			})

			return
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "inline, filename="+hash+".jpg")
	w.Header().Set("Link", "<"+uri+">; rel=\"canonical\"")

	if _, err := w.Write(image); err != nil {
		h.logger.LogAttrs(
			r.Context(),
			slog.LevelError,
			"failed to write response",
			slog.String("url", uri),
			slog.String("error", err.Error()),
		)
	}
}

// IsHexadecimal returns true if the string is a hexadecimal string.
func IsHexadecimal(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}

// NormalizeQueryString returns a normalized query string.
func NormalizeQueryString(query string) (string, error) {
	if query == "" {
		return "", nil
	}

	parsedQuery, err := url.ParseQuery(query)
	if err != nil {
		return "", fmt.Errorf("failed to parse query: %w", err)
	}

	return parsedQuery.Encode(), nil
}
