package main

import (
	"context"
	"log/slog"
	"net/http"
	"path/filepath"
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

	mux := setupRoutes(app)
	srv := server.New(
		server.WithAddr(app.Flags.Addr),
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
