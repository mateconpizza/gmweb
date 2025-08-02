package api

import (
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/mateconpizza/gmweb/internal/files"
	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/responder"
)

var ErrPathNotFound = errors.New("path not found")

type HandlerOptFn func(*handlerOpt)

type handlerOpt struct {
	repoLoader    func(string) (*models.BookmarkModel, error)
	defaultDBName string
	appInfo       any
	dataDir       string // dataDir path where the database are found, the home app.
	cacheDir      string // dataDir path where the database are found.
}

type Handler struct {
	*handlerOpt
}

func WithRepoLoader(fn func(string) (*models.BookmarkModel, error)) HandlerOptFn {
	return func(o *handlerOpt) {
		o.repoLoader = fn
	}
}

func WithDefaultDBName(s string) HandlerOptFn {
	return func(o *handlerOpt) {
		o.defaultDBName = s
	}
}

func WithInformation(info any) HandlerOptFn {
	return func(o *handlerOpt) {
		o.appInfo = info
	}
}

func WithDataDir(path string) HandlerOptFn {
	return func(o *handlerOpt) {
		o.dataDir = path
	}
}

func WithCacheDir(s string) HandlerOptFn {
	return func(o *handlerOpt) {
		o.cacheDir = s
	}
}

func NewHandler(opts ...HandlerOptFn) *Handler {
	ao := &handlerOpt{}
	for _, opt := range opts {
		opt(ao)
	}

	return &Handler{
		handlerOpt: ao,
	}
}

func dbStats(h *Handler, path string) (*responder.RepoStatsResponse, error) {
	repo, err := h.repoLoader(path)
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", filepath.Base(path))
		return nil, err
	}
	defer repo.Close()

	return &responder.RepoStatsResponse{
		Name:      repo.Name(),
		Bookmarks: repo.CountRecords("bookmarks"),
		Tags:      repo.CountRecords("tags"),
		Favorites: repo.CountFavorites(),
	}, nil
}

func extractDBName(r *http.Request, def string) string {
	dbName := r.URL.Query().Get("db")
	if dbName == "" {
		slog.Debug("no database name provided, using main database")
		dbName = def
	}

	return files.EnsureSuffix(dbName, ".db")
}
