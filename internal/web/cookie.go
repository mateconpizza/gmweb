package web

import (
	"log/slog"
	"net/http"

	"github.com/mateconpizza/gmweb/ui"
)

func setThemeCookie(w http.ResponseWriter, themeName string) {
	cookie := &http.Cookie{
		Name:     "user_theme",
		Value:    themeName,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365, // 1 year
		HttpOnly: false,              // Allow JavaScript access
		Secure:   false,              // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func getThemeModeFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("theme_mode")
	if err == nil && (cookie.Value == "dark" || cookie.Value == "light") {
		slog.Debug("themeMode from cookie:", "theme", cookie.Value)
		return cookie.Value
	}

	slog.Debug("themeMode from cookie:", "fallback", "light")
	return "light"
}

func getThemeFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("user_theme")
	if err != nil {
		slog.Debug("theme from cookie:", "theme", "default")
		return ui.DefaultColorsCSS
	}

	slog.Debug("theme from cookie:", "theme", cookie.Value)
	return cookie.Value
}
