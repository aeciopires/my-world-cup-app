// Package handlers wires HTTP routes to the data store and services, and
// renders the HTML templates.
package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/models"
	"github.com/aeciopires/my-world-cup-app/web"
)

// Renderer parses and caches the application's HTML templates.
type Renderer struct {
	pages map[string]*template.Template
}

var templateFuncs = template.FuncMap{
	"inc": func(i int) int { return i + 1 },
}

var pageFiles = map[string]string{
	"home":     "web/templates/home.html",
	"groups":   "web/templates/groups.html",
	"knockout": "web/templates/knockout.html",
	"matches":  "web/templates/matches.html",
	"links":    "web/templates/links.html",
	"stats":    "web/templates/stats.html",
}

// NewRenderer parses all page templates against the shared layout.
func NewRenderer() (*Renderer, error) {
	pages := make(map[string]*template.Template, len(pageFiles))
	for name, file := range pageFiles {
		tmpl, err := template.New("layout").Funcs(templateFuncs).ParseFS(web.Templates, "templates/layout.html", trimWebPrefix(file))
		if err != nil {
			return nil, err
		}
		pages[name] = tmpl
	}
	return &Renderer{pages: pages}, nil
}

func trimWebPrefix(path string) string {
	const prefix = "web/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		return path[len(prefix):]
	}
	return path
}

// baseData is embedded by every page's template data to supply the fields
// used by the shared layout.
type baseData struct {
	PageTitle   string
	ActiveNav   string
	LastUpdated string
	DataSource  string
	Flags       map[string]string
}

func newBaseData(store *data.Store, title, active string) baseData {
	tournament, lastUpdated, source := store.Snapshot()
	formatted := "never"
	if !lastUpdated.IsZero() {
		formatted = lastUpdated.Format(time.RFC1123)
	}
	return baseData{
		PageTitle:   title,
		ActiveNav:   active,
		LastUpdated: formatted,
		DataSource:  source,
		Flags:       teamFlags(tournament),
	}
}

// teamFlags maps each team's name to its flag emoji, so templates can look
// up a flag by the team name string carried on matches/standings without
// needing the full Team record.
func teamFlags(t models.Tournament) map[string]string {
	flags := make(map[string]string, len(t.Teams))
	for _, team := range t.Teams {
		flags[team.Name] = team.FlagIcon
	}
	return flags
}

func (r *Renderer) render(w http.ResponseWriter, page string, data any) {
	tmpl, ok := r.pages[page]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("template render failed", "page", page, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
