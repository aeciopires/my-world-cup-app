package services

import (
	"testing"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

func ft(home, away int) *models.Score {
	return &models.Score{Home: home, Away: away}
}

func TestGroupStandings_ComputesTableFromResults(t *testing.T) {
	tournament := models.Tournament{
		Groups: []models.Group{
			{Name: "Group A", Teams: []string{"Alpha", "Beta", "Gamma", "Delta"}},
		},
		Matches: []models.Match{
			{Team1: "Alpha", Team2: "Beta", Group: "Group A", Date: "2026-06-01", Num: 1, FullTime: ft(2, 0)},
			{Team1: "Gamma", Team2: "Delta", Group: "Group A", Date: "2026-06-01", Num: 2, FullTime: ft(1, 1)},
			{Team1: "Alpha", Team2: "Gamma", Group: "Group A", Date: "2026-06-05", Num: 3, FullTime: ft(1, 1)},
			{Team1: "Beta", Team2: "Delta", Group: "Group A", Date: "2026-06-05", Num: 4}, // unplayed
		},
	}

	tables := GroupStandings(tournament)
	if len(tables) != 1 {
		t.Fatalf("len(tables) = %d, want 1", len(tables))
	}

	standings := tables[0].Standings
	if len(standings) != 4 {
		t.Fatalf("len(standings) = %d, want 4", len(standings))
	}

	// Alpha: W(2-0) + D(1-1) = 4 pts, GD +2
	// Gamma: D(1-1) + D(1-1) = 2 pts, GD 0
	// Delta: D(1-1) = 1 pt, GD 0 (unplayed match excluded)
	// Beta: L(0-2) = 0 pts, GD -2
	want := []struct {
		team   string
		points int
	}{
		{"Alpha", 4},
		{"Gamma", 2},
		{"Delta", 1},
		{"Beta", 0},
	}
	for i, w := range want {
		if standings[i].Team != w.team {
			t.Errorf("standings[%d].Team = %q, want %q", i, standings[i].Team, w.team)
		}
		if standings[i].Points != w.points {
			t.Errorf("standings[%d].Points = %d, want %d", i, standings[i].Points, w.points)
		}
	}

	alpha := standings[0]
	if alpha.Played != 2 || alpha.Won != 1 || alpha.Drawn != 1 || alpha.Lost != 0 {
		t.Errorf("Alpha record = %+v", alpha)
	}
	if alpha.GoalsFor != 3 || alpha.GoalsAgainst != 1 || alpha.GoalDifference != 2 {
		t.Errorf("Alpha goals = %+v", alpha)
	}

	beta := standings[3]
	if beta.Played != 1 || beta.Lost != 1 {
		t.Errorf("Beta record = %+v (unplayed match should not count)", beta)
	}
}

func TestGroupStandings_TieBreakByGoalDifferenceThenAlphabetical(t *testing.T) {
	tournament := models.Tournament{
		Groups: []models.Group{
			{Name: "Group A", Teams: []string{"Zulu", "Alpha"}},
		},
		Matches: []models.Match{
			// Both teams end with 3 points from a single win each against
			// an external opponent not in this group's standings map, so
			// only goal difference determines order via a head-to-head draw.
			{Team1: "Zulu", Team2: "Alpha", Group: "Group A", Date: "2026-06-01", Num: 1, FullTime: ft(1, 1)},
		},
	}

	tables := GroupStandings(tournament)
	standings := tables[0].Standings
	if standings[0].Team != "Alpha" || standings[1].Team != "Zulu" {
		t.Errorf("expected alphabetical tie-break Alpha before Zulu, got %s, %s", standings[0].Team, standings[1].Team)
	}
}

func TestKnockoutStage_GroupsByRoundInOrder(t *testing.T) {
	tournament := models.Tournament{
		Matches: []models.Match{
			{Round: "Final", Num: 104, Team1: "A", Team2: "B"},
			{Round: "Matchday 1", Num: 1, Team1: "C", Team2: "D"},
			{Round: "Semi-final", Num: 102, Team1: "E", Team2: "F"},
			{Round: "Semi-final", Num: 101, Team1: "G", Team2: "H"},
		},
	}

	rounds := KnockoutStage(tournament)
	if len(rounds) != 2 {
		t.Fatalf("len(rounds) = %d, want 2 (Semi-final, Final; Matchday 1 excluded)", len(rounds))
	}
	if rounds[0].Name != "Semi-final" || rounds[1].Name != "Final" {
		t.Errorf("round order = %v", []string{rounds[0].Name, rounds[1].Name})
	}
	if rounds[0].Matches[0].Num != 101 {
		t.Errorf("expected Semi-final matches sorted by Num, got first Num=%d", rounds[0].Matches[0].Num)
	}
}
