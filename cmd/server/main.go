// Command server runs the My World Cup App HTTP server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aeciopires/my-world-cup-app/internal/config"
	"github.com/aeciopires/my-world-cup-app/internal/data"
	"github.com/aeciopires/my-world-cup-app/internal/handlers"
	"github.com/aeciopires/my-world-cup-app/internal/metrics"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func run() error {
	cfg := config.Load()

	client := data.NewClient(cfg.SourceURLs, cfg.FetchTimeout)
	store := data.NewStore(client)

	// Kick off an initial refresh in the background so startup isn't blocked
	// on the live data source; the embedded fallback serves requests until
	// it completes.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.FetchTimeout)
		defer cancel()
		err := store.Refresh(ctx)
		metrics.RecordRefresh(err)
		if err != nil {
			slog.Warn("initial data refresh failed, serving fallback snapshot", "error", err)
		}
	}()

	router, err := handlers.NewRouter(store, cfg.FetchTimeout)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
