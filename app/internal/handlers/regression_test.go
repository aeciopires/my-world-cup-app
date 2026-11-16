package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, router http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestKnockoutPage_VenueLinksToHostCity(t *testing.T) {
	// The Match Details tables used to link .Ground through StadiumLinks
	// (keyed by stadium name), which never matches a host-city string.
	body := get(t, newTestRouter(t), "/knockout").Body.String()
	if !strings.Contains(body, "canadamexicousa2026/mexico/mexico-city") {
		t.Error("knockout Match Details should link the Mexico City venue to its FIFA host-city page")
	}
}

func TestKnockoutPage_ShowsPenaltiesForShootOuts(t *testing.T) {
	// Fallback data: Round of 32, Germany 1-1 Paraguay (a.e.t.), 3-4 on pens.
	body := get(t, newTestRouter(t), "/knockout").Body.String()
	if !containsAll(body, "(pens 3-4)", "a.e.t.") {
		t.Error("knockout page should mark extra time and show the shoot-out score")
	}
}

func TestMatchesPage_TeamFilterExcludesPlaceholders(t *testing.T) {
	body := get(t, newTestRouter(t), "/matches").Body.String()
	if strings.Contains(body, `<option value="W74"`) {
		t.Error(`team filter offers the knockout placeholder "W74" as a team`)
	}
}

func TestResponses_IncludeSecurityHeaders(t *testing.T) {
	rec := get(t, newTestRouter(t), "/")
	for header, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestMetrics_UnmatchedRoutesShareOneLabel(t *testing.T) {
	router := newTestRouter(t)
	for _, path := range []string{"/wp-admin/setup.php", "/.env", "/does/not/exist"} {
		if rec := get(t, router, path); rec.Code != http.StatusNotFound {
			t.Fatalf("GET %s: status = %d, want 404", path, rec.Code)
		}
	}

	body := get(t, router, "/metrics").Body.String()
	if strings.Contains(body, `path="/wp-admin/setup.php"`) || strings.Contains(body, `path="/.env"`) {
		t.Error("unmatched request paths leaked into the metrics path label (unbounded cardinality)")
	}
	if !strings.Contains(body, `path="unmatched"`) {
		t.Error(`expected 404s to be recorded under path="unmatched"`)
	}
}
