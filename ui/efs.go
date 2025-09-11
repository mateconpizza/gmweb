// Package ui contains files for rendering web app.
package ui

import (
	"embed"
	"path/filepath"
	"slices"
)

//go:embed "templates" "static"
var Files embed.FS

const (
	CacheFavicon     string = "/cache/favicon/"               // Path to read favicons
	ColorSchemes     string = "static/css/colorshemes"        // Path to colorschemes files
	DefaultColorsCSS string = "default-colors.css"            // Default colors
	TemplateGlob     string = "templates/**/*.gohtml"         // Template files
	ColorSchemesJSON string = "static/json/colorschemes.json" // File used to generate QRCode
)

func SupportedColorschemes() []string {
	entries, err := Files.ReadDir(ColorSchemes)
	if err != nil {
		return []string{}
	}

	var themes []string
	for _, entry := range entries {
		if !entry.IsDir() {
			themes = append(themes, entry.Name())
		}
	}

	return themes
}

func IsValidColorscheme(s string) bool {
	f := filepath.Base(s)
	return slices.Contains(SupportedColorschemes(), f)
}

func IsValidColorschemeMode(m string) bool {
	return m == "dark" || m == "light"
}
