package services

import (
	"cmp"
	"maps"
	"slices"
	"strings"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// AllMatches returns a copy of every match sorted by date, then by match
// number. The tournament's own slice is never reordered, since it is shared
// by every concurrent request reading the same snapshot.
func AllMatches(t models.Tournament) []models.Match {
	return sortMatchesByDate(slices.Clone(t.Matches))
}

// UpcomingMatches returns the next n unplayed matches, ordered by date.
func UpcomingMatches(t models.Tournament, n int) []models.Match {
	upcoming := filterMatches(t.Matches, func(m models.Match) bool { return !m.Played() })
	sortMatchesByDate(upcoming)
	return firstN(upcoming, n)
}

// RecentResults returns the last n played matches, most recent first.
func RecentResults(t models.Tournament, n int) []models.Match {
	played := filterMatches(t.Matches, models.Match.Played)
	slices.SortStableFunc(played, func(a, b models.Match) int {
		return cmp.Or(cmp.Compare(b.Date, a.Date), cmp.Compare(a.Num, b.Num))
	})
	return firstN(played, n)
}

// FilterMatches narrows matches down to those matching the given round,
// group, and/or team (each compared case-insensitively). An empty filter
// value matches every match for that dimension; team matches either side of
// the fixture. The input order is preserved.
func FilterMatches(matches []models.Match, round, group, team string) []models.Match {
	return filterMatches(matches, func(m models.Match) bool {
		switch {
		case round != "" && !strings.EqualFold(m.Round, round):
			return false
		case group != "" && !strings.EqualFold(m.Group, group):
			return false
		case team != "" && !strings.EqualFold(m.Team1, team) && !strings.EqualFold(m.Team2, team):
			return false
		default:
			return true
		}
	})
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
//
// When the tournament carries a team list, only real team names are
// offered: knockout fixtures not yet decided use placeholders such as
// "W74" (winner of match 74) or "2A" (runner-up of Group A), which would
// otherwise clutter the team dropdown.
func MatchFilterOptions(t models.Tournament) FilterOptions {
	knownTeams := make(map[string]bool, len(t.Teams))
	for _, team := range t.Teams {
		knownTeams[team.Name] = true
	}
	isTeam := func(name string) bool {
		return name != "" && (len(knownTeams) == 0 || knownTeams[name])
	}

	seenRounds := make(map[string]bool, len(knockoutRoundOrder))
	rounds := make([]string, 0, len(knockoutRoundOrder))
	groupSet := make(map[string]struct{})
	teamSet := make(map[string]struct{})
	for _, m := range AllMatches(t) {
		if m.Round != "" && !seenRounds[m.Round] {
			seenRounds[m.Round] = true
			rounds = append(rounds, m.Round)
		}
		if m.Group != "" {
			groupSet[m.Group] = struct{}{}
		}
		for _, team := range []string{m.Team1, m.Team2} {
			if isTeam(team) {
				teamSet[team] = struct{}{}
			}
		}
	}

	return FilterOptions{
		Rounds: rounds,
		Groups: slices.Sorted(maps.Keys(groupSet)),
		Teams:  slices.Sorted(maps.Keys(teamSet)),
	}
}

// filterMatches returns a new slice with the matches satisfying keep, in
// input order. It never aliases the input's backing array.
func filterMatches(matches []models.Match, keep func(models.Match) bool) []models.Match {
	out := make([]models.Match, 0, len(matches))
	for _, m := range matches {
		if keep(m) {
			out = append(out, m)
		}
	}
	return out
}

// firstN returns at most the first n elements of s (none for n <= 0).
func firstN[T any](s []T, n int) []T {
	return s[:max(0, min(n, len(s)))]
}
