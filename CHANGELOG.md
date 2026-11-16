<!-- TOC -->

- [Changelog](#changelog)
  - [1.4.0 - 2026-11-16](#140---2026-11-16)
    - [Added](#added)
    - [Changed](#changed)
    - [Fixed](#fixed)
  - [1.3.0 - 2026-07-11](#130---2026-07-11)
    - [Added](#added-1)
    - [Changed](#changed-1)
  - [1.2.0 - 2026-07-06](#120---2026-07-06)
    - [Added](#added-2)
    - [Changed](#changed-2)
    - [Fixed](#fixed-1)
  - [1.1.0 - 2026-07-05](#110---2026-07-05)
    - [Added](#added-3)
    - [Changed](#changed-3)
    - [Fixed](#fixed-2)
  - [1.0.0 - 2026-06-30](#100---2026-06-30)
    - [Added](#added-4)
    - [Planned / Future Improvements](#planned--future-improvements)

<!-- TOC -->

# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## 1.4.0 - 2026-11-16

### Added

- `mise.toml` pinning the development toolchain: Go 1.27.4, golangci-lint 2.15.0, Helm 4.3.2, helm-docs 1.14.2, kind 0.33.0, kubectl 1.36.2.
- `make install-deps` (via `scripts/deps.sh install`): installs base packages (apt on Ubuntu, Xcode CLT on macOS), mise, and the pinned toolchain. Docker remains a separate install.
- `make lint` (golangci-lint v2, config in `app/.golangci.yml`, including the `modernize` and `gosec` analyzers), now part of `make check`.
- `make helm-template`.
- `FETCH_TIMEOUT` environment variable (Go duration, default `10s`).
- `REQUIREMENTS.md`: supported platforms (Ubuntu 22.04/24.04/26.04 amd64, macOS amd64/arm64) and installation instructions.
- `docs/`: a beginner's guide to the architecture and features, with analogies and Mermaid diagrams.
- `result` template func rendering a match's final score, an `a.e.t.` marker, and the penalty shoot-out score.
- Security headers (`X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`) on every response; `HEALTHCHECK` in the `Dockerfile`.
- Helm chart: pod `seccompProfile: RuntimeDefault`, explicit `runAsUser`/`runAsGroup` 65532, `serviceAccount.automountToken` (default `false`).
- `GET /readyz` readiness endpoint (JSON with the data source in use; `503` once shutdown starts) and `SHUTDOWN_DRAIN_DELAY` to keep serving while load balancers notice it. Helm chart: readiness probe on `/readyz`, new `startupProbe` on `/healthz`, `terminationGracePeriodSeconds`, and `SHUTDOWN_DRAIN_DELAY=5s` by default.
- OpenTelemetry tracing (`app/internal/telemetry`), enabled by the standard `OTEL_EXPORTER_OTLP_ENDPOINT` variable: a server span per request (named after the route; probes and `/metrics` excluded), a `data.Refresh` span with one client span per download, W3C trace-context propagation, and spans flushed on shutdown.
- Trace exemplars on `http_requests_total` and `http_request_duration_seconds` (OpenMetrics exposition enabled on `/metrics`).
- `LOG_FORMAT` (`json` default, or `text`); log lines carry `trace_id`/`span_id` when a trace context exists; the access log gained `route` and `duration_ms`.
- Local observability stack (`docker-compose.observability.yml`, `observability/`): Grafana 13.3.1, Prometheus 3.16.0, Loki 3.7.10, Tempo 3.1.1, Jaeger 2.23.0, and Grafana Alloy 1.22.0, with cross-linked data sources (exemplars → traces, logs ↔ traces, spans → metrics) and three provisioned dashboards (Metrics, Logs, Traces). New targets `make observability-up`, `observability-down`, `observability-logs`, and `observability-validate`.
- The binary reports its version (`-X main.version`, from the `VERSION` file) at startup and as the traces' `service.version`.

### Changed

- **Helm 4 is now required**; the chart fails to render under Helm 3 with an explicit error. Chart version bumped to `2.0.0`.
- Go 1.25 → 1.27; `prometheus/client_golang` 1.23.2 → 1.25.0 (and transitive dependencies).
- Docker images: builder `golang:1.27-alpine` (runs natively and cross-compiles per `TARGETARCH`), runtime `distroless/static-debian13:nonroot`; module downloads cached and verified.
- `make check-deps` rewritten (`scripts/deps.sh check`): detects the platform, checks base packages, mise and every pinned tool version, and Docker/Compose/Buildx. All Makefile targets run tools through `mise exec`.
- The four source files are downloaded in parallel (`sync.WaitGroup.Go`), each capped at 10 MiB.
- Only one data refresh runs at a time; concurrent `POST /refresh` calls get `429 Too Many Requests` with `Retry-After`.
- `data.Store.Snapshot()` returns a `data.Snapshot` struct, and page handlers build the layout data from that same snapshot.
- Templates render into a buffer before writing, so template errors produce a clean 500.
- HTTP server sets `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`; the initial refresh is canceled on shutdown.
- **Logs are now JSON by default** (previously slog's default text output); set `LOG_FORMAT=text` for the old human-oriented style. The access log's `duration` field is now `duration_ms` (float milliseconds).
- `data_refresh_total{outcome}` series are exported at 0 from startup.
- `docker-compose.yml` sets an explicit Compose project name (`my-world-cup-app`).
- Code modernized per the JetBrains Modern Go Guidelines (`slices`/`maps`/`cmp`, `min`, `cmp.Or`, `strings.Cut`, `errors.Is`, `t.Context()`, `new(expr)`, ...).

### Fixed

- Knockout matches decided in **extra time** had no winner (no bracket highlight) and extra-time goals were missing from `/stats` team records; the score shown was the 90-minute one.
- **Own goals** were credited to the player who scored them, listing them as a scorer for the opposing team on `/stats`.
- Knockout "Match Details" venues linked through `StadiumLinks` (keyed by stadium name) and therefore never linked; they now use `CityLinks`.
- `http_requests_total`/`http_request_duration_seconds` used the raw URL path as the `path` label for unmatched requests (unbounded cardinality); they're now labeled `unmatched`.
- The `/matches` team filter listed knockout placeholders such as `W74` as teams.
- Top-scorer order was nondeterministic for players with the same name and goal count; `TopScorers` panicked on a negative limit.
- `app.js` stopped initializing every button when `localStorage` was unavailable (private browsing / blocked site data).
- The `Dockerfile` dependency-download layer ignored `go.sum`.

## 1.3.0 - 2026-07-11

### Added

- `make docker-build-multiarch`: builds a multi-arch image (`linux/amd64` + `linux/arm64`, covering Linux and macOS on both Intel and Apple Silicon) via Docker Buildx, to validate the cross-arch build without pushing anywhere.
- `make docker-push`: builds and pushes a multi-arch image to a registry, interactively prompting for registry username, password/access token (hidden input, piped to `docker login --password-stdin`), repository name, and image tag.
- `make docker-buildx-setup`: creates/reuses the `docker-container`-driver Buildx builder the two targets above require (the default `docker` driver can't produce multi-platform output).
- `docker buildx` added to `make check-deps`' tool checks.
- Root `VERSION` file: the single source of truth for the app's release version, read by the `Makefile` into `APP_VERSION`.
- `make helm-sync-version`: writes `VERSION`'s content into `charts/my-world-cup-app/Chart.yaml`'s `appVersion` (portable `sed -i.bak` for both GNU/Linux and BSD/macOS sed); run automatically by `make docker-push` and `make helm-docs` before every publish/docs regeneration, so `appVersion` can't go stale.
- `Dockerfile`'s runtime stage now accepts an `APP_VERSION` build-arg (default `dev`) and sets it as the `org.opencontainers.image.version` image label.
- `CONTRIBUTING.md` now reminds contributors to bump `VERSION` (and add a `CHANGELOG.md` entry) for release-worthy changes.

### Changed

- `DOCKER_TAG` now defaults to the version in the `VERSION` file instead of a hardcoded `latest`; `DOCKER_PLATFORMS` (default `linux/amd64,linux/arm64`) is also an overridable `Makefile` variable, e.g. `make docker-push DOCKER_PLATFORMS=linux/amd64`.
- `make docker-build` / `make docker-up` (`docker compose build`/`up`) and `docker-compose.yml` now pass `APP_VERSION` through as a build-arg too, so a plain Compose-based build carries the same version label as the Buildx-based targets.

## 1.2.0 - 2026-07-06

### Added

- "Expand all matches" / "Collapse all matches" button on the Groups page (`#toggle-all-matches` in `web/templates/groups.html`, wired up by `initGroupMatchesToggle` in `web/static/js/app.js`) toggles every group's match-results `<details class="group-matches">` at once instead of requiring one click per group.

### Changed

- Home page's Upcoming Matches / Recent Results cards moved from a side-by-side `.grid-2` layout (each card capped at ~538px even on wide desktops) to a new full-width `.home-grid` single column, so venue names have room to stay on one line at common desktop resolutions.

### Fixed

- Match venues (home page, `/matches`) now link to the official fifa.com **host-city** page instead of a stadium-name lookup that never matched: `models.Match.Ground` carries the host city (e.g. "Mexico City"), not the stadium name ("Estadio Azteca"), so rendering switched from `StadiumLinks` to `CityLinks` (`internal/handlers/render.go`, `web/templates/home.html`, `web/templates/matches.html`).
- Reworked responsive layout across the Groups, Home, and Matches pages so tables and cards fit without a horizontal scrollbar at common phone/tablet/desktop widths:
  - `.card` now sets `min-width: 0` so a wide table can no longer blow out its CSS grid track (and the whole page) instead of scrolling within its own card — grid items default to a content-based minimum width, not `0`.
  - `.group-grid` sizes cards with `minmax(min(460px, 100%), 1fr)` instead of a bare `minmax(460px, 1fr)`: `auto-fill`/`auto-fit` only ever reduce column *count*, never the `minmax` minimum, so the un-capped version still forced a 460px-wide column (and page-wide horizontal scroll) on any phone screen.
  - `.table` no longer forces `min-width: max-content`, which was defeating the existing `overflow-wrap`/`word-break` rules and always triggering `.table-responsive`'s scrollbar instead of letting long cell content wrap.
  - New `.nowrap-cell` utility keeps short fixed-format values (dates, round, group) from wrapping character-by-character once table columns are free to shrink.
  - `.standings-table` (group tables) and the new `.matches-table` (fixture-list tables on the home, `/matches`, and group match-detail views) restack into team-first blocks with labeled fields below 640px instead of scrolling sideways, using new `data-label` attributes on the affected table cells.

## 1.1.0 - 2026-07-05

### Added

- Official fifa.com links for every team, stadium, and host city (`baseData.TeamLinks`/`StadiumLinks`/`CityLinks`, built in `internal/handlers/fifa_links.go`), rendered next to team names, match venues, and the Stadiums table on the FIFA Links page.
- Three more official resources on the FIFA Links page: Official Match Ball, Official Posters, and Mascots.
- Official FIFA article on the FIFA Links page explaining the group stage format and tie-breaking rules ("Groups: How Teams Qualify & Tie-Breakers").
- Graphical knockout bracket on `/knockout`: rounds connected with bracket lines, winning teams highlighted, penalty shoot-outs annotated, and the match for third place rendered as its own standalone fixture — shown alongside the existing round-by-round detail tables (`internal/handlers/pages.go`, `web/templates/knockout.html`).
- Matches page filtering: `/matches` now accepts combinable `round`, `group`, and `team` query parameters (`services.FilterMatches`/`MatchFilterOptions` in `internal/services/matches.go`), with a filter form, a "Showing X of Y matches" summary, and a "Clear filters" link.

### Changed

- Refreshed the dark/light theme color palettes and general UI polish: sticky header with shadow, pill-style active nav highlighting, card shadows, button hover/focus states, and table row hover (`web/static/css/style.css`).
- Team flags now render as bordered flag "chip" badges (new `flag` template func) instead of a bare emoji; match results render as a colored score pill (new `score` template func), with unplayed matches shown as a "scheduled" badge.
- Updated the screenshots in `images/` to reflect the refreshed UI.

### Fixed

- `models.Match.Winner()` now accounts for a penalty shoot-out when full time ends level, so knockout draws decided on penalties correctly report a winner (previously only the full-time score was checked). This backs the new bracket's winner highlighting.

## 1.0.0 - 2026-06-30

### Added

- Initial release of My World Cup App.
- Go 1.25 web server (standard library only) with clean architecture: `models`, `data`, `services`, `handlers`.
- Group standings, knockout stage, and full match list pages, rendered server-side with `html/template`.
- Live data fetching from [openfootball/worldcup.json](https://github.com/openfootball/worldcup.json) on startup and via a manual "Update data" action (`POST /refresh`), with an embedded fallback snapshot for offline/degraded operation.
- Dark/light theme toggle with `localStorage` persistence.
- Links page with official FIFA World Cup 2026 resources (teams, standings, stadiums, articles, scores & fixtures, Club World Cup 2025) and the official Spotify playlist.
- Unit tests for JSON parsing/normalization and standings/knockout computation; HTTP integration tests for all routes.
- Dockerfile (multi-stage, distroless runtime) and Docker Compose setup.
- Makefile with `run`, `build`, `test`, `test-coverage`, `fmt`, `vet`, `docker-*` targets.
- `/metrics` endpoint exposing Prometheus metrics (`internal/metrics`): HTTP request counts and latency, and data refresh outcome/timestamp. First third-party dependency (`prometheus/client_golang`), added specifically for this.
- `/stats` page and `internal/services/stats.go`: top scorers (aggregated goal tally per player) and overall team statistics (played/won/drawn/lost/goals/points across group and knockout stages).
- Helm chart (`charts/my-world-cup-app`) for Kubernetes deployment, with liveness/readiness probes on `/healthz`, Prometheus scrape annotations, and optional Ingress/HorizontalPodAutoscaler.
- Unit tests for the stats service and additional handler integration tests for `/stats` and `/metrics`.

### Planned / Future Improvements

- Periodic automatic background refresh (in addition to startup and manual refresh).
- Optional persistent cache (e.g. local file) to survive restarts without a network call.
