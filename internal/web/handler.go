// Package web provides HTTP handlers and utilities for rendering HTML pages
// and serving static assets.
package web

import (
	"embed"
	"html/template"
	"log"
	"log/slog"

	"github.com/mateconpizza/gmweb/internal/application"
	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/router"
)

type OptFn func(*Opt)

type Opt struct {
	files        *embed.FS
	repoLoader   func(string) (models.Repo, error)
	cacheDir     string // dataDir path where the database are found.
	itemsPerPage int
	appCfg       *application.Config
	logger       *slog.Logger
	router       *router.Router
}

type Handler struct {
	*Opt
	template     *template.Template
	qrImgSize    int
	colorschemes []string
}

func WithFiles(f *embed.FS) OptFn {
	return func(o *Opt) {
		o.files = f
	}
}

func WithDevMode(enabled bool) OptFn {
	return func(o *Opt) {
		devMode = enabled
	}
}

func WithRepoLoader(fn func(string) (models.Repo, error)) OptFn {
	return func(o *Opt) {
		o.repoLoader = fn
	}
}

func WithCacheDir(s string) OptFn {
	return func(o *Opt) {
		o.cacheDir = s
	}
}

func WithItemsPerPage(n int) OptFn {
	return func(o *Opt) {
		o.itemsPerPage = n
	}
}

func WithCfg(info *application.Config) OptFn {
	return func(o *Opt) {
		o.appCfg = info
	}
}

func WithLogger(l *slog.Logger) OptFn {
	return func(o *Opt) {
		o.logger = l
	}
}

func WithRoutes(r *router.Router) OptFn {
	return func(o *Opt) {
		o.router = r
	}
}

func NewHandler(opts ...OptFn) *Handler {
	wo := &Opt{}
	for _, opt := range opts {
		opt(wo)
	}

	tmpl, err := createMainTemplate(wo.files)
	if err != nil {
		log.Fatal(err)
	}

	themes, err := getColorschemesNames(wo.files)
	if err != nil {
		log.Fatal(err)
	}

	return &Handler{
		template:     tmpl,
		Opt:          wo,
		qrImgSize:    512,
		colorschemes: themes,
	}
}
