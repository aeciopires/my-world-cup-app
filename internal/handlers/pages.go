package handlers

import (
	"net/http"

	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/models"
	"github.com/aeciopires/my-world-cup-app/internal/services"
)

const upcomingAndRecentLimit = 8
const topScorersLimit = 15

// PageHandlers holds the dependencies needed to render tournament pages.
type PageHandlers struct {
	store    *data.Store
	renderer *Renderer
}

// NewPageHandlers creates a PageHandlers instance.
func NewPageHandlers(store *data.Store, renderer *Renderer) *PageHandlers {
	return &PageHandlers{store: store, renderer: renderer}
}

type homeData struct {
	baseData
	TournamentName string
	Upcoming       []models.Match
	Recent         []models.Match
}

// Home renders the dashboard page.
func (h *PageHandlers) Home(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	h.renderer.render(w, "home", homeData{
		baseData:       newBaseData(h.store, "Home", "home"),
		TournamentName: tournament.Name,
		Upcoming:       services.UpcomingMatches(tournament, upcomingAndRecentLimit),
		Recent:         services.RecentResults(tournament, upcomingAndRecentLimit),
	})
}

type groupsData struct {
	baseData
	Tables []services.GroupTable
}

// Groups renders group standings and results.
func (h *PageHandlers) Groups(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	h.renderer.render(w, "groups", groupsData{
		baseData: newBaseData(h.store, "Groups", "groups"),
		Tables:   services.GroupStandings(tournament),
	})
}

type knockoutData struct {
	baseData
	Rounds []services.KnockoutRound
}

// Knockout renders the knockout stage bracket.
func (h *PageHandlers) Knockout(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	h.renderer.render(w, "knockout", knockoutData{
		baseData: newBaseData(h.store, "Knockout Stage", "knockout"),
		Rounds:   services.KnockoutStage(tournament),
	})
}

type matchesData struct {
	baseData
	Matches []models.Match
}

// Matches renders the full match list.
func (h *PageHandlers) Matches(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	h.renderer.render(w, "matches", matchesData{
		baseData: newBaseData(h.store, "Matches", "matches"),
		Matches:  services.AllMatches(tournament),
	})
}

type externalLink struct {
	Title       string
	URL         string
	Description string
}

type linksData struct {
	baseData
	Links    []externalLink
	Stadiums []models.Stadium
}

var fifaLinks = []externalLink{
	{Title: "Teams", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/teams", Description: "Official list of qualified national teams."},
	{Title: "Standings", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/standings", Description: "Official FIFA group standings."},
	{Title: "Scores & Fixtures", URL: "https://www.fifa.com/pt/tournaments/mens/worldcup/canadamexicousa2026/scores-fixtures?country=BR&wtw-filter=ALL", Description: "Official match schedule and results."},
	{Title: "Stadiums", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/stadiums", Description: "Venues hosting the 2026 tournament."},
	{Title: "Official Match Ball", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/official-match-ball", Description: "The official match ball of the FIFA World Cup 2026."},
	{Title: "Official Posters", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/official-posters", Description: "Official tournament and host city posters."},
	{Title: "Mascots", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/mascots", Description: "Meet the official mascots of the FIFA World Cup 2026."},
	{Title: "Tournament Home (PT)", URL: "https://www.fifa.com/pt/tournaments/mens/worldcup/canadamexicousa2026", Description: "Official FIFA World Cup 2026 tournament page."},
	{Title: "Articles & News", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/articles/", Description: "Latest FIFA World Cup 2026 news and articles."},
	{Title: "FIFA Club World Cup 2025", URL: "https://www.fifa.com/en/tournaments/mens/club-world-cup/usa-2025", Description: "Official FIFA Club World Cup 2025 tournament page."},
	{Title: "FIFA Sound", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/fifa-sound", Description: "Official soundtrack and music hub for the tournament."},
	{Title: "Official Playlist (Spotify)", URL: "https://open.spotify.com/playlist/58BwtjH93yTBGpGIIIWzkm?si=158a76e6684c473d&nd=1&dlsi=f2a56174aed64294", Description: "Official FIFA World Cup 2026 Spotify playlist."},
}

// Links renders the external FIFA resources page.
func (h *PageHandlers) Links(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	h.renderer.render(w, "links", linksData{
		baseData: newBaseData(h.store, "FIFA Links", "links"),
		Links:    fifaLinks,
		Stadiums: tournament.Stadiums,
	})
}

type statsData struct {
	baseData
	TopScorers    []services.PlayerStat
	TeamStandings []models.Standing
}

// Stats renders player and team statistics aggregated across the whole
// tournament (group and knockout stages).
func (h *PageHandlers) Stats(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	h.renderer.render(w, "stats", statsData{
		baseData:      newBaseData(h.store, "Statistics", "stats"),
		TopScorers:    services.TopScorers(tournament, topScorersLimit),
		TeamStandings: services.TeamStandings(tournament),
	})
}
