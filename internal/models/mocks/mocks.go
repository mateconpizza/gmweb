package mocks

import (
	"context"
	"errors"

	"github.com/mateconpizza/gm/pkg/bookmark"
)

var ErrMock = errors.New("mock error")

var (
	TagsCount = map[string]int{"tag1": 1, "tag2": 2, "tag3": 3}

	Bookmark = &bookmark.Bookmark{
		ID:    4,
		URL:   "https://golang.org/",
		Title: "The Go Programming Language",
		Tags:  "golang,awesome",
		Desc:  "Go is an open source programming language that makes it simple to build secure,\nscalable systems.",
	}

	Bookmarks = []*bookmark.Bookmark{
		{ID: 1, URL: "https://example.com", Title: "Example", VisitCount: 0},
		{ID: 2, URL: "https://otherexample.com", Title: "Other Example", VisitCount: 0},
		{ID: 3, URL: "https://nonexistent.com", Title: "Non Existent"},
	}

	BookmarkJSONValid = bookmark.BookmarkJSON{
		ID: 1, URL: "https://example.com", Title: "Example", VisitCount: 0,
	}

	BookmarkJSONInvalid = bookmark.BookmarkJSON{
		ID: 1, URL: "", Title: "Example", VisitCount: 0,
	}
)

type Mock struct {
	Fail              bool
	MockSetVisitCount func(ctx context.Context, bID int) error
	Records           []*bookmark.Bookmark
	TagsCount         map[string]int
	MockHas           func(url string) (*bookmark.Bookmark, bool)
}

func (m *Mock) All(ctx context.Context) ([]*bookmark.Bookmark, error) { return m.Records, nil }

func (m *Mock) ByID(ctx context.Context, id int) (*bookmark.Bookmark, error) {
	for _, b := range m.Records {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, bookmark.ErrBookmarkNotFound
}

func (m *Mock) Has(ctx context.Context, url string) (*bookmark.Bookmark, bool) {
	if m.MockHas != nil {
		return m.MockHas(url)
	}

	for _, b := range m.Records {
		if b.URL == url {
			return b, true
		}
	}
	return nil, false
}

func (m *Mock) Count(ctx context.Context, table string) int { return 5 }

func (m *Mock) CountFavorites(ctx context.Context) int { return 2 }

func (m *Mock) CountTags(ctx context.Context) (map[string]int, error) { return m.TagsCount, nil }

func (m *Mock) Close() {}

func (m *Mock) Name() string { return "mock" }

func (m *Mock) Fullpath() string { return "/mock" }

func (m *Mock) Init(ctx context.Context) error { return nil }

func (m *Mock) InsertOne(ctx context.Context, b *bookmark.Bookmark) (int64, error) { return 0, nil }

func (m *Mock) UpdateOne(ctx context.Context, b *bookmark.Bookmark) error { return nil }

func (m *Mock) SetFavorite(ctx context.Context, b *bookmark.Bookmark) error { return nil }

func (m *Mock) AddVisit(ctx context.Context, bID int) error {
	if m.MockSetVisitCount != nil {
		return m.MockSetVisitCount(ctx, bID)
	}
	if m.Fail {
		return ErrMock
	}
	return nil
}

func (m *Mock) DeleteMany(ctx context.Context, bs []*bookmark.Bookmark) error { return nil }

func New() *Mock {
	return &Mock{}
}
