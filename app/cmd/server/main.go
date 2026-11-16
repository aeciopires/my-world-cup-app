// Command server runs the My World Cup App HTTP server.
package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/config"
	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/handlers"
	"github.com/aeciopires/my-world-cup-app/internal/metrics"
	"github.com/aeciopires/my-world-cup-app/internal/telemetry"
)

// version is the app release, stamped at build time with
// -ldflags "-X main.version=..." (see the Makefile and Dockerfile). It's
// reported as the service.version of every trace.
var version = "dev"

const (
	healthcheckTimeout = 3 * time.Second
	shutdownTimeout    = 10 * time.Second
	// telemetryFlushTimeout bounds how long exiting waits for buffered
	// spans to reach the collector.
	telemetryFlushTimeout = 5 * time.Second
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(healthcheck())
	}

	if err := run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

// healthcheck performs a local GET /healthz and returns a process exit code,
// used as the Docker HEALTHCHECK command since the distroless runtime image
// has no shell or curl/wget available.
func healthcheck() int {
	port := cmp.Or(os.Getenv("PORT"), config.DefaultPort)
	ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
	defer cancel()
	// Always loopback; only the port comes from the (trusted) environment.
	url := "http://" + net.JoinHostPort("127.0.0.1", port) + "/healthz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) //nolint:gosec // G704: fixed loopback host
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck failed:", err)
		return 1
	}
	resp, err := http.DefaultClient.Do(req) //nolint:gosec // G704: fixed loopback host
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck failed:", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "healthcheck failed: status", resp.StatusCode)
		return 1
	}
	return 0
}

func run() error {
	// Log JSON from the very first line (config warnings included), then
	// switch to text if LOG_FORMAT asks for it.
	slog.SetDefault(telemetry.NewLogger(os.Stdout, true))
	cfg := config.Load()
	if cfg.LogFormat != config.LogFormatJSON {
		slog.SetDefault(telemetry.NewLogger(os.Stdout, false))
	}

	// ctx is canceled on SIGINT/SIGTERM, which both starts the graceful
	// shutdown below and aborts an initial refresh still in flight.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTelemetry, err := telemetry.Setup(ctx, version)
	if err != nil {
		return err
	}
	defer func() {
		// The signal context is already canceled here; flush with a fresh one.
		flushCtx, cancel := context.WithTimeout(context.Background(), telemetryFlushTimeout)
		defer cancel()
		if err := shutdownTelemetry(flushCtx); err != nil {
			slog.Warn("flushing telemetry failed", "error", err)
		}
	}()

	client := data.NewClient(cfg.SourceURLs, cfg.FetchTimeout)
	store := data.NewStore(client)

	// Kick off an initial refresh in the background so startup isn't blocked
	// on the live data source; the embedded fallback serves requests until
	// it completes.
	go func() {
		refreshCtx, cancel := context.WithTimeout(ctx, cfg.FetchTimeout)
		defer cancel()
		err := store.Refresh(refreshCtx)
		metrics.RecordRefresh(err)
		if err != nil {
			slog.WarnContext(refreshCtx, "initial data refresh failed, serving fallback snapshot", "error", err)
		}
	}()

	health := handlers.NewHealth()
	router, err := handlers.NewRouter(store, cfg.FetchTimeout, health)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		// POST /refresh waits up to FetchTimeout for the upstream, so the
		// write deadline must comfortably exceed it.
		WriteTimeout: cfg.FetchTimeout + 15*time.Second,
		IdleTimeout:  2 * time.Minute,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "port", cfg.Port, "version", version, "tracing", telemetry.Enabled())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// Fail readiness first and keep serving for the drain delay, so the
		// load balancer stops sending new requests before we stop accepting.
		health.SetDraining()
		if cfg.ShutdownDrainDelay > 0 {
			slog.Info("draining before shutdown", "delay", cfg.ShutdownDrainDelay.String())
			time.Sleep(cfg.ShutdownDrainDelay)
		}
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
