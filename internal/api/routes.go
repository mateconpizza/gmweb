package api

import (
	"encoding/json"
	"net/http"

	"github.com/mateconpizza/gmweb/internal/middleware"
)

// Routes registers the routes for the API.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api", h.root)
	mux.HandleFunc("POST /api/scrape", h.scrapeData)
	mux.HandleFunc("POST /api/qr", h.genQR)
	mux.HandleFunc("POST /api/qr/png", h.genQRPNG)

	// Middleware
	requireIDAndDB := func(fn http.HandlerFunc) http.Handler {
		return middleware.RequireIDAndDB(fn)
	}
	requireDBPath := func(fn http.HandlerFunc) http.Handler {
		return middleware.RequireDBPath(fn)
	}

	// Records
	mux.Handle("GET /api/{db}/bookmarks/{id}", requireIDAndDB(h.recordByID))
	mux.Handle("GET /api/{db}/bookmarks/tags", requireDBPath(h.allTags))
	mux.Handle("POST /api/{db}/bookmarks/new", requireDBPath(h.newRecord))
	mux.Handle("PUT /api/{db}/bookmarks/{id}/favorite", requireIDAndDB(h.toggleFavorite))
	mux.Handle("PUT /api/{db}/bookmarks/{id}/visit", requireIDAndDB(h.addVisit))
	mux.Handle("PUT /api/{db}/bookmarks/{id}/update", requireIDAndDB(h.updateRecord))
	mux.Handle("DELETE /api/{db}/bookmarks/{id}/delete", requireIDAndDB(h.deleteRecord))

	// Repositories
	mux.Handle("GET /api/{db}/info", requireDBPath(h.dbInfo))
	mux.Handle("POST /api/{db}/new", requireDBPath(h.dbCreate))
	mux.Handle("DELETE /api/{db}/delete", requireDBPath(h.dbDelete))

	// Database|Repository
	mux.HandleFunc("GET /api/repo/list", h.dbList)
	mux.HandleFunc("GET /api/repo/all", h.dbInfoAll)
}

func (h *Handler) root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(h.appInfo)
}

// dbList returns the repo availables list.
func (h *Handler) dbList(w http.ResponseWriter, _ *http.Request) {}

// dbInfoAll returns all repo stats.
func (h *Handler) dbInfoAll(w http.ResponseWriter, _ *http.Request) {}

// dbInfo returns repo stats.
func (h *Handler) dbInfo(w http.ResponseWriter, r *http.Request) {}

// scrapeData scrapes a URL and returns its data.
func (h *Handler) scrapeData(w http.ResponseWriter, r *http.Request) {}

// genQR generates a QR-code as a Base64.
func (h *Handler) genQR(w http.ResponseWriter, r *http.Request) {}

// genQR generates a QR-code as a PNG file.
func (h *Handler) genQRPNG(w http.ResponseWriter, r *http.Request) {}

// dbDelete deletes (rename) the given repo name.
func (h *Handler) dbDelete(w http.ResponseWriter, r *http.Request) {}

// dbCreate creates a new repository.
func (h *Handler) dbCreate(w http.ResponseWriter, r *http.Request) {}

// recordByID returns a record with the given ID.
func (h *Handler) recordByID(w http.ResponseWriter, r *http.Request) {}

// newRecord creates a new record in the repository.
func (h *Handler) newRecord(w http.ResponseWriter, r *http.Request) {}

// updateRecord updates the given record.
func (h *Handler) updateRecord(w http.ResponseWriter, r *http.Request) {}

// deleteRecord deletes the given record id.
func (h *Handler) deleteRecord(w http.ResponseWriter, r *http.Request) {}

// addVisit adds a visit to the bookmark in the repository.
func (h *Handler) addVisit(w http.ResponseWriter, r *http.Request) {}

// toggleFavorite toggles record favorite status.
func (h *Handler) toggleFavorite(w http.ResponseWriter, r *http.Request) {}

// allTags return all tags and counts.
//
// tagName: n count
func (h *Handler) allTags(w http.ResponseWriter, r *http.Request) {}

// allTags return all tags and counts.
//
// tagName: n count
func (h *Handler) tags(w http.ResponseWriter, r *http.Request) {}
