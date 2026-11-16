package data

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	serve := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }
	}
	mux.HandleFunc("/matches.json", serve(testMatchesJSON))
	mux.HandleFunc("/groups.json", serve(testGroupsJSON))
	mux.HandleFunc("/teams.json", serve(testTeamsJSON))
	mux.HandleFunc("/stadiums.json", serve(testStadiumsJSON))
	mux.HandleFunc("/missing.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/huge.json", serve(strings.Repeat(" ", maxSourceBytes+1)))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func testURLs(srv *httptest.Server) SourceURLs {
	return SourceURLs{
		Matches:  srv.URL + "/matches.json",
		Groups:   srv.URL + "/groups.json",
		Teams:    srv.URL + "/teams.json",
		Stadiums: srv.URL + "/stadiums.json",
	}
}

func TestNewStore_SeedsFromFallback(t *testing.T) {
	store := NewStore(NewClient(SourceURLs{}, time.Second))

	snap := store.Snapshot()
	if snap.Source != SourceFallback {
		t.Errorf("Source = %q, want %q", snap.Source, SourceFallback)
	}
	if snap.LastUpdated.IsZero() {
		t.Error("expected LastUpdated to be set after seeding")
	}
	if len(snap.Tournament.Matches) == 0 {
		t.Error("expected fallback tournament to have matches")
	}
}

func TestStore_Refresh_Success(t *testing.T) {
	srv := newTestServer(t)
	store := NewStore(NewClient(testURLs(srv), time.Second))

	if err := store.Refresh(t.Context()); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	snap := store.Snapshot()
	if snap.Source != SourceLive {
		t.Errorf("Source = %q, want %q", snap.Source, SourceLive)
	}
	if snap.Tournament.Name != "World Cup Test" {
		t.Errorf("Name = %q, want World Cup Test", snap.Tournament.Name)
	}
	if store.LastError() != "" {
		t.Errorf("LastError() = %q, want empty after a successful refresh", store.LastError())
	}
}

func TestStore_Refresh_FailureKeepsPreviousSnapshot(t *testing.T) {
	srv := newTestServer(t)
	urls := testURLs(srv)
	urls.Matches = srv.URL + "/missing.json"
	store := NewStore(NewClient(urls, time.Second))

	before := store.Snapshot()

	if err := store.Refresh(t.Context()); err == nil {
		t.Fatal("expected Refresh() to return an error for a 404 response")
	}

	after := store.Snapshot()
	if after.Source != before.Source {
		t.Errorf("Source changed from %q to %q after failed refresh", before.Source, after.Source)
	}
	if !after.LastUpdated.Equal(before.LastUpdated) {
		t.Error("LastUpdated changed after failed refresh")
	}
	if !strings.Contains(store.LastError(), "fetch matches") {
		t.Errorf("LastError() = %q, want it to name the failing source", store.LastError())
	}
}

func TestStore_Refresh_RejectsConcurrentRefresh(t *testing.T) {
	store := NewStore(NewClient(SourceURLs{}, time.Second))

	// Simulate a refresh already in flight.
	store.refreshMu.Lock()
	defer store.refreshMu.Unlock()

	if err := store.Refresh(t.Context()); !errors.Is(err, ErrRefreshInProgress) {
		t.Fatalf("Refresh() error = %v, want ErrRefreshInProgress", err)
	}
}

func TestClient_Fetch_RejectsOversizedResponse(t *testing.T) {
	srv := newTestServer(t)
	urls := testURLs(srv)
	urls.Teams = srv.URL + "/huge.json"

	_, err := NewClient(urls, 5*time.Second).Fetch(t.Context())
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("Fetch() error = %v, want a size-limit error", err)
	}
}
