// Package handlers wires HTTP routes to the data store and services, and
// renders the HTML templates.
package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
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
	"linked": func(links map[string]string, name string) template.HTML {
		if url, ok := links[name]; ok && url != "" {
			return template.HTML(`<a href="` + template.HTMLEscapeString(url) + `" target="_blank" rel="noopener noreferrer">` + template.HTMLEscapeString(name) + `</a>`)
		}
		return template.HTML(template.HTMLEscapeString(name))
	},
	// flag renders a team's flag emoji inside a styled chip, so every page
	// gets the same flag presentation without repeating markup per template.
	"flag": func(flags map[string]string, name string) template.HTML {
		return template.HTML(`<span class="flag">` + template.HTMLEscapeString(flags[name]) + `</span>`)
	},
	// score renders a full-time result as a styled pill badge.
	"score": func(home, away int) template.HTML {
		return template.HTML(`<span class="score">` + strconv.Itoa(home) + ` - ` + strconv.Itoa(away) + `</span>`)
	},
	// pairs groups a round's matches two at a time (in bracket order) so the
	// knockout template can render each pair as a bracket "elbow" that feeds
	// into a single match in the next round. A trailing single match (the
	// Final, which has no sibling) is returned as its own group of one.
	"pairs": func(matches []models.Match) [][]models.Match {
		groups := make([][]models.Match, 0, (len(matches)+1)/2)
		for i := 0; i < len(matches); i += 2 {
			end := i + 2
			if end > len(matches) {
				end = len(matches)
			}
			groups = append(groups, matches[i:end])
		}
		return groups
	},
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
	PageTitle    string
	ActiveNav    string
	LastUpdated  string
	DataSource   string
	Flags        map[string]string
	TeamLinks    map[string]string
	StadiumLinks map[string]string
	CityLinks    map[string]string
}

func newBaseData(store *data.Store, title, active string) baseData {
	tournament, lastUpdated, source := store.Snapshot()
	formatted := "never"
	if !lastUpdated.IsZero() {
		formatted = lastUpdated.Format(time.RFC1123)
	}
	return baseData{
		PageTitle:    title,
		ActiveNav:    active,
		LastUpdated:  formatted,
		DataSource:   source,
		Flags:        teamFlags(tournament),
		TeamLinks:    teamLinks(tournament),
		StadiumLinks: stadiumLinks(tournament),
		CityLinks:    cityLinks(tournament),
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
