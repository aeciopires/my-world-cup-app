package handlers_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/aeciopires/my-world-cup-app/internal/handlers"
	"github.com/aeciopires/my-world-cup-app/internal/telemetry"
)

func TestReadyz_ReportsReadyThenDraining(t *testing.T) {
	health := handlers.NewHealth()
	router, err := handlers.NewRouter(newTestStore(t), time.Second, health)
	if err != nil {
		t.Fatal(err)
	}

	get := func() (int, map[string]string) {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("readyz body is not JSON: %v (%q)", err, rec.Body.String())
		}
		return rec.Code, body
	}

	code, body := get()
	if code != http.StatusOK || body["status"] != "ready" || body["data_source"] != "fallback" {
		t.Errorf("before draining: %d %v, want 200 ready/fallback", code, body)
	}

	health.SetDraining()
	code, body = get()
	if code != http.StatusServiceUnavailable || body["status"] != "draining" {
		t.Errorf("after SetDraining: %d %v, want 503 draining", code, body)
	}

	// Liveness must stay green while draining, or Kubernetes would kill
	// the pod before in-flight requests finish.
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("healthz while draining = %d, want 200", rec.Code)
	}
}

// An incoming W3C traceparent must reach the request log even with tracing
// disabled (no-op tracer), so logs can be correlated with upstream traces.
func TestRequestLog_CarriesIncomingTraceID(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(telemetry.NewLogger(&buf, true))
	t.Cleanup(func() { slog.SetDefault(prev) })

	router := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/groups", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	router.ServeHTTP(httptest.NewRecorder(), req)

	var requestLine string
	for line := range strings.Lines(buf.String()) {
		if strings.Contains(line, `"msg":"request"`) {
			requestLine = line
		}
	}
	if !strings.Contains(requestLine, `"trace_id":"4bf92f3577b34da6a3ce929d0e0e4736"`) ||
		!strings.Contains(requestLine, `"route":"/groups"`) {
		t.Errorf("request log missing trace_id/route: %q", requestLine)
	}
}

func TestMetrics_OpenMetricsExposesTraceExemplars(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	router.ServeHTTP(httptest.NewRecorder(), req)

	scrape := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	scrape.Header.Set("Accept", "application/openmetrics-text; version=1.0.0")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, scrape)

	if !strings.Contains(rec.Body.String(), `trace_id="0af7651916cd43dd8448eb211c80319c"`) {
		t.Errorf("no exemplar with the request's trace_id in OpenMetrics output")
	}
}
