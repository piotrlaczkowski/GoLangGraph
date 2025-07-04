# Contributing to GoLangGraph

Thank you for your interest in contributing to GoLangGraph! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)
- [Community](#community)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please be respectful, inclusive, and constructive in all interactions.

### Our Standards

- **Be respectful**: Treat everyone with respect and kindness
- **Be inclusive**: Welcome people of all backgrounds and experience levels
- **Be constructive**: Focus on what is best for the community
- **Be patient**: Help others learn and grow
- **Be collaborative**: Work together to achieve common goals

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Docker and Docker Compose (for integration tests)
- Make (for build automation)

### First Time Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:

   ```bash
   git clone https://github.com/YOUR_USERNAME/GoLangGraph.git
   cd GoLangGraph
   ```

3. **Add upstream remote**:

   ```bash
   git remote add upstream https://github.com/piotrlaczkowski/GoLangGraph.git
   ```

4. **Install dependencies**:

   ```bash
   go mod tidy
   ```

5. **Verify setup**:

   ```bash
   make test
   ```

## Development Setup

### Local Development Environment

```bash
# Start development services (PostgreSQL, Redis)
make dev-up

# Run tests
make test

# Run linting
make lint

# Format code
make fmt

# Build the project
make build

# Stop development services
make dev-down
```

### Environment Variables

Create a `.env` file for local development:

```bash
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=golanggraph_dev
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# LLM Provider Configuration
OPENAI_API_KEY=your_openai_api_key
GEMINI_API_KEY=your_gemini_api_key
OLLAMA_BASE_URL=http://localhost:11434
```

### Project Structure

```
├── pkg/                    # Core packages
│   ├── core/              # Graph execution engine
│   ├── agent/             # AI agent framework
│   ├── llm/               # LLM provider integrations
│   ├── persistence/       # Database and storage
│   ├── tools/             # Built-in tools
│   ├── server/            # HTTP server
│   └── debug/             # Debugging utilities
├── cmd/                   # Command-line applications
├── examples/              # Example applications
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
├── .github/               # GitHub workflows
└── Makefile              # Build automation
```

## Making Changes

### Branching Strategy

1. **Create a feature branch** from `main`:

   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the guidelines below

3. **Commit your changes** with clear messages:

   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

4. **Push to your fork**:

   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes

**Examples:**

```
feat(core): add conditional edge support
fix(persistence): resolve connection pool leak
docs(readme): update installation instructions
test(agent): add comprehensive agent tests
```

### Types of Contributions

#### 🐛 Bug Fixes

- Fix existing functionality that isn't working correctly
- Include reproduction steps in the issue
- Add tests to prevent regression

#### ✨ New Features

- Add new functionality to the project
- Discuss in an issue before implementing large features
- Include documentation and tests

#### 📚 Documentation

- Improve existing documentation
- Add new documentation for features
- Fix typos and improve clarity

#### 🧪 Testing

- Add missing tests
- Improve test coverage
- Add integration tests

#### 🏗️ Infrastructure

- Improve build process
- Update CI/CD pipelines
- Enhance development tools

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/core -v
go test ./pkg/agent -v
go test ./pkg/persistence -v

# Run tests with coverage
make test-coverage

# Run integration tests (requires running services)
make test-integration

# Run benchmarks
make benchmark

# Run race detection
make test-race
```

### Writing Tests

#### Unit Tests

- Test individual functions and methods
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Aim for high test coverage

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "hello",
            expected: "HELLO",
            wantErr:  false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("Function() = %v, expected %v", result, tt.expected)
            }
        })
    }
}
```

#### Integration Tests

- Test component interactions
- Use real databases when possible
- Clean up resources after tests

```go
//go:build integration

func TestDatabaseIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    // Test integration
    // ...
}
```

#### Benchmark Tests

- Measure performance of critical paths
- Use `testing.B` for benchmarks

```go
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Function("test input")
    }
}
```

## Code Style

### Go Code Standards

We follow standard Go conventions with some additions:

#### Formatting

- Use `gofmt` for formatting
- Use `goimports` for import organization
- Run `make fmt` before committing

#### Naming Conventions

- Use descriptive names for variables and functions
- Follow Go naming conventions (PascalCase for exported, camelCase for unexported)
- Use meaningful package names

#### Error Handling

- Always handle errors explicitly
- Use wrapped errors for context: `fmt.Errorf("operation failed: %w", err)`
- Return errors as the last return value

```go
// Good
func ProcessData(data string) (string, error) {
    if data == "" {
        return "", fmt.Errorf("data cannot be empty")
    }
    
    result, err := transform(data)
    if err != nil {
        return "", fmt.Errorf("failed to transform data: %w", err)
    }
    
    return result, nil
}
```

#### Documentation

- Add godoc comments for all exported functions, types, and packages
- Use complete sentences in comments
- Include examples for complex functions

```go
// ProcessData transforms the input data according to the specified rules.
// It returns the transformed data or an error if the transformation fails.
//
// Example:
//   result, err := ProcessData("hello world")
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println(result) // Output: HELLO WORLD
func ProcessData(data string) (string, error) {
    // ...
}
```

### Linting

We use `golangci-lint` with a comprehensive configuration:

```bash
# Run linting
make lint

# Auto-fix issues where possible
make lint-fix
```

### Code Organization

#### Package Structure

- Keep packages focused and cohesive
- Avoid circular dependencies
- Use internal packages for implementation details

#### File Organization

- Group related functionality in the same file
- Keep files reasonably sized (< 500 lines)
- Use meaningful file names

#### Interface Design

- Keep interfaces small and focused
- Define interfaces where they're used, not where they're implemented
- Use composition over inheritance

## Submitting Changes

### Pull Request Process

1. **Ensure your branch is up to date**:

   ```bash
   git checkout main
   git pull upstream main
   git checkout feature/your-feature-name
   git rebase main
   ```

2. **Run the full test suite**:

   ```bash
   make test
   make lint
   ```

3. **Create a pull request** with:
   - Clear title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - List any breaking changes

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. **Automated checks** must pass (CI/CD, linting, tests)
2. **Code review** by maintainers
3. **Address feedback** and update PR
4. **Final approval** and merge

### Merge Criteria

- All CI checks pass
- Code review approval from maintainer
- No merge conflicts
- Documentation updated if needed
- Tests added for new functionality

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH`
- Major: Breaking changes
- Minor: New features (backward compatible)
- Patch: Bug fixes (backward compatible)

### Release Workflow

1. **Prepare release**:
   - Update version numbers
   - Update CHANGELOG.md
   - Create release branch

2. **Test release**:
   - Run full test suite
   - Test examples and documentation
   - Verify builds on all platforms

3. **Create release**:
   - Tag the release
   - Build and publish artifacts
   - Update documentation

4. **Announce release**:
   - GitHub release notes
   - Community announcements

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and discussions
- **Pull Requests**: Code contributions and reviews

### Getting Help

- Check existing issues and documentation
- Search previous discussions
- Create a new issue with detailed information
- Be patient and respectful when asking for help

### Reporting Issues

When reporting bugs, please include:

1. **Environment information**:
   - Go version
   - Operating system
   - GoLangGraph version

2. **Reproduction steps**:
   - Minimal code example
   - Expected behavior
   - Actual behavior

3. **Additional context**:
   - Error messages
   - Logs
   - Screenshots if applicable

### Feature Requests

When requesting features:

1. **Describe the problem** you're trying to solve
2. **Explain the proposed solution**
3. **Consider alternatives** you've thought of
4. **Provide use cases** and examples

## Recognition

We appreciate all contributions! Contributors will be:

- Listed in the project's contributors
- Mentioned in release notes for significant contributions
- Invited to join the maintainer team for sustained contributions

## Questions?

If you have any questions about contributing, please:

1. Check this document first
2. Search existing issues and discussions
3. Create a new discussion or issue
4. Reach out to maintainers

Thank you for contributing to GoLangGraph! 🚀
