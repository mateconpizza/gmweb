package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// RequestParams holds the query parameters from the request.
type RequestParams struct {
	CurrentDB string
	Debug     bool
	Favorites bool
	FilterBy  string
	Letter    string
	ReturnURL string
	Page      int
	Query     string
	Tag       string
}

func (p *RequestParams) with() *ParamsBuilder {
	return &ParamsBuilder{params: *p}
}

// queryValues returns the query parameters as a map.
func (p *RequestParams) queryValues() url.Values {
	q := url.Values{}

	if p.Debug {
		q.Set("debug", "1")
	}
	if p.Favorites {
		q.Set("favorites", "true")
	}
	if p.FilterBy != "" {
		q.Set("filter", p.FilterBy)
	}
	if p.Letter != "" {
		q.Set("letter", p.Letter)
	}
	if p.Query != "" {
		q.Set("q", p.Query)
	}
	if p.Tag != "" {
		q.Set("tag", p.Tag)
	}

	return q
}

// paginationURL returns a URL with preserved filters and the given page number.
func (p *RequestParams) paginationURL(path string, page int) string {
	q := p.queryValues()
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}

	return path + "?" + q.Encode()
}

func (p *RequestParams) IsFilterActive(name string) bool {
	return p.FilterBy == name
}

// baseURL constructs the base URL for pagination, preserving query/tag/filter.
func (p *RequestParams) baseURL(path string) string {
	return path + "?" + p.queryValues().Encode()
}

func parseRequestParams(r *http.Request) *RequestParams {
	q := r.URL.Query()

	debug := q.Get("debug") == "1"
	filterBy := q.Get("filter")
	letter := q.Get("letter")
	queryStr := q.Get("q")
	tag := q.Get("tag")
	returnURL := q.Get("returnTo")

	// tag selected from dropdown autocmp
	if tag == "" && strings.HasPrefix(queryStr, "#") {
		tag = queryStr
		queryStr = ""
	}

	// parse page
	currentPage, err := strconv.Atoi(q.Get("page"))
	if err != nil || currentPage < 1 {
		currentPage = 1
	}

	pb := &ParamsBuilder{}
	return pb.
		Tag(tag).
		Query(queryStr).
		Filter(filterBy).
		Return(returnURL).
		Letter(letter).
		Page(currentPage).
		Debug(debug).
		BuildParams()
}

type ParamsBuilder struct {
	params RequestParams
}

func (b *ParamsBuilder) Query(q string) *ParamsBuilder {
	b.params.Query = q
	return b
}

func (b *ParamsBuilder) Filter(filter string) *ParamsBuilder {
	b.params.FilterBy = filter
	return b
}

func (b *ParamsBuilder) Return(q string) *ParamsBuilder {
	b.params.ReturnURL = q
	return b
}

func (b *ParamsBuilder) Tag(tag string) *ParamsBuilder {
	b.params.Tag = tag
	return b
}

func (b *ParamsBuilder) Page(page int) *ParamsBuilder {
	b.params.Page = page
	return b
}

func (b *ParamsBuilder) Debug(debug bool) *ParamsBuilder {
	b.params.Debug = debug
	return b
}

func (b *ParamsBuilder) Favorite(favorite bool) *ParamsBuilder {
	b.params.Favorites = favorite
	return b
}

func (b *ParamsBuilder) Letter(letter string) *ParamsBuilder {
	b.params.Letter = letter
	return b
}

func (b *ParamsBuilder) Database(database string) *ParamsBuilder {
	b.params.CurrentDB = database
	return b
}

func (b *ParamsBuilder) Build(path string) string {
	return b.params.paginationURL(path, b.params.Page)
}

func (b *ParamsBuilder) BuildParams() *RequestParams {
	c := b.params
	return &c
}
