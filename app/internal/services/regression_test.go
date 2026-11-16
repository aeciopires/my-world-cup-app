package services

import (
	"slices"
	"testing"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// Regression tests for scoring rules that are easy to get wrong with the
// openfootball data model: own goals, extra time, and knockout placeholders.

func TestTopScorers_ExcludesOwnGoals(t *testing.T) {
	tournament := models.Tournament{
		Matches: []models.Match{{
			Team1: "Qatar", Team2: "Switzerland", FullTime: ft(1, 1),
			// An own goal is listed under the team it counts for (Qatar),
			// but scored by an opposing (Swiss) player.
			Goals1: []models.Goal{{Name: "Miro Muheim", OwnGoal: true}},
			Goals2: []models.Goal{goal("Breel Embolo")},
		}},
	}

	scorers := TopScorers(tournament, 10)
	if len(scorers) != 1 || scorers[0].Name != "Breel Embolo" {
		t.Fatalf("TopScorers() = %+v, want only Breel Embolo (own goals excluded)", scorers)
	}
}

func TestTopScorers_NonPositiveLimitReturnsEmpty(t *testing.T) {
	for _, n := range []int{0, -1} {
		if got := TopScorers(testTournamentForStats(), n); len(got) != 0 {
			t.Errorf("TopScorers(n=%d) = %+v, want empty", n, got)
		}
	}
}

func TestTopScorers_TieBreaksByTeamForSameName(t *testing.T) {
	tournament := models.Tournament{
		Matches: []models.Match{{
			Team1: "Zeta", Team2: "Alpha", FullTime: ft(1, 1),
			Goals1: []models.Goal{goal("Silva")},
			Goals2: []models.Goal{goal("Silva")},
		}},
	}
	// Run several times: map iteration order is randomized, so a missing
	// tie-break would show up as a flaky order.
	for range 20 {
		scorers := TopScorers(tournament, 10)
		if len(scorers) != 2 || scorers[0].Team != "Alpha" || scorers[1].Team != "Zeta" {
			t.Fatalf("TopScorers() = %+v, want Silva (Alpha) before Silva (Zeta)", scorers)
		}
	}
}

func TestTeamStandings_CountsExtraTimeGoalsButNotPenalties(t *testing.T) {
	tournament := models.Tournament{
		Matches: []models.Match{
			{
				// 1-1 after 90', 2-1 after extra time: Alpha wins.
				Team1: "Alpha", Team2: "Beta",
				FullTime: ft(1, 1), ExtraTime: ft(2, 1),
			},
			{
				// 0-0 after extra time, decided on penalties: a draw for
				// standings purposes; the shoot-out goals don't count.
				Team1: "Gamma", Team2: "Delta",
				FullTime: ft(0, 0), ExtraTime: ft(0, 0), Penalties: ft(5, 4),
			},
		},
	}

	byTeam := make(map[string]models.Standing)
	for _, s := range TeamStandings(tournament) {
		byTeam[s.Team] = s
	}

	alpha := byTeam["Alpha"]
	if alpha.Won != 1 || alpha.GoalsFor != 2 || alpha.GoalsAgainst != 1 || alpha.Points != pointsForWin {
		t.Errorf("Alpha = %+v, want a 2-1 win after extra time", alpha)
	}
	if beta := byTeam["Beta"]; beta.Lost != 1 {
		t.Errorf("Beta = %+v, want a loss", beta)
	}
	gamma := byTeam["Gamma"]
	if gamma.Drawn != 1 || gamma.GoalsFor != 0 || gamma.Points != pointsForDraw {
		t.Errorf("Gamma = %+v, want a 0-0 draw (penalties excluded)", gamma)
	}
}

func TestMatchFilterOptions_ExcludesKnockoutPlaceholders(t *testing.T) {
	tournament := models.Tournament{
		Teams: []models.Team{{Name: "Canada"}, {Name: "Mexico"}},
		Matches: []models.Match{
			{Round: "Matchday 1", Team1: "Canada", Team2: "Mexico", Date: "2026-06-11"},
			{Round: "Round of 16", Team1: "Canada", Team2: "W75", Date: "2026-07-04"},
			{Round: "Final", Team1: "W101", Team2: "W102", Date: "2026-07-19"},
		},
	}

	got := MatchFilterOptions(tournament).Teams
	if want := []string{"Canada", "Mexico"}; !slices.Equal(got, want) {
		t.Errorf("Teams = %v, want %v", got, want)
	}
}

func TestAllMatches_DoesNotReorderTournamentSlice(t *testing.T) {
	tournament := models.Tournament{
		Matches: []models.Match{
			{Num: 2, Date: "2026-06-12"},
			{Num: 1, Date: "2026-06-11"},
		},
	}
	_ = AllMatches(tournament)
	if tournament.Matches[0].Num != 2 {
		t.Error("AllMatches reordered the shared tournament slice in place")
	}
}
