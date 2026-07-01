package services

import (
	"testing"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

func goal(name string) models.Goal {
	return models.Goal{Name: name}
}

func testTournamentForStats() models.Tournament {
	return models.Tournament{
		Matches: []models.Match{
			{
				Team1: "Alpha", Team2: "Beta", FullTime: ft(2, 1),
				Goals1: []models.Goal{goal("Player A"), goal("Player A")},
				Goals2: []models.Goal{goal("Player B")},
			},
			{
				Team1: "Gamma", Team2: "Alpha", FullTime: ft(0, 1),
				Goals2: []models.Goal{goal("Player A")},
			},
			{
				Team1: "Beta", Team2: "Gamma", FullTime: ft(1, 1),
				Goals1: []models.Goal{goal("Player B")},
				Goals2: []models.Goal{goal("Player C")},
			},
			{
				// Unplayed match: goals (if any) should not occur, and it
				// must not contribute to team standings.
				Team1: "Alpha", Team2: "Gamma",
			},
		},
	}
}

func TestTopScorers_AggregatesAndOrdersByGoalsThenName(t *testing.T) {
	scorers := TopScorers(testTournamentForStats(), 10)

	if len(scorers) != 3 {
		t.Fatalf("len(scorers) = %d, want 3", len(scorers))
	}

	want := []PlayerStat{
		{Name: "Player A", Team: "Alpha", Goals: 3},
		{Name: "Player B", Team: "Beta", Goals: 2},
		{Name: "Player C", Team: "Gamma", Goals: 1},
	}
	for i, w := range want {
		if scorers[i] != w {
			t.Errorf("scorers[%d] = %+v, want %+v", i, scorers[i], w)
		}
	}
}

func TestTopScorers_RespectsLimit(t *testing.T) {
	scorers := TopScorers(testTournamentForStats(), 1)
	if len(scorers) != 1 {
		t.Fatalf("len(scorers) = %d, want 1", len(scorers))
	}
	if scorers[0].Name != "Player A" {
		t.Errorf("scorers[0].Name = %q, want Player A (top scorer)", scorers[0].Name)
	}
}

func TestTeamStandings_AggregatesAcrossAllMatches(t *testing.T) {
	standings := TeamStandings(testTournamentForStats())

	if len(standings) != 3 {
		t.Fatalf("len(standings) = %d, want 3 (unplayed match's teams still counted via other matches)", len(standings))
	}

	byTeam := make(map[string]models.Standing, len(standings))
	for _, s := range standings {
		byTeam[s.Team] = s
	}

	alpha := byTeam["Alpha"]
	// Alpha: W(2-1 vs Beta) + W(1-0 vs Gamma) = 6 pts, played 2 (unplayed match excluded)
	if alpha.Played != 2 || alpha.Won != 2 || alpha.Points != 6 {
		t.Errorf("Alpha standing = %+v", alpha)
	}

	beta := byTeam["Beta"]
	// Beta: L(1-2 vs Alpha) + D(1-1 vs Gamma) = 1 pt, played 2
	if beta.Played != 2 || beta.Lost != 1 || beta.Drawn != 1 || beta.Points != 1 {
		t.Errorf("Beta standing = %+v", beta)
	}

	// Standings must be sorted: Alpha (6 pts) before Beta/Gamma.
	if standings[0].Team != "Alpha" {
		t.Errorf("standings[0].Team = %q, want Alpha (highest points)", standings[0].Team)
	}
}
