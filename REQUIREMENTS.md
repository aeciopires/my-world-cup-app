<!-- TOC -->

- [Requirements](#requirements)
  - [Supported platforms](#supported-platforms)
  - [Quick setup](#quick-setup)
  - [Toolchain (managed by mise)](#toolchain-managed-by-mise)
  - [Base packages (installed by make install-deps)](#base-packages-installed-by-make-install-deps)
  - [Container tooling (installed separately)](#container-tooling-installed-separately)
    - [Ubuntu 22.04 / 24.04 / 26.04](#ubuntu-2204--2404--2604)
    - [macOS (Intel and Apple Silicon)](#macos-intel-and-apple-silicon)
  - [Activating mise in your shell](#activating-mise-in-your-shell)
  - [Runtime requirements](#runtime-requirements)
  - [Updating tool versions](#updating-tool-versions)
  - [Troubleshooting](#troubleshooting)

<!-- TOC -->

# Requirements

Everything you need installed to develop, test, build, and deploy **My World Cup App**.

## Supported platforms

| Operating system | Architecture             | Status    |
|------------------|--------------------------|-----------|
| Ubuntu 22.04 LTS | amd64 (`x86_64`)         | Supported |
| Ubuntu 24.04 LTS | amd64 (`x86_64`)         | Supported |
| Ubuntu 26.04 LTS | amd64 (`x86_64`)         | Supported |
| macOS            | amd64 (Intel)            | Supported |
| macOS            | arm64 (Apple Silicon)    | Supported |

Other Linux distributions and architectures will probably work (every tool below ships for them), but they aren't tested. `make check-deps` / `make install-deps` print a warning on an unsupported platform and carry on.

The **container images** the project builds are `linux/amd64` and `linux/arm64`, so they run on all of the platforms above (Docker Desktop on macOS runs Linux containers).

## Quick setup

```bash
git clone https://github.com/aeciopires/my-world-cup-app.git
cd my-world-cup-app

make install-deps   # base packages + mise + every tool pinned in mise.toml
make check-deps     # verify everything (also checks Docker)
make check          # fmt-check + vet + lint + tests
```

`make install-deps` runs [`scripts/deps.sh install`](scripts/deps.sh), which:

1. Installs the base packages (`git`, `make`, `curl`) — via `apt-get` on Ubuntu (asks for `sudo`), or the Xcode Command Line Tools on macOS.
2. Installs [mise](https://mise.jdx.dev) if it isn't present — via Homebrew on macOS when available, otherwise the official `https://mise.run` installer (into `~/.local/bin`).
3. Runs `mise trust` + `mise install` to install the exact tool versions pinned in [`mise.toml`](mise.toml).
4. Runs `make check-deps` to show the final state.

It never installs Docker (see [Container tooling](#container-tooling-installed-separately)).

## Toolchain (managed by mise)

The single source of truth for these versions is [`mise.toml`](mise.toml). Every `Makefile` target runs them through `mise exec`, so the pinned versions are used even if mise isn't activated in your shell.

| Tool | Pinned version | Used for |
|------|----------------|----------|
| [Go](https://go.dev/) | 1.27.4 | Building, running, and testing the app (`app/go.mod` declares `go 1.27.0`) |
| [golangci-lint](https://golangci-lint.run/) | 2.15.0 | `make lint` / `make check` (config: [`app/.golangci.yml`](app/.golangci.yml)) |
| [Helm](https://helm.sh/) | 4.3.2 | Linting/rendering/installing the chart. **Helm 4 is required — Helm 3 is no longer supported** (the chart fails to render with an explicit error under Helm 3) |
| [helm-docs](https://github.com/norwoodj/helm-docs) | 1.14.2 | Regenerating `helm/my-world-cup-app/README.md` (`make helm-docs`) |
| [kind](https://kind.sigs.k8s.io/) | 0.33.0 | Local Kubernetes cluster for testing the chart (`make kind-load`) |
| [kubectl](https://kubernetes.io/docs/reference/kubectl/) | 1.36.2 | Talking to the cluster (port-forwarding, inspecting pods) |

## Base packages (installed by `make install-deps`)

| Tool | Ubuntu | macOS |
|------|--------|-------|
| `git` | `apt-get install git` | Xcode Command Line Tools |
| `make` | `apt-get install make` (GNU Make) | Xcode Command Line Tools (GNU Make 3.81 is enough) |
| `curl` | `apt-get install curl` | Built in |

## Container tooling (installed separately)

The local observability stack (`make observability-up`) needs only Docker + Compose: every component runs as a container. Plan for about 1.5 GB of RAM and free host ports 3000, 3100, 3200, 9090, 12345, 16686 (each overridable, e.g. `GRAFANA_PORT=13000`). `make observability-validate` also uses `python3` (preinstalled on Ubuntu; part of the Xcode Command Line Tools on macOS) to check the dashboard JSON. On macOS, Docker Desktop's default socket (`/var/run/docker.sock`) is what Alloy reads container logs from.

Docker needs a system daemon (Linux) or Docker Desktop (macOS) and per-distribution repositories, so it's deliberately left out of `make install-deps`. `make check-deps` still verifies it.

| Tool | Minimum | Used for |
|------|---------|----------|
| Docker Engine / Docker Desktop | 24+ | `make docker-build`, `make docker-up`, `make kind-load` |
| Docker Compose | v2.20+ (the `docker compose` plugin) | `make docker-up` / `docker-down` / `docker-logs`, `make observability-*` |
| Docker Buildx | v0.10+ | `make docker-build-multiarch`, `make docker-push` |

### Ubuntu 22.04 / 24.04 / 26.04

Follow the official guide: <https://docs.docker.com/engine/install/ubuntu/> (the `docker-ce`, `docker-buildx-plugin`, and `docker-compose-plugin` packages), then let your user run Docker without `sudo`:

```bash
sudo usermod -aG docker "$USER"
newgrp docker   # or log out and back in
```

### macOS (Intel and Apple Silicon)

Install [Docker Desktop](https://docs.docker.com/desktop/setup/install/mac-install/) (bundles Compose and Buildx). Lighter alternatives such as [OrbStack](https://orbstack.dev/) or [Colima](https://github.com/abiosoft/colima) work too, as long as `docker compose` and `docker buildx` are available.

## Activating mise in your shell

The `Makefile` doesn't need it, but to use `go`, `helm`, etc. directly in your terminal, activate mise once:

```bash
# bash (Ubuntu default)
echo 'eval "$(mise activate bash)"' >> ~/.bashrc

# zsh (macOS default)
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc
```

Open a new terminal; inside the repository, `go version` should then print `go1.27.4`.

## Runtime requirements

To just **run** the built app (binary or container) you need none of the above except Docker (for the image). The app:

- needs outbound HTTPS to `raw.githubusercontent.com` to fetch live data — **optional**: without it, it serves the embedded snapshot;
- listens on `PORT` (default `8080`);
- uses ~10–30 MiB of RAM and negligible CPU (the Helm chart requests 50m CPU / 32Mi memory).

Supported browsers: any current Chrome, Edge, Firefox, or Safari (the UI uses plain CSS custom properties and ES5 JavaScript).

## Updating tool versions

1. Edit the version in [`mise.toml`](mise.toml) and run `make install-deps`.
2. For Go, also update `app/go.mod`'s `go` line (`go -C app mod edit -go=X.Y.0`) and the `golang:X.Y-alpine` builder image in the [`Dockerfile`](Dockerfile).
3. Update the tables in this file and run `make check` (and `make helm-lint` for Helm).

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `mise: command not found` after `make install-deps` | `~/.local/bin` isn't on your `PATH` yet — open a new terminal or `export PATH="$HOME/.local/bin:$PATH"` |
| `mise ERROR ... is not trusted` | Run `mise trust` in the repository root |
| `chart my-world-cup-app requires Helm 4` | You're running Helm 3 from outside mise — use `make helm-*` targets or activate mise |
| `permission denied ... /var/run/docker.sock` (Ubuntu) | Add your user to the `docker` group (see above) |
| `docker buildx` missing | Install the `docker-buildx-plugin` package (Ubuntu) or update Docker Desktop |
