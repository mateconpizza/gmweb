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

	"github.com/mateconpizza/gm/pkg/files"

	"github.com/mateconpizza/gmweb/internal/api"
	"github.com/mateconpizza/gmweb/internal/application"
	"github.com/mateconpizza/gmweb/internal/database"
	"github.com/mateconpizza/gmweb/internal/graceful"
	"github.com/mateconpizza/gmweb/internal/middleware"
	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/router"
	"github.com/mateconpizza/gmweb/internal/server"
	"github.com/mateconpizza/gmweb/internal/web"
	"github.com/mateconpizza/gmweb/ui"
)

// setupRoutes configures and returns the main HTTP router with all handlers.
func setupRoutes(app *application.App) *http.ServeMux {
	r := router.New("{db}")
	mux := http.NewServeMux()

	apiHandler := api.NewHandler(
		api.WithRepoLoader(database.Get),
		api.WithAppInfo(app.Cfg.Info),
		api.WithDataDir(app.Flags.Path),
		api.WithCacheDir(app.Cfg.CacheDir),
		api.WithLogger(app.Log),
		api.WithRoutes(r),
	)
	apiHandler.Routes(mux)

	// FIX: inject the `app` struct?
	webHandler := web.NewHandler(
		web.WithRepoLoader(database.Get),
		web.WithCacheDir(app.Cfg.CacheDir),
		web.WithFiles(&ui.Files),
		web.WithItemsPerPage(app.Server.ItemsPerPage),
		web.WithCfg(app.Cfg),
		web.WithLogger(app.Log),
		web.WithRoutes(r),
		web.WithDevMode(app.Flags.DevMode),
	)
	webHandler.Routes(mux)

	return mux
}

func setupRepos(app *application.App) error {
	paths, err := files.FindByExtList(app.Cfg.DataDir, ".db")
	if err != nil {
		return err
	}

	// first run?
	if len(paths) == 0 {
		dbPath := files.EnsureSuffix(filepath.Join(app.Cfg.DataDir, app.Cfg.MainDB), ".db")
		app.Log.Debug("first run: creating main database")
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

	app.Log = logger
	app.Log.Debug("app paths", "data", app.Cfg.DataDir, "cache", app.Cfg.CacheDir, "log", logFile)

	if err := setupRepos(app); err != nil {
		return err
	}

	middle := []server.Middleware{
		middleware.Logging,
		middleware.PanicRecover,
	}

	if !app.Flags.DevMode {
		middle = append(middle, middleware.CommonHeaders, middleware.NoSurf)
	}

	mux := setupRoutes(app)
	srv := server.New(
		server.WithAddr(app.Flags.Addr),
		server.WithLogger(app.Log),
		server.WithMux(mux),
		server.WithMiddleware(middle...),
		server.WithTLS(app.Server.CertFile, app.Server.KeyFile),
	)

	cleanups := []graceful.CleanupFunc{
		func() error {
			database.CloseAll()
			return nil
		},
		func() error {
			app.Log.Info("shutting down HTTP server")
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

	return f, logger, nil
}
