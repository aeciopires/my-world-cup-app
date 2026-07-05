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

// Goal represents a single goal scored during a match.
type Goal struct {
	Name    string
	Minute  string
	Penalty bool
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

// Winner returns the name of the winning team, or "" for a draw/unplayed
// match. For a knockout fixture that finishes level after full time, the
// penalty shoot-out (if recorded) decides the winner.
func (m Match) Winner() string {
	if m.FullTime == nil {
		return ""
	}
	if m.FullTime.Home > m.FullTime.Away {
		return m.Team1
	}
	if m.FullTime.Away > m.FullTime.Home {
		return m.Team2
	}
	if m.Penalties != nil {
		if m.Penalties.Home > m.Penalties.Away {
			return m.Team1
		}
		if m.Penalties.Away > m.Penalties.Home {
			return m.Team2
		}
	}
	return ""
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
