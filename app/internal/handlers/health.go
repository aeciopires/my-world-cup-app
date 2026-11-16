package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/data"
)

// Health tracks whether this instance should receive new traffic. It backs
// the two probe endpoints:
//
//   - GET /healthz (liveness): "is the process alive?" Always 200 while the
//     server can answer at all. Failing it makes Kubernetes restart the pod,
//     so it deliberately checks nothing else.
//   - GET /readyz (readiness): "should I get traffic?" 200 normally, 503 once
//     shutdown has begun (SetDraining), so a load balancer stops routing new
//     requests here while in-flight ones finish.
//
// Readiness doesn't depend on the live data source: the embedded fallback
// snapshot means the app always has something to serve, and failing every
// replica's readiness during an openfootball outage would take the whole
// site down instead of serving slightly stale data.
type Health struct {
	draining atomic.Bool
}

// NewHealth returns a Health that reports ready.
func NewHealth() *Health {
	return &Health{}
}

// SetDraining makes /readyz report 503 from now on.
func (h *Health) SetDraining() {
	h.draining.Store(true)
}

type readyResponse struct {
	Status      string `json:"status"`
	DataSource  string `json:"data_source"`
	LastUpdated string `json:"last_updated"`
}

// readyz reports readiness plus which data snapshot is being served.
func (h *Health) readyz(store *data.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := store.Snapshot()
		resp := readyResponse{
			Status:      "ready",
			DataSource:  snap.Source,
			LastUpdated: snap.LastUpdated.Format(time.RFC3339),
		}
		status := http.StatusOK
		if h.draining.Load() {
			resp.Status = "draining"
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.DebugContext(r.Context(), "writing readiness response failed", "error", err)
		}
	}
}

// healthz is the liveness probe.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
