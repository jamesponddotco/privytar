// Package server provides a simple HTTP server for the Privytar service.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.sr.ht/~jamesponddotco/privytar/internal/cache"
	"git.sr.ht/~jamesponddotco/privytar/internal/config"
	"git.sr.ht/~jamesponddotco/privytar/internal/endpoint"
	"git.sr.ht/~jamesponddotco/privytar/internal/fetch"
	"git.sr.ht/~jamesponddotco/privytar/internal/perror"
	"git.sr.ht/~jamesponddotco/privytar/internal/server/handler"
	"git.sr.ht/~jamesponddotco/privytar/internal/server/middleware"
	"git.sr.ht/~jamesponddotco/xstd-go/xcrypto/xtls"
	"git.sr.ht/~jamesponddotco/xstd-go/xnet/xhttp/xmiddleware"
)

// Server represents a Privytar server.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

// New creates a new Privytar server.
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	cert, err := tls.LoadX509KeyPair(cfg.Server.TLS.Certificate, cfg.Server.TLS.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	var tlsConfig *tls.Config

	if cfg.Server.TLS.Version == "1.3" {
		tlsConfig = xtls.ModernServerConfig()
	}

	if cfg.Server.TLS.Version == "1.2" {
		tlsConfig = xtls.IntermediateServerConfig()
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	middlewares := []func(http.Handler) http.Handler{
		func(h http.Handler) http.Handler { return xmiddleware.PanicRecovery(logger, h) },
		func(h http.Handler) http.Handler { return xmiddleware.UserAgent(logger, h) },
		func(h http.Handler) http.Handler { return middleware.AcceptRequests(logger, h) },
		func(h http.Handler) http.Handler { return xmiddleware.PrivacyPolicy(cfg.Service.PrivacyPolicy, h) },
		func(h http.Handler) http.Handler { return xmiddleware.TermsOfService(cfg.Service.TermsOfService, h) },
		middleware.CORS,
	}

	if cfg.Server.LogRequests {
		accessLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		middlewares = append(middlewares, func(h http.Handler) http.Handler {
			return xmiddleware.AccessLog(accessLogger, h)
		})
	}

	var (
		cacheInstance = cache.New(cfg.Server.CacheCapacity, cfg.Server.CacheTTL)
		fetchInstance = fetch.New(cfg.Service.Name, cfg.Service.Contact)
		avatarHandler = handler.NewAvatarHandler(cfg.Service.Homepage, fetchInstance, cacheInstance, logger)
	)

	mux := http.NewServeMux()
	mux.HandleFunc(endpoint.Root, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case endpoint.Root:
			http.Redirect(w, r, cfg.Service.Homepage, http.StatusMovedPermanently)
		default:
			perror.JSON(r.Context(), w, logger, perror.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: "Page not found. Check the URL and try again.",
			})
		}
	})

	mux.Handle(endpoint.Avatar, middleware.Chain(avatarHandler, middlewares...))

	httpServer := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
	}, nil
}

// Start starts the Privytar server.
func (s *Server) Start() error {
	var (
		sigint            = make(chan os.Signal, 1)
		shutdownCompleted = make(chan struct{})
	)

	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.LogAttrs(
				ctx,
				slog.LevelError,
				"failed to shutdown server",
				slog.String("error", err.Error()),
			)
		}

		close(shutdownCompleted)
	}()

	if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	<-shutdownCompleted

	return nil
}

// Stop gracefully shuts down the Privytar server.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
