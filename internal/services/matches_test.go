package services

import (
	"testing"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

func testTournamentForMatches() models.Tournament {
	return models.Tournament{
		Matches: []models.Match{
			{Num: 1, Date: "2026-06-11", Team1: "A", Team2: "B", FullTime: ft(2, 0)},
			{Num: 2, Date: "2026-06-12", Team1: "C", Team2: "D", FullTime: ft(1, 1)},
			{Num: 3, Date: "2026-06-30", Team1: "E", Team2: "F"},
			{Num: 4, Date: "2026-07-01", Team1: "G", Team2: "H"},
		},
	}
}

func TestAllMatches_SortedByDate(t *testing.T) {
	matches := AllMatches(testTournamentForMatches())
	if len(matches) != 4 {
		t.Fatalf("len(matches) = %d, want 4", len(matches))
	}
	for i := 1; i < len(matches); i++ {
		if matches[i-1].Date > matches[i].Date {
			t.Errorf("matches not sorted: %s before %s", matches[i-1].Date, matches[i].Date)
		}
	}
}

func TestUpcomingMatches_ExcludesPlayedAndRespectsLimit(t *testing.T) {
	upcoming := UpcomingMatches(testTournamentForMatches(), 1)
	if len(upcoming) != 1 {
		t.Fatalf("len(upcoming) = %d, want 1", len(upcoming))
	}
	if upcoming[0].Played() {
		t.Error("expected only unplayed matches")
	}
	if upcoming[0].Date != "2026-06-30" {
		t.Errorf("Date = %q, want earliest upcoming 2026-06-30", upcoming[0].Date)
	}
}

func TestRecentResults_ExcludesUnplayedAndOrdersMostRecentFirst(t *testing.T) {
	recent := RecentResults(testTournamentForMatches(), 10)
	if len(recent) != 2 {
		t.Fatalf("len(recent) = %d, want 2", len(recent))
	}
	if !recent[0].Played() || !recent[1].Played() {
		t.Error("expected only played matches")
	}
	if recent[0].Date != "2026-06-12" {
		t.Errorf("Date = %q, want most recent 2026-06-12 first", recent[0].Date)
	}
}
