package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/handlers"
)

// newTestStore returns a Store seeded from the embedded fallback data,
// pointed at a client whose URLs never resolve, so tests exercise the
// router against a stable, deterministic dataset without network access.
func newTestStore(t *testing.T) *data.Store {
	t.Helper()
	client := data.NewClient(data.SourceURLs{}, time.Second)
	return data.NewStore(client)
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	router, err := handlers.NewRouter(newTestStore(t), time.Second)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	return router
}

func TestRoutes_ReturnOK(t *testing.T) {
	router := newTestRouter(t)

	routes := []string{"/", "/groups", "/knockout", "/matches", "/links", "/stats", "/healthz", "/metrics"}
	for _, path := range routes {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s: status = %d, want 200; body: %s", path, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHomePage_ContainsTournamentContent(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !containsAll(body, "Upcoming Matches", "Recent Results", "My World Cup App") {
		t.Errorf("home page body missing expected content: %s", body)
	}
}

func TestLinksPage_ContainsAllFIFALinks(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	body := rec.Body.String()
	required := []string{
		"fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/teams",
		"fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/standings",
		"fifa.com/pt/tournaments/mens/worldcup/canadamexicousa2026/scores-fixtures",
		"fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/stadiums",
		"fifa.com/en/tournaments/mens/club-world-cup/usa-2025",
		"fifa.com/en/tournaments/mens/worldcup/canadamexicousa2026/fifa-sound",
		"open.spotify.com/playlist",
	}
	if !containsAll(body, required...) {
		t.Errorf("links page missing one or more required FIFA links: %s", body)
	}
}

func TestGroupsPage_ContainsStandingsHeaders(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !containsAll(body, "Group A", "Pts", "GD") {
		t.Errorf("groups page missing expected standings content: %s", body)
	}
}

func TestRefresh_FailedFetchReturnsBadGatewayButServerStaysUp(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// The unresolvable source URLs (empty strings) must fail, but the
	// handler should report the failure via HTTP status rather than panic.
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rec.Code)
	}

	// Confirm the app still serves pages after a failed refresh.
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("home page after failed refresh: status = %d, want 200", rec2.Code)
	}
}

func TestStatsPage_ContainsExpectedSections(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !containsAll(body, "Top Scorers", "Team Overview") {
		t.Errorf("stats page missing expected sections: %s", body)
	}
}

func TestMetricsEndpoint_ExposesExpectedSeries(t *testing.T) {
	router := newTestRouter(t)

	// Generate some traffic so the counters have at least one sample.
	for _, path := range []string{"/", "/groups"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics: status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !containsAll(body, "http_requests_total", "http_request_duration_seconds", "data_refresh_total") {
		t.Errorf("/metrics missing expected series: %s", body)
	}
}

func TestStaticAssets_AreServed(t *testing.T) {
	router := newTestRouter(t)

	for _, path := range []string{"/static/css/style.css", "/static/js/app.js"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", path, rec.Code)
		}
	}
}

func containsAll(haystack string, needles ...string) bool {
	for _, n := range needles {
		if !strings.Contains(haystack, n) {
			return false
		}
	}
	return true
}
