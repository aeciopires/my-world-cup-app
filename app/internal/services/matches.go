package services

import (
	"sort"
	"strings"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// AllMatches returns every match sorted by date, then by match number.
func AllMatches(t models.Tournament) []models.Match {
	matches := make([]models.Match, len(t.Matches))
	copy(matches, t.Matches)
	sortMatchesByRoundAndDate(matches)
	return matches
}

// UpcomingMatches returns the next n unplayed matches, ordered by date.
func UpcomingMatches(t models.Tournament, n int) []models.Match {
	upcoming := make([]models.Match, 0)
	for _, m := range t.Matches {
		if !m.Played() {
			upcoming = append(upcoming, m)
		}
	}
	sortMatchesByRoundAndDate(upcoming)
	if len(upcoming) > n {
		upcoming = upcoming[:n]
	}
	return upcoming
}

// RecentResults returns the last n played matches, most recent first.
func RecentResults(t models.Tournament, n int) []models.Match {
	played := make([]models.Match, 0)
	for _, m := range t.Matches {
		if m.Played() {
			played = append(played, m)
		}
	}
	sortMatchesByRoundAndDate(played)
	sort.SliceStable(played, func(i, j int) bool { return played[i].Date > played[j].Date })
	if len(played) > n {
		played = played[:n]
	}
	return played
}

// FilterMatches narrows matches down to those matching the given round,
// group, and/or team (each compared case-insensitively). An empty filter
// value matches every match for that dimension; team matches either side of
// the fixture. The input order is preserved.
func FilterMatches(matches []models.Match, round, group, team string) []models.Match {
	filtered := make([]models.Match, 0, len(matches))
	for _, m := range matches {
		if round != "" && !strings.EqualFold(m.Round, round) {
			continue
		}
		if group != "" && !strings.EqualFold(m.Group, group) {
			continue
		}
		if team != "" && !strings.EqualFold(m.Team1, team) && !strings.EqualFold(m.Team2, team) {
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered
}

// FilterOptions holds the distinct values available to filter the matches
// page, for populating its round/group/team dropdowns.
type FilterOptions struct {
	Rounds []string
	Groups []string
	Teams  []string
}

// MatchFilterOptions returns the distinct rounds, groups, and teams present
// across the tournament's matches. Rounds are returned in schedule order
// (group matchdays before knockout rounds); groups and teams are sorted
// alphabetically.
func MatchFilterOptions(t models.Tournament) FilterOptions {
	ordered := AllMatches(t)

	seenRounds := make(map[string]bool, len(knockoutRoundOrder))
	rounds := make([]string, 0, len(knockoutRoundOrder))
	groupSet := make(map[string]bool)
	teamSet := make(map[string]bool)
	for _, m := range ordered {
		if m.Round != "" && !seenRounds[m.Round] {
			seenRounds[m.Round] = true
			rounds = append(rounds, m.Round)
		}
		if m.Group != "" {
			groupSet[m.Group] = true
		}
		teamSet[m.Team1] = true
		teamSet[m.Team2] = true
	}

	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	teams := make([]string, 0, len(teamSet))
	for team := range teamSet {
		teams = append(teams, team)
	}
	sort.Strings(teams)

	return FilterOptions{Rounds: rounds, Groups: groups, Teams: teams}
}
