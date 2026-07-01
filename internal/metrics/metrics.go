// Package metrics exposes Prometheus instrumentation for HTTP requests and
// the data refresh workflow.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Registry is the dedicated Prometheus registry for this application,
// deliberately separate from the global default registry.
var Registry = prometheus.NewRegistry()

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds, labeled by method and path.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	dataRefreshTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "data_refresh_total",
			Help: "Total number of tournament data refresh attempts, labeled by outcome.",
		},
		[]string{"outcome"},
	)

	dataLastRefreshTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "data_last_refresh_timestamp_seconds",
			Help: "Unix timestamp of the last successful tournament data refresh.",
		},
	)
)

func init() {
	Registry.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		dataRefreshTotal,
		dataLastRefreshTimestamp,
	)
}

// ObserveRequest records a completed HTTP request's status and duration.
func ObserveRequest(method, path string, status int, duration time.Duration) {
	statusLabel := statusLabelFor(status)
	httpRequestsTotal.WithLabelValues(method, path, statusLabel).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordRefresh records the outcome of a data refresh attempt. On success
// it also updates the last-refresh timestamp gauge.
func RecordRefresh(err error) {
	if err != nil {
		dataRefreshTotal.WithLabelValues("failure").Inc()
		return
	}
	dataRefreshTotal.WithLabelValues("success").Inc()
	dataLastRefreshTimestamp.Set(float64(time.Now().Unix()))
}

func statusLabelFor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
