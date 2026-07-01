package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

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
		httpClient: &http.Client{Timeout: timeout},
		urls:       urls,
	}
}

// Fetch downloads the four source JSON files. It fails fast on the first error.
func (c *Client) Fetch(ctx context.Context) (sourceFiles, error) {
	matches, err := c.get(ctx, c.urls.Matches)
	if err != nil {
		return sourceFiles{}, fmt.Errorf("fetch matches: %w", err)
	}
	groups, err := c.get(ctx, c.urls.Groups)
	if err != nil {
		return sourceFiles{}, fmt.Errorf("fetch groups: %w", err)
	}
	teams, err := c.get(ctx, c.urls.Teams)
	if err != nil {
		return sourceFiles{}, fmt.Errorf("fetch teams: %w", err)
	}
	stadiums, err := c.get(ctx, c.urls.Stadiums)
	if err != nil {
		return sourceFiles{}, fmt.Errorf("fetch stadiums: %w", err)
	}
	return sourceFiles{Matches: matches, Groups: groups, Teams: teams, Stadiums: stadiums}, nil
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
	return io.ReadAll(resp.Body)
}
