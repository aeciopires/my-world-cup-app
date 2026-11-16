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

func TestMatch_Winner_DecidedInExtraTime(t *testing.T) {
	// openfootball's "et" score is cumulative: 1-1 after 90 minutes, 2-1
	// after 120 means Alpha won in extra time, with no shoot-out.
	m := models.Match{
		Team1:     "Alpha",
		Team2:     "Beta",
		FullTime:  new(models.Score{Home: 1, Away: 1}),
		ExtraTime: new(models.Score{Home: 2, Away: 1}),
	}
	if got := m.Winner(); got != "Alpha" {
		t.Errorf("Winner() = %q, want Alpha (won in extra time)", got)
	}
}

func TestMatch_FinalScore(t *testing.T) {
	ft := models.Score{Home: 1, Away: 1}
	et := models.Score{Home: 2, Away: 1}
	tests := []struct {
		name  string
		match models.Match
		want  *models.Score
	}{
		{name: "unplayed", match: models.Match{}, want: nil},
		{name: "full time only", match: models.Match{FullTime: &ft}, want: &ft},
		{name: "after extra time", match: models.Match{FullTime: &ft, ExtraTime: &et}, want: &et},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.FinalScore()
			switch {
			case got == nil && tt.want == nil:
			case got == nil || tt.want == nil || *got != *tt.want:
				t.Errorf("FinalScore() = %v, want %v", got, tt.want)
			}
		})
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
