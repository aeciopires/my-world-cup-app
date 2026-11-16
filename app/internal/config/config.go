// Package config loads application configuration from the environment.
package config

import (
	"cmp"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/data"
)

// DefaultPort is the HTTP port used when PORT is unset.
const DefaultPort = "8080"

const (
	defaultFetchTimeout = 10 * time.Second
	defaultMatchesURL   = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.json"
	defaultGroupsURL    = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.groups.json"
	defaultTeamsURL     = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.teams.json"
	defaultStadiumsURL  = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.stadiums.json"
)

// Log output formats accepted in LOG_FORMAT.
const (
	LogFormatText = "text"
	LogFormatJSON = "json"
)

// Config holds runtime configuration for the server.
type Config struct {
	Port         string
	FetchTimeout time.Duration
	// LogFormat is LogFormatJSON (the default: one JSON object per line,
	// parsed by log collectors such as Loki) or LogFormatText (logfmt-style,
	// easier to read in a terminal).
	LogFormat string
	// ShutdownDrainDelay is how long the server keeps serving after a
	// shutdown signal while /readyz already reports 503, giving a load
	// balancer (e.g. Kubernetes endpoints) time to stop routing new
	// requests here before connections are closed. 0 disables it.
	ShutdownDrainDelay time.Duration
	SourceURLs         data.SourceURLs
}

// Load builds a Config from environment variables, falling back to defaults.
func Load() Config {
	return Config{
		Port:               cmp.Or(os.Getenv("PORT"), DefaultPort),
		FetchTimeout:       durationEnv("FETCH_TIMEOUT", defaultFetchTimeout),
		LogFormat:          logFormatEnv("LOG_FORMAT"),
		ShutdownDrainDelay: durationEnv("SHUTDOWN_DRAIN_DELAY", 0),
		SourceURLs: data.SourceURLs{
			Matches:  cmp.Or(os.Getenv("WORLDCUP_MATCHES_URL"), defaultMatchesURL),
			Groups:   cmp.Or(os.Getenv("WORLDCUP_GROUPS_URL"), defaultGroupsURL),
			Teams:    cmp.Or(os.Getenv("WORLDCUP_TEAMS_URL"), defaultTeamsURL),
			Stadiums: cmp.Or(os.Getenv("WORLDCUP_STADIUMS_URL"), defaultStadiumsURL),
		},
	}
}

// logFormatEnv returns LogFormatJSON (the default) or LogFormatText. An
// unknown value falls back to JSON with a warning.
func logFormatEnv(key string) string {
	switch raw := strings.ToLower(os.Getenv(key)); raw {
	case "", LogFormatJSON:
		return LogFormatJSON
	case LogFormatText:
		return LogFormatText
	default:
		slog.Warn("invalid log format in environment, using default", "key", key, "value", raw, "default", LogFormatJSON)
		return LogFormatJSON
	}
}

// durationEnv parses key as a Go duration (e.g. "15s"). Unset, invalid, or
// non-positive values fall back to def, logging a warning for the latter two.
func durationEnv(key string, def time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		slog.Warn("invalid duration in environment, using default", "key", key, "value", raw, "default", def.String())
		return def
	}
	return d
}
