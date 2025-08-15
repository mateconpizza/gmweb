// Package api contains the API server implementation.
package api

import (
	"net/http"
)

type OptFn func(*apiOpt)

type apiOpt struct {
	router *http.ServeMux
}

type API struct {
	*apiOpt
}

func WithMux(r *http.ServeMux) OptFn {
	return func(o *apiOpt) {
		o.router = r
	}
}

func New(opts ...OptFn) *API {
	o := &apiOpt{}
	for _, optFn := range opts {
		optFn(o)
	}

	return &API{
		apiOpt: o,
	}
}
