package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/metrics"
)

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
func (h *RefreshHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	err := h.store.Refresh(ctx)
	metrics.RecordRefresh(err)
	_, lastUpdated, source := h.store.Snapshot()

	resp := refreshResponse{
		OK:          err == nil,
		LastUpdated: lastUpdated.Format(time.RFC3339),
		Source:      source,
	}

	status := http.StatusOK
	if err != nil {
		resp.Error = err.Error()
		status = http.StatusBadGateway
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
