package data

import "testing"

const testMatchesJSON = `{
  "name": "World Cup Test",
  "matches": [
    {
      "num": 1,
      "round": "Matchday 1",
      "date": "2026-06-11",
      "time": "13:00 UTC-6",
      "team1": "Alpha",
      "team2": "Beta",
      "score": {"ft": [2, 0], "ht": [1, 0]},
      "goals1": [{"name": "Player A", "minute": "9"}],
      "goals2": [],
      "group": "Group A",
      "ground": "Test Stadium"
    },
    {
      "num": 2,
      "round": "Round of 32",
      "date": "2026-06-30",
      "time": "19:00 UTC-6",
      "team1": "Gamma",
      "team2": "Delta",
      "group": "",
      "ground": "Test Stadium 2"
    }
  ]
}`

const testGroupsJSON = `{
  "name": "World Cup Test",
  "groups": [
    {"name": "Group A", "teams": ["Alpha", "Beta"]}
  ]
}`

const testTeamsJSON = `[
  {"name": "Alpha", "continent": "Test", "flag_icon": "🏳", "fifa_code": "ALP", "group": "A", "confed": "TEST"},
  {"name": "Beta", "name_normalised": "Beta Republic", "continent": "Test", "flag_icon": "🏳", "fifa_code": "BET", "group": "A", "confed": "TEST"}
]`

const testStadiumsJSON = `{
  "name": "World Cup Test",
  "stadiums": [
    {"city": "Testville", "timezone": "UTC-6", "cc": "tv", "name": "Test Stadium", "capacity": 50000, "coords": "0,0"}
  ]
}`

func testSources() sourceFiles {
	return sourceFiles{
		Matches:  []byte(testMatchesJSON),
		Groups:   []byte(testGroupsJSON),
		Teams:    []byte(testTeamsJSON),
		Stadiums: []byte(testStadiumsJSON),
	}
}

func TestParse_Matches(t *testing.T) {
	tournament, err := parse(testSources())
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	if got, want := tournament.Name, "World Cup Test"; got != want {
		t.Errorf("Name = %q, want %q", got, want)
	}
	if got, want := len(tournament.Matches), 2; got != want {
		t.Fatalf("len(Matches) = %d, want %d", got, want)
	}

	played := tournament.Matches[0]
	if !played.Played() {
		t.Error("expected first match to be played")
	}
	if played.FullTime == nil || played.FullTime.Home != 2 || played.FullTime.Away != 0 {
		t.Errorf("FullTime = %+v, want {2 0}", played.FullTime)
	}
	if played.Winner() != "Alpha" {
		t.Errorf("Winner() = %q, want Alpha", played.Winner())
	}
	if len(played.Goals1) != 1 || played.Goals1[0].Name != "Player A" {
		t.Errorf("Goals1 = %+v", played.Goals1)
	}

	unplayed := tournament.Matches[1]
	if unplayed.Played() {
		t.Error("expected second match to be unplayed")
	}
	if unplayed.Winner() != "" {
		t.Errorf("Winner() = %q, want empty for unplayed match", unplayed.Winner())
	}
}

func TestParse_TeamsAndGroups(t *testing.T) {
	tournament, err := parse(testSources())
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	if len(tournament.Groups) != 1 || tournament.Groups[0].Name != "Group A" {
		t.Fatalf("Groups = %+v", tournament.Groups)
	}
	if len(tournament.Groups[0].Teams) != 2 {
		t.Fatalf("Group A Teams = %+v", tournament.Groups[0].Teams)
	}

	if len(tournament.Teams) != 2 {
		t.Fatalf("len(Teams) = %d, want 2", len(tournament.Teams))
	}
	beta := tournament.Teams[1]
	if beta.DisplayName() != "Beta Republic" {
		t.Errorf("DisplayName() = %q, want Beta Republic", beta.DisplayName())
	}
	alpha := tournament.Teams[0]
	if alpha.DisplayName() != "Alpha" {
		t.Errorf("DisplayName() = %q, want Alpha (fallback to Name)", alpha.DisplayName())
	}
}

func TestParse_Stadiums(t *testing.T) {
	tournament, err := parse(testSources())
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	if len(tournament.Stadiums) != 1 || tournament.Stadiums[0].City != "Testville" {
		t.Fatalf("Stadiums = %+v", tournament.Stadiums)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	src := testSources()
	src.Matches = []byte("not json")
	if _, err := parse(src); err == nil {
		t.Fatal("expected error for invalid matches JSON, got nil")
	}
}

func TestFallbackSources_ParseSuccessfully(t *testing.T) {
	tournament, err := parse(fallbackSources())
	if err != nil {
		t.Fatalf("embedded fallback failed to parse: %v", err)
	}
	if len(tournament.Matches) == 0 {
		t.Error("expected embedded fallback to contain matches")
	}
	if len(tournament.Groups) == 0 {
		t.Error("expected embedded fallback to contain groups")
	}
	if len(tournament.Teams) == 0 {
		t.Error("expected embedded fallback to contain teams")
	}
	if len(tournament.Stadiums) == 0 {
		t.Error("expected embedded fallback to contain stadiums")
	}
}
