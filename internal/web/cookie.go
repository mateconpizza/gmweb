package web

import (
	"log/slog"
	"net/http"
	"strconv"
)

type jar struct {
	vimMode      string
	compactMode  string
	themeCurrent string
	themeMode    string
	itemsPerPage string
}

type cookieType struct {
	jar     *jar
	oneYear int
}

func (c *cookieType) set(w http.ResponseWriter, key, value string) {
	ck := &http.Cookie{
		Name:     key,
		Value:    value,
		Path:     "/",
		MaxAge:   c.oneYear,
		HttpOnly: false, // Allow JavaScript access
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, ck)
}

func (c *cookieType) get(r *http.Request, key, def string) string {
	ck, err := r.Cookie(key)
	if err != nil {
		slog.Debug("cookies: missing cookie, using default", "key", key, "default", def)
		return def
	}
	return ck.Value
}

func (c *cookieType) getBool(r *http.Request, key string, def bool) bool {
	val := c.get(r, key, strconv.FormatBool(def))
	b, err := strconv.ParseBool(val)
	if err != nil {
		slog.Debug("cookies: parse bool failed", "key", key, "val", val, "default", def)
		return def
	}
	return b
}

func (c *cookieType) getInt(r *http.Request, key string, def int) int {
	val := c.get(r, key, strconv.Itoa(def))
	n, err := strconv.Atoi(val)
	if err != nil {
		slog.Debug("cookies: parse int failed", "key", key, "val", val, "default", def)
		return def
	}
	return n
}

var cookie = &cookieType{
	jar: &jar{
		vimMode:      "vim_mode",
		compactMode:  "compact_mode",
		themeCurrent: "user_theme",
		themeMode:    "theme_mode",
		itemsPerPage: "items_per_page",
	},
	oneYear: 60 * 60 * 24 * 365,
}
