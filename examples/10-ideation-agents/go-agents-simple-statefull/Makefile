# Makefile for GoLangGraph Go-Agents-Simple Example
# Comprehensive build, test, and deployment automation

# Configuration
APP_NAME := go-agents-simple
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DOCKER_IMAGE := golanggraph/$(APP_NAME)
DOCKER_TAG := $(VERSION)
PORT := 8080
OLLAMA_PORT := 11434

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: clean deps build test

# Help target
.PHONY: help
help:
	@echo "🚀 GoLangGraph Go-Agents-Simple Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  📦 Build & Development:"
	@echo "    build         - Build the application"
	@echo "    clean         - Clean build artifacts"
	@echo "    deps          - Download dependencies"
	@echo "    tidy          - Tidy dependencies"
	@echo "    fmt           - Format code"
	@echo "    vet           - Run go vet"
	@echo "    lint          - Run linters"
	@echo ""
	@echo "  🧪 Testing:"
	@echo "    test          - Run tests"
	@echo "    test-verbose  - Run tests with verbose output"
	@echo "    test-coverage - Run tests with coverage"
	@echo "    benchmark     - Run benchmarks"
	@echo ""
	@echo "  🏃 Running:"
	@echo "    run           - Run the application"
	@echo "    run-dev       - Run in development mode"
	@echo "    start-ollama  - Start Ollama service"
	@echo "    stop-ollama   - Stop Ollama service"
	@echo ""
	@echo "  🎮 Quick Play Commands:"
	@echo "    play          - Build and run locally (background)"
	@echo "    play-docker   - Build and run with Docker"
	@echo "    dev           - Clean, build, and run locally"
	@echo "    dev-docker    - Clean, build, and run with Docker"
	@echo "    stop-docker   - Stop Docker container"
	@echo "    quick-test    - Quick functionality test with Docker"
	@echo ""
	@echo "  🐳 Docker:"
	@echo "    docker-build  - Build Docker image"
	@echo "    docker-run    - Run Docker container"
	@echo "    docker-stop   - Stop Docker container"
	@echo "    docker-push   - Push Docker image"
	@echo "    docker-clean  - Clean Docker images"
	@echo ""
	@echo "  🚢 Deployment:"
	@echo "    deploy-local  - Deploy minimal (app only)"
	@echo "    deploy-full   - Deploy with Redis and monitoring"
	@echo "    deploy-monitoring - Deploy with monitoring only"
	@echo "    deploy-stop   - Stop docker-compose deployment"
	@echo "    deploy-logs   - Show deployment logs"
	@echo ""
	@echo "  ✅ Validation:"
	@echo "    health-check  - Check application health"
	@echo "    test-endpoints- Test all API endpoints"
	@echo "    test-monitoring- Test monitoring stack (Prometheus + Grafana)"
	@echo "    smoke-test    - Run smoke tests"
	@echo ""
	@echo "  🧹 Maintenance:"
	@echo "    update        - Update dependencies"
	@echo "    security      - Run security checks"
	@echo "    audit         - Audit dependencies"

# Build targets
.PHONY: build
build: deps
	@echo "🔨 Building $(APP_NAME)..."
	@mkdir -p bin/
	GOWORK=off $(GOBUILD) $(LDFLAGS) -o bin/$(APP_NAME) .
	@echo "✅ Build complete: bin/$(APP_NAME)"

.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out
	rm -f coverage.html
	@echo "✅ Clean complete"

.PHONY: deps
deps:
	@echo "📦 Downloading dependencies..."
	$(GOGET) -d ./...
	@echo "✅ Dependencies downloaded"

.PHONY: tidy
tidy:
	@echo "🔧 Tidying dependencies..."
	$(GOMOD) tidy
	@echo "✅ Dependencies tidied"

.PHONY: fmt
fmt:
	@echo "✨ Formatting code..."
	$(GOFMT) -w .
	@echo "✅ Code formatted"

.PHONY: vet
vet:
	@echo "🔍 Running go vet..."
	$(GOVET) ./...
	@echo "✅ Go vet complete"

.PHONY: lint
lint: fmt vet
	@echo "🔍 Running additional linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, skipping..."; \
	fi
	@echo "✅ Linting complete"

# Testing targets
.PHONY: test
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...
	@echo "✅ Tests complete"

.PHONY: test-verbose
test-verbose:
	@echo "🧪 Running tests (verbose)..."
	$(GOTEST) -v -race ./...
	@echo "✅ Verbose tests complete"

.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

.PHONY: benchmark
benchmark:
	@echo "⚡ Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...
	@echo "✅ Benchmarks complete"

# Running targets
.PHONY: run
run: build
	@echo "🚀 Starting $(APP_NAME)..."
	./bin/$(APP_NAME)

.PHONY: run-dev
run-dev: deps
	@echo "🚀 Starting $(APP_NAME) in development mode..."
	GOWORK=off $(GOCMD) run .

.PHONY: start-ollama
start-ollama:
	@echo "🤖 Starting Ollama..."
	@if ! pgrep ollama >/dev/null; then \
		ollama serve & \
		sleep 3; \
		echo "✅ Ollama started"; \
	else \
		echo "ℹ️  Ollama already running"; \
	fi
	@echo "📥 Pulling required models..."
	ollama pull llama3.2 || echo "⚠️  Failed to pull llama3.2"
	ollama pull gemma3:1b || echo "⚠️  Failed to pull gemma3:1b"

.PHONY: stop-ollama
stop-ollama:
	@echo "🛑 Stopping Ollama..."
	@pkill ollama || echo "ℹ️  Ollama not running"
	@echo "✅ Ollama stopped"

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	cd ../../.. && docker build -f examples/10-ideation-agents/go-agents-simple/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

.PHONY: docker-run
docker-run: docker-build
	@echo "🐳 Running Docker container..."
	docker run -d --name $(APP_NAME) \
		-p $(PORT):8080 \
		--restart unless-stopped \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ Docker container running on http://localhost:$(PORT)"

.PHONY: docker-stop
docker-stop:
	@echo "🛑 Stopping Docker container..."
	docker stop $(APP_NAME) || echo "Container not running"
	docker rm $(APP_NAME) || echo "Container not found"
	@echo "✅ Docker container stopped"

.PHONY: docker-push
docker-push: docker-build
	@echo "📤 Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	@echo "✅ Docker image pushed"

.PHONY: docker-clean
docker-clean:
	@echo "🧹 Cleaning Docker images..."
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	docker rmi $(DOCKER_IMAGE):latest || true
	docker system prune -f
	@echo "✅ Docker cleanup complete"

# Deployment targets
.PHONY: deploy-local
deploy-local: docker-compose.yml
	@echo "🚢 Deploying locally with docker-compose (minimal)..."
	docker-compose up -d
	@echo "✅ Local deployment running"
	@echo "🌐 Web UI: http://localhost:$(PORT)/"
	@echo "🛝 API Playground: http://localhost:$(PORT)/playground"
	@echo "❤️  Health: http://localhost:$(PORT)/health"

.PHONY: deploy-full
deploy-full: docker-compose.yml
	@echo "🚢 Deploying with all services (Redis, Monitoring)..."
	docker-compose --profile full --profile monitoring up -d
	@echo "✅ Full deployment running"
	@echo "🌐 Web UI: http://localhost:$(PORT)/"
	@echo "🛝 API Playground: http://localhost:$(PORT)/playground"
	@echo "❤️  Health: http://localhost:$(PORT)/health"
	@echo "📊 Prometheus: http://localhost:9091/"
	@echo "📈 Grafana: http://localhost:3001/ (admin/admin)"
	@echo "🗄️  Redis: localhost:6380"

.PHONY: deploy-monitoring
deploy-monitoring: docker-compose.yml
	@echo "🚢 Deploying with monitoring stack..."
	docker-compose --profile monitoring up -d
	@echo "✅ Monitoring deployment running"
	@echo "🌐 Web UI: http://localhost:$(PORT)/"
	@echo "📊 Prometheus: http://localhost:9091/"
	@echo "📈 Grafana: http://localhost:3001/ (admin/admin)"

.PHONY: deploy-stop
deploy-stop:
	@echo "🛑 Stopping local deployment..."
	docker-compose down
	@echo "✅ Local deployment stopped"

.PHONY: deploy-logs
deploy-logs:
	@echo "📋 Showing deployment logs..."
	docker-compose logs -f

# Validation targets
.PHONY: health-check
health-check:
	@echo "❤️  Checking application health..."
	@curl -f http://localhost:$(PORT)/health || echo "❌ Health check failed"
	@echo "✅ Health check complete"

.PHONY: test-endpoints
test-endpoints:
	@echo "🧪 Testing API endpoints..."
	@./scripts/test_endpoints.sh || echo "❌ Endpoint tests failed"
	@echo "✅ Endpoint tests complete"

.PHONY: test-monitoring
test-monitoring:
	@echo "🔍 Testing monitoring stack..."
	@./scripts/test_monitoring.sh || echo "❌ Monitoring tests failed"
	@echo "✅ Monitoring tests complete"

.PHONY: smoke-test
smoke-test: health-check
	@echo "💨 Running smoke tests..."
	@curl -s http://localhost:$(PORT)/agents | jq . || echo "❌ Agents endpoint failed"
	@curl -s http://localhost:$(PORT)/capabilities | jq . || echo "❌ Capabilities endpoint failed"
	@echo "✅ Smoke tests complete"

# Maintenance targets
.PHONY: update
update:
	@echo "🔄 Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

.PHONY: security
security:
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed, skipping..."; \
	fi
	@echo "✅ Security check complete"

.PHONY: audit
audit:
	@echo "🔍 Auditing dependencies..."
	$(GOMOD) download
	@if command -v nancy >/dev/null 2>&1; then \
		go list -json -m all | nancy sleuth; \
	else \
		echo "⚠️  nancy not installed, skipping vulnerability scan..."; \
	fi
	@echo "✅ Audit complete"

# Special targets
.PHONY: install-tools
install-tools:
	@echo "🛠️  Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/nancy@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✅ Development tools installed"

# Quick development commands
.PHONY: dev
dev: clean build run
	@echo "🎉 Development cycle complete!"

.PHONY: dev-docker
dev-docker: clean docker-build docker-run
	@echo "🎉 Docker development cycle complete!"

.PHONY: play
play: clean build
	@echo "🎮 Starting play mode..."
	@echo "🌐 Building and running application..."
	./bin/$(APP_NAME) &
	@sleep 3
	@echo ""
	@echo "🎉 Ready to play!"
	@echo "🌐 Web UI: http://localhost:8080/"
	@echo "🛝 API Playground: http://localhost:8080/playground"
	@echo "❤️  Health: http://localhost:8080/health"
	@echo ""
	@echo "💡 Test commands:"
	@echo "   curl http://localhost:8080/health"
	@echo "   curl -X POST http://localhost:8080/api/designer -H 'Content-Type: application/json' -d '{\"message\": \"Design a treehouse\"}'"
	@echo ""
	@echo "🛑 To stop: pkill $(APP_NAME)"

.PHONY: play-docker
play-docker: docker-build
	@echo "🎮 Starting Docker play mode..."
	@docker stop $(APP_NAME) 2>/dev/null || true
	@docker rm $(APP_NAME) 2>/dev/null || true
	@docker run --rm -d --name $(APP_NAME) \
		-p $(PORT):8080 \
		-e OLLAMA_ENDPOINT=http://host.docker.internal:11434 \
		--add-host host.docker.internal:host-gateway \
		$(DOCKER_IMAGE):$(DOCKER_TAG)
	@sleep 5
	@echo ""
	@echo "🎉 Docker ready to play!"
	@echo "🌐 Web UI: http://localhost:$(PORT)/"
	@echo "🛝 API Playground: http://localhost:$(PORT)/playground"
	@echo "❤️  Health: http://localhost:$(PORT)/health"
	@echo ""
	@echo "💡 Test commands:"
	@echo "   curl http://localhost:$(PORT)/health"
	@echo "   curl -X POST http://localhost:$(PORT)/api/designer -H 'Content-Type: application/json' -d '{\"message\": \"Design a treehouse\"}'"
	@echo ""
	@echo "🛑 To stop: make stop-docker"

.PHONY: stop-docker
stop-docker:
	@echo "🛑 Stopping Docker play mode..."
	@docker stop $(APP_NAME) 2>/dev/null || echo "Container not running"
	@echo "✅ Docker stopped"

.PHONY: quick-test
quick-test: play-docker
	@echo "🧪 Running quick functionality test..."
	@sleep 2
	@curl -f http://localhost:$(PORT)/health > /dev/null && echo "✅ Health check passed" || echo "❌ Health check failed"
	@curl -f -X POST http://localhost:$(PORT)/api/designer -H 'Content-Type: application/json' -d '{"message": "Quick test"}' > /dev/null && echo "✅ Agent test passed" || echo "❌ Agent test failed"
	@echo "🎉 Quick test complete!"

.PHONY: quick-start
quick-start: start-ollama build run

.PHONY: full-test
full-test: clean deps lint test test-coverage benchmark
	@echo "🎉 Full test suite complete!"

.PHONY: release
release: clean deps lint test docker-build
	@echo "🚀 Release $(VERSION) ready!"
	@echo "📦 Binary: bin/$(APP_NAME)"
	@echo "🐳 Docker: $(DOCKER_IMAGE):$(DOCKER_TAG)"
