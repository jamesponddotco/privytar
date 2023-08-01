package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"git.sr.ht/~jamesponddotco/privytar/internal/cache"
	"git.sr.ht/~jamesponddotco/privytar/internal/fetch"
	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
	"git.sr.ht/~jamesponddotco/xstd-go/xhash/xfnv"
	"go.uber.org/zap"
)

// AvatarHandler is the HTTP handler for the /avatar endpoint.
type AvatarHandler struct {
	fetchClient *fetch.Client
	cache       *cache.Cache
	logger      *zap.Logger
	homepage    string
}

// NewAvatarHandler returns a new AvatarHandler instance.
func NewAvatarHandler(
	homepage string,
	fetchClient *fetch.Client,
	cacheInstance *cache.Cache,
	logger *zap.Logger,
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

	if len(hash) != 32 || !IsHexadecimal(hash) {
		h.logger.Error("Invalid hash format", zap.String("hash", hash))

		perror.JSON(w, h.logger, perror.ErrorResponse{
			Message: "Invalid hash format",
			Code:    http.StatusBadRequest,
		})

		return
	}

	normalizedQuery, err := NormalizeQueryString(r.URL.RawQuery)
	if err != nil {
		h.logger.Error("Failed to normalize query string", zap.Error(err))

		perror.JSON(w, h.logger, perror.ErrorResponse{
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
			h.logger.Error("failed to get image from cache", zap.String("cacheKey", cacheKey), zap.Error(err))

			perror.JSON(w, h.logger, perror.ErrorResponse{
				Message: "Failed to get image from cache",
				Code:    http.StatusInternalServerError,
			})

			return
		}

		// Image not found in cache. Fetch from Gravatar.com.
		image, err = h.fetchClient.Remote(r.Context(), uri)
		if err != nil {
			h.logger.Error("failed to fetch image from Gravatar", zap.String("url", uri), zap.Error(err))

			perror.JSON(w, h.logger, perror.ErrorResponse{
				Message: "Failed to fetch Gravatar image",
				Code:    http.StatusBadGateway,
			})

			return
		}

		if err := h.cache.Set(cacheKey, image); err != nil {
			h.logger.Error("failed to save image to cache", zap.String("cacheKey", cacheKey), zap.Error(err))

			perror.JSON(w, h.logger, perror.ErrorResponse{
				Message: "Failed to save image to cache",
				Code:    http.StatusInternalServerError,
			})

			return
		}
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", "inline, filename="+hash+".jpg")

	if _, err := w.Write(image); err != nil {
		h.logger.Error("failed to write response", zap.String("url", uri), zap.Error(err))
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
