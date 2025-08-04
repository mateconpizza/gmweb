package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mateconpizza/gmweb/internal/api"
	"github.com/mateconpizza/gmweb/internal/application"
	"github.com/mateconpizza/gmweb/internal/database"
	"github.com/mateconpizza/gmweb/internal/files"
	"github.com/mateconpizza/gmweb/internal/graceful"
	"github.com/mateconpizza/gmweb/internal/middleware"
	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/server"
	"github.com/mateconpizza/gmweb/internal/web"
)

// setupRoutes configures and returns the main HTTP router with all handlers.
func setupRoutes(app *application.App) *http.ServeMux {
	router := http.NewServeMux()

	apiHandler := api.NewHandler(
		api.WithRepoLoader(database.Get),
		api.WithAppInfo(app.Cfg.Info),
		api.WithDataDir(app.Flags.Path),
		api.WithCacheDir(app.Cfg.CacheDir),
	)
	apiHandler.Routes(router)

	// FIX: inject the `app` struct?
	webHandler := web.NewHandler(
		web.WithRepoLoader(database.Get),
		web.WithCacheDir(app.Cfg.CacheDir),
		web.WithStaticFiles(app.Server.StaticFiles),
		web.WithTemplates(app.Server.TemplatesFiles),
		web.WithItemsPerPage(app.Server.ItemsPerPage),
		web.WithCfg(app.Cfg),
	)
	webHandler.Routes(router)

	return router
}

func setupRepos(app *application.App) error {
	paths, err := files.FindByExtList(app.Cfg.DataDir, ".db")
	if err != nil {
		return err
	}

	// first run?
	if len(paths) == 0 {
		slog.Debug("first run: create main database")
		dbPath := filepath.Join(app.Cfg.DataDir, app.Cfg.MainDB)
		_, err := models.Initialize(context.Background(), dbPath)
		if err != nil {
			return err
		}

		paths = append(paths, dbPath)
	}

	for _, p := range paths {
		database.Register(files.StripSuffixes(filepath.Base(p)), p)
	}

	return nil
}

// run starts the server and handles graceful shutdown.
func run(app *application.App) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := setupRepos(app); err != nil {
		return err
	}

	logFile := filepath.Join(app.Flags.Path, "log.json")
	f, logger, err := setupLogger(logFile, app.Flags.Verbose)
	if err != nil {
		return err
	}

	defer func() {
		slog.Info("closing logfile")
		if err := f.Close(); err != nil {
			slog.Error("failed to close logfile", "err", err)
		}
	}()

	mux := setupRoutes(app)
	srv := server.New(
		server.WithAddr(app.Flags.Addr),
		server.WithLogger(logger),
		server.WithMux(mux),
		server.WithMiddleware(
			middleware.CommonHeaders,
			middleware.Logging,
			middleware.PanicRecover,
		),
		server.WithTLS(app.Server.CertFile, app.Server.KeyFile),
	)

	cleanups := []graceful.CleanupFunc{
		func() error {
			database.CloseAll()
			return nil
		},
		func() error {
			slog.Info("shutting down HTTP server")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			return srv.Shutdown(shutdownCtx)
		},
	}

	graceful.Listen(ctx, cancel, cleanups...)

	return srv.Start()
}

func setupLogger(fn string, verbosity int) (*os.File, *slog.Logger, error) {
	levels := []slog.Level{
		slog.LevelError,
		slog.LevelWarn,
		slog.LevelInfo,
		slog.LevelDebug,
	}
	level := levels[min(max(verbosity, 0), len(levels)-1)]

	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, files.FilePerm)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	multiWriter := io.MultiWriter(f, os.Stdout)

	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				return slog.String("time", a.Value.Time().Format(time.RFC3339))
			case slog.LevelKey:
				return slog.String("level", strings.ToLower(a.Value.String()))
			case slog.SourceKey:
				if source, ok := a.Value.Any().(*slog.Source); ok {
					dir, file := filepath.Split(source.File)
					shortFile := filepath.Join(filepath.Base(filepath.Clean(dir)), file)
					return slog.String("src", fmt.Sprintf("%s:%d", shortFile, source.Line))
				}
			case slog.MessageKey:
				return a
			}
			return a
		},
	}))

	slog.SetDefault(logger)
	slog.Debug("logging initialized", "level", level)

	return f, logger, nil
}
