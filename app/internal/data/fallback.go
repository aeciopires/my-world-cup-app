package data

import _ "embed"

//go:embed fallback/worldcup.json
var fallbackMatches []byte

//go:embed fallback/worldcup.groups.json
var fallbackGroups []byte

//go:embed fallback/worldcup.teams.json
var fallbackTeams []byte

//go:embed fallback/worldcup.stadiums.json
var fallbackStadiums []byte

// fallbackSources returns the bundled snapshot used when live fetching fails
// or before the first successful refresh completes.
func fallbackSources() sourceFiles {
	return sourceFiles{
		Matches:  fallbackMatches,
		Groups:   fallbackGroups,
		Teams:    fallbackTeams,
		Stadiums: fallbackStadiums,
	}
}
