# GoLangGraph Makefile
# Comprehensive build and test automation

# Variables
BINARY_NAME=golanggraph
MAIN_PATH=./cmd/golanggraph
PKG_PATH=./pkg/...
EXAMPLES_PATH=./examples/...
DOCS_PATH=./docs

# Docker variables
POSTGRES_CONTAINER=golanggraph-postgres
REDIS_CONTAINER=golanggraph-redis
OLLAMA_CONTAINER=golanggraph-ollama

# Test variables
TEST_TIMEOUT=10m
COVERAGE_OUT=coverage.out
COVERAGE_HTML=coverage.html

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help
help: ## Display this help screen
	@echo "$(BLUE)GoLangGraph - AI Agent Framework$(NC)"
	@echo "$(BLUE)================================$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: install
install: ## Install dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	go mod download
	go mod tidy

.PHONY: build
build: ## Build the binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p bin/
	CGO_ENABLED=0 go build -ldflags="-w -s" -trimpath -o bin/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-all
build-all: ## Build all binaries and examples
	@echo "$(BLUE)Building all binaries...$(NC)"
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(BLUE)Building examples...$(NC)"
	go build -o bin/examples/ $(EXAMPLES_PATH)

.PHONY: run
run: build ## Run the main application
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	./bin/$(BINARY_NAME)

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)
	go clean -cache
	go clean -testcache

##@ Testing

.PHONY: test
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v -timeout $(TEST_TIMEOUT) $(PKG_PATH)

.PHONY: test-short
test-short: ## Run tests in short mode (skip integration tests)
	@echo "$(BLUE)Running short tests...$(NC)"
	go test -v -short -timeout $(TEST_TIMEOUT) $(PKG_PATH)

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "$(BLUE)Running tests with race detector...$(NC)"
	go test -v -race -timeout $(TEST_TIMEOUT) $(PKG_PATH)

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -v -coverprofile=$(COVERAGE_OUT) -covermode=atomic $(PKG_PATH)
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(NC)"

.PHONY: test-integration
test-integration: docker-up ## Run integration tests with databases
	@echo "$(BLUE)Running integration tests...$(NC)"
	@echo "$(YELLOW)Waiting for databases to be ready...$(NC)"
	sleep 10
	go test -v -timeout $(TEST_TIMEOUT) -tags=integration $(PKG_PATH)

.PHONY: test-examples
test-examples: ## Test all examples
	@echo "$(BLUE)Testing examples...$(NC)"
	go test -v $(EXAMPLES_PATH)

.PHONY: test-cli
test-cli: ## Test CLI functionality
	@echo "$(BLUE)Testing CLI functionality...$(NC)"
	go test -v ./test/cli_test.go

.PHONY: test-enhanced
test-enhanced: test test-cli ## Run enhanced test suite including CLI tests
	@echo "$(GREEN)Enhanced test suite completed!$(NC)"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	go test -v -bench=. -benchmem $(PKG_PATH)

##@ Local Development with Ollama

.PHONY: ollama-setup
ollama-setup: ## Setup Ollama with gemma2:2b model
	@echo "$(BLUE)Setting up Ollama...$(NC)"
	@if ! command -v ollama &> /dev/null; then \
		echo "$(RED)Ollama not found. Please install Ollama first.$(NC)"; \
		echo "$(YELLOW)Visit: https://ollama.ai/download$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Pulling gemma2:2b model...$(NC)"
	ollama pull gemma2:2b
	@echo "$(GREEN)Ollama setup complete!$(NC)"

.PHONY: ollama-start
ollama-start: ## Start Ollama service
	@echo "$(BLUE)Starting Ollama service...$(NC)"
	ollama serve &
	@echo "$(GREEN)Ollama service started!$(NC)"

.PHONY: test-local
test-local: ollama-setup docker-up ## Run end-to-end tests with local Ollama and databases
	@echo "$(BLUE)Running local end-to-end tests...$(NC)"
	@echo "$(YELLOW)Waiting for services to be ready...$(NC)"
	sleep 15
	OLLAMA_HOST=http://localhost:11434 \
	POSTGRES_HOST=localhost \
	REDIS_HOST=localhost \
	go test -v -timeout $(TEST_TIMEOUT) -tags=e2e $(PKG_PATH)

.PHONY: demo-local
demo-local: ollama-setup docker-up build ## Run local demo with all services
	@echo "$(BLUE)Running local demo...$(NC)"
	@echo "$(YELLOW)Waiting for services to be ready...$(NC)"
	sleep 10
	OLLAMA_HOST=http://localhost:11434 \
	POSTGRES_HOST=localhost \
	REDIS_HOST=localhost \
	go run examples/database_persistence_demo.go

##@ Docker Services

.PHONY: docker-up
docker-up: ## Start PostgreSQL and Redis containers
	@echo "$(BLUE)Starting Docker services...$(NC)"
	@docker run -d --name $(POSTGRES_CONTAINER) \
		-e POSTGRES_DB=golanggraph \
		-e POSTGRES_USER=testuser \
		-e POSTGRES_PASSWORD=testpass \
		-p 5432:5432 \
		postgres:15-alpine || true
	@docker run -d --name $(REDIS_CONTAINER) \
		-p 6379:6379 \
		redis:7-alpine || true
	@echo "$(GREEN)Docker services started!$(NC)"

.PHONY: docker-down
docker-down: ## Stop and remove Docker containers
	@echo "$(BLUE)Stopping Docker services...$(NC)"
	@docker stop $(POSTGRES_CONTAINER) $(REDIS_CONTAINER) 2>/dev/null || true
	@docker rm $(POSTGRES_CONTAINER) $(REDIS_CONTAINER) 2>/dev/null || true
	@echo "$(GREEN)Docker services stopped!$(NC)"

.PHONY: docker-logs
docker-logs: ## Show Docker container logs
	@echo "$(BLUE)PostgreSQL logs:$(NC)"
	@docker logs $(POSTGRES_CONTAINER) 2>/dev/null || echo "$(RED)PostgreSQL container not running$(NC)"
	@echo "$(BLUE)Redis logs:$(NC)"
	@docker logs $(REDIS_CONTAINER) 2>/dev/null || echo "$(RED)Redis container not running$(NC)"

##@ Code Quality

.PHONY: lint
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not found. Installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt $(PKG_PATH)
	go fmt $(EXAMPLES_PATH)

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet $(PKG_PATH)

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "$(BLUE)Tidying go modules...$(NC)"
	go mod tidy

.PHONY: security
security: ## Run security scan
	@echo "$(BLUE)Running security scan...$(NC)"
	@if command -v gosec &> /dev/null; then \
		gosec $(PKG_PATH); \
	else \
		echo "$(YELLOW)gosec not found. Installing...$(NC)"; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
		gosec $(PKG_PATH); \
	fi

##@ Pre-commit

.PHONY: pre-commit-install pre-commit-run pre-commit-update pre-commit-clean
pre-commit-install: ## Install pre-commit hooks
	@echo "$(BLUE)Installing pre-commit hooks...$(NC)"
	@if command -v pre-commit &> /dev/null; then \
		pre-commit install; \
		pre-commit install --hook-type commit-msg; \
		echo "$(GREEN)Pre-commit hooks installed!$(NC)"; \
	else \
		echo "$(YELLOW)pre-commit not found. Installing...$(NC)"; \
		pip install pre-commit; \
		pre-commit install; \
		pre-commit install --hook-type commit-msg; \
		echo "$(GREEN)Pre-commit hooks installed!$(NC)"; \
	fi

pre-commit-run: ## Run pre-commit hooks on all files
	@echo "$(BLUE)Running pre-commit hooks on all files...$(NC)"
	@if command -v pre-commit &> /dev/null; then \
		pre-commit run --all-files; \
	else \
		echo "$(RED)pre-commit not found. Please install: make pre-commit-install$(NC)"; \
	fi

pre-commit-update: ## Update pre-commit hooks
	@echo "$(BLUE)Updating pre-commit hooks...$(NC)"
	@if command -v pre-commit &> /dev/null; then \
		pre-commit autoupdate; \
		echo "$(GREEN)Pre-commit hooks updated!$(NC)"; \
	else \
		echo "$(RED)pre-commit not found. Please install: make pre-commit-install$(NC)"; \
	fi

pre-commit-clean: ## Clean pre-commit cache
	@echo "$(BLUE)Cleaning pre-commit cache...$(NC)"
	@if command -v pre-commit &> /dev/null; then \
		pre-commit clean; \
		echo "$(GREEN)Pre-commit cache cleaned!$(NC)"; \
	else \
		echo "$(RED)pre-commit not found. Please install: make pre-commit-install$(NC)"; \
	fi

##@ Documentation

.PHONY: docs docs-serve docs-build docs-deploy godoc
docs: ## Generate Go documentation
	@echo "$(BLUE)Generating Go documentation...$(NC)"
	@mkdir -p $(DOCS_PATH)/api
	@go doc -all ./... > $(DOCS_PATH)/api/generated.md
	@echo "$(GREEN)Go documentation generated in $(DOCS_PATH)/api/generated.md$(NC)"

docs-serve: ## Serve MkDocs documentation locally
	@echo "$(BLUE)Starting MkDocs development server...$(NC)"
	@if command -v mkdocs &> /dev/null; then \
		mkdocs serve --dev-addr=127.0.0.1:8000; \
	else \
		echo "$(YELLOW)MkDocs not found. Installing...$(NC)"; \
		pip install mkdocs-material; \
		mkdocs serve --dev-addr=127.0.0.1:8000; \
	fi

docs-build: ## Build MkDocs documentation
	@echo "$(BLUE)Building MkDocs documentation...$(NC)"
	@if command -v mkdocs &> /dev/null; then \
		mkdocs build; \
	else \
		echo "$(YELLOW)MkDocs not found. Installing...$(NC)"; \
		pip install mkdocs-material; \
		mkdocs build; \
	fi

docs-deploy: ## Deploy documentation to GitHub Pages
	@echo "$(BLUE)Deploying documentation to GitHub Pages...$(NC)"
	@if command -v mkdocs &> /dev/null; then \
		mkdocs gh-deploy --force; \
	else \
		echo "$(YELLOW)MkDocs not found. Installing...$(NC)"; \
		pip install mkdocs-material; \
		mkdocs gh-deploy --force; \
	fi

godoc: ## Start Go documentation server
	@echo "$(BLUE)Starting Go documentation server...$(NC)"
	@if command -v godoc &> /dev/null; then \
		echo "$(GREEN)Go documentation available at http://localhost:6060/pkg/github.com/piotrlaczkowski/GoLangGraph/$(NC)"; \
		godoc -http=:6060; \
	else \
		echo "$(YELLOW)godoc not found. Installing...$(NC)"; \
		go install golang.org/x/tools/cmd/godoc@latest; \
		echo "$(GREEN)Go documentation available at http://localhost:6060/pkg/github.com/piotrlaczkowski/GoLangGraph/$(NC)"; \
		godoc -http=:6060; \
	fi

.PHONY: docs-generate
docs-generate: ## Generate static documentation
	@echo "$(BLUE)Generating static documentation...$(NC)"
	@mkdir -p $(DOCS_PATH)/api
	@if command -v godoc &> /dev/null; then \
		godoc -html github.com/piotrlaczkowski/golanggraph/pkg/core > $(DOCS_PATH)/api/core.html; \
		godoc -html github.com/piotrlaczkowski/golanggraph/pkg/persistence > $(DOCS_PATH)/api/persistence.html; \
		godoc -html github.com/piotrlaczkowski/golanggraph/pkg/llm > $(DOCS_PATH)/api/llm.html; \
		godoc -html github.com/piotrlaczkowski/golanggraph/pkg/agent > $(DOCS_PATH)/api/agent.html; \
		echo "$(GREEN)Documentation generated in $(DOCS_PATH)/api/$(NC)"; \
	else \
		echo "$(RED)godoc not found. Please install: go install golang.org/x/tools/cmd/godoc@latest$(NC)"; \
	fi

##@ CLI Commands

.PHONY: cli-init
cli-init: build ## Initialize new project with CLI
	@echo "$(BLUE)Initializing new project...$(NC)"
	./bin/$(BINARY_NAME) init test-project --template=basic

.PHONY: cli-dev
cli-dev: build ## Start development server
	@echo "$(BLUE)Starting development server...$(NC)"
	./bin/$(BINARY_NAME) dev --host=localhost --port=8080

.PHONY: cli-docker-build
cli-docker-build: build ## Build Docker container
	@echo "$(BLUE)Building Docker container...$(NC)"
	./bin/$(BINARY_NAME) docker build --tag=golanggraph-agent:latest

.PHONY: cli-docker-build-distroless
cli-docker-build-distroless: build ## Build distroless Docker container
	@echo "$(BLUE)Building distroless Docker container...$(NC)"
	./bin/$(BINARY_NAME) docker build --distroless --tag=golanggraph-agent:distroless

.PHONY: cli-validate
cli-validate: build ## Validate configuration
	@echo "$(BLUE)Validating configuration...$(NC)"
	./bin/$(BINARY_NAME) validate configs/agent-config.yaml

.PHONY: cli-test-all
cli-test-all: cli-init cli-validate cli-docker-build ## Test all CLI commands
	@echo "$(GREEN)All CLI commands tested successfully!$(NC)"

##@ Docker Operations

.PHONY: docker-build-agent
docker-build-agent: ## Build agent Docker image
	@echo "$(BLUE)Building agent Docker image...$(NC)"
	docker build -f Dockerfile.agent -t golanggraph-agent:latest .

.PHONY: docker-build-distroless
docker-build-distroless: ## Build distroless agent Docker image
	@echo "$(BLUE)Building distroless agent Docker image...$(NC)"
	docker build -f Dockerfile.distroless -t golanggraph-agent:distroless .

.PHONY: docker-run-agent
docker-run-agent: ## Run agent Docker container
	@echo "$(BLUE)Running agent Docker container...$(NC)"
	docker run -p 8080:8080 --name golanggraph-agent -d golanggraph-agent:latest

.PHONY: docker-stop-agent
docker-stop-agent: ## Stop agent Docker container
	@echo "$(BLUE)Stopping agent Docker container...$(NC)"
	docker stop golanggraph-agent || true
	docker rm golanggraph-agent || true

.PHONY: docker-logs-agent
docker-logs-agent: ## Show agent Docker container logs
	@echo "$(BLUE)Agent container logs:$(NC)"
	docker logs golanggraph-agent

.PHONY: docker-agent-complete
docker-agent-complete: docker-build-agent docker-run-agent ## Build and run agent container
	@echo "$(GREEN)Agent container is running on http://localhost:8080$(NC)"

##@ Examples

.PHONY: example-quick
example-quick: build ## Run quick start example
	@echo "$(BLUE)Running quick start example...$(NC)"
	go run examples/quick_start_demo.go

.PHONY: example-simple
example-simple: build ## Run simple agent example
	@echo "$(BLUE)Running simple agent example...$(NC)"
	go run examples/simple_agent.go

.PHONY: example-minimal
example-minimal: build ## Run minimal demo
	@echo "$(BLUE)Running minimal demo...$(NC)"
	go run examples/ultimate_minimal_demo.go

.PHONY: example-persistence
example-persistence: docker-up ## Run persistence example
	@echo "$(BLUE)Running persistence example...$(NC)"
	@echo "$(YELLOW)Waiting for databases to be ready...$(NC)"
	sleep 10
	go run examples/database_persistence_demo.go

.PHONY: example-ollama
example-ollama: ## Run Ollama demo with Gemma 3:1B
	@echo "$(BLUE)Building and running Ollama demo...$(NC)"
	@go build -o bin/ollama-demo ./cmd/examples
	@echo "$(GREEN)Ollama demo built successfully$(NC)"
	@echo "$(YELLOW)Running Ollama demo (requires Ollama + Gemma 3:1B)...$(NC)"
	@./bin/ollama-demo

.PHONY: test-ollama
test-ollama: ## Test Ollama integration end-to-end
	@echo "$(BLUE)Running comprehensive Ollama integration tests...$(NC)"
	@./scripts/test-ollama-demo.sh

.PHONY: test-ollama-setup
test-ollama-setup: ## Test Ollama setup only (no demo execution)
	@echo "$(BLUE)Testing Ollama setup and configuration...$(NC)"
	@./scripts/test-ollama-demo.sh check-only

##@ CI/CD

.PHONY: ci-test
ci-test: ## Run CI tests
	@echo "$(BLUE)Running CI tests...$(NC)"
	go test -v -race -coverprofile=$(COVERAGE_OUT) -covermode=atomic $(PKG_PATH)

.PHONY: ci-build
ci-build: ## Build for CI
	@echo "$(BLUE)Building for CI...$(NC)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: ci-lint
ci-lint: ## Run linter for CI
	@echo "$(BLUE)Running linter for CI...$(NC)"
	golangci-lint run --timeout=5m

##@ Release

.PHONY: build-release
build-release: ## Build release binaries for multiple platforms
	@echo "$(BLUE)Building release binaries...$(NC)"
	@mkdir -p bin/release
	GOOS=linux GOARCH=amd64 go build -o bin/release/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o bin/release/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o bin/release/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o bin/release/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "$(GREEN)Release binaries built in bin/release/$(NC)"

.PHONY: version
version: ## Show version information
	@echo "$(BLUE)Version information:$(NC)"
	@go version
	@echo "$(BLUE)GoLangGraph version:$(NC)"
	@./bin/$(BINARY_NAME) version 2>/dev/null || echo "$(YELLOW)Build the binary first with 'make build'$(NC)"

##@ All-in-one Commands

.PHONY: dev-setup
dev-setup: install ollama-setup docker-up ## Complete development setup
	@echo "$(GREEN)Development environment setup complete!$(NC)"
	@echo "$(BLUE)You can now run:$(NC)"
	@echo "  make test-local    - Run end-to-end tests"
	@echo "  make demo-local    - Run local demo"
	@echo "  make test          - Run unit tests"

.PHONY: full-test
full-test: test-short test-race test-coverage benchmark ## Run all tests
	@echo "$(GREEN)All tests completed!$(NC)"

.PHONY: check
check: fmt vet lint security test-short ## Run all checks
	@echo "$(GREEN)All checks passed!$(NC)"

.PHONY: dev-clean
dev-clean: clean docker-down ## Clean everything including Docker containers
	@echo "$(GREEN)Development environment cleaned!$(NC)"

##@ Stateful Agents System

STATEFUL_DIR=./examples/10-ideation-agents/go-agents-simple-statefull
STATEFUL_BINARY=stateful-agents
COMPOSE_FILE=docker-compose.local.yml

.PHONY: stateful-setup
stateful-setup: ## Set up the stateful agents system with database
	@echo "$(BLUE)Setting up stateful agents system...$(NC)"
	@cd $(STATEFUL_DIR) && \
	echo "$(YELLOW)1. Cleaning up existing containers...$(NC)" && \
	docker stop golanggraph-postgres golanggraph-redis 2>/dev/null || true && \
	docker rm golanggraph-postgres golanggraph-redis 2>/dev/null || true && \
	echo "$(YELLOW)2. Starting PostgreSQL and Redis...$(NC)" && \
	docker-compose -f $(COMPOSE_FILE) up -d postgres redis && \
	echo "$(YELLOW)3. Waiting for PostgreSQL to initialize...$(NC)" && \
	sleep 15 && \
	echo "$(YELLOW)4. Verifying database setup...$(NC)" && \
	docker exec golanggraph-postgres psql -U golanggraph -d golanggraph_stateful -c "SELECT 'Database ready!' as status;" && \
	echo "$(GREEN)✅ Stateful agents system setup complete!$(NC)"

.PHONY: stateful-build
stateful-build: ## Build the stateful agents application
	@echo "$(BLUE)Building stateful agents application...$(NC)"
	@cd $(STATEFUL_DIR) && \
	echo "$(YELLOW)Building $(STATEFUL_BINARY)...$(NC)" && \
	GOWORK=off go mod tidy && \
	GOWORK=off go build -o $(STATEFUL_BINARY) . && \
	echo "$(GREEN)✅ Build complete: $(STATEFUL_BINARY)$(NC)"

.PHONY: stateful-run
stateful-run: ## Run the stateful agents application
	@echo "$(BLUE)Starting stateful agents application...$(NC)"
	@cd $(STATEFUL_DIR) && \
	echo "$(YELLOW)Checking Ollama connection...$(NC)" && \
	curl -s http://localhost:11434/api/tags > /dev/null && \
	echo "$(GREEN)✅ Ollama is running$(NC)" && \
	echo "$(YELLOW)Starting application...$(NC)" && \
	OLLAMA_ENDPOINT=http://localhost:11434 \
	POSTGRES_HOST=localhost POSTGRES_PORT=5432 \
	POSTGRES_DB=golanggraph_stateful POSTGRES_USER=golanggraph \
	POSTGRES_PASSWORD=stateful_password_2024 POSTGRES_SSL_MODE=disable \
	LOG_LEVEL=info \
	./$(STATEFUL_BINARY)

.PHONY: stateful-test
stateful-test: ## Test the stateful agents system
	@echo "$(BLUE)Testing stateful agents system...$(NC)"
	@cd $(STATEFUL_DIR) && \
	echo "$(YELLOW)Running comprehensive tests...$(NC)" && \
	chmod +x test-system.sh && \
	./test-system.sh

.PHONY: stateful-dev
stateful-dev: stateful-setup stateful-build stateful-run ## Set up, build, and run the stateful agents system

.PHONY: stateful-test-full
stateful-test-full: stateful-setup stateful-build ## Full setup and testing
	@echo "$(BLUE)Running full stateful agents test...$(NC)"
	@cd $(STATEFUL_DIR) && \
	echo "$(YELLOW)Starting application in background...$(NC)" && \
	OLLAMA_ENDPOINT=http://localhost:11434 \
	POSTGRES_HOST=localhost POSTGRES_PORT=5432 \
	POSTGRES_DB=golanggraph_stateful POSTGRES_USER=golanggraph \
	POSTGRES_PASSWORD=stateful_password_2024 POSTGRES_SSL_MODE=disable \
	LOG_LEVEL=info \
	./$(STATEFUL_BINARY) > app.log 2>&1 & \
	echo $$! > app.pid && \
	sleep 5 && \
	echo "$(YELLOW)Running tests...$(NC)" && \
	chmod +x test-system.sh && \
	./test-system.sh && \
	echo "$(YELLOW)Stopping application...$(NC)" && \
	kill `cat app.pid` 2>/dev/null || true && \
	rm -f app.pid && \
	echo "$(GREEN)✅ Full test complete!$(NC)"

.PHONY: stateful-logs
stateful-logs: ## Show logs from the stateful agents system
	@echo "$(BLUE)Showing Docker logs...$(NC)"
	@cd $(STATEFUL_DIR) && \
	docker-compose -f $(COMPOSE_FILE) logs -f

.PHONY: stateful-status
stateful-status: ## Check status of the stateful agents system
	@echo "$(BLUE)Checking stateful agents system status...$(NC)"
	@cd $(STATEFUL_DIR) && \
	echo "$(YELLOW)Docker containers:$(NC)" && \
	docker ps --filter name=golanggraph --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" && \
	echo "$(YELLOW)Application health:$(NC)" && \
	curl -s http://localhost:8080/health | jq . 2>/dev/null || echo "Application not responding" && \
	echo "$(YELLOW)Database connection:$(NC)" && \
	docker exec golanggraph-postgres psql -U golanggraph -d golanggraph_stateful -c "SELECT count(*) as tables FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null || echo "Database not accessible"

.PHONY: stateful-clean
stateful-clean: ## Clean up stateful agents system
	@echo "$(BLUE)Cleaning up stateful agents system...$(NC)"
	@cd $(STATEFUL_DIR) && \
	echo "$(YELLOW)Stopping containers...$(NC)" && \
	docker-compose -f $(COMPOSE_FILE) down -v && \
	echo "$(YELLOW)Removing binary...$(NC)" && \
	rm -f $(STATEFUL_BINARY) app.log app.pid && \
	echo "$(GREEN)✅ Cleanup complete!$(NC)"

.PHONY: ollama-check
ollama-check: ## Check if Ollama is running and show available models
	@echo "$(BLUE)Checking Ollama status...$(NC)"
	@curl -s http://localhost:11434/api/tags | jq -r '.models[] | .name' 2>/dev/null && \
	echo "$(GREEN)✅ Ollama is running with models above$(NC)" || \
	echo "$(RED)❌ Ollama is not running. Please start Ollama first.$(NC)"

# Default target
.DEFAULT_GOAL := help
