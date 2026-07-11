// Package config loads application configuration from the environment.
package config

import (
	"os"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/data"
)

const (
	defaultPort         = "8080"
	defaultFetchTimeout = 10 * time.Second
	defaultMatchesURL   = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.json"
	defaultGroupsURL    = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.groups.json"
	defaultTeamsURL     = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.teams.json"
	defaultStadiumsURL  = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.stadiums.json"
)

// Config holds runtime configuration for the server.
type Config struct {
	Port         string
	FetchTimeout time.Duration
	SourceURLs   data.SourceURLs
}

// Load builds a Config from environment variables, falling back to defaults.
func Load() Config {
	return Config{
		Port:         envOr("PORT", defaultPort),
		FetchTimeout: defaultFetchTimeout,
		SourceURLs: data.SourceURLs{
			Matches:  envOr("WORLDCUP_MATCHES_URL", defaultMatchesURL),
			Groups:   envOr("WORLDCUP_GROUPS_URL", defaultGroupsURL),
			Teams:    envOr("WORLDCUP_TEAMS_URL", defaultTeamsURL),
			Stadiums: envOr("WORLDCUP_STADIUMS_URL", defaultStadiumsURL),
		},
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
