package models

import (
	"context"

	"github.com/mateconpizza/gm/pkg/bookmark"
)

// Reader provides methods to read from the repository.
type Reader interface {
	// All returns all bookmarks.
	All() ([]*bookmark.Bookmark, error)

	// ByID returns a bookmark by its ID.
	ByID(id int) (*bookmark.Bookmark, error)

	// Has returns a bookmark by its URL and a boolean indicating if it exists.
	Has(url string) (*bookmark.Bookmark, bool)

	// Count returns the number of records in the given table.
	Count(table string) int

	// CountFavorites returns the number of favorite records.
	CountFavorites() int

	// CountTags returns tags and their counts.
	CountTags() (map[string]int, error)
}

// Writer provides methods to write, update to the repository.
type Writer interface {
	// InsertOne inserts a bookmark.
	InsertOne(ctx context.Context, b *bookmark.Bookmark) (int64, error)

	// Update updates an existing bookmark.
	Update(ctx context.Context, newB, oldB *bookmark.Bookmark) error

	// SetFavorite sets a bookmark as favorite.
	SetFavorite(ctx context.Context, b *bookmark.Bookmark) error

	// AddVisitAndUpdateCount adds a visit to a bookmark and updates its count.
	AddVisitAndUpdateCount(ctx context.Context, bID int) error

	// DeleteMany deletes multiple bookmarks.
	DeleteMany(ctx context.Context, bs []*bookmark.Bookmark) error
}

type Repo interface {
	Reader
	Writer

	// Close closes the repository.
	Close()
}
