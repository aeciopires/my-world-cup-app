.DEFAULT_GOAL := help
SHELL := /bin/bash

APP_NAME := my-world-cup-app
APP_DIR := app
BIN_DIR := bin
CHART_DIR := helm/$(APP_NAME)
PORT ?= 8080
NAMESPACE ?= $(APP_NAME)
KIND_CLUSTER ?= kind-multinodes
# The single source of truth for the app's release version — bump this file
# (see CONTRIBUTING.md) to change the Docker image tag built/pushed below
# and the value `make helm-sync-version` writes into the Helm chart's
# appVersion.
VERSION_FILE := VERSION
APP_VERSION := $(shell cat $(VERSION_FILE) 2>/dev/null || echo 0.0.0)
DOCKER_TAG ?= $(APP_VERSION)
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_BUILDER := $(APP_NAME)-builder
# Stamps the release into the binary (reported as the traces' service.version).
GO_LDFLAGS := -s -w -X main.version=$(APP_VERSION)
# Local observability stack: the app plus Grafana/Prometheus/Loki/Tempo/
# Jaeger/Alloy (see docker-compose.observability.yml and observability/).
# Host ports are overridable when they clash, e.g. `make observability-up GRAFANA_PORT=13000`.
GRAFANA_PORT ?= 3000
PROMETHEUS_PORT ?= 9090
LOKI_PORT ?= 3100
TEMPO_PORT ?= 3200
JAEGER_PORT ?= 16686
ALLOY_PORT ?= 12345
OBS_COMPOSE := PORT=$(PORT) GRAFANA_PORT=$(GRAFANA_PORT) PROMETHEUS_PORT=$(PROMETHEUS_PORT) LOKI_PORT=$(LOKI_PORT) \
	TEMPO_PORT=$(TEMPO_PORT) JAEGER_PORT=$(JAEGER_PORT) ALLOY_PORT=$(ALLOY_PORT) \
	docker compose -f docker-compose.yml -f docker-compose.observability.yml

# Run toolchain commands through mise so the versions pinned in mise.toml are
# used even when mise isn't activated in the user's shell. Falls back to
# whatever is on PATH when mise isn't installed (run `make install-deps`).
MISE := $(shell command -v mise 2>/dev/null || { test -x $(HOME)/.local/bin/mise && echo $(HOME)/.local/bin/mise; })
RUN := $(if $(MISE),$(MISE) exec --,)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

.PHONY: check-deps
check-deps: ## Check that every required tool (base packages, mise-pinned toolchain, Docker) is installed
	@scripts/deps.sh check

.PHONY: install-deps
install-deps: ## Install base packages, mise, and the toolchain pinned in mise.toml (Docker is installed separately)
	@scripts/deps.sh install

.PHONY: run
run: ## Run the application locally
	PORT=$(PORT) $(RUN) go -C $(APP_DIR) run -ldflags="-X main.version=$(APP_VERSION)" ./cmd/server

.PHONY: build
build: ## Build the server binary into bin/
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(RUN) go -C $(APP_DIR) build -trimpath -ldflags="$(GO_LDFLAGS)" -o $(CURDIR)/$(BIN_DIR)/$(APP_NAME) ./cmd/server

.PHONY: test
test: ## Run all tests
	$(RUN) go -C $(APP_DIR) test ./... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	$(RUN) go -C $(APP_DIR) test ./... -cover -coverprofile=$(CURDIR)/coverage.out
	$(RUN) go -C $(APP_DIR) tool cover -func=$(CURDIR)/coverage.out

.PHONY: fmt
fmt: ## Format source code
	$(RUN) gofmt -w $(APP_DIR)

.PHONY: fmt-check
fmt-check: ## Check source code formatting
	@test -z "$$($(RUN) gofmt -l $(APP_DIR))" || (echo "Unformatted files:"; $(RUN) gofmt -l $(APP_DIR); exit 1)

.PHONY: vet
vet: ## Run go vet
	$(RUN) go -C $(APP_DIR) vet ./...

.PHONY: lint
lint: ## Run golangci-lint (config: app/.golangci.yml, includes the modernize analyzer)
	cd $(APP_DIR) && $(RUN) golangci-lint run ./...

.PHONY: tidy
tidy: ## Tidy go.mod/go.sum
	$(RUN) go -C $(APP_DIR) mod tidy

.PHONY: check
check: fmt-check vet lint test ## Run formatting, vet, lint, and tests

.PHONY: docker-build
docker-build: ## Build the Docker image (tagged with the VERSION file's version)
	APP_VERSION=$(APP_VERSION) docker compose build

.PHONY: docker-up
docker-up: ## Start the application via Docker Compose
	APP_VERSION=$(APP_VERSION) docker compose up -d --build

.PHONY: docker-down
docker-down: ## Stop and remove the Docker Compose services
	docker compose down

.PHONY: docker-logs
docker-logs: ## Tail the application container logs
	docker compose logs -f app

.PHONY: observability-up
observability-up: ## Start the app plus the observability stack (Grafana, Prometheus, Loki, Tempo, Jaeger, Alloy) via Docker Compose
	APP_VERSION=$(APP_VERSION) $(OBS_COMPOSE) up -d --build
	@echo ""
	@echo "  App         http://localhost:$(PORT)"
	@echo "  Grafana     http://localhost:$(GRAFANA_PORT)   (dashboards: folder 'My World Cup App')"
	@echo "  Prometheus  http://localhost:$(PROMETHEUS_PORT)"
	@echo "  Jaeger      http://localhost:$(JAEGER_PORT)"
	@echo "  Alloy       http://localhost:$(ALLOY_PORT)"

.PHONY: observability-down
observability-down: ## Stop the app and observability stack (add VOLUMES=1 to also delete stored metrics/logs/traces)
	$(OBS_COMPOSE) down $(if $(VOLUMES),--volumes,)

.PHONY: observability-logs
observability-logs: ## Tail the logs of every container in the observability stack
	$(OBS_COMPOSE) logs -f

.PHONY: observability-validate
observability-validate: ## Validate the observability Compose file, Alloy/Prometheus/Loki/Tempo configs, and Grafana dashboard JSON (needs Docker)
	@scripts/observability-validate.sh

.PHONY: docker-buildx-setup
docker-buildx-setup: ## Create (or reuse) the Docker Buildx builder used for multi-arch images
	@docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1 || docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --use
	@docker buildx inspect $(BUILDX_BUILDER) --bootstrap >/dev/null

.PHONY: docker-build-multiarch
docker-build-multiarch: docker-buildx-setup ## Build a multi-arch image (linux/amd64 + linux/arm64, runs on Linux and macOS/Intel+Apple Silicon) without pushing, to validate the build for both platforms
	docker buildx build --builder $(BUILDX_BUILDER) --platform $(DOCKER_PLATFORMS) --build-arg APP_VERSION=$(APP_VERSION) -t $(APP_NAME):$(DOCKER_TAG) .

.PHONY: docker-push
docker-push: docker-buildx-setup helm-sync-version ## Build and push a multi-arch image (linux/amd64 + linux/arm64), tagged from the VERSION file by default; interactively prompts for registry username, password/token, and repository name
	@read -r -p "Docker registry username: " DOCKER_USER; \
	read -r -s -p "Docker registry password or access token: " DOCKER_PASS; echo; \
	read -r -p "Repository name (e.g. docker.io/<user>/$(APP_NAME) or ghcr.io/<user>/$(APP_NAME)): " DOCKER_REPO; \
	read -r -p "Image tag [$(DOCKER_TAG)]: " DOCKER_TAG_INPUT; \
	if [ -z "$$DOCKER_USER" ] || [ -z "$$DOCKER_PASS" ] || [ -z "$$DOCKER_REPO" ]; then \
		echo "Username, password, and repository name are all required. Aborting."; \
		exit 1; \
	fi; \
	TAG=$${DOCKER_TAG_INPUT:-$(DOCKER_TAG)}; \
	REGISTRY=$$(echo "$$DOCKER_REPO" | cut -d/ -f1); \
	case "$$REGISTRY" in \
		*.*|*:*) ;; \
		*) REGISTRY="" ;; \
	esac; \
	printf '%s\n' "$$DOCKER_PASS" | docker login $$REGISTRY --username "$$DOCKER_USER" --password-stdin; \
	echo "Building and pushing $$DOCKER_REPO:$$TAG for $(DOCKER_PLATFORMS)..."; \
	docker buildx build --builder $(BUILDX_BUILDER) --platform $(DOCKER_PLATFORMS) --build-arg APP_VERSION=$$TAG -t "$$DOCKER_REPO:$$TAG" --push .

.PHONY: helm-sync-version
helm-sync-version: ## Write the VERSION file's version into the Helm chart's appVersion (helm/my-world-cup-app/Chart.yaml)
	@sed -i.bak "s/^appVersion:.*/appVersion: \"$(APP_VERSION)\"/" $(CHART_DIR)/Chart.yaml && rm -f $(CHART_DIR)/Chart.yaml.bak
	@echo "$(CHART_DIR)/Chart.yaml appVersion set to $(APP_VERSION)"

.PHONY: helm-lint
helm-lint: ## Lint the Helm chart (Helm 4)
	$(RUN) helm lint $(CHART_DIR)
	$(RUN) helm lint $(CHART_DIR) --set ingress.enabled=true --set autoscaling.enabled=true

.PHONY: helm-template
helm-template: ## Render the Helm chart locally, including the optional ingress/HPA templates
	$(RUN) helm template $(APP_NAME) $(CHART_DIR) --set ingress.enabled=true --set autoscaling.enabled=true

.PHONY: helm-docs
helm-docs: helm-sync-version ## Sync appVersion from VERSION, then regenerate the Helm chart README (helm/*/README.md) via helm-docs
	$(RUN) helm-docs --chart-search-root helm

.PHONY: kind-load
kind-load: docker-build ## Load the local Docker image into the kind cluster (KIND_CLUSTER, default kind-multinodes)
	$(RUN) kind load docker-image $(APP_NAME):$(APP_VERSION) --name $(KIND_CLUSTER)

.PHONY: helm-install
helm-install: ## Install/upgrade the app into Kubernetes via Helm 4 (namespace: NAMESPACE, default app name)
	$(RUN) helm upgrade --install $(APP_NAME) $(CHART_DIR) -n $(NAMESPACE) --create-namespace --wait --timeout 5m

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall the Helm release from Kubernetes
	$(RUN) helm uninstall $(APP_NAME) -n $(NAMESPACE)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out
