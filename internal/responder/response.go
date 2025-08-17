package responder

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ResponseData struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

type ResponseError struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
}

type FetchDataResponse struct {
	Title      string   `json:"title"`
	Desc       string   `json:"desc"`
	Tags       []string `json:"tags"`
	FaviconURL string   `json:"favicon_url"`
	ArchiveURL string   `json:"archive_url"`
}

type QRCodeResponse struct {
	URL     string `json:"url"`
	Base64  string `json:"base64"`
	MIME    string `json:"mime"`
	Message string `json:"message,omitempty"`
}

type QRCodeRequest struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
}

type RepoStatsResponse struct {
	Name      string `json:"name"`
	Bookmarks int    `json:"bookmarks"`
	Tags      int    `json:"tags"`
	Favorites int    `json:"favorites"`
}

func EncodeErrJSON(w http.ResponseWriter, statusCode int, err string) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(&ResponseError{Error: err, StatusCode: statusCode}); err != nil {
		slog.Error("encoding error", "error", err)
	}
}

func ServerErr(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	slog.Error(err.Error(), "method", method, "uri", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func ServerCustomErr(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	slog.Error(err.Error(), "method", method, "uri", uri)
	http.Error(w, err.Error(), statusCode)
}

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	ct := w.Header().Get("Content-Type")
	if ct == "" || ct != "application/json" {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("fetching response: failed to encode JSON", "error", err, "data", data)
		EncodeErrJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}
