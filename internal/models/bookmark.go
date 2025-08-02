// Package models contains the models for the application.
package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/mateconpizza/gm/pkg/bookmark"
	"github.com/mateconpizza/gm/pkg/db"
	"github.com/mateconpizza/gm/pkg/repository"
)

var (
	ErrRecordDuplicate = errors.New("record already exists")
	ErrDBNotFound      = errors.New("database not found")
	ErrURLEmpty        = errors.New("URL cannot be empty")
)

type BookmarkModel struct {
	store *db.SQLite
}

var (
	toDBModel   = repository.ToDBModel
	fromDBModel = repository.FromDBModel
)

func (bm *BookmarkModel) Insert(ctx context.Context, b *bookmark.Bookmark) (int64, error) {
	b.GenChecksum()
	return bm.store.InsertOne(ctx, toDBModel(b))
}

func (bm *BookmarkModel) Update(ctx context.Context, newB, oldB *bookmark.Bookmark) error {
	return bm.store.Update(ctx, toDBModel(newB), repository.ToDBModel(oldB))
}

func (bm *BookmarkModel) MarkAsFavorite(ctx context.Context, b *bookmark.Bookmark) error {
	return bm.store.SetFavorite(ctx, toDBModel(b))
}

func (bm *BookmarkModel) AddVisit(ctx context.Context, bID int) error {
	return bm.store.AddVisitAndUpdateCount(ctx, bID)
}

func (bm *BookmarkModel) Has(url string) (*bookmark.Bookmark, bool) {
	b, ok := bm.store.Has(url)
	if !ok {
		return nil, ok
	}
	return fromDBModel(b), ok
}

func (bm *BookmarkModel) Delete(ctx context.Context, bs []*bookmark.Bookmark) error {
	bookmarks := make([]*db.BookmarkModel, len(bs))
	for i, m := range bs {
		bookmarks[i] = toDBModel(m)
	}

	// delete records from main table.
	if err := bm.store.DeleteMany(ctx, bookmarks); err != nil {
		return fmt.Errorf("deleting records: %w", err)
	}
	// reorder IDs from main table to avoid gaps.
	if err := bm.store.ReorderIDs(ctx); err != nil {
		return fmt.Errorf("reordering IDs: %w", err)
	}
	// recover space after deletion.
	if err := bm.store.Vacuum(ctx); err != nil {
		return fmt.Errorf("vacuum: %w", err)
	}

	return nil
}

func (bm *BookmarkModel) ByID(bID int) (*bookmark.Bookmark, error) {
	b, err := bm.store.ByID(bID)
	if err != nil {
		return nil, err
	}

	return fromDBModel(b), nil
}

func (bm *BookmarkModel) All() ([]*bookmark.Bookmark, error) {
	dbModels, err := bm.store.All()
	if err != nil {
		return nil, err
	}

	// Translate the slice of db models to a slice of domain models.
	bookmarks := make([]*bookmark.Bookmark, len(dbModels))
	for i, m := range dbModels {
		bookmarks[i] = fromDBModel(m)
	}

	return bookmarks, nil
}

func (bm *BookmarkModel) Close() {
	bm.store.Close()
}

func (bm *BookmarkModel) Name() string {
	return bm.store.Name()
}

func (bm *BookmarkModel) CountRecords(table string) int {
	return bm.store.CountRecordsFrom(table)
}

func (bm *BookmarkModel) CountTags() (map[string]int, error) {
	return bm.store.TagsCounter()
}

func (bm *BookmarkModel) CountFavorites() int {
	return bm.store.CountFavorites()
}

func New(dsn string) (*BookmarkModel, error) {
	r, err := db.New(dsn)
	if err != nil {
		return nil, err
	}

	return &BookmarkModel{store: r}, nil
}

func Initialize(ctx context.Context, dsn string) (*BookmarkModel, error) {
	r, err := db.Init(dsn)
	if err != nil {
		return nil, err
	}

	if err := r.Init(ctx); err != nil {
		return nil, err
	}

	return &BookmarkModel{store: r}, nil
}
