// Package api contains the API server implementation.
package api

import (
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	ErrWMNoVersionAvailable = errors.New("wayback machine: no version available")
	ErrServerAlreadyRunning = errors.New("server is already running")
)

type Middleware func(http.Handler) http.Handler

type OptFn func(*apiOpt)

type apiOpt struct {
	addr        string
	router      *http.ServeMux
	middlewares []Middleware
	certFile    string
	keyFile     string
}

type API struct {
	*http.Server
	isActive int32
	*apiOpt
}

func WithMux(r *http.ServeMux) OptFn {
	return func(o *apiOpt) {
		o.router = r
	}
}

func WithAddr(addr string) OptFn {
	return func(o *apiOpt) {
		o.addr = addr
	}
}

func WithMiddleware(mw ...Middleware) OptFn {
	return func(o *apiOpt) {
		o.middlewares = append(o.middlewares, mw...)
	}
}

func WithTLS(certFile, keyFile string) OptFn {
	return func(o *apiOpt) {
		o.certFile = certFile
		o.keyFile = keyFile
	}
}

func New(opts ...OptFn) *API {
	o := &apiOpt{
		router: http.NewServeMux(),
		addr:   ":8080",
	}

	for _, optFn := range opts {
		optFn(o)
	}

	var router http.Handler = o.router
	for i := len(o.middlewares) - 1; i >= 0; i-- {
		router = o.middlewares[i](router)
	}

	return &API{
		Server: &http.Server{
			Addr:         o.addr,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		apiOpt: o,
	}
}

func (a *API) Start() error {
	if !atomic.CompareAndSwapInt32(&a.isActive, 0, 1) {
		return ErrServerAlreadyRunning
	}

	slog.Info("starting server", "addr", a.addr)
	return a.ListenAndServeTLS(a.certFile, a.keyFile)
}
