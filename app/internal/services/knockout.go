package services

import (
	"sort"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// KnockoutRound groups the matches belonging to a single knockout stage round.
type KnockoutRound struct {
	Name    string
	Matches []models.Match
}

// knockoutRoundOrder defines the display order of knockout stage rounds,
// matching the round names used by the openfootball data source.
var knockoutRoundOrder = []string{
	"Round of 32",
	"Round of 16",
	"Quarter-final",
	"Semi-final",
	"Match for third place",
	"Final",
}

// KnockoutStage groups the tournament's knockout matches into ordered rounds.
// Matches whose round is not part of the known knockout stage (i.e. group
// stage "Matchday N" rounds) are excluded.
func KnockoutStage(t models.Tournament) []KnockoutRound {
	matchesByRound := make(map[string][]models.Match)
	for _, m := range t.Matches {
		matchesByRound[m.Round] = append(matchesByRound[m.Round], m)
	}

	rounds := make([]KnockoutRound, 0, len(knockoutRoundOrder))
	for _, name := range knockoutRoundOrder {
		matches := matchesByRound[name]
		if len(matches) == 0 {
			continue
		}
		sort.SliceStable(matches, func(i, j int) bool { return matches[i].Num < matches[j].Num })
		rounds = append(rounds, KnockoutRound{Name: name, Matches: matches})
	}
	return rounds
}
