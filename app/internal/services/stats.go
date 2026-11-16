package services

import (
	"cmp"
	"slices"

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
// descending, then alphabetically by name, then by team. Own goals are
// excluded: they count for the opposing team's score but are not credited
// to the scorer, as in FIFA's Golden Boot ranking. A non-positive n returns
// an empty result.
func TopScorers(t models.Tournament, n int) []PlayerStat {
	if n <= 0 {
		return []PlayerStat{}
	}

	type key struct{ name, team string }
	tally := make(map[key]int)
	count := func(goals []models.Goal, team string) {
		for _, g := range goals {
			if g.OwnGoal {
				continue
			}
			tally[key{g.Name, team}]++
		}
	}

	for _, m := range t.Matches {
		count(m.Goals1, m.Team1)
		count(m.Goals2, m.Team2)
	}

	scorers := make([]PlayerStat, 0, len(tally))
	for k, goals := range tally {
		scorers = append(scorers, PlayerStat{Name: k.name, Team: k.team, Goals: goals})
	}

	// Name+Team is the map key, so this order is total and deterministic
	// even though map iteration order is random.
	slices.SortFunc(scorers, func(a, b PlayerStat) int {
		return cmp.Or(
			cmp.Compare(b.Goals, a.Goals),
			cmp.Compare(a.Name, b.Name),
			cmp.Compare(a.Team, b.Team),
		)
	})

	return scorers[:min(n, len(scorers))]
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
		accumulateResult(getOrCreate(m.Team1), getOrCreate(m.Team2), m)
	}

	return collectStandings(byTeam)
}
