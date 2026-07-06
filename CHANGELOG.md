<!-- TOC -->

- [Changelog](#changelog)
  - [1.2.0 - 2026-07-06](#120---2026-07-06)
    - [Added](#added)
    - [Changed](#changed)
    - [Fixed](#fixed)
  - [1.1.0 - 2026-07-05](#110---2026-07-05)
    - [Added](#added-1)
    - [Changed](#changed-1)
    - [Fixed](#fixed-1)
  - [1.0.0 - 2026-06-30](#100---2026-06-30)
    - [Added](#added-2)
    - [Planned / Future Improvements](#planned--future-improvements)

<!-- TOC -->

# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

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
