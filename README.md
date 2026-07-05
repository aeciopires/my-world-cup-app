<!-- TOC -->

- [My World Cup App](#my-world-cup-app)
  - [Features](#features)
  - [Screenshots](#screenshots)
  - [Technology Stack](#technology-stack)
  - [Architecture](#architecture)
    - [Data flow](#data-flow)
    - [Request routing](#request-routing)
  - [Directory Structure](#directory-structure)
  - [Getting Started](#getting-started)
    - [Software Requirements](#software-requirements)
    - [Makefile Targets](#makefile-targets)
    - [Run locally](#run-locally)
    - [Run tests](#run-tests)
    - [Build a binary](#build-a-binary)
    - [Run with Docker](#run-with-docker)
    - [Run with Helm](#run-with-helm)
    - [Configuration](#configuration)
  - [Data Refresh Behavior](#data-refresh-behavior)
  - [Standings Calculation](#standings-calculation)
  - [Metrics](#metrics)
  - [Testing](#testing)
  - [Contributing](#contributing)
  - [Developer](#developer)
  - [License](#license)

<!-- TOC -->

# My World Cup App

A lightweight Go web application that displays the FIFA World Cup 2026 (Canada/Mexico/USA) group standings, match results, and knockout stage, plus curated links to official FIFA resources. Data is fetched live on startup and on demand, with no database required.

## Features

- **Group standings** — computed on the fly from match results (played, won, drawn, lost, goals, goal difference, points).
- **Knockout stage** — a graphical bracket (Round of 32 through the Final, plus the match for third place) with connector lines and winner highlighting, alongside round-by-round detail tables with date/venue/result.
- **Match list** — every fixture with date, round, group, venue, and result, filterable by round, group, and/or team via query parameters (combinable).
- **Live data refresh** — data is fetched from [openfootball/worldcup.json](https://github.com/openfootball/worldcup.json) on startup and via the "Update data" button; an embedded snapshot is used as a fallback if the live source is unreachable.
- **Statistics** — top scorers and overall team records (played/won/drawn/lost/goals/points) aggregated across group and knockout stage matches.
- **Dark / light theme** — toggle persisted in the browser via `localStorage`.
- **Official FIFA links** — stadiums, teams, standings, articles, scores & fixtures, official match ball, posters, mascots, Club World Cup 2025, and the official FIFA Sound playlist; team names, stadiums, and host cities across the app link to their official fifa.com pages.
- **Observability** — `/healthz` health check and a Prometheus-compatible `/metrics` endpoint (HTTP request counts/latency, data refresh outcomes).

## Screenshots

<table>
  <tr>
    <td align="center" width="50%">
      <a href="images/a.png"><img src="images/a.png" alt="Home page - dark theme"></a>
      <br><sub>Home — dark theme</sub>
    </td>
    <td align="center" width="50%">
      <a href="images/b.png"><img src="images/b.png" alt="Home page - light theme"></a>
      <br><sub>Home — light theme</sub>
    </td>
  </tr>
  <tr>
    <td align="center" width="50%">
      <a href="images/c.png"><img src="images/c.png" alt="Group standings - dark theme"></a>
      <br><sub>Group Standings — dark theme</sub>
    </td>
    <td align="center" width="50%">
      <a href="images/d.png"><img src="images/d.png" alt="Knockout stage - light theme"></a>
      <br><sub>Knockout Stage — light theme</sub>
    </td>
  </tr>
  <tr>
    <td align="center" width="50%">
      <a href="images/e.png"><img src="images/e.png" alt="All matches - dark theme"></a>
      <br><sub>All Matches — dark theme</sub>
    </td>
    <td align="center" width="50%">
      <a href="images/f.png"><img src="images/f.png" alt="Statistics - light theme"></a>
      <br><sub>Statistics (top scorers &amp; team overview) — light theme</sub>
    </td>
  </tr>
  <tr>
    <td align="center" width="50%">
      <a href="images/g.png"><img src="images/g.png" alt="FIFA links - dark theme"></a>
      <br><sub>Official FIFA Links — dark theme</sub>
    </td>
    <td align="center" width="50%">
      <a href="images/h.png"><img src="images/h.png" alt="FIFA links and stadiums - light theme"></a>
      <br><sub>FIFA Links &amp; Stadiums — light theme</sub>
    </td>
  </tr>
</table>

## Technology Stack

- [Go](https://go.dev/) 1.25 — standard library for the web layer (`net/http`, `html/template`, `encoding/json`, `embed`); no web framework.
- [prometheus/client_golang](https://github.com/prometheus/client_golang) — the only third-party dependency, used solely for `/metrics` instrumentation (`internal/metrics`).
- Vanilla CSS (custom properties for theming) and vanilla JavaScript (no build step, no client framework).
- Docker / Docker Compose for containerized runs; a Helm chart for Kubernetes deployment.
- `go test` for unit and integration tests.

## Architecture

The application follows a clean, layered architecture:

```
cmd/server        entrypoint: wiring, HTTP server lifecycle
internal/config    environment-driven configuration
internal/models    domain types (Team, Group, Match, Stadium, Standing, Tournament)
internal/data      HTTP client, JSON parsing/normalization, thread-safe in-memory store
internal/services  business logic: group standings, knockout grouping, match/stats queries
internal/handlers  HTTP handlers, routing, template rendering
internal/metrics   Prometheus instrumentation (HTTP requests, data refresh outcomes)
web/               embedded HTML templates and static assets (CSS/JS)
charts/            Helm chart for Kubernetes deployment
```

- **Handlers** depend on **services** and the **data store**, never the other way around.
- **Services** are pure functions operating on **models**, independent of HTTP or the data source — easy to unit test.
- **Data** owns fetching, parsing, and caching; it exposes a `Store` with `Snapshot()` and `Refresh()`.
- No database: the `Store` holds the current `Tournament` in memory behind a `sync.RWMutex`. Standings are recomputed from match results on every request rather than persisted.

### Data flow

```mermaid
flowchart TD
    subgraph Startup
        A[main.go] --> B[data.NewStore]
        B --> C[Seed from embedded fallback JSON]
        A --> D[Background goroutine: store.Refresh]
    end

    subgraph Refresh["Data Refresh (startup or /refresh)"]
        D --> E[data.Client.Fetch]
        E -->|success| F[parse: JSON -> models.Tournament]
        F --> G[Store.set: update in-memory snapshot]
        E -->|failure| H[Keep previous snapshot, log warning]
    end

    subgraph Request["HTTP Request"]
        I[Browser] --> J[handlers.Router]
        J --> K[Store.Snapshot]
        K --> L[services: GroupStandings / KnockoutStage / AllMatches / Stats]
        L --> M[html/template render]
        M --> I
        J -.-> P[metrics.ObserveRequest]
    end

    subgraph UI["Update data button"]
        N[Click 'Update data'] --> O[POST /refresh]
        O --> E
    end

    G -.-> Q[metrics.RecordRefresh]
    H -.-> Q
    Q --> R["/metrics (Prometheus exposition)"]
```

### Request routing

| Method | Path        | Handler              | Description                          |
|--------|-------------|-----------------------|---------------------------------------|
| GET    | `/`         | `pages.Home`          | Dashboard: upcoming/recent matches    |
| GET    | `/groups`   | `pages.Groups`        | Group standings and results           |
| GET    | `/knockout` | `pages.Knockout`      | Knockout stage bracket + round detail |
| GET    | `/matches`  | `pages.Matches`       | Match list, filterable by `round`/`group`/`team` query params |
| GET    | `/stats`    | `pages.Stats`         | Top scorers and team statistics       |
| GET    | `/links`    | `pages.Links`         | Official FIFA/Spotify links, stadiums |
| POST   | `/refresh`  | `refresh.Refresh`     | Triggers a live data refresh          |
| GET    | `/healthz`  | `healthz`             | Health check (used by Docker/Helm)    |
| GET    | `/metrics`  | Prometheus handler    | Prometheus metrics exposition         |
| GET    | `/static/*` | embedded file server  | CSS/JS assets                         |

## Directory Structure

```
my-world-cup-app/
├── cmd/server/main.go
├── internal/
│   ├── config/
│   ├── models/
│   ├── data/
│   │   ├── client.go
│   │   ├── fallback.go
│   │   ├── fallback/            # embedded snapshot JSON (openfootball 2026 data)
│   │   ├── parser.go
│   │   └── store.go
│   ├── services/                # standings, knockout, matches, stats
│   ├── handlers/
│   └── metrics/                  # Prometheus counters/histograms
├── web/
│   ├── assets.go                # go:embed directives
│   ├── templates/
│   └── static/{css,js}/
├── charts/my-world-cup-app/      # Helm chart
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── README.md
├── CHANGELOG.md
└── CLAUDE.md
```

## Getting Started

### Software Requirements

| Tool                                                     | Minimum Version | Required For                          |
|-----------------------------------------------------------|------------------|-----------------------------------------|
| [Go](https://go.dev/doc/install)                           | 1.25             | Building/running/testing locally        |
| [Make](https://www.gnu.org/software/make/)                 | any              | Running the `Makefile` shortcuts        |
| [Git](https://git-scm.com/downloads)                       | any              | Cloning the repository, contributing    |
| [Docker](https://docs.docker.com/get-docker/)               | 24+              | Building/running the container image    |
| [Docker Compose](https://docs.docker.com/compose/install/) | v2 (plugin)      | Local containerized run (`docker compose`) |
| [Helm](https://helm.sh/docs/intro/install/)                 | 3.x              | Installing/linting the Helm chart       |
| [helm-docs](https://github.com/norwoodj/helm-docs#installation) | 1.x        | Regenerating `charts/my-world-cup-app/README.md` |

Run `make check-deps` to verify which of these are installed on your machine; it prints installation instructions for anything missing.

See also [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution workflow.

### Makefile Targets

Run `make help` (or just `make`, since `help` is the default goal) to print this list from the terminal.

| Target                | Description                                                                  |
|------------------------|--------------------------------------------------------------------------------|
| `make help`            | Show the list of available targets                                           |
| `make check-deps`      | Verify required development/runtime tools are installed                      |
| `make run`             | Run the application locally                                                  |
| `make build`           | Build the server binary into `bin/`                                          |
| `make test`            | Run all tests                                                                |
| `make test-coverage`   | Run tests with a coverage report                                             |
| `make fmt`             | Format source code                                                           |
| `make fmt-check`       | Check source code formatting                                                 |
| `make vet`             | Run `go vet`                                                                 |
| `make tidy`            | Tidy `go.mod`/`go.sum`                                                       |
| `make check`           | Run formatting, vet, and tests (`fmt-check` + `vet` + `test`)                |
| `make docker-build`    | Build the Docker image                                                       |
| `make docker-up`       | Start the application via Docker Compose                                     |
| `make docker-down`     | Stop and remove the Docker Compose services                                  |
| `make docker-logs`     | Tail the application container logs                                          |
| `make helm-lint`       | Lint the Helm chart                                                          |
| `make helm-docs`       | Regenerate the Helm chart README (`charts/*/README.md`) via helm-docs        |
| `make helm-install`    | Install/upgrade the app into Kubernetes via Helm (namespace: `NAMESPACE`, default app name) |
| `make helm-uninstall`  | Uninstall the Helm release from Kubernetes                                   |
| `make clean`           | Remove build artifacts                                                       |

### Run locally

```bash
make run
# or
PORT=8080 go run ./cmd/server
```

Then open http://localhost:8080.

### Run tests

```bash
make test              # go test ./... -v
make test-coverage     # with coverage report
```

### Build a binary

```bash
make build              # outputs bin/my-world-cup-app
```

### Run with Docker

```bash
make docker-build       # docker compose build
make docker-up          # docker compose up -d --build
make docker-logs        # tail logs
make docker-down        # stop and remove
```

The container serves the app on `PORT` (default `8080`), mapped to the host via `docker-compose.yml`.

### Run with Helm

A Helm chart is provided at `charts/my-world-cup-app` for Kubernetes deployment:

```bash
helm lint charts/my-world-cup-app
helm template my-world-cup-app charts/my-world-cup-app   # render manifests locally
helm install my-world-cup-app charts/my-world-cup-app --set image.repository=<your-registry>/my-world-cup-app --set image.tag=<tag>
```

The chart deploys a single `Deployment` + `Service` (ClusterIP by default), wires `/healthz` as the liveness/readiness probe, and pre-populates `prometheus.io/scrape`, `prometheus.io/port`, and `prometheus.io/path` pod annotations so a cluster Prometheus can auto-discover `/metrics`. Ingress and HPA are included but disabled by default (`ingress.enabled` / `autoscaling.enabled` in `values.yaml`).

### Configuration

All configuration is via environment variables (see `internal/config/config.go`). There are no required variables — every one of them has a working default.

| Variable                | Default                                            | Description                |
|--------------------------|-----------------------------------------------------|------------------------------|
| `PORT`                   | `8080`                                              | HTTP listen port            |
| `WORLDCUP_MATCHES_URL`   | openfootball `2026/worldcup.json`                   | Match data source            |
| `WORLDCUP_GROUPS_URL`    | openfootball `2026/worldcup.groups.json`            | Group assignments source     |
| `WORLDCUP_TEAMS_URL`     | openfootball `2026/worldcup.teams.json`             | Team metadata source          |
| `WORLDCUP_STADIUMS_URL`  | openfootball `2026/worldcup.stadiums.json`          | Stadium data source           |

## Data Refresh Behavior

1. On startup, the app seeds itself from an **embedded snapshot** of the four openfootball JSON files (bundled at build time via `go:embed`), so it can serve pages immediately.
2. A background goroutine performs an initial **live refresh** against the configured URLs.
3. Clicking **"Update data"** in the UI (or `POST /refresh`) triggers a synchronous live refresh.
4. If a live fetch fails (network issue, rate limit, etc.), the previous snapshot is kept and the failure is logged — the app never serves a broken or empty page.

## Standings Calculation

Group tables (and the overall team statistics on `/stats`) are computed from played matches using standard football scoring (3 points for a win, 1 for a draw). Ties are broken by: points → goal difference → goals for → alphabetical order. This is a simplified tie-break; it does not implement FIFA's full head-to-head/fair-play rules.

## Metrics

`/metrics` exposes a dedicated Prometheus registry (`internal/metrics`), not the global default one:

- `http_requests_total{method,path,status}` — request counts, `status` bucketed as `2xx`/`3xx`/`4xx`/`5xx`.
- `http_request_duration_seconds{method,path}` — request latency histogram.
- `data_refresh_total{outcome}` — count of refresh attempts, `outcome` = `success`/`failure`.
- `data_last_refresh_timestamp_seconds` — Unix timestamp of the last successful refresh.

## Testing

- `internal/data`: JSON parsing/normalization tests, plus store tests covering successful refresh, failed refresh (previous snapshot retained), and fallback seeding.
- `internal/services`: standings, knockout grouping, top-scorer, and team-statistics computation tests with known inputs/expected outputs.
- `internal/handlers`: HTTP integration tests (`httptest`) covering every route (including `/stats` and `/metrics`), static asset serving, and refresh failure handling.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the fork/branch/PR workflow and recommended editor setup.

## Developer

Aecio dos Santos Pires

- Linkedin: https://www.linkedin.com/in/aeciopires/
- Site: http://aeciopires.com/

## License

See [LICENSE](LICENSE).
