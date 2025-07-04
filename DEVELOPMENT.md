# 🛠️ Development Guide

Welcome to GoLangGraph development! This guide will help you set up your development environment and contribute effectively.

## 📋 Prerequisites

- **Go**: Version 1.23 or later
- **Docker**: For running integration tests
- **Make**: For build automation
- **Git**: For version control

### Optional Tools

- **golangci-lint**: For code linting
- **Ollama**: For local LLM testing
- **pre-commit**: For git hooks

## 🚀 Quick Setup

1. **Clone the repository**:

   ```bash
   git clone https://github.com/piotrlaczkowski/GoLangGraph.git
   cd GoLangGraph
   ```

2. **Initialize Go workspace**:

   ```bash
   go work init
   go work use .
   go work use ./examples/*
   ```

3. **Install dependencies**:

   ```bash
   make install
   ```

4. **Run tests**:

   ```bash
   make test
   ```

5. **Build the project**:

   ```bash
   make build
   ```

## 🏗️ Project Structure

```
GoLangGraph/
├── 📁 pkg/              # Core library packages
│   ├── agent/           # AI agent framework
│   ├── core/            # Graph execution engine
│   ├── llm/             # LLM provider integrations
│   ├── persistence/     # Database and storage
│   └── tools/           # Built-in tools
├── 📁 cmd/              # CLI applications
├── 📁 examples/         # Usage examples (each with own go.mod)
├── 📁 docs/             # Documentation source
├── 📁 test/             # Integration tests
└── 📁 scripts/          # Build and utility scripts
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

# Test with local Ollama
make test-local
```

### Example Tests

```bash
# Test all examples
make test-examples
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

### Start Services

```bash
# Start PostgreSQL and Redis
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

### Local Ollama Setup

```bash
# Setup Ollama with required models
make ollama-setup

# Start Ollama service
make ollama-start
```

## 📦 Adding Dependencies

1. **Add to main module**:

   ```bash
   go get github.com/new/dependency
   go mod tidy
   ```

2. **Update workspace**:

   ```bash
   go work sync
   ```

3. **Test compatibility**:

   ```bash
   make test-all
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

### Generate Documentation

```bash
# Start documentation server
make docs-serve

# Build documentation
make docs-build

# Deploy documentation
make docs-deploy
```

### API Documentation

```bash
# Generate Go docs
godoc -http=:6060
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

## ❓ Troubleshooting

### Common Issues

**Go workspace issues**:

```bash
# Reinitialize workspace
rm go.work
make workspace-init
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

## 📞 Getting Help

- 💬 **Discussions**: GitHub Discussions
- 🐛 **Issues**: GitHub Issues
- 📧 **Email**: [dev@golanggraph.dev]
- 📖 **Docs**: [docs.golanggraph.dev]

## 🤝 Contributing Guidelines

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for detailed contribution guidelines.

Happy coding! 🚀
