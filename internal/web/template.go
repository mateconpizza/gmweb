package web

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/justinas/nosurf"
	"github.com/mateconpizza/gm/pkg/bookmark"

	"github.com/mateconpizza/gmweb/internal/application"
	"github.com/mateconpizza/gmweb/internal/files"
	"github.com/mateconpizza/gmweb/internal/helpers"
	"github.com/mateconpizza/gmweb/internal/router"
	"github.com/mateconpizza/gmweb/ui"
)

var ErrThemeNotFound = errors.New("theme not found")

// TemplateData holds all data needed for template rendering.
type TemplateData struct {
	App           *application.Config
	Bookmark      *bookmark.Bookmark
	Bookmarks     []*bookmark.Bookmark
	CurrentYear   int
	Form          any
	FormHasErrors bool
	PageTitle     string
	CurrentURI    string
	Pagination    PaginationInfo
	Params        *RequestParams
	Routes        *router.WebRouter
	TagGroups     map[string][]string
	CSRFToken     string

	// Theme
	Colorschemes           []string
	CurrentTheme           string
	CurrentColorscheme     string
	CurrentColorschemeMode string

	// URLs
	URL         *URLs
	CurrentPath string
}

func newTemplateData(dbName string) *TemplateData {
	return &TemplateData{
		Routes:      router.New(dbName).Web,
		CurrentYear: time.Now().Year(),
	}
}

type URLs struct {
	Base           string
	Newest         string
	Oldest         string
	LastVisited    string
	Favorites      string
	MoreVisits     string
	ClearTag       string
	ClearQuery     string
	ExtensionFrame string
}

var templateFuncs = template.FuncMap{
	"title":             helpers.TitleFirstLetter,
	"itoa":              strconv.Itoa,
	"formatDate":        helpers.FormatDate,
	"formatTimestamp":   helpers.FormatTimestamp,
	"TagsWithPoundList": helpers.TagsWithPoundList,
	"tagsWithPound":     helpers.TagsWithPound,
	"RelativeISOTime":   helpers.RelativeISOTime,
	"sortCurrentDB":     helpers.SortCurrentDB,
	"shortStr":          func(s string) string { return helpers.ShortStr(s, 80) },
	"isNew":             isWithinLastWeek,
	"stripSuffix":       files.StripSuffixes,
	"now":               func() int64 { return time.Now().UnixNano() },
	"add":               func(a, b int) int { return a + b },
	"sub":               func(a, b int) int { return a - b },
	"tagURL": func(p *RequestParams, tag string, path string) string {
		return p.with().Tag(tag).Build(path)
	},
	"letterToggleURL": func(p *RequestParams, targetLetter, path string) string {
		if p.Letter == targetLetter {
			return p.with().Letter("").Build(path)
		}
		return p.with().Letter(targetLetter).Build(path)
	},
	"seq": func(start, end int) []int {
		if start > end {
			return nil
		}
		s := make([]int, end-start+1)
		for i := range s {
			s[i] = start + i
		}
		return s
	},
}

// Theme represents a single theme with its name and color schemes.
type Theme struct {
	Name  string      `json:"name"`
	Dark  ColorScheme `json:"dark"`
	Light ColorScheme `json:"light"`
}

// ColorScheme represents the color settings for a theme.
type ColorScheme struct {
	Bg string `json:"bg"`
	Fg string `json:"fg"`
}

// TemplateContext holds the context for template rendering.
type TemplateContext struct {
	App        *application.Config
	Request    *http.Request
	Bookmarks  []*bookmark.Bookmark
	Params     *RequestParams
	Routes     *router.Router
	BaseURL    string
	TagsFn     func() []string
	Pagination PaginationInfo
}

// PaginationInfo holds pagination-related data.
type PaginationInfo struct {
	CurrentPage    int
	TotalPages     int
	ItemsPerPage   int
	TotalBookmarks int
	StartIndex     int
	EndIndex       int
}

// buildIndexTemplateData constructs the data structure for template rendering.
func buildIndexTemplateData(ctx *TemplateContext) *TemplateData {
	r := ctx.Request
	p := ctx.Params

	return &TemplateData{
		App:         ctx.App,
		Bookmarks:   ctx.Bookmarks,
		Params:      p,
		PageTitle:   ctx.App.Name + ": bookmarks",
		CurrentYear: time.Now().Year(),
		Pagination:  ctx.Pagination,
		TagGroups:   helpers.GroupTagsByLetter(ctx.TagsFn()),
		CSRFToken:   nosurf.Token(r),
		CurrentPath: r.URL.Path,
		CurrentURI:  r.RequestURI,
		Routes:      ctx.Routes.Web.SetDB(p.CurrentDB),
		URL:         buildURLs(p, r),
	}
}

func buildURLs(p *RequestParams, r *http.Request) *URLs {
	path := r.URL.Path
	return &URLs{
		Base:           p.baseURL(path),
		LastVisited:    filterToggleURL(p, "last_visit", path),
		Newest:         filterToggleURL(p, "newest", path),
		Oldest:         filterToggleURL(p, "oldest", path),
		Favorites:      filterToggleURL(p, "favorites", path),
		MoreVisits:     filterToggleURL(p, "more_visits", path),
		ClearTag:       p.with().Tag("").Build(path),
		ClearQuery:     p.with().Query("").Page(1).Build(path),
		ExtensionFrame: r.URL.Query().Get("url"),
	}
}

func createMainTemplate(templates *embed.FS) (*template.Template, error) {
	return template.New("pages/base").Funcs(templateFuncs).ParseFS(templates, ui.TemplateGlob)
}

func getColorschemesNames(staticFiles *embed.FS) ([]string, error) {
	entries, err := staticFiles.ReadDir(ui.ColorSchemes)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, entry := range entries {
		if !entry.IsDir() {
			themes = append(themes, entry.Name())
		}
	}

	return themes, nil
}

func getCurrentTheme(content []byte, name string) (*Theme, error) {
	var themes []Theme
	err := json.Unmarshal(content, &themes)
	if err != nil {
		log.Fatal("error unmarshalling JSON: %w", err)
	}

	themeMap := make(map[string]*Theme, len(themes))
	for _, theme := range themes {
		themeMap[theme.Name] = &theme
	}

	theme, ok := themeMap[name]
	if !ok {
		return nil, fmt.Errorf("%w: '%s'", ErrThemeNotFound, name)
	}

	return theme, nil
}

func isWithinLastWeek(dateString string) bool {
	// Parsear la cadena de texto a un objeto time.Time.
	// Se usa time.RFC3339 para manejar el formato "2025-08-10T01:32:58Z".
	t, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		fmt.Printf("Error al parsear la fecha '%s': %v\n", dateString, err)
		return false
	}

	// Obtiene el tiempo actual del sistema.
	now := time.Now()

	// Calcula el tiempo de hace exactamente 7 días.
	sevenDaysAgo := now.AddDate(0, 0, -7)

	// Compara la fecha a chequear con la de hace 7 días.
	return t.After(sevenDaysAgo) || t.Equal(sevenDaysAgo)
}
