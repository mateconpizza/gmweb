package router

import "fmt"

// API provides type-safe API route functions.
type API struct {
	// General endpoints
	Index    func() string
	Scrape   func() string
	Health   func() string
	GenQR    func() string
	GenQRPNG func() string

	// Import endpoints
	ImportHTML     func() string
	ImportRepoJSON func() string
	ImportRepoGPG  func() string

	// Repository endpoints
	RepoList   func() string
	RepoAll    func() string
	RepoNew    func() string
	RepoInfo   func() string
	RepoDelete func() string

	// Bookmark endpoints
	Tags               func() string
	NewBookmark        func() string
	InternetArchiveURL func() string
	BookmarkByID       func(id string) string
	ToggleFavorite     func(id string) string
	AddVisit           func(id string) string
	UpdateBookmark     func(id string) string
	DeleteBookmark     func(id string) string
	CheckStatus        func(id string) string
	Notes              func(id string) string
}

// NewAPIRoutes creates type-safe route functions for a given database.
func NewAPIRoutes(db string) *API {
	if db == "" {
		panic("database name cannot be empty")
	}

	basePath := func(path string) string {
		return fmt.Sprintf("/api/%s%s", db, path)
	}

	bookmarksPath := func(path string) string {
		return fmt.Sprintf("/api/%s/bookmarks%s", db, path)
	}

	return &API{
		// General endpoints
		Index:    func() string { return "/api" },
		Health:   func() string { return "/api/health" },
		Scrape:   func() string { return "/api/scrape" },
		GenQR:    func() string { return "/api/qr" },
		GenQRPNG: func() string { return "/api/qr/png" },

		// Import endpoints
		ImportHTML:     func() string { return basePath("/import/html") },
		ImportRepoJSON: func() string { return basePath("/import/repojson") },
		ImportRepoGPG:  func() string { return basePath("/import/repogpg") },

		// Repository endpoints
		RepoList:   func() string { return "/api/repo/list" },
		RepoAll:    func() string { return "/api/repo/all" },
		RepoNew:    func() string { return fmt.Sprintf("/api/%s/new", db) },
		RepoInfo:   func() string { return basePath("/info") },
		RepoDelete: func() string { return basePath("/delete") },

		// Bookmark endpoints
		Tags:               func() string { return bookmarksPath("/tags") },
		NewBookmark:        func() string { return bookmarksPath("/new") },
		InternetArchiveURL: func() string { return "/api/archive" },
		BookmarkByID:       func(id string) string { return bookmarksPath("/" + id) },
		ToggleFavorite:     func(id string) string { return bookmarksPath("/" + id + "/favorite") },
		AddVisit:           func(id string) string { return bookmarksPath("/" + id + "/visit") },
		UpdateBookmark:     func(id string) string { return bookmarksPath("/" + id + "/update") },
		DeleteBookmark:     func(id string) string { return bookmarksPath("/" + id + "/delete") },
		CheckStatus:        func(id string) string { return bookmarksPath("/" + id + "/status") },
		Notes:              func(id string) string { return bookmarksPath("/" + id + "/notes") },
	}
}
