// Package data fetches, parses, and caches World Cup 2026 tournament data.
package data

import (
	"encoding/json"
	"fmt"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// rawMatchesFile mirrors the shape of worldcup.json from openfootball/worldcup.json.
type rawMatchesFile struct {
	Name    string     `json:"name"`
	Matches []rawMatch `json:"matches"`
}

type rawMatch struct {
	Num    int       `json:"num"`
	Round  string    `json:"round"`
	Date   string    `json:"date"`
	Time   string    `json:"time"`
	Team1  string    `json:"team1"`
	Team2  string    `json:"team2"`
	Group  string    `json:"group"`
	Ground string    `json:"ground"`
	Score  *rawScore `json:"score"`
	Goals1 []rawGoal `json:"goals1"`
	Goals2 []rawGoal `json:"goals2"`
}

type rawScore struct {
	FT []int `json:"ft"`
	HT []int `json:"ht"`
	ET []int `json:"et"`
	P  []int `json:"p"`
}

type rawGoal struct {
	Name    string `json:"name"`
	Minute  string `json:"minute"`
	Penalty bool   `json:"penalty"`
}

// rawGroupsFile mirrors worldcup.groups.json.
type rawGroupsFile struct {
	Name   string     `json:"name"`
	Groups []rawGroup `json:"groups"`
}

type rawGroup struct {
	Name  string   `json:"name"`
	Teams []string `json:"teams"`
}

// rawTeam mirrors an entry of worldcup.teams.json.
type rawTeam struct {
	Name           string `json:"name"`
	NameNormalised string `json:"name_normalised"`
	Continent      string `json:"continent"`
	FlagIcon       string `json:"flag_icon"`
	FIFACode       string `json:"fifa_code"`
	Group          string `json:"group"`
	Confederation  string `json:"confed"`
}

// rawStadiumsFile mirrors worldcup.stadiums.json.
type rawStadiumsFile struct {
	Name     string       `json:"name"`
	Stadiums []rawStadium `json:"stadiums"`
}

type rawStadium struct {
	City     string `json:"city"`
	Timezone string `json:"timezone"`
	CC       string `json:"cc"`
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Coords   string `json:"coords"`
}

// sourceFiles bundles the raw bytes of the four source JSON documents.
type sourceFiles struct {
	Matches  []byte
	Groups   []byte
	Teams    []byte
	Stadiums []byte
}

// parse decodes and normalizes the raw source files into a Tournament.
func parse(src sourceFiles) (models.Tournament, error) {
	var matchesFile rawMatchesFile
	if err := json.Unmarshal(src.Matches, &matchesFile); err != nil {
		return models.Tournament{}, fmt.Errorf("parse matches: %w", err)
	}

	var groupsFile rawGroupsFile
	if err := json.Unmarshal(src.Groups, &groupsFile); err != nil {
		return models.Tournament{}, fmt.Errorf("parse groups: %w", err)
	}

	var rawTeams []rawTeam
	if err := json.Unmarshal(src.Teams, &rawTeams); err != nil {
		return models.Tournament{}, fmt.Errorf("parse teams: %w", err)
	}

	var stadiumsFile rawStadiumsFile
	if err := json.Unmarshal(src.Stadiums, &stadiumsFile); err != nil {
		return models.Tournament{}, fmt.Errorf("parse stadiums: %w", err)
	}

	tournament := models.Tournament{
		Name:     matchesFile.Name,
		Teams:    normalizeTeams(rawTeams),
		Groups:   normalizeGroups(groupsFile.Groups),
		Matches:  normalizeMatches(matchesFile.Matches),
		Stadiums: normalizeStadiums(stadiumsFile.Stadiums),
	}
	return tournament, nil
}

func normalizeTeams(raw []rawTeam) []models.Team {
	teams := make([]models.Team, 0, len(raw))
	for _, t := range raw {
		teams = append(teams, models.Team{
			Name:           t.Name,
			NameNormalised: t.NameNormalised,
			Continent:      t.Continent,
			FlagIcon:       t.FlagIcon,
			FIFACode:       t.FIFACode,
			Group:          t.Group,
			Confederation:  t.Confederation,
		})
	}
	return teams
}

func normalizeGroups(raw []rawGroup) []models.Group {
	groups := make([]models.Group, 0, len(raw))
	for _, g := range raw {
		groups = append(groups, models.Group{
			Name:  g.Name,
			Teams: g.Teams,
		})
	}
	return groups
}

func normalizeMatches(raw []rawMatch) []models.Match {
	matches := make([]models.Match, 0, len(raw))
	for _, m := range raw {
		match := models.Match{
			Num:    m.Num,
			Round:  m.Round,
			Date:   m.Date,
			Time:   m.Time,
			Team1:  m.Team1,
			Team2:  m.Team2,
			Group:  m.Group,
			Ground: m.Ground,
			Goals1: normalizeGoals(m.Goals1),
			Goals2: normalizeGoals(m.Goals2),
		}
		if m.Score != nil {
			match.FullTime = scoreFromPair(m.Score.FT)
			match.HalfTime = scoreFromPair(m.Score.HT)
			match.ExtraTime = scoreFromPair(m.Score.ET)
			match.Penalties = scoreFromPair(m.Score.P)
		}
		matches = append(matches, match)
	}
	return matches
}

func normalizeGoals(raw []rawGoal) []models.Goal {
	if len(raw) == 0 {
		return nil
	}
	goals := make([]models.Goal, 0, len(raw))
	for _, g := range raw {
		goals = append(goals, models.Goal{
			Name:    g.Name,
			Minute:  g.Minute,
			Penalty: g.Penalty,
		})
	}
	return goals
}

func scoreFromPair(pair []int) *models.Score {
	if len(pair) != 2 {
		return nil
	}
	return &models.Score{Home: pair[0], Away: pair[1]}
}

func normalizeStadiums(raw []rawStadium) []models.Stadium {
	stadiums := make([]models.Stadium, 0, len(raw))
	for _, s := range raw {
		stadiums = append(stadiums, models.Stadium{
			City:     s.City,
			Country:  s.CC,
			Timezone: s.Timezone,
			Name:     s.Name,
			Capacity: s.Capacity,
			Coords:   s.Coords,
		})
	}
	return stadiums
}
