// Package web provides HTTP handlers and utilities for rendering HTML pages
// and serving static assets.
package web

import (
	"embed"
	"html/template"
	"log"

	"github.com/mateconpizza/gmweb/internal/application"
	"github.com/mateconpizza/gmweb/internal/models"
)

type OptFn func(*Opt)

type Opt struct {
	templates    *embed.FS
	static       *embed.FS
	repoLoader   func(string) (models.Repo, error)
	cacheDir     string // dataDir path where the database are found.
	itemsPerPage int
	appCfg       *application.Config
}

type Handler struct {
	*Opt
	template     *template.Template
	qrImgSize    int
	colorschemes []string
	routes       *webRoutes
}

func WithTemplates(t *embed.FS) OptFn {
	return func(o *Opt) {
		o.templates = t
	}
}

func WithStaticFiles(s *embed.FS) OptFn {
	return func(o *Opt) {
		o.static = s
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

func NewHandler(opts ...OptFn) *Handler {
	wo := &Opt{}
	for _, opt := range opts {
		opt(wo)
	}

	tmpl, err := createMainTemplate(wo.templates)
	if err != nil {
		log.Fatal(err)
	}

	themes, err := getColorschemesNames(wo.static)
	if err != nil {
		log.Fatal(err)
	}

	return &Handler{
		template:     tmpl,
		Opt:          wo,
		qrImgSize:    512,
		colorschemes: themes,
		routes:       &webRoutes{},
	}
}

type webRoutes struct{}

func (w *webRoutes) Index(s string) string { return "/" + s }

func (w *webRoutes) All(db string) string { return "/web/" + db + "/bookmarks/all" }

func (w *webRoutes) New(db string) string { return "/web/" + db + "/bookmarks/new" }

func (w *webRoutes) Detail(db, bID string) string { return "/web/" + db + "/bookmarks/detail/" + bID }

func (w *webRoutes) View(db, bID string) string { return "/web/" + db + "/bookmarks/view/" + bID }

func (w *webRoutes) Edit(db, bID string) string { return "/web/" + db + "/bookmarks/edit/" + bID }

func (w *webRoutes) QRCode(db, bID string) string { return "/web/" + db + "/bookmarks/qr/" + bID }

func (w *webRoutes) UserSignup() string { return "/user/signup" }

func (w *webRoutes) UserLogin() string { return "/user/login" }

func (w *webRoutes) UserLogout() string { return "/user/logout" }

func (w *webRoutes) Favicon() string { return "/static/img/favicon.png" }
