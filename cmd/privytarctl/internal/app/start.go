package app

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"git.sr.ht/~jamesponddotco/privytar/internal/config"
	"git.sr.ht/~jamesponddotco/privytar/internal/server"
	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

// ErrServerRunning is returned when the server is already running.
const ErrServerRunning xerrors.Error = "server is already running"

// StartAction is the action for the start command.
func StartAction(ctx *cli.Context) error {
	cfg, err := config.LoadConfig(ctx.String("config"))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	logger, err := zap.NewProduction()
	if err != nil && !errors.Is(err, syscall.ENOTTY) {
		return fmt.Errorf("%w", err)
	}

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
