.DEFAULT_GOAL := help
SHELL := /bin/bash

APP_NAME := my-world-cup-app
BIN_DIR := bin
CHART_DIR := charts/$(APP_NAME)
PORT ?= 8080
NAMESPACE ?= $(APP_NAME)
# The single source of truth for the app's release version — bump this file
# (see CONTRIBUTING.md) to change the Docker image tag built/pushed below
# and the value `make helm-sync-version` writes into the Helm chart's
# appVersion.
VERSION_FILE := VERSION
APP_VERSION := $(shell cat $(VERSION_FILE) 2>/dev/null || echo 0.0.0)
DOCKER_TAG ?= $(APP_VERSION)
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_BUILDER := $(APP_NAME)-builder

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: check-deps
check-deps: ## Verify required development/runtime tools are installed
	@missing=0; \
	echo "Checking required tools..."; \
	check() { \
		if command -v "$$1" >/dev/null 2>&1; then \
			printf "  [OK]      %-15s %s\n" "$$1" "$$(command -v $$1)"; \
		else \
			printf "  [MISSING] %-15s install: %s\n" "$$1" "$$2"; \
			missing=1; \
		fi; \
	}; \
	check git "https://git-scm.com/downloads"; \
	check go "https://go.dev/doc/install"; \
	check docker "https://docs.docker.com/get-docker/"; \
	check helm "https://helm.sh/docs/intro/install/"; \
	check helm-docs "https://github.com/norwoodj/helm-docs#installation"; \
	if docker compose version >/dev/null 2>&1; then \
		printf "  [OK]      %-15s %s\n" "docker compose" "$$(docker compose version --short 2>/dev/null)"; \
	else \
		printf "  [MISSING] %-15s install: %s\n" "docker compose" "https://docs.docker.com/compose/install/"; \
		missing=1; \
	fi; \
	if docker buildx version >/dev/null 2>&1; then \
		printf "  [OK]      %-15s %s\n" "docker buildx" "$$(docker buildx version 2>/dev/null)"; \
	else \
		printf "  [MISSING] %-15s install: %s\n" "docker buildx" "https://docs.docker.com/build/architecture/#buildx"; \
		missing=1; \
	fi; \
	echo ""; \
	if [ "$$missing" -eq 1 ]; then \
		echo "One or more required tools are missing. Install them using the links above."; \
		exit 1; \
	fi; \
	echo "All required tools are installed."

.PHONY: run
run: ## Run the application locally
	PORT=$(PORT) go run ./cmd/server

.PHONY: build
build: ## Build the server binary into bin/
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) ./cmd/server

.PHONY: test
test: ## Run all tests
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: fmt
fmt: ## Format source code
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Check source code formatting
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files:"; gofmt -l .; exit 1)

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: tidy
tidy: ## Tidy go.mod/go.sum
	go mod tidy

.PHONY: check
check: fmt-check vet test ## Run formatting, vet, and tests

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
	echo "$$DOCKER_PASS" | docker login $$REGISTRY --username "$$DOCKER_USER" --password-stdin; \
	echo "Building and pushing $$DOCKER_REPO:$$TAG for $(DOCKER_PLATFORMS)..."; \
	docker buildx build --builder $(BUILDX_BUILDER) --platform $(DOCKER_PLATFORMS) --build-arg APP_VERSION=$$TAG -t "$$DOCKER_REPO:$$TAG" --push .

.PHONY: helm-sync-version
helm-sync-version: ## Write the VERSION file's version into the Helm chart's appVersion (charts/my-world-cup-app/Chart.yaml)
	@sed -i.bak "s/^appVersion:.*/appVersion: \"$(APP_VERSION)\"/" $(CHART_DIR)/Chart.yaml && rm -f $(CHART_DIR)/Chart.yaml.bak
	@echo "$(CHART_DIR)/Chart.yaml appVersion set to $(APP_VERSION)"

.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	helm lint $(CHART_DIR)

.PHONY: helm-docs
helm-docs: helm-sync-version ## Sync appVersion from VERSION, then regenerate the Helm chart README (charts/*/README.md) via helm-docs
	helm-docs --chart-search-root charts

.PHONY: helm-install
helm-install: ## Install/upgrade the app into Kubernetes via Helm (namespace: NAMESPACE, default app name)
	helm upgrade --install $(APP_NAME) $(CHART_DIR) -n $(NAMESPACE) --create-namespace

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall the Helm release from Kubernetes
	helm uninstall $(APP_NAME) -n $(NAMESPACE)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) coverage.out
