// Package metrics exposes Prometheus instrumentation for HTTP requests and
// the data refresh workflow.
package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
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
	// Export both outcomes as 0 from the start: a series that only appears
	// at its first increment makes increase()/rate() miss that first event.
	dataRefreshTotal.WithLabelValues("success")
	dataRefreshTotal.WithLabelValues("failure")
}

// ObserveRequest records a completed HTTP request's status and duration.
// When ctx carries a sampled trace, its ID is attached as an exemplar, so a
// Grafana latency graph can link a data point straight to that trace
// (exemplars are only exposed in the OpenMetrics format).
func ObserveRequest(ctx context.Context, method, path string, status int, duration time.Duration) {
	counter := httpRequestsTotal.WithLabelValues(method, path, statusLabelFor(status))
	histogram := httpRequestDuration.WithLabelValues(method, path)

	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsSampled() {
		counter.Inc()
		histogram.Observe(duration.Seconds())
		return
	}
	exemplar := prometheus.Labels{"trace_id": sc.TraceID().String()}
	counter.(prometheus.ExemplarAdder).AddWithExemplar(1, exemplar)
	histogram.(prometheus.ExemplarObserver).ObserveWithExemplar(duration.Seconds(), exemplar)
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
