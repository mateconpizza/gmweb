// Package middleware contains middleware functions for the HTTP server.
package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mateconpizza/gmweb/internal/database"
	"github.com/mateconpizza/gmweb/internal/responder"
)

var (
	ErrRepoNotProvided = errors.New("repo not provided")
	ErrRepoInvalidPath = errors.New("invalid repo path")
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		if strings.HasPrefix(r.URL.Path, "/static/") || strings.HasPrefix(r.URL.Path, "/cache/favicon") {
			return
		}

		slog.Info(
			"request",
			"status",
			wrapped.statusCode,
			"method",
			r.Method,
			"path",
			r.URL.Path,
			"duration",
			time.Since(start),
		)
	})
}

func CommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().
			Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Server", "Go")
		next.ServeHTTP(w, r)
	})
}

func RequireDBPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbParam := r.PathValue("db")

		if err := validateDBParam(dbParam); err != nil {
			slog.Error("db validation failed", "error", err, "dbParam", dbParam)
			responder.EncodeErrJSON(w, http.StatusBadRequest, err.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireIDAndDB(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			responder.EncodeErrJSON(w, http.StatusBadRequest, "id not provided")
			return
		}

		bID, err := strconv.Atoi(idStr)
		if err != nil || bID < 1 {
			responder.EncodeErrJSON(w, http.StatusNotFound, "not found")
			return
		}

		dbParam := r.PathValue("db")
		if err := validateDBParam(dbParam); err != nil {
			slog.Error("db validation failed", "error", err, "path", dbParam)
			responder.EncodeErrJSON(w, http.StatusBadRequest, err.Error())
			return
		}

		next.ServeHTTP(w, r)
	})
}

func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		usernameHash := sha256.Sum256([]byte(username))
		passwordHash := sha256.Sum256([]byte(password))
		expectedUsernameHash := sha256.Sum256([]byte("pepe"))
		expectedPasswordHash := sha256.Sum256([]byte("hongo"))

		usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
		passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

		if usernameMatch && passwordMatch {
			next.ServeHTTP(w, r)
			return
		}
	})
}

func PanicRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func validateDBParam(dbParam string) error {
	if dbParam == "" {
		return ErrRepoNotProvided
	}

	// Path traversal protection
	cleanDB := filepath.Clean(dbParam)
	if cleanDB != dbParam || strings.Contains(dbParam, "/") {
		return ErrRepoInvalidPath
	}

	// Business logic validation
	if !database.IsValid(cleanDB) {
		return database.ErrDBNotAllowed
	}

	return nil
}
