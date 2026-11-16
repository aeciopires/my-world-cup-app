// Package handlers wires HTTP routes to the data store and services, and
// renders the HTML templates.
package handlers

import (
	"bytes"
	"fmt"
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

// scorePill renders a result as a styled pill badge.
func scorePill(home, away int) template.HTML {
	return template.HTML(`<span class="score">` + strconv.Itoa(home) + ` - ` + strconv.Itoa(away) + `</span>`)
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
	// score renders a single scoreline as a styled pill badge.
	"score": scorePill,
	// result renders a match's complete outcome: the final score pill (after
	// extra time, if played), an "a.e.t." marker, and the penalty shoot-out
	// score when there was one. Unplayed matches get the "scheduled" badge.
	"result": func(m models.Match) template.HTML {
		final := m.FinalScore()
		if final == nil {
			return `<span class="badge-scheduled">scheduled</span>`
		}
		out := scorePill(final.Home, final.Away)
		if m.WentToExtraTime() {
			out += ` <span class="muted result-note">a.e.t.</span>`
		}
		if p := m.Penalties; p != nil {
			out += template.HTML(fmt.Sprintf(` <span class="muted result-note">(pens %d-%d)</span>`, p.Home, p.Away))
		}
		return out
	},
	// pairs groups a round's matches two at a time (in bracket order) so the
	// knockout template can render each pair as a bracket "elbow" that feeds
	// into a single match in the next round. A trailing single match (the
	// Final, which has no sibling) is returned as its own group of one.
	"pairs": func(matches []models.Match) [][]models.Match {
		groups := make([][]models.Match, 0, (len(matches)+1)/2)
		for i := 0; i < len(matches); i += 2 {
			groups = append(groups, matches[i:min(i+2, len(matches))])
		}
		return groups
	},
}

// pageFiles maps each page name to its template, relative to web.Templates.
var pageFiles = map[string]string{
	"home":     "templates/home.html",
	"groups":   "templates/groups.html",
	"knockout": "templates/knockout.html",
	"matches":  "templates/matches.html",
	"links":    "templates/links.html",
	"stats":    "templates/stats.html",
}

// NewRenderer parses all page templates against the shared layout.
func NewRenderer() (*Renderer, error) {
	pages := make(map[string]*template.Template, len(pageFiles))
	for name, file := range pageFiles {
		tmpl, err := template.New("layout").Funcs(templateFuncs).ParseFS(web.Templates, "templates/layout.html", file)
		if err != nil {
			return nil, fmt.Errorf("parse template %q: %w", name, err)
		}
		pages[name] = tmpl
	}
	return &Renderer{pages: pages}, nil
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

// newBaseData builds the shared layout data from the same snapshot the page
// handler renders, so the header metadata, flags and links can never come
// from a different refresh than the page body.
func newBaseData(snap data.Snapshot, title, active string) baseData {
	formatted := "never"
	if !snap.LastUpdated.IsZero() {
		formatted = snap.LastUpdated.Format(time.RFC1123)
	}
	t := snap.Tournament
	return baseData{
		PageTitle:    title,
		ActiveNav:    active,
		LastUpdated:  formatted,
		DataSource:   snap.Source,
		Flags:        teamFlags(t),
		TeamLinks:    teamLinks(t),
		StadiumLinks: stadiumLinks(t),
		CityLinks:    cityLinks(t),
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

// render executes the page template into a buffer first, so a template
// error results in a clean 500 instead of a half-written page followed by
// an error message (and a "superfluous WriteHeader" warning).
func (r *Renderer) render(w http.ResponseWriter, page string, data any) {
	tmpl, ok := r.pages[page]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		slog.Error("template render failed", "page", page, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		slog.Debug("writing response failed", "page", page, "error", err)
	}
}
