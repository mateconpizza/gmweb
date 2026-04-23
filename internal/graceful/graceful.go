// Package graceful
package graceful

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	// cleanupFuncs holds functions to be executed before program termination.
	// Functions are executed in reverse order of registration (LIFO).
	cleanupFuncs []func() error

	// cleanupMu protects concurrent access to cleanupFuncs.
	mu sync.Mutex

	// done is closed when the shutdown sequence has completed.
	done = make(chan struct{})
)

// Listen sets up graceful shutdown with cleanup callbacks.
func Listen(ctx context.Context, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		defer close(done)

		select {
		case sig := <-sigChan:
			fmt.Println()
			slog.Info("[graceful] received interruption signal", "signal", sig)
			run()
			cancel()
		case <-ctx.Done():
			slog.Debug("[graceful] interrupt handler canceled by context")
		}
	}()
}

// Register adds a cleanup function to be called during shutdown.
// Functions are called in reverse order of registration (LIFO).
func Register(fn func() error) {
	mu.Lock()
	defer mu.Unlock()
	cleanupFuncs = append(cleanupFuncs, fn)
}

// Wait blocks until the shutdown sequence has completed.
func Wait() {
	<-done
}

// run executes all registered cleanup functions in reverse order.
func run() {
	mu.Lock()
	defer mu.Unlock()
	for i := len(cleanupFuncs) - 1; i >= 0; i-- {
		if err := cleanupFuncs[i](); err != nil {
			slog.Debug("[graceful] cleanup error", "err", err)
		}
	}
}
