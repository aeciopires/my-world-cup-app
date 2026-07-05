package models_test

import (
	"testing"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

func TestMatch_Winner_DecidedInFullTime(t *testing.T) {
	m := models.Match{Team1: "Alpha", Team2: "Beta", FullTime: &models.Score{Home: 2, Away: 0}}
	if got := m.Winner(); got != "Alpha" {
		t.Errorf("Winner() = %q, want Alpha", got)
	}
}

func TestMatch_Winner_UnplayedMatch(t *testing.T) {
	m := models.Match{Team1: "Alpha", Team2: "Beta"}
	if got := m.Winner(); got != "" {
		t.Errorf("Winner() = %q, want empty for unplayed match", got)
	}
}

func TestMatch_Winner_DrawWithoutPenaltiesRecorded(t *testing.T) {
	m := models.Match{Team1: "Alpha", Team2: "Beta", FullTime: &models.Score{Home: 1, Away: 1}}
	if got := m.Winner(); got != "" {
		t.Errorf("Winner() = %q, want empty for a draw with no penalty shoot-out recorded", got)
	}
}

func TestMatch_Winner_PenaltiesDecideADraw(t *testing.T) {
	m := models.Match{
		Team1:     "Alpha",
		Team2:     "Beta",
		FullTime:  &models.Score{Home: 1, Away: 1},
		Penalties: &models.Score{Home: 3, Away: 4},
	}
	if got := m.Winner(); got != "Beta" {
		t.Errorf("Winner() = %q, want Beta (won on penalties)", got)
	}
}
