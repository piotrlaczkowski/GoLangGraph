# 🛠️ Development Guide

Welcome to GoLangGraph development! This guide will help you set up your development environment and contribute effectively.

## 📋 Prerequisites

- **Go**: Version 1.21 or later
- **Docker**: For running integration tests with PostgreSQL and Redis
- **Make**: For build automation
- **Git**: For version control

### Optional Tools

- **golangci-lint**: For code linting (will be installed automatically)
- **Ollama**: For local LLM testing with Gemma models
- **pre-commit**: For git hooks

## 🚀 Quick Setup

1. **Clone the repository**:

   ```bash
   git clone https://github.com/piotrlaczkowski/GoLangGraph.git
   cd GoLangGraph
   ```

2. **Install dependencies**:

   ```bash
   make install
   ```

3. **Run tests**:

   ```bash
   make test
   ```

4. **Build the project**:

   ```bash
   make build
   ```

5. **Set up local development environment**:

   ```bash
   # Start PostgreSQL and Redis containers
   make docker-up
   
   # Set up Ollama with required models
   make ollama-setup
   
   # Run comprehensive tests
   make test-local
   ```

## 🏗️ Project Structure

```
GoLangGraph/
├── 📁 pkg/              # Core library packages
│   ├── agent/           # AI agent implementations (Chat, ReAct, Tool)
│   ├── builder/         # Quick builder patterns for rapid development
│   ├── core/            # Graph execution engine and state management
│   ├── debug/           # Debugging and visualization tools
│   ├── llm/             # LLM provider integrations (OpenAI, Ollama, Gemini)
│   ├── persistence/     # Database integration and checkpointing
│   ├── server/          # HTTP server and WebSocket support
│   └── tools/           # Built-in tools and tool registry
├── 📁 cmd/              # CLI applications
│   ├── golanggraph/     # Main CLI application
│   └── examples/        # Example CLI tools
├── 📁 examples/         # Working examples (each with own go.mod)
│   ├── 01-basic-chat/
│   ├── 02-react-agent/
│   ├── 03-multi-agent/
│   ├── 04-rag-system/
│   ├── 05-streaming/
│   ├── 06-persistence/
│   ├── 07-tools-integration/
│   ├── 08-production-ready/
│   └── 09-workflow-graph/
├── 📁 docs/             # Documentation source
├── 📁 test/             # Integration tests
├── 📁 scripts/          # Build and utility scripts
├── 📁 .github/          # GitHub Actions workflows
├── go.work              # Go workspace configuration
├── Makefile             # Build automation
└── README.md            # Project overview
```

## 🧪 Testing Strategy

### Unit Tests

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run short tests (skip integration)
make test-short

# Run with race detector
make test-race
```

### Integration Tests

```bash
# Start dependencies and run integration tests
make test-integration

# Run comprehensive local tests with all services
make test-local
```

### Example Tests

```bash
# Test all examples
make test-examples

# Run local demo with all services
make demo-local
```

### Enhanced Testing

```bash
# Run enhanced test suite including CLI tests
make test-enhanced

# Run benchmarks
make benchmark
```

## 🔧 Development Workflow

### 1. **Branch Naming Convention**

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring

### 2. **Commit Messages**

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

feat(agent): add new tool integration
fix(core): resolve graph execution deadlock
docs(readme): update installation instructions
refactor(llm): simplify provider interface
```

### 3. **Code Quality**

```bash
# Run linter
make lint

# Format code
make fmt

# Security scan
make security-scan

# Run all quality checks
make quality
```

### 4. **Pre-commit Hooks**

Install pre-commit hooks:

```bash
pip install pre-commit
pre-commit install
```

## 🐳 Docker Development

### Database Services

```bash
# Start PostgreSQL and Redis containers
make docker-up

# View container logs
make docker-logs

# Stop and remove containers
make docker-down
```

### Local LLM Setup

```bash
# Install and setup Ollama with gemma2:2b model
make ollama-setup

# Start Ollama service
make ollama-start
```

## 📦 Go Workspace Management

The project uses Go workspaces to manage multiple modules:

```bash
# The workspace is already configured in go.work
# It includes the main module and all examples

# To sync workspace
go work sync

# To add a new example
go work use ./examples/new-example
```

## 🔍 Building and Running

### Build Commands

```bash
# Build main binary
make build

# Build all binaries and examples
make build-all

# Run the main application
make run

# Clean build artifacts
make clean
```

### Running Examples

```bash
# Examples are in the workspace, run them directly
cd examples/01-basic-chat
go run main.go

# Or run from root
go run ./examples/01-basic-chat/main.go
```

## 🔍 Debugging

### Enable Debug Logging

```bash
export LOG_LEVEL=debug
export GOLANGGRAPH_DEBUG=true
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.

# Memory profiling
go test -memprofile=mem.prof -bench=.

# View profiles
go tool pprof cpu.prof
```

## 📚 Documentation

### Local Documentation

```bash
# Generate Go docs
godoc -http=:6060

# View at http://localhost:6060/pkg/github.com/piotrlaczkowski/GoLangGraph/
```

### MkDocs Documentation

```bash
# Install dependencies
pip install -r requirements.txt

# Serve documentation locally
mkdocs serve

# Build documentation
mkdocs build
```

## 🚢 Release Process

1. **Update version**:

   ```bash
   git tag v1.x.x
   ```

2. **Push tag**:

   ```bash
   git push origin v1.x.x
   ```

3. **GitHub Actions** will automatically:
   - Run tests
   - Build binaries
   - Create release
   - Deploy documentation

## 🧰 Available Make Targets

Run `make help` to see all available targets:

```bash
make help
```

Key targets include:

- **Development**: `install`, `build`, `run`, `clean`
- **Testing**: `test`, `test-coverage`, `test-integration`, `test-local`
- **Quality**: `lint`, `fmt`, `security-scan`, `quality`
- **Docker**: `docker-up`, `docker-down`, `docker-logs`
- **Ollama**: `ollama-setup`, `ollama-start`, `demo-local`

## ❓ Troubleshooting

### Common Issues

**Go workspace issues**:

```bash
# Reinitialize workspace
go work sync
```

**Docker permission issues**:

```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
```

**Module resolution issues**:

```bash
# Clean and rebuild
make clean
go clean -modcache
make install
```

**Ollama connection issues**:

```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Restart Ollama
make ollama-start
```

**Test failures**:

```bash
# Run tests with verbose output
go test -v ./pkg/...

# Run specific test
go test -v -run TestSpecificFunction ./pkg/core
```

## 📝 Adding New Features

### 1. **Adding a New Agent Type**

1. Create agent implementation in `pkg/agent/`
2. Add tests in `pkg/agent/`
3. Update documentation
4. Add example in `examples/`

### 2. **Adding a New Tool**

1. Implement tool interface in `pkg/tools/`
2. Register in tool registry
3. Add tests
4. Update documentation

### 3. **Adding a New LLM Provider**

1. Implement provider interface in `pkg/llm/`
2. Add configuration options
3. Add tests with mocked responses
4. Update documentation

## 📞 Getting Help

- 💬 **Discussions**: [GitHub Discussions](https://github.com/piotrlaczkowski/GoLangGraph/discussions)
- 🐛 **Issues**: [GitHub Issues](https://github.com/piotrlaczkowski/GoLangGraph/issues)
- 📧 **Email**: Support via GitHub Issues
- 📖 **Docs**: Browse the `/docs` directory

## 🤝 Contributing Guidelines

Please read [CONTRIBUTING.md](https://github.com/piotrlaczkowski/GoLangGraph/blob/main/CONTRIBUTING.md) for detailed contribution guidelines.

### Development Checklist

Before submitting a PR:

- [ ] Code follows Go conventions
- [ ] Tests pass: `make test`
- [ ] Linting passes: `make lint`
- [ ] Documentation updated
- [ ] Examples work with changes
- [ ] Integration tests pass: `make test-integration`

Happy coding! 🚀
