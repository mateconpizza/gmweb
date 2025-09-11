package models

import (
	"context"

	"github.com/mateconpizza/gm/pkg/bookmark"
)

// Reader provides methods to read from the repository.
type Reader interface {
	// All returns all bookmarks.
	All(ctx context.Context) ([]*bookmark.Bookmark, error)

	// ByID returns a bookmark by its ID.
	ByID(ctx context.Context, id int) (*bookmark.Bookmark, error)

	// Has returns a bookmark by its URL and a boolean indicating if it exists.
	Has(ctx context.Context, url string) (*bookmark.Bookmark, bool)

	// Count returns the number of records in the given table.
	Count(ctx context.Context, table string) int

	// CountFavorites returns the number of favorite records.
	CountFavorites(ctx context.Context) int

	// CountTags returns tags and their counts.
	CountTags(ctx context.Context) (map[string]int, error)
}

// Writer provides methods to write, update to the repository.
type Writer interface {
	// InsertOne inserts a bookmark.
	InsertOne(ctx context.Context, b *bookmark.Bookmark) (int64, error)

	// InsertMany inserts multiple bookmarks.
	InsertMany(ctx context.Context, bs []*bookmark.Bookmark) error

	// Update updates an existing bookmark.
	UpdateOne(ctx context.Context, b *bookmark.Bookmark) error

	// UpdateNotes updates the bookmak's notes.
	UpdateNotes(ctx context.Context, bID int, notes string) error

	// SetFavorite sets a bookmark as favorite.
	SetFavorite(ctx context.Context, b *bookmark.Bookmark) error

	// AddVisitAndUpdateCount adds a visit to a bookmark and updates its count.
	AddVisit(ctx context.Context, bID int) error

	// DeleteMany deletes multiple bookmarks.
	DeleteMany(ctx context.Context, bs []*bookmark.Bookmark) error
}

type Repo interface {
	Reader
	Writer

	// Close closes the repository.
	Close()
}
