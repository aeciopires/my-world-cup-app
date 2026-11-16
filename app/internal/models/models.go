// Package models defines the core domain types for the World Cup 2026 tournament.
package models

// Team represents a national team competing in the tournament.
type Team struct {
	Name           string
	NameNormalised string
	Continent      string
	FlagIcon       string
	FIFACode       string
	Group          string
	Confederation  string
}

// DisplayName returns the normalised name when available, falling back to Name.
func (t Team) DisplayName() string {
	if t.NameNormalised != "" {
		return t.NameNormalised
	}
	return t.Name
}

// Group represents a group stage group (e.g. "Group A") and its teams.
type Group struct {
	Name  string
	Teams []string
}

// Goal represents a single goal scored during a match. Goals are listed
// under the team they count for, so for an own goal Name is a player of the
// *opposing* team.
type Goal struct {
	Name    string
	Minute  string
	Penalty bool
	OwnGoal bool
}

// Score holds the goals scored by each side at a given stage of the match.
type Score struct {
	Home int
	Away int
}

// Match represents a single fixture, played or upcoming.
type Match struct {
	Num       int
	Round     string
	Date      string
	Time      string
	Team1     string
	Team2     string
	Group     string
	Ground    string
	Goals1    []Goal
	Goals2    []Goal
	FullTime  *Score
	HalfTime  *Score
	ExtraTime *Score
	Penalties *Score
}

// Played reports whether the match has a recorded full-time result.
func (m Match) Played() bool {
	return m.FullTime != nil
}

// WentToExtraTime reports whether the match has a recorded extra-time score.
func (m Match) WentToExtraTime() bool {
	return m.ExtraTime != nil
}

// FinalScore returns the score at the end of play (excluding any penalty
// shoot-out): the extra-time score when the match went to extra time,
// otherwise the full-time score. It returns nil for an unplayed match.
//
// openfootball reports the extra-time score cumulatively (it already
// includes the goals scored in the first 90 minutes), so it can be used
// as-is rather than added to FullTime.
func (m Match) FinalScore() *Score {
	if m.FullTime == nil {
		return nil
	}
	if m.ExtraTime != nil {
		return m.ExtraTime
	}
	return m.FullTime
}

// Winner returns the name of the winning team, or "" for a draw/unplayed
// match. The score after extra time (if played) decides the winner; for a
// knockout fixture still level after that, the penalty shoot-out (if
// recorded) decides it.
func (m Match) Winner() string {
	final := m.FinalScore()
	if final == nil {
		return ""
	}
	if w := m.sideWinning(*final); w != "" {
		return w
	}
	if m.Penalties != nil {
		return m.sideWinning(*m.Penalties)
	}
	return ""
}

// sideWinning returns the team ahead in s, or "" when s is level.
func (m Match) sideWinning(s Score) string {
	switch {
	case s.Home > s.Away:
		return m.Team1
	case s.Away > s.Home:
		return m.Team2
	default:
		return ""
	}
}

// Stadium represents a venue hosting matches.
type Stadium struct {
	City     string
	Country  string
	Timezone string
	Name     string
	Capacity int
	Coords   string
}

// Standing represents a single row of a group table.
type Standing struct {
	Team           string
	Played         int
	Won            int
	Drawn          int
	Lost           int
	GoalsFor       int
	GoalsAgainst   int
	GoalDifference int
	Points         int
}

// Tournament is the full in-memory snapshot of the World Cup data.
type Tournament struct {
	Name     string
	Teams    []Team
	Groups   []Group
	Matches  []Match
	Stadiums []Stadium
}
