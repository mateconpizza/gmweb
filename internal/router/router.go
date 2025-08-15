// Package router provides URL path generation for web and API endpoints with
// database-specific routing.
package router

import "fmt"

type Router struct {
	Web *WebRouter
	API *APIRouter
}

func New(db string) *Router {
	return &Router{
		Web: &WebRouter{db: db},
		API: &APIRouter{db: db},
	}
}

// SetRepo updates the repository name for all routes.
func (r *Router) SetRepo(db string) *Router {
	r.Web.db = db
	r.API.db = db

	return r
}

// WebRouter handles web page routing.
type WebRouter struct {
	db string
}

// SetDB updates the database name for web routes.
func (w *WebRouter) SetDB(db string) *WebRouter {
	w.db = db
	return w
}

func (w *WebRouter) Index(s string) string { return "/" + s }
func (w *WebRouter) All() string { return w.bookmarksPath("/all") }
func (w *WebRouter) New() string { return w.bookmarksPath("/new") }
func (w *WebRouter) NewFrame() string { return w.bookmarksPath("/frame") }
func (w *WebRouter) Detail(id string) string { return w.bookmarksPath("/detail/" + id) }
func (w *WebRouter) View(id string) string { return w.bookmarksPath("/view/" + id) }
func (w *WebRouter) Edit(id string) string { return w.bookmarksPath("/edit/" + id) }
func (w *WebRouter) QRCode(id string) string { return w.bookmarksPath("/qr/" + id) }
func (w *WebRouter) UserSignup() string { return "/user/signup" }
func (w *WebRouter) UserLogin() string { return "/user/login" }
func (w *WebRouter) UserLogout() string { return "/user/logout" }
func (w *WebRouter) Favicon() string { return "/static/img/favicon.png" }
func (w *WebRouter) bookmarksPath(path string) string {
	return fmt.Sprintf("/web/%s/bookmarks%s", w.db, path)
}

// APIRouter handles API endpoint routing.
type APIRouter struct {
	db string
}

// SetDB updates the database name for API routes.
func (a *APIRouter) SetDB(db string) *APIRouter {
	a.db = db
	return a
}

// General endpoints.
func (a *APIRouter) Index() string { return "/api" }
func (a *APIRouter) Scrape() string { return "/api/scrape" }
func (a *APIRouter) GenQR() string { return "/api/qr" }
func (a *APIRouter) GenQRPNG() string { return "/api/qr/png" }

// Repository endpoints.
func (a *APIRouter) RepoList() string { return "/api/repo/list" }
func (a *APIRouter) RepoAll() string { return "/api/repo/all" }
func (a *APIRouter) RepoNew() string { return fmt.Sprintf("/api/%s/new", a.db) }
func (a *APIRouter) RepoInfo() string { return a.basePath("/info") }
func (a *APIRouter) RepoDelete() string { return a.basePath("/delete") }

// Bookmark endpoints.
func (a *APIRouter) Tags() string { return a.bookmarksPath("/tags") }
func (a *APIRouter) NewBookmark() string { return a.bookmarksPath("/new") }
func (a *APIRouter) GetByID(id string) string { return a.bookmarksPath("/" + id) }
func (a *APIRouter) ToggleFavorite(id string) string { return a.bookmarksPath("/" + id + "/favorite") }
func (a *APIRouter) AddVisit(id string) string { return a.bookmarksPath("/" + id + "/visit") }
func (a *APIRouter) UpdateBookmark(id string) string { return a.bookmarksPath("/" + id + "/update") }
func (a *APIRouter) DeleteBookmark(id string) string { return a.bookmarksPath("/" + id + "/delete") }
func (a *APIRouter) CheckStatus(id string) string { return a.bookmarksPath("/" + id + "/status") }

func (a *APIRouter) InternetArchiveURL(id string) string {
	return a.bookmarksPath("/" + id + "/archive")
}

func (a *APIRouter) basePath(path string) string {
	return fmt.Sprintf("/api/%s%s", a.db, path)
}

func (a *APIRouter) bookmarksPath(path string) string {
	return fmt.Sprintf("/api/%s/bookmarks%s", a.db, path)
}
