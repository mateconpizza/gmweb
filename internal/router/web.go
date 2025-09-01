package router

// TODO: drop this...

import "fmt"

// WebRouter handles web page routing.
type WebRouter struct{ db string }

// SetDB updates the database name for web routes.
func (w *WebRouter) SetDB(db string) *WebRouter {
	w.db = db
	return w
}

func (w *WebRouter) Index(s string) string { return "/" + s }

// All Bookmarks.
func (w *WebRouter) All() string             { return w.bookmarksPath("/all") }
func (w *WebRouter) New() string             { return w.bookmarksPath("/new") }
func (w *WebRouter) NewFrame() string        { return w.bookmarksPath("/frame") }
func (w *WebRouter) Detail(id string) string { return w.bookmarksPath("/detail/" + id) }
func (w *WebRouter) View(id string) string   { return w.bookmarksPath("/view/" + id) }
func (w *WebRouter) Edit(id string) string   { return w.bookmarksPath("/edit/" + id) }
func (w *WebRouter) Import() string          { return w.bookmarksPath("/import") }
func (w *WebRouter) Export() string          { return w.bookmarksPath("/export") }
func (w *WebRouter) QRCode(id string) string { return w.bookmarksPath("/qr/" + id) }
func (w *WebRouter) Favicon() string         { return "/static/img/favicon.png" }
func (w *WebRouter) bookmarksPath(path string) string {
	return fmt.Sprintf("/web/%s/bookmarks%s", w.db, path)
}

func NewWebRoutes(db string) *WebRouter {
	if db == "" {
		panic("database name cannot be empty")
	}
	return &WebRouter{db: db}
}
