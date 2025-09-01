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
	"github.com/mateconpizza/gm/pkg/files"
	"github.com/mateconpizza/gm/pkg/scraper"

	"github.com/mateconpizza/gmweb/internal/helpers"
	"github.com/mateconpizza/gmweb/internal/models"
)

var (
	ErrInvalidURLFormat = errors.New("invalid data URL format")
	ErrNonOKStatus      = errors.New("non-OK HTTP status")
)

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
	if files.Exists(savePath) && files.SizeBytes(savePath) > 0 {
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
	if files.Exists(savePath) && files.SizeBytes(savePath) > 0 {
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
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("closing request body", "error", err)
		}
	}()

	// Check HTTP status code
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s: %w", r.StatusCode, r.Status, ErrNonOKStatus)
	}

	// Create destination file
	out, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := out.Close(); err != nil {
			slog.Error("closing destination file", "error", err)
		}
	}()

	// Copy response body to file
	_, err = io.Copy(out, r.Body)
	if err != nil {
		// Clean up the file if copy failed
		_ = os.Remove(savePath)
		return "", err
	}

	return savePath, nil
}

// processFavicons updates bookmark favicon URLs, downloads favicons,
// and scrapes missing favicon URLs from a list of bookmarks.
func processFavicons(r models.Repo, destPath, staticPath string, bs []*bookmark.Bookmark) {
	var (
		wg       sync.WaitGroup
		scraped  = make(chan *bookmark.Bookmark)
		download = make(chan *bookmark.Bookmark)
	)

	// Update bookmark's favicon URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		for b := range scraped {
			if err := r.UpdateOne(context.Background(), b); err != nil {
				slog.Error("db update failed", "url", b.URL, "err", err)
				continue
			}
			download <- b // forward to download stage
		}
		close(download)
	}()

	// Download the favicons
	wg.Add(1)
	go func() {
		defer wg.Done()
		for b := range download {
			if b.FaviconURL == "" {
				continue
			}
			localFavicon, err := downloadFavicon(destPath, b.URL, b.FaviconURL)
			if err != nil {
				slog.Error("favicon fetch failed", "url", b.URL, "err", err)
				continue
			}
			favicon := filepath.Base(localFavicon)
			b.FaviconLocal = staticPath + favicon
			// could also update DB here if needed
		}
	}()

	// Producer: scrape missing FaviconURLs
	go func() {
		defer close(scraped)
		var scrapeWg sync.WaitGroup
		for _, b := range bs {
			if b.FaviconLocal != "" {
				continue
			}

			// If URL missing → scrape
			if b.FaviconURL == "" {
				scrapeWg.Add(1)
				go func(b *bookmark.Bookmark) {
					defer scrapeWg.Done()
					scrapeFaviconURL(b)
					scraped <- b
				}(b)
			} else {
				// Already has FaviconURL, just forward directly to download stage
				download <- b
			}
		}
		scrapeWg.Wait()
	}()

	wg.Wait()
}

func scrapeFaviconURL(b *bookmark.Bookmark) {
	sc := scraper.New(b.URL)
	if err := sc.Start(); err != nil {
		return
	}
	fv, err := sc.Favicon()
	if err != nil {
		return
	}
	b.FaviconURL = fv
}
