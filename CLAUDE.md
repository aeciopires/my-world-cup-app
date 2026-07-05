# CLAUDE.md

Guidance for Claude Code (and other contributors) working in this repository.

## Project Summary

My World Cup App is a Go web application displaying FIFA World Cup 2026 group standings, knockout stage results, and match data, plus links to official FIFA resources. It has no database: tournament data is fetched live from [openfootball/worldcup.json](https://github.com/openfootball/worldcup.json) on startup and on demand, cached in memory.

Full details: see `README.md` (architecture, stack, directory structure, Mermaid diagram) and `PLANS/PLAN.md` / `PLANS/TASK.md` / `PLANS/TASK2.md` (original and follow-up requirements).

## Conventions

- **Language**: all code, comments, commit messages, and documentation must be in **en-US**, regardless of the language used in requests.
- **No web framework**: use Go's standard library (`net/http`, `html/template`) only for routing/rendering. This constraint applies to the web layer specifically, not to observability tooling — `prometheus/client_golang` (`internal/metrics`) is an intentional, discussed exception. Do not add further dependencies (router, template library, etc.) without discussing it first.
- **No database**: tournament data lives only in the in-memory `data.Store` (`internal/data/store.go`). Don't introduce persistence without a clear reason.
- **Clean architecture layering**: `handlers` → `services` → `models`; `data` is a separate concern owning fetch/parse/cache. Handlers should not contain business logic (standings math, sorting rules); that belongs in `services`.
- **Standings/knockout/stats logic is pure**: `internal/services` functions take a `models.Tournament` and return computed values with no I/O — keep them that way so they stay trivially testable. `TeamStandings` (in `stats.go`) reuses the same `accumulateResult`/`sortStandings` helpers as `GroupStandings` (in `standings.go`) rather than duplicating the scoring math — extend those shared helpers instead of copy-pasting if you add another aggregation view.
- **Metrics**: `internal/metrics` owns a dedicated `prometheus.Registry` (not the global default). `metrics.RecordRefresh(err)` must be called at every `store.Refresh(ctx)` call site (`cmd/server/main.go`'s startup goroutine, `internal/handlers/refresh.go`) — if you add another refresh trigger, call it there too. `metrics.ObserveRequest` is wired in automatically via `metricsMiddleware` in `internal/handlers/router.go`.

## Working in this Repo

- Run `make check` (fmt-check + vet + test) before considering a change done. Run `make check-deps` if you're unsure whether your machine has everything installed (`git`, `go`, `docker`, `docker compose`, `helm`, `helm-docs`) — it reports what's missing with install links instead of failing mid-task.
- **Keep the Makefile and README.md in sync**: every `.PHONY` target in the `Makefile` must have a `## description` comment (used by `make help`) and a matching row in README.md's "Makefile Targets" table (`README.md#makefile-targets`). If you add, rename, or remove a target, update both. Same goes for the "Software Requirements" table if you add a new required tool, and the "Configuration" env var table (`README.md#configuration`) if you add/change an `internal/config` environment variable.
- The fork/branch/PR contribution workflow lives in `CONTRIBUTING.md`, not in this file — don't duplicate it here.
- New source JSON fields from openfootball should be added to the `raw*` structs in `internal/data/parser.go`, then mapped into `internal/models` — never expose the raw JSON structs outside the `data` package.
- The embedded fallback data lives in `internal/data/fallback/*.json` — a snapshot of the real openfootball 2026 files, embedded via `go:embed` so the app has content even with no network access. If the upstream schema changes, refresh this snapshot too.
- Templates live in `web/templates/`, embedded via `web/assets.go`. Every page template must `{{define "content"}}` and is rendered against `web/templates/layout.html`. Shared layout data comes from `handlers.baseData` (`internal/handlers/render.go`).
- Static assets (`web/static/css`, `web/static/js`) are embedded and served at `/static/*`; there is no build step (no bundler, no npm).
- **Team flags**: `baseData.Flags` (`internal/handlers/render.go`, built by `teamFlags()`) is a `map[string]string` of team name → flag emoji, populated from `Tournament.Teams` and available on every page. Any template showing a team name should render it as `<span class="team">{{flag $.Flags .TeamField}} {{.TeamField}}</span>` — use `$.Flags` (not `.Flags`) since `$` always refers to the root page data even inside nested `{{range}}` blocks. The `flag` template func (`internal/handlers/render.go`) wraps the emoji in a `<span class="flag">` chip (bordered, shadowed box — see `web/static/css/style.css`) so it reads as a small flag badge instead of a bare character; don't inline `{{index $.Flags ...}}` directly in markup anymore.
- **FIFA links for teams/stadiums/cities**: `baseData.TeamLinks`, `.StadiumLinks`, and `.CityLinks` (`internal/handlers/fifa_links.go`) are `map[string]string` keyed by team name, stadium name, and host city respectively, pointing at the matching official fifa.com page. The fifa.com slugs are looked up from manually maintained tables (`fifaTeamSlugs`, `fifaHostCitySlugs`, `fifaHostCountrySlugs`) since they don't always match the openfootball name or the stadium's own (sometimes sponsor-renamed) name — key stadium/host-city lookups off `Stadium.City`, not `Stadium.Name`. Render any of these with the `linked` template func, e.g. `{{linked $.TeamLinks .TeamField}}` or `{{linked $.StadiumLinks .Ground}}`, which falls back to plain text if no slug is mapped.
- **Match scores**: render a played match's result with the `score` template func, e.g. `{{score .FullTime.Home .FullTime.Away}}`, which wraps it in a `<span class="score">` pill (see `web/static/css/style.css`). Unplayed matches should use `<span class="badge-scheduled">scheduled</span>` rather than bare text, for the same reason.
- **Tables**: every `<table class="table">` must be wrapped in `<div class="table-responsive">...</div>` (see `web/static/css/style.css`) so wide tables scroll horizontally inside their card instead of overflowing the page. Cells containing a team name (or two, for "Team1 vs Team2") should use `<td class="team-cell">` with the team(s) wrapped in `<span class="team">` as above, so the flag+name pair doesn't line-break.
- **Theme colors**: dark/light palettes are CSS custom properties on `:root`/`[data-theme="dark"]` and `[data-theme="light"]` in `web/static/css/style.css` (`--bg`, `--bg-elevated`, `--bg-elevated-2`, `--text`, `--text-muted`, `--border`, `--accent`, `--accent-hover`, `--accent-contrast`, `--accent-2`/`--accent-2-contrast` for the score pill, `--link`, `--table-stripe`, `--shadow`/`--shadow-sm`). Add new colors as variables in both blocks rather than hardcoding hex values in component rules, so the theme toggle (`web/static/js/app.js`, `data-theme` attribute) keeps working for every element.

## Testing

- `internal/data`: parser tests (`parser_test.go`) validate JSON normalization edge cases; store tests (`store_test.go`) validate refresh success/failure behavior using `httptest`.
- `internal/services`: table-driven tests with hand-built `models.Tournament` fixtures — no need to hit the real data source.
- `internal/handlers`: integration tests spin up the real router (`handlers.NewRouter`) against a store seeded from the embedded fallback, and assert on rendered HTML via `httptest`.
- Run `make test` or `make test-coverage`.

## Docker

- `Dockerfile` is a multi-stage build producing a static binary run on a `distroless/static-debian12:nonroot` image (no shell). The Docker healthcheck relies on `cmd/server`'s `-healthcheck` flag (a local `GET /healthz`), since there's no `curl`/`wget` in the runtime image — keep that flag working if you touch `main.go`.
- `make docker-build` / `make docker-up` / `make docker-down` wrap `docker compose`.

## Helm

- `charts/my-world-cup-app` is a standard Helm chart (`Chart.yaml`, `values.yaml`, `templates/`). Probes point at `/healthz`; pod annotations pre-declare `prometheus.io/scrape|port|path` for `/metrics` auto-discovery.
- Validate changes with `make helm-lint` (or `helm lint charts/my-world-cup-app`) and `helm template charts/my-world-cup-app` (re-render with `--set ingress.enabled=true --set autoscaling.enabled=true` to exercise the optional templates) before committing.
- `charts/my-world-cup-app/README.md` is **generated, not hand-written** — it's produced by `make helm-docs` (wraps [helm-docs](https://github.com/norwoodj/helm-docs)) from `charts/my-world-cup-app/README.md.gotmpl` plus the `# -- description` head-comments in `values.yaml`. If you add/rename/remove a `values.yaml` key, add a matching `# -- ...` comment directly above it and re-run `make helm-docs` rather than editing the chart's `README.md` by hand.
- `make helm-install` / `make helm-uninstall` wrap `helm upgrade --install` / `helm uninstall` against the `NAMESPACE` variable (defaults to the app name).

## Common Pitfalls

- `go:embed` patterns cannot use `..` — this is why templates/static assets live under `web/` with their own `embed.go` (`web/assets.go`), rather than being embedded directly from `internal/handlers`.
- The openfootball match JSON omits the `score` field entirely for matches not yet played; `models.Match.Played()` checks `FullTime != nil`, not an empty score — don't assume `score.ft` is always present.
- Group standings tie-break is simplified (points → goal difference → goals for → alphabetical) and does **not** implement FIFA's head-to-head rule. This is documented in the UI and README; if implementing full tie-break rules, update both.
