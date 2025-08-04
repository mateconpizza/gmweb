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
	repoLoader   func(string) (*models.BookmarkModel, error)
	cacheDir     string // dataDir path where the database are found.
	itemsPerPage int
	appCfg       *application.Config
}

type Handler struct {
	*Opt
	template     *template.Template
	qrImgSize    int
	colorschemes []string
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

func WithRepoLoader(fn func(string) (*models.BookmarkModel, error)) OptFn {
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
	}
}
