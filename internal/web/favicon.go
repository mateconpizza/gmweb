package web

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mateconpizza/gm/pkg/bookmark"

	"github.com/mateconpizza/gmweb/internal/files"
	"github.com/mateconpizza/gmweb/internal/helpers"
)

var ErrInvalidURLFormat = errors.New("invalid data URL format")

func setHeaders(r *http.Request) {
	r.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0")
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	r.Header.Set("Accept-Language", "en-US,en;q=0.5")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Upgrade-Insecure-Requests", "1")
	r.Header.Set("Sec-Fetch-Dest", "document")
	r.Header.Set("Sec-Fetch-Mode", "navigate")
	r.Header.Set("Sec-Fetch-Site", "none")
}

// downloadFavicon fetches the favicon for a given URL and stores it locally.
// Returns the local path to use in HTML.
func downloadFavicon(destPath, bURL, faviconURL string) (string, error) {
	if faviconURL == "" {
		return faviconURL, nil
	}

	// Hash filename from domain (avoid collisions)
	hashDomain, err := helpers.HashDomain(bURL)
	if err != nil {
		return "", err
	}

	// Handle data URLs
	if strings.HasPrefix(faviconURL, "data:") {
		return handleDataURL(destPath, hashDomain, faviconURL)
	}

	// Handle regular URLs
	return handleRegularURL(destPath, hashDomain, faviconURL)
}

// handleDataURL processes data: URLs and saves them as files.
func handleDataURL(destPath, hashDomain, dataURL string) (string, error) {
	// Parse data URL: data:[<mediatype>][;base64],<data>
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("%w: %q", ErrInvalidURLFormat, dataURL)
	}

	header := parts[0]
	data := parts[1]

	c := strings.Contains

	var ext string
	switch {
	case c(header, "image/png"):
		ext = ".png"
	case c(header, "image/jpeg"), c(header, "image/jpg"):
		ext = ".jpg"
	case c(header, "image/gif"):
		ext = ".gif"
	case c(header, "image/svg"):
		ext = ".svg"
	case c(header, "image/webp"):
		ext = ".webp"
	default:
		ext = ".ico"
	}

	filename := hashDomain + ext
	savePath := filepath.Join(destPath, filename)

	// Check if file already exists
	if files.Exists(savePath) {
		return savePath, nil
	}

	// Decode base64 data (if it's base64 encoded)
	var fileData []byte
	var err error

	if strings.Contains(header, "base64") {
		fileData, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 data: %w", err)
		}
	} else {
		// URL-encoded data (less common)
		decoded, err := url.QueryUnescape(data)
		if err != nil {
			return "", fmt.Errorf("failed to decode URL data: %w", err)
		}
		fileData = []byte(decoded)
	}

	// Write to file
	err = os.WriteFile(savePath, fileData, files.FilePerm)
	if err != nil {
		return "", fmt.Errorf("failed to write favicon file: %w", err)
	}

	return savePath, nil
}

// handleRegularURL processes regular HTTP(S) URLs.
func handleRegularURL(destPath, hashDomain, faviconURL string) (string, error) {
	ext := filepath.Ext(faviconURL)
	if ext == "" || len(ext) > 5 {
		ext = ".ico"
	}

	filename := hashDomain + ext
	savePath := filepath.Join(destPath, filename)

	if files.Exists(savePath) {
		return savePath, nil
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, faviconURL, http.NoBody)
	if err != nil {
		return "", err
	}

	setHeaders(req)

	client := &http.Client{Timeout: 5 * time.Second}
	r, err := client.Do(req)
	if err != nil {
		return "", err
	}
	go func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("closing request body", "error", err)
		}
	}()

	// Create destination file
	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	go func() {
		if err := out.Close(); err != nil {
			slog.Error("closing destination file", "error", err)
		}
	}()

	_, err = io.Copy(out, r.Body)
	if err != nil {
		return "", err
	}

	return savePath, nil
}

func loadFavicons(destPath string, bs []*bookmark.Bookmark) error {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []string
	)

	for _, b := range bs {
		if b.FaviconLocal != "" || b.FaviconURL == "" {
			continue
		}

		wg.Add(1)

		go func(b *bookmark.Bookmark) {
			defer wg.Done()

			localFavicon, err := downloadFavicon(destPath, b.URL, b.FaviconURL)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				slog.Error("favicon fetch failed", "url", b.URL, "err", err)
				errs = append(errs, fmt.Sprintf("url %s: %s", b.URL, err.Error()))
			} else {
				favicon := filepath.Base(localFavicon)
				b.FaviconLocal = "/cache/favicon/" + favicon
			}
		}(b)
	}

	wg.Wait()

	for _, e := range errs {
		slog.Error("favicon local", "error", e)
	}

	return nil
}
