package handlers

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/metrics"
	"github.com/aeciopires/my-world-cup-app/web"
)

// NewRouter builds the application's HTTP handler, wiring pages, the
// refresh endpoint, and static asset serving.
func NewRouter(store *data.Store, fetchTimeout time.Duration) (http.Handler, error) {
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
	mux.Handle("GET /metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))

	staticFS := http.FileServerFS(web.Static)
	mux.Handle("GET /static/", staticFS)

	return withMiddleware(mux), nil
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func withMiddleware(next http.Handler) http.Handler {
	return recoverMiddleware(metricsMiddleware(loggingMiddleware(next)))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

// statusRecorder wraps a ResponseWriter to capture the status code written,
// defaulting to 200 if WriteHeader is never called explicitly.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		metrics.ObserveRequest(r.Method, routePattern(r), rec.status, time.Since(start))
	})
}

// routePattern returns the matched mux pattern without its "METHOD " prefix,
// keeping the "path" metric label low-cardinality (e.g. "/static/" rather
// than every distinct static asset URL).
func routePattern(r *http.Request) string {
	pattern := r.Pattern
	if idx := strings.IndexByte(pattern, ' '); idx != -1 {
		return pattern[idx+1:]
	}
	if pattern == "" {
		return r.URL.Path
	}
	return pattern
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "path", r.URL.Path)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
