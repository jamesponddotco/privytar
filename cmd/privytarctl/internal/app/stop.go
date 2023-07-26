package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.sr.ht/~jamesponddotco/privytar/internal/config"
	"github.com/urfave/cli/v2"
)

// StopAction is the action for the stop command.
func StopAction(ctx *cli.Context) error {
	cfg, err := config.LoadConfig(ctx.String("config"))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	pidFileData, err := os.ReadFile(cfg.Server.PID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidFileData)))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err = process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err = os.Remove(cfg.Server.PID); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
