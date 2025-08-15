// Package middleware contains middleware functions for the HTTP server.
package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/justinas/nosurf"

	"github.com/mateconpizza/gmweb/internal/database"
	"github.com/mateconpizza/gmweb/internal/responder"
)

const (
	chromeExtID  = "mllmmjngfojnegaaaidjdmcpaknhjcjb"
	firefoxExtID = "78fc8a7e-8e4f-4ab6-a941-f37cb9617369"
)

var (
	ErrRepoNotProvided = errors.New("repo name not provided")
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
		basePolicy := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
		frameAncestors := fmt.Sprintf(
			"frame-ancestors 'self' chrome-extension://%s moz-extension://%s",
			chromeExtID,
			firefoxExtID,
		)
		cspHeader := fmt.Sprintf("%s; %s", basePolicy, frameAncestors)
		w.Header().Set("Content-Security-Policy", cspHeader)

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Server", "Go")
		next.ServeHTTP(w, r)
	})
}

func RequireDBParam(next http.Handler) http.Handler {
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

func RequireIDAndDBParam(next http.Handler) http.Handler {
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

// NoSurf uses a customized CSRF cookie with the Secure, Path and HttpOnly
// attributes set.
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
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
		return fmt.Errorf("%w: '%s'", database.ErrDBNotFound, dbParam)
	}

	return nil
}
