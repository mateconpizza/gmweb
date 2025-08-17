package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/responder"
	"github.com/mateconpizza/gmweb/internal/router"
)

var ErrPathNotFound = errors.New("path not found")

type HandlerOptFn func(*handlerOpt)

type handlerOpt struct {
	repoLoader func(string) (models.Repo, error)
	appInfo    any
	dataDir    string // dataDir path where the database are found, the home app.
	cacheDir   string // dataDir path where the database are found.
	logger     *slog.Logger
	routes     *router.Router
}

type Handler struct {
	*handlerOpt
}

func WithRepoLoader(fn func(string) (models.Repo, error)) HandlerOptFn {
	return func(o *handlerOpt) {
		o.repoLoader = fn
	}
}

func WithAppInfo(info any) HandlerOptFn {
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

func WithLogger(l *slog.Logger) HandlerOptFn {
	return func(o *handlerOpt) {
		o.logger = l
	}
}

func WithRoutes(r *router.Router) HandlerOptFn {
	return func(o *handlerOpt) {
		o.routes = r
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

func dbStats(r *http.Request, h *Handler, dbKey string) (*responder.RepoStatsResponse, error) {
	repo, err := h.repoLoader(dbKey)
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", dbKey)
		return nil, err
	}

	return &responder.RepoStatsResponse{
		Name:      dbKey,
		Bookmarks: repo.Count(r.Context(), "bookmarks"),
		Tags:      repo.Count(r.Context(), "tags"),
		Favorites: repo.CountFavorites(r.Context()),
	}, nil
}
