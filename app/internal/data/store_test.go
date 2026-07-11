package data

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/matches.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testMatchesJSON))
	})
	mux.HandleFunc("/groups.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testGroupsJSON))
	})
	mux.HandleFunc("/teams.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testTeamsJSON))
	})
	mux.HandleFunc("/stadiums.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testStadiumsJSON))
	})
	mux.HandleFunc("/missing.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}

func TestNewStore_SeedsFromFallback(t *testing.T) {
	client := NewClient(SourceURLs{}, time.Second)
	store := NewStore(client)

	tournament, lastUpdated, source := store.Snapshot()
	if source != "fallback" {
		t.Errorf("source = %q, want fallback", source)
	}
	if lastUpdated.IsZero() {
		t.Error("expected lastUpdated to be set after seeding")
	}
	if len(tournament.Matches) == 0 {
		t.Error("expected fallback tournament to have matches")
	}
}

func TestStore_Refresh_Success(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	client := NewClient(SourceURLs{
		Matches:  srv.URL + "/matches.json",
		Groups:   srv.URL + "/groups.json",
		Teams:    srv.URL + "/teams.json",
		Stadiums: srv.URL + "/stadiums.json",
	}, time.Second)
	store := NewStore(client)

	if err := store.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	tournament, _, source := store.Snapshot()
	if source != "live" {
		t.Errorf("source = %q, want live", source)
	}
	if tournament.Name != "World Cup Test" {
		t.Errorf("Name = %q, want World Cup Test", tournament.Name)
	}
}

func TestStore_Refresh_FailureKeepsPreviousSnapshot(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	client := NewClient(SourceURLs{
		Matches:  srv.URL + "/missing.json",
		Groups:   srv.URL + "/groups.json",
		Teams:    srv.URL + "/teams.json",
		Stadiums: srv.URL + "/stadiums.json",
	}, time.Second)
	store := NewStore(client)

	_, beforeUpdated, beforeSource := store.Snapshot()

	if err := store.Refresh(context.Background()); err == nil {
		t.Fatal("expected Refresh() to return an error for a 404 response")
	}

	_, afterUpdated, afterSource := store.Snapshot()
	if afterSource != beforeSource {
		t.Errorf("source changed from %q to %q after failed refresh", beforeSource, afterSource)
	}
	if !afterUpdated.Equal(beforeUpdated) {
		t.Error("lastUpdated changed after failed refresh")
	}
	if store.LastError() == "" {
		t.Error("expected LastError() to be populated after failed refresh")
	}
}
