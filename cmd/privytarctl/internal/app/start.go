package app

import (
	"fmt"
	"log/slog"
	"os"

	"git.sr.ht/~jamesponddotco/imgdiet-go"
	"git.sr.ht/~jamesponddotco/privytar/internal/config"
	"git.sr.ht/~jamesponddotco/privytar/internal/server"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

// ErrServerRunning is returned when the server is already running.
const ErrServerRunning xerrors.Error = "server is already running"

// StartAction is the action for the start command.
func StartAction(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("%w", ErrConfigPathRequired)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	imgdiet.Start(nil)
	defer imgdiet.Stop()

	srv, err := server.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if _, err = os.Stat(cfg.Server.PID); !os.IsNotExist(err) {
		return ErrServerRunning
	}

	pid := os.Getpid()

	pidFile, err := os.Create(cfg.Server.PID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer pidFile.Close()

	_, err = fmt.Fprintf(pidFile, "%d\n", pid)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := srv.Start(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
