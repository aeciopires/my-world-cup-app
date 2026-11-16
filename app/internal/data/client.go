package data

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// maxSourceBytes caps how much of each source document is read, so a
// misbehaving or malicious upstream can't exhaust the server's memory. The
// real files are well under 1 MiB; 10 MiB leaves ample headroom.
const maxSourceBytes = 10 << 20

// SourceURLs holds the remote locations of the four openfootball JSON files.
type SourceURLs struct {
	Matches  string
	Groups   string
	Teams    string
	Stadiums string
}

// Client fetches tournament source data over HTTP.
type Client struct {
	httpClient *http.Client
	urls       SourceURLs
}

// NewClient creates a Client with the given source URLs and request timeout.
func NewClient(urls SourceURLs, timeout time.Duration) *Client {
	return &Client{
		// otelhttp gives each download its own client span (a child of
		// the data.Refresh span) and forwards the trace context upstream.
		httpClient: &http.Client{Timeout: timeout, Transport: otelhttp.NewTransport(http.DefaultTransport)},
		urls:       urls,
	}
}

// Fetch downloads the four source JSON files concurrently. If any download
// fails, the remaining ones are canceled and all failures are returned
// joined together.
func (c *Client) Fetch(ctx context.Context) (sourceFiles, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		src  sourceFiles
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	targets := []struct {
		name string
		url  string
		dst  *[]byte
	}{
		{"matches", c.urls.Matches, &src.Matches},
		{"groups", c.urls.Groups, &src.Groups},
		{"teams", c.urls.Teams, &src.Teams},
		{"stadiums", c.urls.Stadiums, &src.Stadiums},
	}
	for _, target := range targets {
		wg.Go(func() {
			body, err := c.get(ctx, target.url)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("fetch %s: %w", target.name, err))
				mu.Unlock()
				cancel()
				return
			}
			*target.dst = body // each goroutine writes a distinct field
		})
	}
	wg.Wait()

	if err := errors.Join(errs...); err != nil {
		return sourceFiles{}, err
	}
	return src, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	// Read one byte past the limit so an oversized body is reported as an
	// error instead of being silently truncated into invalid JSON.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSourceBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxSourceBytes {
		return nil, fmt.Errorf("response from %s exceeds %d bytes", url, maxSourceBytes)
	}
	return body, nil
}
