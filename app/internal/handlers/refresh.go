package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/metrics"
)

// refreshRetryAfterSeconds is the Retry-After hint sent when a refresh is
// requested while another one is still running.
const refreshRetryAfterSeconds = "5"

// RefreshHandlers exposes the manual data-refresh endpoint.
type RefreshHandlers struct {
	store   *data.Store
	timeout time.Duration
}

// NewRefreshHandlers creates a RefreshHandlers instance.
func NewRefreshHandlers(store *data.Store, timeout time.Duration) *RefreshHandlers {
	return &RefreshHandlers{store: store, timeout: timeout}
}

type refreshResponse struct {
	OK          bool   `json:"ok"`
	LastUpdated string `json:"last_updated"`
	Source      string `json:"source"`
	Error       string `json:"error,omitempty"`
}

// Refresh triggers a synchronous fetch of the live data source and reports
// the outcome as JSON. On failure, the previous snapshot remains in place.
// If a refresh is already running, it answers 429 Too Many Requests without
// contacting the upstream again.
func (h *RefreshHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	err := h.store.Refresh(ctx)
	inProgress := errors.Is(err, data.ErrRefreshInProgress)
	if !inProgress {
		// A rejected duplicate request isn't a refresh attempt, so it
		// doesn't count toward data_refresh_total.
		metrics.RecordRefresh(err)
	}
	snap := h.store.Snapshot()

	resp := refreshResponse{
		OK:          err == nil,
		LastUpdated: snap.LastUpdated.Format(time.RFC3339),
		Source:      snap.Source,
	}

	status := http.StatusOK
	switch {
	case inProgress:
		resp.Error = err.Error()
		status = http.StatusTooManyRequests
		w.Header().Set("Retry-After", refreshRetryAfterSeconds)
	case err != nil:
		resp.Error = err.Error()
		status = http.StatusBadGateway
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.DebugContext(ctx, "writing refresh response failed", "error", err)
	}
}
