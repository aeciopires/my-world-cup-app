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

// thirdPlaceRoundName is excluded from the graphical bracket (it isn't part
// of the single-elimination tree) and rendered as its own standalone match.
const thirdPlaceRoundName = "Match for third place"

type knockoutData struct {
	baseData
	Rounds        []services.KnockoutRound // full round-by-round detail (with venue/date), unchanged
	BracketRounds []services.KnockoutRound // Rounds minus the third-place match, for the graphical bracket
	ThirdPlace    *services.KnockoutRound
}

// Knockout renders the knockout stage bracket.
func (h *PageHandlers) Knockout(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	rounds := services.KnockoutStage(tournament)
	bracketRounds, thirdPlace := splitThirdPlaceRound(rounds)
	h.renderer.render(w, "knockout", knockoutData{
		baseData:      newBaseData(h.store, "Knockout Stage", "knockout"),
		Rounds:        rounds,
		BracketRounds: bracketRounds,
		ThirdPlace:    thirdPlace,
	})
}

// splitThirdPlaceRound separates the "Match for third place" round (a
// standalone fixture, not part of the single-elimination tree) from the
// rounds that make up the graphical bracket.
func splitThirdPlaceRound(rounds []services.KnockoutRound) (bracket []services.KnockoutRound, thirdPlace *services.KnockoutRound) {
	bracket = make([]services.KnockoutRound, 0, len(rounds))
	for _, round := range rounds {
		if round.Name == thirdPlaceRoundName {
			round := round
			thirdPlace = &round
			continue
		}
		bracket = append(bracket, round)
	}
	return bracket, thirdPlace
}

type matchesData struct {
	baseData
	Matches       []models.Match
	TotalMatches  int
	FilterOptions services.FilterOptions
	Round         string
	Group         string
	Team          string
}

// Matches renders the match list, optionally narrowed by the "round",
// "group", and/or "team" query parameters (all optional, combinable).
func (h *PageHandlers) Matches(w http.ResponseWriter, r *http.Request) {
	tournament, _, _ := h.store.Snapshot()
	all := services.AllMatches(tournament)
	round := r.URL.Query().Get("round")
	group := r.URL.Query().Get("group")
	team := r.URL.Query().Get("team")
	h.renderer.render(w, "matches", matchesData{
		baseData:      newBaseData(h.store, "Matches", "matches"),
		Matches:       services.FilterMatches(all, round, group, team),
		TotalMatches:  len(all),
		FilterOptions: services.MatchFilterOptions(tournament),
		Round:         round,
		Group:         group,
		Team:          team,
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
	{Title: "Groups: How Teams Qualify & Tie-Breakers", URL: "https://www.fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/articles/groups-how-teams-qualify-tie-breakers", Description: "Official explanation of the group stage format and tie-breaking rules."},
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
