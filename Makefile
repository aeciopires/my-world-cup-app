.DEFAULT_GOAL := help

APP_NAME := my-world-cup-app
BIN_DIR := bin
CHART_DIR := charts/$(APP_NAME)
PORT ?= 8080
NAMESPACE ?= $(APP_NAME)

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
docker-build: ## Build the Docker image
	docker compose build

.PHONY: docker-up
docker-up: ## Start the application via Docker Compose
	docker compose up -d --build

.PHONY: docker-down
docker-down: ## Stop and remove the Docker Compose services
	docker compose down

.PHONY: docker-logs
docker-logs: ## Tail the application container logs
	docker compose logs -f app

.PHONY: helm-lint
helm-lint: ## Lint the Helm chart
	helm lint $(CHART_DIR)

.PHONY: helm-docs
helm-docs: ## Regenerate the Helm chart README (charts/*/README.md) via helm-docs
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
