//nolint:errcheck,err113,funlen //test
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/mateconpizza/gm/pkg/bookmark"

	"github.com/mateconpizza/gmweb/internal/middleware"
	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/models/mocks"
	"github.com/mateconpizza/gmweb/internal/responder"
	"github.com/mateconpizza/gmweb/internal/router"
)

func setupHandler(t *testing.T, mock *mocks.Mock) *Handler {
	t.Helper()

	return NewHandler(
		WithRepoLoader(func(string) (models.Repo, error) {
			return mock, nil
		}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithRoutes(router.New("{db}")),
	)
}

func parseResponseErr(t *testing.T, r *http.Response) responder.ResponseError {
	t.Helper()
	var resultErr responder.ResponseError
	_ = json.NewDecoder(r.Body).Decode(&resultErr)
	return resultErr
}

func TestAddVisit_DBError(t *testing.T) {
	t.Parallel()
	mock := mocks.New()
	h := setupHandler(t, mock)
	h.repoLoader = func(name string) (models.Repo, error) {
		return nil, errors.New("load db failed")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/visit?id=1", http.NoBody)
	w := httptest.NewRecorder()

	h.addVisit(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", res.StatusCode)
	}
}

func TestListDB(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	dbFiles := []string{"one.db", "two.db", "three.db"}
	for _, name := range dbFiles {
		fullPath := filepath.Join(dir, name)
		if err := os.WriteFile(fullPath, []byte("dummy"), 0o600); err != nil {
			t.Fatalf("error writing file: %v", err)
		}
	}

	h := NewHandler(
		WithDataDir(dir),
	)

	r := httptest.NewRequest(http.MethodGet, "/api/repo/list", http.NoBody)
	w := httptest.NewRecorder()

	h.dbList(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var result []string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("error decoding json: %v", err)
	}

	expected := map[string]bool{"one.db": true, "two.db": true, "three.db": true}
	if len(result) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(result))
	}
	for _, name := range result {
		if !expected[name] {
			t.Errorf("unexpected db name: %s", name)
		}
	}
}

func TestInfoDB_Success(t *testing.T) {
	t.Parallel()
	mock := mocks.New()
	h := setupHandler(t, mock)

	req := httptest.NewRequest(http.MethodGet, "/api/mock/info", http.NoBody)
	req.SetPathValue("db", mock.Name())
	w := httptest.NewRecorder()

	h.dbInfo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		rErr := parseResponseErr(t, resp)
		t.Fatalf("expected status 200, got %d, message error: %q", resp.StatusCode, rErr.Error)
	}

	var result responder.RepoStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.Name != mock.Name() {
		t.Errorf("expected name 'mock', got %q", result.Name)
	}
	if result.Bookmarks != 5 {
		t.Errorf("expected 5 bookmarks, got %d", result.Bookmarks)
	}
	if result.Tags != 5 {
		t.Errorf("expected 3 tags, got %d", result.Tags)
	}
	if result.Favorites != 2 {
		t.Errorf("expected 2 favorites, got %d", result.Favorites)
	}
}

func TestInfoDB_Error(t *testing.T) {
	t.Parallel()
	mock := mocks.New()
	h := setupHandler(t, mock)

	req := httptest.NewRequest(http.MethodGet, "/api/mock/info", http.NoBody)
	w := httptest.NewRecorder()
	h.dbInfo(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	resErr := parseResponseErr(t, resp)

	if resErr.Error != middleware.ErrRepoNotProvided.Error() {
		t.Fatalf("expected error mesg: %q, got %q", middleware.ErrRepoNotProvided.Error(), resErr.Error)
	}
}

func TestAllTags(t *testing.T) {
	t.Parallel()
	mock := mocks.New()
	mock.TagsCount = mocks.TagsCount
	h := setupHandler(t, mock)

	req := httptest.NewRequest(http.MethodGet, "/api/mock/bookmarks/tags", http.NoBody)
	req.SetPathValue("db", mock.Name())
	w := httptest.NewRecorder()

	h.allTags(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		rErr := parseResponseErr(t, resp)
		t.Fatalf("expected status 200, got %d, message error: %q", resp.StatusCode, rErr.Error)
	}

	var result map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("error decoding json: %v", err)
	}

	wantLen := len(mock.TagsCount)
	if wantLen != len(result) {
		t.Fatalf("want len tags map: %d, got %d", wantLen, len(result))
	}
}

func TestRecordByID_Success(t *testing.T) {
	t.Skip("not full implemented yet")
	mock := mocks.New()
	mock.Records = mocks.Bookmarks
	want := mocks.Bookmark
	wantID := strconv.Itoa(want.ID)
	mock.Records = append(mock.Records, want)

	h := setupHandler(t, mock)

	req := httptest.NewRequest(http.MethodGet, h.routes.API.GetByID(wantID), http.NoBody)
	req.SetPathValue("db", mock.Name())
	req.SetPathValue("id", wantID)

	w := httptest.NewRecorder()

	h.recordByID(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		respErr := parseResponseErr(t, resp)
		t.Fatalf("expected status 200, got %d, message error: %q", resp.StatusCode, respErr.Error)
	}

	var got bookmark.BookmarkJSON
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("error decoding json: %v", err)
	}

	if got.ID != want.ID {
		t.Fatalf("expected id %d, got %d", want.ID, got.ID)
	}
	if got.URL != want.URL {
		t.Fatalf("expected URL %s, got URL %s", want.URL, got.URL)
	}
	if got.Title != want.Title {
		t.Fatalf("expected Title %s, got Title %s", want.Title, got.Title)
	}
	if got.Desc != want.Desc {
		t.Fatalf("expected Desc %s, got Desc %s", want.Desc, got.Desc)
	}
}

func TestNewRecord_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		record      bookmark.BookmarkJSON
		wantErrMesg string
		noTags      bool
		noURL       bool
		withRecords []*bookmark.Bookmark
	}{
		{
			name:        "duplicated record",
			record:      mocks.BookmarkJSONValid,
			withRecords: mocks.Bookmarks,
			wantErrMesg: models.ErrRecordDuplicate.Error(),
		},
		{
			name:        "empty URL",
			record:      mocks.BookmarkJSONValid,
			noURL:       true,
			wantErrMesg: bookmark.ErrBookmarkURLEmpty.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mock := mocks.New()
			if len(tt.withRecords) > 0 {
				mock.Records = tt.withRecords
			}

			h := setupHandler(t, mock)
			w := httptest.NewRecorder()

			if tt.noURL {
				tt.record.URL = ""
			}

			if tt.noTags {
				tt.record.Tags = []string{}
			}

			body, err := json.Marshal(tt.record)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			path := "/api/" + mock.Name() + "/bookmarks/new"
			req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
			req.SetPathValue("db", mock.Name())

			h.newRecord(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("expected status 200, got %d", resp.StatusCode)
			}

			got := parseResponseErr(t, resp)
			if got.Error != tt.wantErrMesg {
				t.Fatalf("expected error message: %q, got %q", tt.wantErrMesg, got.Error)
			}
		})
	}
}

func TestToggleFavorite(t *testing.T) {
	t.Parallel()
	mock := mocks.New()
	mock.Records = mocks.Bookmarks
}
