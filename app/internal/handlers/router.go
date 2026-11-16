package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/metrics"
	"github.com/aeciopires/my-world-cup-app/web"
)

// unmatchedRoute is the metrics "path" label for requests that matched no
// route (404s, 405s). Using the raw URL path there would let any client
// create unbounded label values — e.g. a vulnerability scanner probing
// thousands of paths — and blow up Prometheus' memory.
const unmatchedRoute = "unmatched"

// NewRouter builds the application's HTTP handler, wiring pages, the
// refresh endpoint, health probes, and static asset serving. health backs
// /readyz; the caller keeps it to flag draining during shutdown.
func NewRouter(store *data.Store, fetchTimeout time.Duration, health *Health) (http.Handler, error) {
	renderer, err := NewRenderer()
	if err != nil {
		return nil, err
	}

	pages := NewPageHandlers(store, renderer)
	refresh := NewRefreshHandlers(store, fetchTimeout)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", pages.Home)
	mux.HandleFunc("GET /groups", pages.Groups)
	mux.HandleFunc("GET /knockout", pages.Knockout)
	mux.HandleFunc("GET /matches", pages.Matches)
	mux.HandleFunc("GET /links", pages.Links)
	mux.HandleFunc("GET /stats", pages.Stats)
	mux.HandleFunc("POST /refresh", refresh.Refresh)
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("GET /readyz", health.readyz(store))
	mux.Handle("GET /metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
		// OpenMetrics is the only exposition format that carries exemplars
		// (trace IDs attached to samples); it's used only when the scraper
		// asks for it, so plain-text scrapers are unaffected.
		EnableOpenMetrics: true,
	}))
	mux.Handle("GET /static/", http.FileServerFS(web.Static))

	return withMiddleware(mux), nil
}

// withMiddleware wraps the mux, outermost first: tracing, panic recovery,
// security headers, then one combined access-log + metrics layer.
//
// Tracing is outermost so a recovered panic is still recorded on the span as
// a 500. With tracing disabled (no OTLP endpoint), otelhttp uses the global
// no-op tracer but still extracts an incoming W3C traceparent header, so
// those IDs keep showing up in logs.
func withMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(
		recoverMiddleware(securityHeadersMiddleware(observeMiddleware(next))),
		"http.server",
		otelhttp.WithFilter(traced),
	)
}

// untracedPaths are the probe and scrape endpoints: they're hit every few
// seconds by Docker/Kubernetes/Prometheus and would drown out real traffic
// in Tempo/Jaeger.
var untracedPaths = map[string]bool{"/healthz": true, "/readyz": true, "/metrics": true}

func traced(r *http.Request) bool {
	return !untracedPaths[r.URL.Path]
}

// securityHeadersMiddleware sets conservative browser hardening headers on
// every response.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// statusRecorder wraps a ResponseWriter to capture the status code written,
// defaulting to 200 if WriteHeader is never called explicitly.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	return r.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the underlying writer (for
// Flush, deadlines, etc.) through this wrapper.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// observeMiddleware logs each request, records its Prometheus metrics, and
// names the request's trace span after the matched route. It must wrap the
// mux directly: ServeMux sets r.Pattern on the request it is handed, which
// is how the low-cardinality route ("path" metric label, span name) is
// derived after the handler returns.
func observeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		duration := time.Since(start)

		ctx := r.Context()
		route := routePattern(r)
		span := trace.SpanFromContext(ctx)
		span.SetName(r.Method + " " + route)
		span.SetAttributes(attribute.String("http.route", route))

		metrics.ObserveRequest(ctx, r.Method, route, rec.status, duration)
		slog.InfoContext(ctx, "request",
			"method", r.Method,
			"path", r.URL.Path,
			"route", route,
			"status", rec.status,
			"duration_ms", float64(duration.Microseconds())/1000,
		)
	})
}

// routePattern returns the matched mux pattern without its "METHOD " prefix,
// keeping the "path" metric label low-cardinality (e.g. "/static/" rather
// than every distinct static asset URL). Requests that matched no route
// collapse into the single unmatchedRoute label.
func routePattern(r *http.Request) string {
	if r.Pattern == "" {
		return unmatchedRoute
	}
	if _, path, ok := strings.Cut(r.Pattern, " "); ok {
		return path
	}
	return r.Pattern
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// http.ErrAbortHandler is the documented way to abort a
			// response; net/http handles it silently, so re-panic.
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(rec)
			}
			slog.ErrorContext(r.Context(), "panic recovered", "error", rec, "path", r.URL.Path)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}
