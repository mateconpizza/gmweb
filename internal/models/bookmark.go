// Package models contains the models for the application.
package models

import (
	"context"
	"errors"

	"github.com/mateconpizza/gm/pkg/bookmark"
	"github.com/mateconpizza/gm/pkg/db"
)

var (
	ErrRecordDuplicate = errors.New("record already exists")
	ErrDBNotFound      = errors.New("database not found")
	ErrURLEmpty        = errors.New("URL cannot be empty")
)

type BookmarkModel struct {
	store *db.SQLite
}

func (bm *BookmarkModel) InsertOne(ctx context.Context, b *bookmark.Bookmark) (int64, error) {
	return bm.store.InsertOne(ctx, b)
}

func (bm *BookmarkModel) Update(ctx context.Context, newB, oldB *bookmark.Bookmark) error {
	return bm.store.Update(ctx, newB, oldB)
}

func (bm *BookmarkModel) SetFavorite(ctx context.Context, b *bookmark.Bookmark) error {
	return bm.store.SetFavorite(ctx, b)
}

func (bm *BookmarkModel) AddVisit(ctx context.Context, bID int) error {
	return bm.store.AddVisit(ctx, bID)
}

func (bm *BookmarkModel) Has(ctx context.Context, url string) (*bookmark.Bookmark, bool) {
	b, ok := bm.store.Has(ctx, url)
	if !ok {
		return nil, ok
	}
	return b, ok
}

func (bm *BookmarkModel) DeleteMany(ctx context.Context, bs []*bookmark.Bookmark) error {
	return bm.store.DeleteMany(ctx, bs)
}

func (bm *BookmarkModel) ByID(ctx context.Context, bID int) (*bookmark.Bookmark, error) {
	b, err := bm.store.ByID(ctx, bID)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (bm *BookmarkModel) All(ctx context.Context) ([]*bookmark.Bookmark, error) {
	return bm.store.All(ctx)
}

func (bm *BookmarkModel) Close() {
	bm.store.Close()
}

func (bm *BookmarkModel) Name() string {
	return bm.store.Name()
}

func (bm *BookmarkModel) Count(ctx context.Context, table string) int {
	return bm.store.Count(ctx, table)
}

func (bm *BookmarkModel) CountTags(ctx context.Context) (map[string]int, error) {
	return bm.store.TagsCounter(ctx)
}

func (bm *BookmarkModel) CountFavorites(ctx context.Context) int {
	return bm.store.CountFavorites(ctx)
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
