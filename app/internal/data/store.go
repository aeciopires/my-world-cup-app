package data

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/aeciopires/my-world-cup-app/internal/models"
)

// tracer creates the data-refresh spans. It resolves the global provider
// lazily, so it picks up the provider installed by telemetry.Setup.
var tracer = otel.Tracer("github.com/aeciopires/my-world-cup-app/internal/data")

// Data source labels reported by Snapshot.
const (
	SourceLive     = "live"
	SourceFallback = "fallback"
)

// ErrRefreshInProgress is returned by Refresh when another refresh is
// already running. Concurrent refreshes would only duplicate the same
// upstream downloads, so the extra callers are turned away instead.
var ErrRefreshInProgress = errors.New("data refresh already in progress")

// Snapshot is an immutable, point-in-time view of the store's contents.
// Callers must treat Tournament (and the slices inside it) as read-only:
// the same value is shared by every concurrent request.
type Snapshot struct {
	Tournament  models.Tournament
	LastUpdated time.Time
	Source      string // SourceLive or SourceFallback
}

// Store holds the current tournament snapshot in memory and knows how to
// refresh it from the live source, falling back to the embedded snapshot.
type Store struct {
	client *Client

	// refreshMu ensures only one Refresh talks to the upstream at a time.
	refreshMu sync.Mutex

	mu        sync.RWMutex
	snapshot  Snapshot
	lastError string
}

// NewStore creates a Store seeded with the embedded fallback data so the
// application has content to serve before the first live refresh completes.
func NewStore(client *Client) *Store {
	tournament, err := parse(fallbackSources())
	if err != nil {
		// The embedded fallback is a build-time asset validated by tests;
		// a parse failure here indicates a broken build, not a runtime condition.
		panic("data: embedded fallback failed to parse: " + err.Error())
	}
	s := &Store{client: client}
	s.set(tournament, SourceFallback)
	return s
}

// Snapshot returns the current tournament data along with refresh metadata,
// read atomically so all three fields always belong to the same refresh.
func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

// Refresh fetches the latest data from the live source. On failure the
// previously loaded snapshot (live or fallback) is kept in place. If another
// refresh is already running it returns ErrRefreshInProgress immediately.
func (s *Store) Refresh(ctx context.Context) error {
	if !s.refreshMu.TryLock() {
		return ErrRefreshInProgress
	}
	defer s.refreshMu.Unlock()

	ctx, span := tracer.Start(ctx, "data.Refresh")
	defer span.End()

	tournament, err := s.fetchAndParse(ctx)
	if err != nil {
		s.mu.Lock()
		s.lastError = err.Error()
		s.mu.Unlock()
		span.RecordError(err)
		span.SetStatus(codes.Error, "refresh failed")
		slog.WarnContext(ctx, "data refresh failed, keeping previous snapshot", "error", err)
		return err
	}

	s.set(tournament, SourceLive)
	span.SetAttributes(attribute.Int("worldcup.matches", len(tournament.Matches)))
	slog.InfoContext(ctx, "data refreshed from live source", "matches", len(tournament.Matches))
	return nil
}

func (s *Store) fetchAndParse(ctx context.Context) (models.Tournament, error) {
	src, err := s.client.Fetch(ctx)
	if err != nil {
		return models.Tournament{}, err
	}
	return parse(src)
}

func (s *Store) set(t models.Tournament, source string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = Snapshot{Tournament: t, LastUpdated: time.Now(), Source: source}
	s.lastError = ""
}

// LastError returns the message of the most recent refresh failure, if any.
func (s *Store) LastError() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}
