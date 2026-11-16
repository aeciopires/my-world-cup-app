package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	for _, key := range []string{"PORT", "FETCH_TIMEOUT", "WORLDCUP_MATCHES_URL", "WORLDCUP_GROUPS_URL", "WORLDCUP_TEAMS_URL", "WORLDCUP_STADIUMS_URL", "LOG_FORMAT", "SHUTDOWN_DRAIN_DELAY"} {
		t.Setenv(key, "")
	}

	cfg := Load()
	if cfg.Port != DefaultPort {
		t.Errorf("Port = %q, want %q", cfg.Port, DefaultPort)
	}
	if cfg.FetchTimeout != defaultFetchTimeout {
		t.Errorf("FetchTimeout = %v, want %v", cfg.FetchTimeout, defaultFetchTimeout)
	}
	if cfg.LogFormat != LogFormatJSON {
		t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, LogFormatJSON)
	}
	if cfg.ShutdownDrainDelay != 0 {
		t.Errorf("ShutdownDrainDelay = %v, want 0", cfg.ShutdownDrainDelay)
	}
	if cfg.SourceURLs.Matches != defaultMatchesURL {
		t.Errorf("SourceURLs.Matches = %q, want default", cfg.SourceURLs.Matches)
	}
}

func TestLoad_Overrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("FETCH_TIMEOUT", "30s")
	t.Setenv("WORLDCUP_TEAMS_URL", "http://example.test/teams.json")
	t.Setenv("LOG_FORMAT", "Text")
	t.Setenv("SHUTDOWN_DRAIN_DELAY", "5s")

	cfg := Load()
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.FetchTimeout != 30*time.Second {
		t.Errorf("FetchTimeout = %v, want 30s", cfg.FetchTimeout)
	}
	if cfg.LogFormat != LogFormatText {
		t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, LogFormatText)
	}
	if cfg.ShutdownDrainDelay != 5*time.Second {
		t.Errorf("ShutdownDrainDelay = %v, want 5s", cfg.ShutdownDrainDelay)
	}
	if cfg.SourceURLs.Teams != "http://example.test/teams.json" {
		t.Errorf("SourceURLs.Teams = %q, want override", cfg.SourceURLs.Teams)
	}
}

func TestLoad_InvalidFetchTimeoutFallsBackToDefault(t *testing.T) {
	for _, v := range []string{"soon", "-5s", "0"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("FETCH_TIMEOUT", v)
			if got := Load().FetchTimeout; got != defaultFetchTimeout {
				t.Errorf("FetchTimeout = %v, want default %v", got, defaultFetchTimeout)
			}
		})
	}
}

func TestLoad_InvalidLogFormatFallsBack(t *testing.T) {
	t.Setenv("LOG_FORMAT", "xml")
	if got := Load().LogFormat; got != LogFormatJSON {
		t.Errorf("LogFormat = %q, want %q", got, LogFormatJSON)
	}
}
