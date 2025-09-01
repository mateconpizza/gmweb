package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

var ErrServerAlreadyRunning = errors.New("server is already running")

type Middleware func(http.Handler) http.Handler

// Server represents an HTTP server with middleware support.
type Server struct {
	httpServer *http.Server
	isActive   int32
	certFile   string
	keyFile    string
}

// ServerOptFn defines functional options for Server configuration.
type ServerOptFn func(*serverOpt)

type serverOpt struct {
	addr        string
	router      *http.ServeMux
	logger      *slog.Logger
	middlewares []Middleware
	certFile    string
	keyFile     string
}

// WithMux sets the HTTP router/mux for the server.
func WithMux(r *http.ServeMux) ServerOptFn {
	return func(o *serverOpt) {
		o.router = r
	}
}

// WithAddr sets the server address.
func WithAddr(addr string) ServerOptFn {
	return func(o *serverOpt) {
		o.addr = addr
	}
}

// WithMiddleware adds middleware to the server.
func WithMiddleware(mw ...Middleware) ServerOptFn {
	return func(o *serverOpt) {
		o.middlewares = append(o.middlewares, mw...)
	}
}

// WithTLS configures TLS certificate and key files.
func WithTLS(certFile, keyFile string) ServerOptFn {
	return func(o *serverOpt) {
		o.certFile = certFile
		o.keyFile = keyFile
	}
}

func WithLogger(l *slog.Logger) ServerOptFn {
	return func(o *serverOpt) {
		o.logger = l
	}
}

// New creates a new Server with the given options.
func New(opts ...ServerOptFn) *Server {
	o := &serverOpt{
		router: http.NewServeMux(),
		addr:   ":8080",
	}

	if o.logger == nil {
		o.logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// Apply options
	for _, optFn := range opts {
		optFn(o)
	}

	// Apply middlewares in reverse order (last middleware wraps first)
	var handler http.Handler = o.router
	for i := len(o.middlewares) - 1; i >= 0; i-- {
		handler = o.middlewares[i](handler)
	}

	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         o.addr,
			Handler:      handler,
			ErrorLog:     slog.NewLogLogger(o.logger.Handler(), slog.LevelError),
			TLSConfig:    tlsConfig,
			IdleTimeout:  time.Minute,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		certFile: o.certFile,
		keyFile:  o.keyFile,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	if !atomic.CompareAndSwapInt32(&s.isActive, 0, 1) {
		return ErrServerAlreadyRunning
	}

	fmt.Print("starting server on ")
	slog.Info("starting server", "addr", s.httpServer.Addr)

	// Start server with or without TLS based on configuration
	if s.certFile != "" && s.keyFile != "" {
		fmt.Printf("https://localhost%s\n", s.httpServer.Addr)
		return s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile)
	}

	fmt.Printf("http://localhost%s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// GetAddr returns the server address.
func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}

// IsActive returns whether the server is currently active.
func (s *Server) IsActive() bool {
	return atomic.LoadInt32(&s.isActive) == 1
}
