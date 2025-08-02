// Package graceful
package graceful

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

type CleanupFunc func() error

// Listen sets up graceful shutdown with cleanup callbacks.
func Listen(ctx context.Context, cancel context.CancelFunc, cleanups ...CleanupFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		select {
		case sig := <-sigChan:
			fmt.Println()
			slog.Info("received interruption signal", "signal", sig)
			executeCleanups(cleanups...)
			os.Exit(0)
		case <-ctx.Done():
			slog.Debug("interrupt handler canceled by context")
		}
	}()
}

func executeCleanups(cleanups ...CleanupFunc) {
	for _, fn := range cleanups {
		if err := fn(); err != nil {
			slog.Debug("[graceful] cleanup error", "err", err)
		}
	}
}
