// Package ui contains files for rendering web app.
package ui

import "embed"

//go:embed "templates" "static"
var Files embed.FS

const (
	CacheFavicon     string = "/cache/favicon/"        // Path to read favicons
	ColorSchemes     string = "static/css/colorshemes" // Path to colorschemes files
	DefaultColorsCSS string = "default-colors.css"     // Default colors
	TemplateGlob     string = "templates/**/*.gohtml"  // Template files
)
