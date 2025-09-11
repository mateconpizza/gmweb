package web

import (
	"embed"
	"encoding/json"
	"log/slog"

	"github.com/mateconpizza/gm/pkg/files"

	"github.com/mateconpizza/gmweb/ui"
)

// Theme represents a single theme with its name and color schemes.
type Theme struct {
	Name  string          `json:"name"`
	Dark  ColorSchemeJSON `json:"dark"`
	Light ColorSchemeJSON `json:"light"`
}

// ColorSchemeJSON represents the color settings for a theme.
type ColorSchemeJSON struct {
	Bg string `json:"bg"`
	Fg string `json:"fg"`
}

func getCurrentTheme(content []byte, name string) (*Theme, error) {
	var themes []Theme
	err := json.Unmarshal(content, &themes)
	if err != nil {
		slog.Error("error unmarshalling JSON:", "error", err)
		return nil, err
	}

	themeMap := make(map[string]*Theme, len(themes))
	for _, theme := range themes {
		themeMap[theme.Name] = &theme
	}

	theme, ok := themeMap[name]
	if !ok {
		defaultTheme := files.StripSuffixes(ui.DefaultColorsCSS)
		slog.Error("theme not found", "theme", name, "default", defaultTheme)
		theme = themeMap[defaultTheme]
	}

	return theme, nil
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
