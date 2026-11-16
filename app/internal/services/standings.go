// Package services contains the business logic derived from the raw
// tournament data: standings computation and knockout stage grouping.
package services

import (
	"cmp"
	"slices"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

const (
	pointsForWin  = 3
	pointsForDraw = 1
)

// GroupTable pairs a group with its computed standings, ordered by rank.
type GroupTable struct {
	Group     models.Group
	Standings []models.Standing
	Matches   []models.Match
}

// GroupStandings computes the standings table for every group in the
// tournament, derived from the played matches. Unplayed matches do not
// contribute to the table. Rows are ordered by points, then goal
// difference, then goals for, then alphabetically by team name; this is a
// simplified tie-break that does not implement FIFA's full head-to-head
// rules.
func GroupStandings(t models.Tournament) []GroupTable {
	matchesByGroup := make(map[string][]models.Match)
	for _, m := range t.Matches {
		matchesByGroup[m.Group] = append(matchesByGroup[m.Group], m)
	}

	tables := make([]GroupTable, 0, len(t.Groups))
	for _, g := range t.Groups {
		standingByTeam := make(map[string]*models.Standing, len(g.Teams))
		for _, team := range g.Teams {
			standingByTeam[team] = &models.Standing{Team: team}
		}

		groupMatches := matchesByGroup[g.Name]
		for _, m := range groupMatches {
			if !m.Played() {
				continue
			}
			applyResult(standingByTeam, m)
		}

		tables = append(tables, GroupTable{
			Group:     g,
			Standings: collectStandings(standingByTeam),
			Matches:   sortMatchesByDate(groupMatches),
		})
	}
	return tables
}

func applyResult(byTeam map[string]*models.Standing, m models.Match) {
	home, ok1 := byTeam[m.Team1]
	away, ok2 := byTeam[m.Team2]
	if !ok1 || !ok2 {
		return
	}
	accumulateResult(home, away, m)
}

// accumulateResult applies a played match's result to the two teams'
// running standings. Goals scored in extra time count (the score after
// extra time is used when present), but a penalty shoot-out does not: a
// knockout match level after extra time counts as a draw, as in FIFA's own
// tournament statistics. The caller is responsible for resolving (or
// creating) the Standing entries and for ensuring m.Played() is true.
func accumulateResult(home, away *models.Standing, m models.Match) {
	score := m.FinalScore()

	home.Played++
	away.Played++
	home.GoalsFor += score.Home
	home.GoalsAgainst += score.Away
	away.GoalsFor += score.Away
	away.GoalsAgainst += score.Home

	switch {
	case score.Home > score.Away:
		home.Won++
		home.Points += pointsForWin
		away.Lost++
	case score.Home < score.Away:
		away.Won++
		away.Points += pointsForWin
		home.Lost++
	default:
		home.Drawn++
		away.Drawn++
		home.Points += pointsForDraw
		away.Points += pointsForDraw
	}

	home.GoalDifference = home.GoalsFor - home.GoalsAgainst
	away.GoalDifference = away.GoalsFor - away.GoalsAgainst
}

// collectStandings flattens a team → standing map into a ranked slice.
func collectStandings(byTeam map[string]*models.Standing) []models.Standing {
	standings := make([]models.Standing, 0, len(byTeam))
	for _, s := range byTeam {
		standings = append(standings, *s)
	}
	sortStandings(standings)
	return standings
}

// sortStandings orders rows by points, goal difference, goals for (all
// descending), then team name. Team names are unique, so the order is
// total and deterministic regardless of the input (map iteration) order.
func sortStandings(standings []models.Standing) {
	slices.SortFunc(standings, func(a, b models.Standing) int {
		return cmp.Or(
			cmp.Compare(b.Points, a.Points),
			cmp.Compare(b.GoalDifference, a.GoalDifference),
			cmp.Compare(b.GoalsFor, a.GoalsFor),
			cmp.Compare(a.Team, b.Team),
		)
	})
}

// compareByDate orders matches by date, then by match number.
func compareByDate(a, b models.Match) int {
	return cmp.Or(cmp.Compare(a.Date, b.Date), cmp.Compare(a.Num, b.Num))
}

// sortMatchesByDate sorts matches in place by date, then by match number,
// and returns the same slice for convenience.
func sortMatchesByDate(matches []models.Match) []models.Match {
	slices.SortStableFunc(matches, compareByDate)
	return matches
}
