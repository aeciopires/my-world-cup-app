package services

import (
	"sort"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// PlayerStat is a player's aggregated goal tally across the tournament.
type PlayerStat struct {
	Name  string
	Team  string
	Goals int
}

// TopScorers aggregates goals per player across every match (group and
// knockout stages) and returns the top n, ordered by goals scored
// descending, then alphabetically by name.
func TopScorers(t models.Tournament, n int) []PlayerStat {
	type key struct{ name, team string }
	tally := make(map[key]int)

	for _, m := range t.Matches {
		for _, g := range m.Goals1 {
			tally[key{g.Name, m.Team1}]++
		}
		for _, g := range m.Goals2 {
			tally[key{g.Name, m.Team2}]++
		}
	}

	scorers := make([]PlayerStat, 0, len(tally))
	for k, goals := range tally {
		scorers = append(scorers, PlayerStat{Name: k.name, Team: k.team, Goals: goals})
	}

	sort.SliceStable(scorers, func(i, j int) bool {
		if scorers[i].Goals != scorers[j].Goals {
			return scorers[i].Goals > scorers[j].Goals
		}
		return scorers[i].Name < scorers[j].Name
	})

	if len(scorers) > n {
		scorers = scorers[:n]
	}
	return scorers
}

// TeamStandings aggregates each team's overall record (played, won, drawn,
// lost, goals, points) across every played match in the tournament,
// including both group and knockout stages, unlike GroupStandings which is
// scoped to group-stage matches only. Tie-break order matches
// GroupStandings: points, goal difference, goals for, alphabetical.
func TeamStandings(t models.Tournament) []models.Standing {
	byTeam := make(map[string]*models.Standing)

	getOrCreate := func(team string) *models.Standing {
		s, ok := byTeam[team]
		if !ok {
			s = &models.Standing{Team: team}
			byTeam[team] = s
		}
		return s
	}

	for _, m := range t.Matches {
		if !m.Played() {
			continue
		}
		home := getOrCreate(m.Team1)
		away := getOrCreate(m.Team2)
		accumulateResult(home, away, m)
	}

	standings := make([]models.Standing, 0, len(byTeam))
	for _, s := range byTeam {
		standings = append(standings, *s)
	}
	sortStandings(standings)
	return standings
}
