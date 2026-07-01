// Package web embeds the HTML templates and static assets (CSS/JS) served
// by the application, so the compiled binary is self-contained.
package web

import "embed"

//go:embed templates/*.html
var Templates embed.FS

//go:embed static
var Static embed.FS
