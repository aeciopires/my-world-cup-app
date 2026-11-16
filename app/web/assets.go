// Package web embeds the HTML templates and static assets (CSS/JS) served
// by the application, so the compiled binary is self-contained.
package web

import "embed"

// Templates holds every page template plus the shared layout, under
// "templates/".
//
//go:embed templates/*.html
var Templates embed.FS

// Static holds the CSS/JS assets served at /static/*, under "static/".
//
//go:embed static
var Static embed.FS
