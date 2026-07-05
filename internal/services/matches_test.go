package services

import (
	"testing"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

func testTournamentForMatches() models.Tournament {
	return models.Tournament{
		Matches: []models.Match{
			{Num: 1, Date: "2026-06-11", Round: "Matchday 1", Group: "Group A", Team1: "A", Team2: "B", FullTime: ft(2, 0)},
			{Num: 2, Date: "2026-06-12", Round: "Matchday 1", Group: "Group B", Team1: "C", Team2: "D", FullTime: ft(1, 1)},
			{Num: 3, Date: "2026-06-30", Round: "Round of 16", Team1: "E", Team2: "F"},
			{Num: 4, Date: "2026-07-01", Round: "Round of 16", Team1: "G", Team2: "A"},
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

func TestFilterMatches_NoFiltersReturnsEverything(t *testing.T) {
	all := AllMatches(testTournamentForMatches())
	filtered := FilterMatches(all, "", "", "")
	if len(filtered) != len(all) {
		t.Fatalf("len(filtered) = %d, want %d (all matches)", len(filtered), len(all))
	}
}

func TestFilterMatches_ByRound(t *testing.T) {
	all := AllMatches(testTournamentForMatches())
	filtered := FilterMatches(all, "Round of 16", "", "")
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(filtered))
	}
	for _, m := range filtered {
		if m.Round != "Round of 16" {
			t.Errorf("Round = %q, want Round of 16", m.Round)
		}
	}
}

func TestFilterMatches_ByGroupIsCaseInsensitive(t *testing.T) {
	all := AllMatches(testTournamentForMatches())
	filtered := FilterMatches(all, "", "group a", "")
	if len(filtered) != 1 || filtered[0].Group != "Group A" {
		t.Fatalf("filtered = %+v, want the single Group A match", filtered)
	}
}

func TestFilterMatches_ByTeamMatchesEitherSide(t *testing.T) {
	all := AllMatches(testTournamentForMatches())
	filtered := FilterMatches(all, "", "", "a")
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2 (team A plays as Team1 and Team2)", len(filtered))
	}
}

func TestFilterMatches_CombinedFiltersAreANDed(t *testing.T) {
	all := AllMatches(testTournamentForMatches())
	filtered := FilterMatches(all, "Round of 16", "", "A")
	if len(filtered) != 1 || filtered[0].Num != 4 {
		t.Fatalf("filtered = %+v, want only match #4", filtered)
	}
}

func TestMatchFilterOptions_ReturnsDistinctValuesInExpectedOrder(t *testing.T) {
	opts := MatchFilterOptions(testTournamentForMatches())

	wantRounds := []string{"Matchday 1", "Round of 16"}
	if len(opts.Rounds) != len(wantRounds) || opts.Rounds[0] != wantRounds[0] || opts.Rounds[1] != wantRounds[1] {
		t.Errorf("Rounds = %v, want %v (schedule order, deduplicated)", opts.Rounds, wantRounds)
	}

	wantGroups := []string{"Group A", "Group B"}
	if len(opts.Groups) != len(wantGroups) || opts.Groups[0] != wantGroups[0] || opts.Groups[1] != wantGroups[1] {
		t.Errorf("Groups = %v, want %v (alphabetical)", opts.Groups, wantGroups)
	}

	wantTeams := []string{"A", "B", "C", "D", "E", "F", "G"}
	if len(opts.Teams) != len(wantTeams) {
		t.Fatalf("Teams = %v, want %v (alphabetical, deduplicated)", opts.Teams, wantTeams)
	}
	for i, team := range wantTeams {
		if opts.Teams[i] != team {
			t.Errorf("Teams[%d] = %q, want %q", i, opts.Teams[i], team)
		}
	}
}
