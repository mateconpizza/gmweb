package web

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/mateconpizza/gmweb/internal/middleware"
)

func (h *Handler) Routes(mux *http.ServeMux) {
	requireIDAndDB := func(fn http.HandlerFunc) http.Handler {
		return middleware.RequireIDAndDB(fn)
	}
	requireDB := func(fn http.HandlerFunc) http.Handler {
		return middleware.RequireDBPath(fn)
	}

	r := h.routes
	mux.Handle("GET "+r.All("{db}"), requireDB(h.index))
	mux.Handle("GET "+r.New("{db}"), requireDB(h.newRecord))
	mux.Handle("GET "+r.Detail("{db}", "{id}"), requireIDAndDB(h.recordDetail))
	mux.Handle("GET "+r.Edit("{db}", "{id}"), requireIDAndDB(h.editRecord))
	mux.Handle("GET "+r.QRCode("{db}", "{id}"), requireIDAndDB(h.showQR))

	// static and cache files
	staticFS, err := fs.Sub(h.static, "ui/static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	iconsPath := filepath.Join(h.cacheDir, "favicon")
	mux.Handle("/cache/favicon/", http.StripPrefix("/cache/favicon/", http.FileServer(http.Dir(iconsPath))))
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Load all bookmarks...")
}

func (h *Handler) showQR(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Show bookmark's QRCode...")
}

func (h *Handler) editRecord(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Show Edit bookmark's form...")
}

func (h *Handler) recordDetail(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Show Detail bookmark's form...")
}

func (h *Handler) newRecord(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Show New bookmark's form...")
}

func (h *Handler) renderPage(w http.ResponseWriter, r *http.Request, n int, page string, d *TemplateData) {
	buf := new(bytes.Buffer)

	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our responder.ServerErr() helper
	// and then return.
	err := h.template.ExecuteTemplate(buf, page, d)
	if err != nil {
		responder.ServerErr(w, r, err)
		return
	}

	// If the template is written to the buffer without any errors, we are safe
	// to go ahead and write the HTTP status code to http.ResponseWriter.
	w.WriteHeader(n)

	// Write the contents of the buffer to the http.ResponseWriter. Note: this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.
	if _, err := buf.WriteTo(w); err != nil {
		responder.ServerErr(w, r, err)
		return
	}
}
