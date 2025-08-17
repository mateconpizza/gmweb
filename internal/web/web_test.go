//nolint:funlen //test
package web

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mateconpizza/gmweb/internal/application"
	"github.com/mateconpizza/gmweb/internal/database"
	"github.com/mateconpizza/gmweb/internal/models"
	"github.com/mateconpizza/gmweb/internal/models/mocks"
	"github.com/mateconpizza/gmweb/internal/router"
	"github.com/mateconpizza/gmweb/ui"
)

func setupHandler(t *testing.T, mock *mocks.Mock) *Handler {
	t.Helper()

	app := application.New()
	database.Register(mock.Name(), "")

	return NewHandler(
		WithRepoLoader(func(string) (models.Repo, error) {
			return mock, nil
		}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithFiles(&ui.Files),
		WithRoutes(router.New("{db}")),
		WithItemsPerPage(10),
		WithCfg(app.Cfg),
	)
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	t.Helper()
	ts := httptest.NewServer(h)
	return &testServer{ts}
}

//nolint:gocritic,unparam //helper
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	t.Helper()

	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rs.Body.Close(); err != nil {
			t.Error("gen QRCode: closing request body", "error", err)
		}
	}()

	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

func TestIndex(t *testing.T) {
	t.Parallel()
	m := mocks.New()
	m.Records = mocks.Bookmarks
	h := setupHandler(t, m)
	mux := http.NewServeMux()
	h.Routes(mux)

	ts := newTestServer(t, mux)
	defer ts.Close()

	h.router.SetRepo(m.Name())
	code, _, body := ts.get(t, h.router.Web.All())
	_ = body

	if code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", code)
	}

	b := mocks.Bookmarks[0]

	if !strings.Contains(body, b.URL) {
		t.Errorf("expected to contain: %q", b.URL)
	}
	if !strings.Contains(body, b.Desc) {
		t.Errorf("expected to contain: %q", b.Desc)
	}
	if !strings.Contains(body, b.Tags) {
		t.Errorf("expected to contain: %q", b.Tags)
	}
}

func TestEdit(t *testing.T) {
	t.Parallel()
	m := mocks.New()
	m.Records = mocks.Bookmarks

	h := setupHandler(t, m)
	mux := http.NewServeMux()
	h.Routes(mux)

	ts := newTestServer(t, mux)
	defer ts.Close()

	h.router.SetRepo(m.Name())

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid-ID",
			urlPath:  h.router.Web.Edit("1"),
			wantCode: http.StatusOK,
			wantBody: mocks.Bookmarks[0].URL,
		},
		{
			name:     "Non-existent-ID",
			urlPath:  h.router.Web.Edit("100000"),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "Negative-ID",
			urlPath:  h.router.Web.Edit("-1"),
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Decimal-ID",
			urlPath:  h.router.Web.Edit("1.23"),
			wantCode: http.StatusNotFound,
		},
		{
			name:     "String-ID",
			urlPath:  h.router.Web.Edit("foo"),
			wantCode: http.StatusNotFound,
		},
		{
			name:     "Empty-ID",
			urlPath:  h.router.Web.Edit(""),
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			t.Parallel()

			if code != tt.wantCode {
				t.Fatalf("expected code %d, got %d", tt.wantCode, code)
			}

			if tt.wantBody != "" {
				if !strings.Contains(body, tt.wantBody) {
					t.Errorf("expected to contain: %q", tt.wantBody)
				}
			}
		})
	}
}
