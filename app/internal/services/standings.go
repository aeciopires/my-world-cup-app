// Package services contains the business logic derived from the raw
// tournament data: standings computation and knockout stage grouping.
package services

import (
	"sort"

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

		standings := make([]models.Standing, 0, len(standingByTeam))
		for _, s := range standingByTeam {
			standings = append(standings, *s)
		}
		sortStandings(standings)

		sortMatchesByRoundAndDate(groupMatches)

		tables = append(tables, GroupTable{Group: g, Standings: standings, Matches: groupMatches})
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
// running standings. The caller is responsible for resolving (or creating)
// the Standing entries and for ensuring m.Played() is true.
func accumulateResult(home, away *models.Standing, m models.Match) {
	home.Played++
	away.Played++
	home.GoalsFor += m.FullTime.Home
	home.GoalsAgainst += m.FullTime.Away
	away.GoalsFor += m.FullTime.Away
	away.GoalsAgainst += m.FullTime.Home

	switch {
	case m.FullTime.Home > m.FullTime.Away:
		home.Won++
		home.Points += pointsForWin
		away.Lost++
	case m.FullTime.Home < m.FullTime.Away:
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

func sortStandings(standings []models.Standing) {
	sort.SliceStable(standings, func(i, j int) bool {
		a, b := standings[i], standings[j]
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.GoalDifference != b.GoalDifference {
			return a.GoalDifference > b.GoalDifference
		}
		if a.GoalsFor != b.GoalsFor {
			return a.GoalsFor > b.GoalsFor
		}
		return a.Team < b.Team
	})
}

func sortMatchesByRoundAndDate(matches []models.Match) {
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Date != matches[j].Date {
			return matches[i].Date < matches[j].Date
		}
		return matches[i].Num < matches[j].Num
	})
}
