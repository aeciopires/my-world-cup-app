# Glossary

[← Back to the index](README.md)

| Term | Meaning |
|------|---------|
| **a.e.t.** | "After extra time" — a knockout match that needed the extra 30 minutes |
| **Alloy (Grafana Alloy)** | A telemetry collector: receives traces, reads container logs, and forwards them to the right storage (Tempo, Jaeger, Loki) |
| **Binary** | The single executable file produced by `go build` |
| **Container** | A running instance of a Docker image, isolated from the rest of the machine |
| **Context (`context.Context`)** | A Go value carrying a deadline/cancellation signal through a chain of calls ("stop working, nobody's waiting anymore") |
| **Counter / gauge / histogram** | Prometheus metric types: only-up number / up-and-down number / distribution in buckets |
| **Distroless** | A container base image with only what's needed to run a program — no shell, no package manager |
| **Embedded files (`go:embed`)** | Files copied into the compiled binary at build time |
| **Endpoint / route** | A method + path the server answers, e.g. `GET /groups` |
| **Exemplar** | A sample request attached to a metric data point, carrying its trace ID, so a graph can link to the exact trace |
| **Fallback snapshot** | The copy of the openfootball data embedded in the binary, used when the live download fails |
| **Goroutine** | A lightweight thread managed by the Go runtime; started with `go f()` or `wg.Go(f)` |
| **Graceful shutdown / draining** | Stopping without dropping requests: fail readiness, wait for traffic to move away, finish in-flight requests, then exit |
| **Grafana** | A web app for dashboards and data exploration across many data sources (Prometheus, Loki, Tempo, Jaeger, ...) |
| **Handler** | A function that receives an HTTP request and writes the response |
| **Helm / chart** | Kubernetes' package manager / a package of Kubernetes templates. This project requires Helm 4 |
| **HTTP status codes** | `200` OK · `404` not found · `429` too many requests · `500` server bug · `502` upstream (openfootball) failed · `503` not ready (shutting down) |
| **Idiomatic / modern Go** | Code written the way the Go community currently recommends (see [chapter 9](09-dev-workflow.md#writing-modern-go)) |
| **Integration test** | A test exercising several layers together (here: router + handlers + services + templates) |
| **Jaeger** | An open-source tracing backend with its own web UI for searching and comparing traces |
| **JSON log** | A log line written as a JSON object (`{"level":"INFO","msg":"request",...}`), easy for programs like Loki to query |
| **JSON** | A text format for data (`{"team1": "Mexico"}`) — what openfootball publishes |
| **kind** | "Kubernetes in Docker": a full cluster running in containers on your laptop |
| **Kubernetes** | A system that runs and supervises containers across machines |
| **Label (metrics)** | A key/value dimension on a metric, e.g. `status="2xx"`; each distinct combination is a separate time series |
| **Linter** | A program that reads code and warns about bugs, bad style, or security issues |
| **Liveness / readiness** | Two health questions: "is the process alive?" (`/healthz`, failing → restart) vs. "should it get traffic?" (`/readyz`, failing → no new requests) |
| **Loki** | Grafana's log database: stores log lines indexed by a few labels and queried with LogQL |
| **Middleware** | A wrapper around a handler that runs before/after it (logging, security headers, crash recovery) |
| **mise** | A tool-version manager: installs the versions pinned in `mise.toml` |
| **Mutex / RWMutex** | Locks that stop goroutines from changing shared data at the same time; an RWMutex allows many readers *or* one writer |
| **openfootball** | The open-data project publishing the World Cup JSON files this app reads |
| **OpenTelemetry (OTel)** | The vendor-neutral standard (APIs, SDKs, protocol) for producing traces, metrics, and logs |
| **OTLP** | The OpenTelemetry Protocol: how spans travel from the app to a collector (here over HTTP, port 4318) |
| **Own goal** | A goal a player scores into their own net; it counts for the other team |
| **Panic** | A Go run-time crash (e.g. reading past the end of a slice); recovered by `recoverMiddleware` |
| **Parse** | Turning raw text (JSON) into structured values (Go structs) |
| **Placeholder (team)** | Stand-in names like `W74` ("winner of match 74") or `2A` ("runner-up of group A") for undecided knockout slots |
| **Pod** | The smallest unit Kubernetes runs: one or more containers |
| **Prometheus** | A monitoring system that periodically reads `/metrics` endpoints and stores the numbers |
| **Pure function** | A function whose output depends only on its inputs, with no I/O or side effects |
| **Query parameter** | The `?key=value` part of a URL, e.g. `/matches?team=Brazil` |
| **Slice** | Go's growable list type, e.g. `[]models.Match` |
| **Snapshot** | The store's current data plus when it was loaded and from where (`live` / `fallback`) |
| **Span** | One timed step inside a trace (e.g. "GET /groups", "data.Refresh", one download), with a start time, duration, and attributes |
| **Struct** | A Go type grouping named fields, e.g. `models.Match` |
| **Structured metadata (Loki)** | Per-line key/value data stored next to a log line without being an index label, ideal for high-cardinality values like `trace_id` |
| **Template** | An HTML file with `{{ ... }}` placeholders that Go fills with data |
| **Tempo** | Grafana's trace database; it can also turn spans into metrics and a service graph |
| **Tie-break** | The rule deciding the order of teams with equal points |
| **Trace / trace ID** | The full story of one request across all the steps it took; the trace ID is the random identifier shared by all its spans, logs, and exemplars |
