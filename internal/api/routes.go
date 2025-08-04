package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mateconpizza/gm/pkg/bookmark"
	"github.com/mateconpizza/gm/pkg/scraper"

	"github.com/mateconpizza/gmweb/internal/database"
	"github.com/mateconpizza/gmweb/internal/files"
	"github.com/mateconpizza/gmweb/internal/helpers"
	"github.com/mateconpizza/gmweb/internal/middleware"
	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/qr"
	"github.com/mateconpizza/gmweb/internal/responder"
)

// Routes registers the routes for the API.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api", h.index)
	mux.HandleFunc("POST /api/scrape", h.scrapeData)
	mux.HandleFunc("POST /api/qr", h.genQR)
	mux.HandleFunc("POST /api/qr/png", h.genQRPNG)

	// Database|Repository
	mux.HandleFunc("GET /api/repo/list", h.dbList)
	mux.HandleFunc("GET /api/repo/all", h.dbInfoAll)

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
	mux.HandleFunc("POST /api/{db}/new", h.dbCreate)
	mux.Handle("DELETE /api/{db}/delete", requireDBPath(h.dbDelete))
}

func (h *Handler) index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(h.appInfo); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
	}
}

// dbList returns the repo availables list.
func (h *Handler) dbList(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// FIX: add `.enc` extension
	paths, err := files.FindByExtList(h.dataDir, ".db")
	if err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	dbs := make([]string, 0, len(paths))
	for i := range paths {
		dbs = append(dbs, filepath.Base(paths[i]))
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(dbs); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// dbInfoAll returns all repo stats.
func (h *Handler) dbInfoAll(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := make([]*responder.RepoStatsResponse, 0, len(database.Valid))
	for k := range database.Valid {
		stat, err := dbStats(h, k)
		if err != nil {
			responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		stats = append(stats, stat)
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// dbInfo returns repo stats.
func (h *Handler) dbInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats := &responder.RepoStatsResponse{
		Name:      dbName,
		Bookmarks: repo.Count("bookmarks"),
		Tags:      repo.Count("tags"),
		Favorites: repo.CountFavorites(),
	}

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// scrapeData scrapes a URL and returns its data.
func (h *Handler) scrapeData(w http.ResponseWriter, r *http.Request) {
	// TODO: add URL validation
	w.Header().Set("Content-Type", "application/json")

	u := r.URL.Query().Get("url")
	if u == "" {
		slog.Debug("fetching URL", "error", "empty URL")
		responder.EncodeErrJSON(w, http.StatusBadRequest, "empty URL")
		return
	}

	slog.Debug("fetching URL", "url", u)

	sc := scraper.New(u)
	if err := sc.Start(); err != nil {
		slog.Error("scrape new bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	b := bookmark.NewJSON()
	b.Title, _ = sc.Title()
	b.Desc, _ = sc.Desc()
	k, _ := sc.Keywords()
	b.Tags = strings.Split(k, ",")
	b.FaviconURL, _ = sc.Favicon()

	w.WriteHeader(http.StatusOK)

	responsePayload := &responder.FetchDataResponse{
		Title:      b.Title,
		Desc:       b.Desc,
		Tags:       b.Tags,
		FaviconURL: b.FaviconURL,
	}

	err := json.NewEncoder(w).Encode(responsePayload)
	if err != nil {
		slog.Error("fetching response: failed to encode JSON", "error", err, "url", u)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
	}
}

// genQR generates a QR-code as a Base64.
func (h *Handler) genQR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req := &responder.QRCodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		slog.Error("gen QRCode", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("gen QRCode: closing request body", "error", err)
		}
	}()

	q := qr.New(req.URL, req.Size)
	if err := q.Generate(); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	img, err := q.ImageBase64()
	if err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := &responder.QRCodeResponse{
		URL:    req.URL,
		Base64: img,
		MIME:   "image/png",
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("gen QRCode: failed to encode JSON", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// genQR generates a QR-code as a PNG file.
func (h *Handler) genQRPNG(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req := &responder.QRCodeRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		slog.Error("gen QRCode", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("gen QRCode: closing request body", "error", err)
		}
	}()

	q := qr.New(req.URL, req.Size)
	if err := q.Generate(); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	img, err := q.ImagePNG("", "")
	if err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(img.Bytes()); err != nil {
		w.Header().Set("Content-Type", "application/json")
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
	}
}

// dbDelete deletes (rename) the given repo name.
func (h *Handler) dbDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	repo, err := h.repoLoader(r.PathValue("db"))
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", r.PathValue("db"))
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	repo.Close()
	database.Forget(r.PathValue("db"))

	// FIX: should i use os.Root?
	dbName := files.EnsureSuffix(r.PathValue("db"), ".db")
	dbPath := filepath.Clean(filepath.Join(h.dataDir, dbName))
	if err := files.Rename(dbPath, dbName+".bk"); err != nil {
		slog.Error("renaming repo", "err", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := &responder.ResponseData{
		Message:    "database deleted: " + dbName,
		StatusCode: http.StatusOK,
	}

	database.Forget(dbName)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// dbCreate creates a new repository.
func (h *Handler) dbCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbParam := r.PathValue("db")
	if dbParam == "" {
		slog.Error("create database: no name provided")
		responder.EncodeErrJSON(w, http.StatusInternalServerError, "no name provide")
		return
	}

	newDBName := files.EnsureSuffix(r.PathValue("db"), ".db")
	dbPath := filepath.Join(h.dataDir, newDBName)
	dbPath = filepath.Clean(dbPath)
	_, err := models.Initialize(context.Background(), dbPath)
	if err != nil {
		slog.Error("creating database", "error", err, "db", newDBName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	database.Register(r.PathValue("db"), dbPath)

	w.WriteHeader(http.StatusCreated)
	res := &responder.ResponseData{
		Message:    fmt.Sprintf("New database %q successfully created", newDBName),
		StatusCode: http.StatusOK,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
	}
}

// recordByID returns a record with the given ID.
func (h *Handler) recordByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	idStr := r.PathValue("id")
	bID, _ := strconv.Atoi(idStr)
	b, err := repo.ByID(bID)
	if err != nil {
		slog.Error("deleting bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := json.NewEncoder(w).Encode(b); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// newRecord creates a new record in the repository.
func (h *Handler) newRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	bj := bookmark.NewJSON()
	if err := json.NewDecoder(r.Body).Decode(bj); err != nil {
		slog.Error("creating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("creating bookmark: closing request body", "error", err)
		}
	}()

	if bj.URL == "" {
		responder.EncodeErrJSON(w, http.StatusBadRequest, models.ErrURLEmpty.Error())
		return
	}

	bj.URL = strings.TrimSuffix(bj.URL, "/")
	if _, exists := repo.Has(bj.URL); exists {
		slog.Error("creating bookmark", "error", models.ErrRecordDuplicate)
		responder.EncodeErrJSON(w, http.StatusBadRequest, models.ErrRecordDuplicate.Error())
		return
	}

	rawURL := bj.URL
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		responder.EncodeErrJSON(w, http.StatusBadRequest, err.Error())
	}

	newB := bookmark.NewFromJSON(bj)
	newB.URL = u.String()

	if err := bookmark.Validate(newB); err != nil {
		slog.Error("creating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := repo.InsertOne(context.Background(), newB); err != nil {
		slog.Error("creating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	res := &responder.ResponseData{
		Message:    "New bookmark successfully created",
		StatusCode: http.StatusOK,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.Error("creating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
	}
}

// updateRecord updates the given record.
func (h *Handler) updateRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	bj := bookmark.NewJSON()
	if err := json.NewDecoder(r.Body).Decode(bj); err != nil {
		slog.Error("updating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			slog.Error("updating bookmark: closing request body", "error", err)
		}
	}()

	newB := bookmark.NewFromJSON(bj)
	oldB, err := repo.ByID(bj.ID)
	if err != nil {
		slog.Error("updating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := bookmark.Validate(newB); err != nil {
		slog.Error("updating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if bytes.Equal(newB.Buffer(), oldB.Buffer()) {
		w.WriteHeader(http.StatusBadRequest)
		responder.EncodeErrJSON(w, http.StatusBadRequest, "no changes found")
		return
	}

	newB.CreatedAt = oldB.CreatedAt
	newB.LastVisit = oldB.LastVisit
	newB.Favorite = oldB.Favorite
	newB.VisitCount = oldB.VisitCount
	if oldB.URL == newB.URL {
		newB.FaviconURL = oldB.FaviconURL
		newB.FaviconLocal = oldB.FaviconLocal
	}

	if err := repo.Update(context.Background(), newB, oldB); err != nil {
		slog.Error("updating bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := &responder.ResponseData{
		Message:    "Bookmark updated successfully!",
		StatusCode: http.StatusOK,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
	}
}

// deleteRecord deletes the given record id.
func (h *Handler) deleteRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("delete: repo loader", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	idStr := r.PathValue("id")
	bID, _ := strconv.Atoi(idStr)

	b, err := repo.ByID(bID)
	if err != nil {
		slog.Error("delete: getting by ID", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	slog.Debug("delete: bookmark", "id", b.ID)

	if err := repo.DeleteMany(context.Background(), []*bookmark.Bookmark{b}); err != nil {
		slog.Error("deleting bookmark", "error", err)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	res := &responder.ResponseData{Message: "Bookmark deleted successfully!", StatusCode: http.StatusOK}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// addVisit adds a visit to the bookmark in the repository.
func (h *Handler) addVisit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("toggle favorite", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	idStr := r.PathValue("id")
	bID, _ := strconv.Atoi(idStr)

	if err := repo.AddVisitAndUpdateCount(context.Background(), bID); err != nil {
		slog.Error("visit update", "error", err, "id", bID)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)

	res := &responder.ResponseData{
		Message:    fmt.Sprintf("Bookmark id=%d: visited on database %q", bID, dbName),
		StatusCode: http.StatusOK,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// toggleFavorite toggles record favorite status.
func (h *Handler) toggleFavorite(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("toggle favorite", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	idStr := r.PathValue("id")
	bID, _ := strconv.Atoi(idStr)
	b, err := repo.ByID(bID)
	if err != nil {
		slog.Error("toggle favorite", "error", err, "id", bID)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	b.Favorite = !b.Favorite
	status := "Unfavorited"
	if b.Favorite {
		status = "Favorited"
	}

	if err := repo.SetFavorite(context.Background(), b); err != nil {
		slog.Error("toggle favorite", "error", err, "id", bID)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	res := &responder.ResponseData{
		Message:    fmt.Sprintf("Bookmark %q: %s", helpers.ShortStr(b.URL, 60), status),
		StatusCode: http.StatusOK,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		slog.Error("toggle favorite", "error", err, "id", bID)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// allTags return all tags and counts.
//
// tagName: n count.
func (h *Handler) allTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbName := r.PathValue("db")
	repo, err := h.repoLoader(dbName)
	if err != nil {
		slog.Error("listing bookmarks", "error", err, "db", dbName)
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	tags, err := repo.CountTags()
	if err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		responder.EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}
