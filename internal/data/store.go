package data

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// Store holds the current tournament snapshot in memory and knows how to
// refresh it from the live source, falling back to the embedded snapshot.
type Store struct {
	client *Client

	mu          sync.RWMutex
	tournament  models.Tournament
	lastUpdated time.Time
	lastSource  string // "live" or "fallback"
	lastError   string
}

// NewStore creates a Store seeded with the embedded fallback data so the
// application has content to serve before the first live refresh completes.
func NewStore(client *Client) *Store {
	s := &Store{client: client}
	tournament, err := parse(fallbackSources())
	if err != nil {
		// The embedded fallback is a build-time asset validated by tests;
		// a parse failure here indicates a broken build, not a runtime condition.
		panic("data: embedded fallback failed to parse: " + err.Error())
	}
	s.set(tournament, "fallback", nil)
	return s
}

// Snapshot returns the current tournament data along with refresh metadata.
func (s *Store) Snapshot() (models.Tournament, time.Time, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tournament, s.lastUpdated, s.lastSource
}

// Refresh fetches the latest data from the live source. On failure the
// previously loaded snapshot (live or fallback) is kept in place.
func (s *Store) Refresh(ctx context.Context) error {
	src, err := s.client.Fetch(ctx)
	if err != nil {
		s.mu.Lock()
		s.lastError = err.Error()
		s.mu.Unlock()
		slog.Warn("data refresh failed, keeping previous snapshot", "error", err)
		return err
	}

	tournament, err := parse(src)
	if err != nil {
		s.mu.Lock()
		s.lastError = err.Error()
		s.mu.Unlock()
		slog.Warn("data refresh parse failed, keeping previous snapshot", "error", err)
		return err
	}

	s.set(tournament, "live", nil)
	slog.Info("data refreshed from live source", "matches", len(tournament.Matches))
	return nil
}

func (s *Store) set(t models.Tournament, source string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tournament = t
	s.lastUpdated = time.Now()
	s.lastSource = source
	if err != nil {
		s.lastError = err.Error()
	} else {
		s.lastError = ""
	}
}

// LastError returns the message of the most recent refresh failure, if any.
func (s *Store) LastError() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}
