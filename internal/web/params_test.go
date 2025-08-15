package web

import (
	"net/url"
	"reflect"
	"testing"
)

func TestQueryValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    RequestParams
		want url.Values
	}{
		{
			name: "all fields set",
			p: RequestParams{
				Query:     "go",
				Tag:       "dev",
				FilterBy:  "title",
				Favorites: true,
				CurrentDB: "test.db",
				Letter:    "g",
				Debug:     true,
			},
			want: url.Values{
				"q":         {"go"},
				"tag":       {"dev"},
				"filter":    {"title"},
				"favorites": {"true"},
				"letter":    {"g"},
				"debug":     {"1"},
			},
		},
		{
			name: "minimal fields",
			p:    RequestParams{Query: "test"},
			want: url.Values{"q": {"test"}},
		},
	}

	//nolint:paralleltest //test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.queryValues()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
