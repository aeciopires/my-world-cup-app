# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.0.0] - 2026-06-30

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

## [Unreleased]

### Planned / Future Improvements

- Implement full FIFA tie-break rules (head-to-head results, fair play points) for group standings.
- Add a graphical knockout bracket view.
- Support filtering the matches page by round/group/team via query parameters.
- Periodic automatic background refresh (in addition to startup and manual refresh).
- Optional persistent cache (e.g. local file) to survive restarts without a network call.
- Internationalization (additional UI languages beyond en-US).
