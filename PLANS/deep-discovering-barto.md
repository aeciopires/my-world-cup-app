# My World Cup App — TASK2.md Additions Implementation Plan

## Context

The base app from `PLAN.md`/`TASK.md` is complete and verified (groups, knockout, matches, links, theme toggle, live data refresh with fallback, unit + integration tests, Dockerfile/Compose, Makefile, README/CHANGELOG/CLAUDE docs — all committed to disk, `make check` green, manual + Docker verification done).

A new `TASK2.md` appeared mid-session with additional requirements (still en-US, same non-functional constraints as before):
- A `/metrics` endpoint (health check path `/healthz` already exists from the base build).
- A Helm chart for Kubernetes deployment.
- A page showing statistics by player and by national team.

User confirmed (via AskUserQuestion): use `prometheus/client_golang` for metrics — the industry-standard client, worth the first external dependency for this project (previously zero-dependency by design; that constraint was specifically about the web layer, not observability tooling).

## Metrics (`/metrics`)

- New package `internal/metrics/metrics.go`: a dedicated `prometheus.Registry` (not the global default, to keep it explicit) exposing:
  - `http_requests_total{method,path,status}` (CounterVec)
  - `http_request_duration_seconds{method,path}` (HistogramVec)
  - `data_refresh_total{outcome}` (CounterVec, outcome = `success`|`failure`)
  - `data_last_refresh_timestamp_seconds` (Gauge)
  - Helper functions: `ObserveRequest(method, path string, status int, duration time.Duration)`, `RecordRefresh(err error)` (sets outcome + gauge on success).
- `internal/handlers/router.go`: replace `loggingMiddleware` with a `metricsMiddleware` that wraps it (keep the existing slog line, add the Prometheus observation using a `http.ResponseWriter` wrapper to capture status code — reuse the pattern already in `recoverMiddleware`/`loggingMiddleware`, just add one more wrapping middleware in `withMiddleware`). Register `GET /metrics` via `promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{})`.
- Call `metrics.RecordRefresh(err)` at the two existing refresh call sites: `cmd/server/main.go`'s startup goroutine and `internal/handlers/refresh.go`'s `Refresh` handler (both already call `store.Refresh(ctx)` and check `err`).
- `go get github.com/prometheus/client_golang` — first entry in `go.mod`/`go.sum`.

## Player & Team Statistics Page (`/stats`)

- New `internal/services/stats.go`:
  - `PlayerStat{Name, Team string; Goals int}` and `TopScorers(t models.Tournament, limit int) []PlayerStat` — iterate every match's `Goals1` (credited to `Team1`) and `Goals2` (credited to `Team2`), aggregate by player name, sort by goals desc then name asc, truncate to `limit`.
  - `TeamStat{Team string; Played, Won, Drawn, Lost, GoalsFor, GoalsAgainst, GoalDifference, Points int}` and `TeamStandings(t models.Tournament) []TeamStat` — same aggregation logic as `GroupStandings` in `internal/services/standings.go` but across **all** played matches (group + knockout stages), not grouped/limited to group members. Extract the shared per-match accumulation into a small reusable helper (`applyResult` in `standings.go` operates on `map[string]*models.Standing`; add an equivalent for the flat all-teams case, or generalize `applyResult` to accept any team names — reuse rather than duplicate scoring math).
  - Sort team standings by points desc, goal difference desc, goals for desc, name asc (same tie-break already documented for groups).
- `internal/handlers/pages.go`: add `Stats(w, r)` handler + `statsData` struct (embeds `baseData`, `TopScorers []services.PlayerStat`, `TeamStandings []services.TeamStat`).
- `web/templates/stats.html`: two-table layout (Top Scorers: Rank/Player/Team/Goals; Team Overview: Team/P/W/D/L/GF/GA/GD/Pts), following the existing `groups.html`/`matches.html` markup conventions.
- `web/templates/layout.html`: add a "Stats" nav link (same pattern as the other nav `<a>` tags) and wire `pages.Stats` in `router.go` as `GET /stats`.

## Helm Chart

New `charts/my-world-cup-app/` (standard Helm chart layout):
```
charts/my-world-cup-app/
├── Chart.yaml
├── values.yaml
├── .helmignore
└── templates/
    ├── _helpers.tpl
    ├── deployment.yaml
    ├── service.yaml
    ├── serviceaccount.yaml
    ├── ingress.yaml       # disabled by default
    ├── hpa.yaml           # disabled by default
    └── NOTES.txt
```
- `values.yaml`: `image.repository` (default `my-world-cup-app`), `image.tag` (default `.Chart.AppVersion`), `image.pullPolicy`; `replicaCount: 1`; `service.type: ClusterIP`, `service.port: 8080`; `env` map for `PORT`/`WORLDCUP_*_URL` overrides; `resources` (requests/limits, sensible small defaults); `livenessProbe`/`readinessProbe` on `GET /healthz`; `podAnnotations` pre-populated with `prometheus.io/scrape: "true"`, `prometheus.io/port: "8080"`, `prometheus.io/path: /metrics`; `ingress.enabled: false`; `autoscaling.enabled: false`.
- `deployment.yaml` renders env vars from `values.env`, wires the probes and annotations.
- Verify with `helm lint charts/my-world-cup-app` and `helm template charts/my-world-cup-app` (check `helm` CLI availability first; if absent, at minimum validate the rendered YAML is well-formed via `helm template` if installable, otherwise note manual review in the summary).

## Testing

- `internal/services/stats_test.go`: table-driven tests for `TopScorers` (multiple scorers, tie-break by name, limit truncation) and `TeamStandings` (aggregation across group + knockout matches, tie-break order) — mirrors the style of `internal/services/standings_test.go`.
- `internal/handlers/router_test.go`: extend `TestRoutes_ReturnOK` with `/stats` and `/metrics`; add a `TestMetricsEndpoint_ExposesExpectedSeries` asserting the response body contains `http_requests_total` and `data_refresh_total` after making a couple of requests; add a `TestStatsPage_ContainsExpectedSections` asserting `"Top Scorers"` and `"Team Overview"` (or equivalent headings) appear.
- Run `make check` (fmt-check + vet + test) after changes.

## Documentation Updates

- `README.md`: add `/metrics` and `/stats` rows to the routing table; add a "Helm" subsection under Getting Started (`helm install my-world-cup-app ./charts/my-world-cup-app`); update the Technology Stack section to note `prometheus/client_golang` as the one external dependency; update directory structure tree to include `internal/metrics/` and `charts/`.
- `CHANGELOG.md`: add a new `[1.0.0]` entry (or `[Unreleased]` promoted) listing metrics, stats page, and Helm chart additions.
- `CLAUDE.md`: update the "No web framework" bullet to clarify the zero-dependency rule applied to the web layer only, note the new `internal/metrics` package and its two call sites (`main.go`, `handlers/refresh.go`), and note the Helm chart location/how to test it locally.

## Verification

1. `go get github.com/prometheus/client_golang && go mod tidy` succeeds (requires network — already confirmed available in this environment).
2. `make check` — fmt, vet, all tests (existing + new) pass.
3. `make run`; manually confirm in browser: `/stats` renders top scorers and team standings with sensible data; `/metrics` returns Prometheus exposition text including the four custom series after a few requests and a refresh.
4. `make docker-build && make docker-up` (on an alternate port if 8080 is occupied, as done previously) — confirm `/metrics` and `/stats` work inside the container too, then `make docker-down`.
5. `helm lint charts/my-world-cup-app` and `helm template charts/my-world-cup-app` render without errors (if `helm` CLI is available in this environment; otherwise document as a manual follow-up in the final summary).
