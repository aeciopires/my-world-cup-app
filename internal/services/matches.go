package services

import (
	"sort"

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
