#!/bin/bash

# GoLangGraph CI/CD Setup Script
# This script sets up the complete CI/CD pipeline including documentation and pre-commit hooks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository. Please run this script from the project root."
        exit 1
    fi
    
    print_status "Git repository detected"
}

# Setup GitHub repository settings
setup_github_settings() {
    print_status "Setting up GitHub repository settings..."
    
    # Check if GitHub CLI is available
    if command -v gh &> /dev/null; then
        print_status "GitHub CLI detected. Configuring repository settings..."
        
        # Enable GitHub Pages
        gh api repos/:owner/:repo --method PATCH --field has_pages=true || print_warning "Could not enable GitHub Pages via API"
        
        # Set up branch protection (optional)
        read -p "Do you want to set up branch protection for main branch? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            gh api repos/:owner/:repo/branches/main/protection \
                --method PUT \
                --field required_status_checks='{"strict":true,"contexts":["CI"]}' \
                --field enforce_admins=true \
                --field required_pull_request_reviews='{"required_approving_review_count":1}' \
                --field restrictions=null || print_warning "Could not set up branch protection"
        fi
    else
        print_warning "GitHub CLI not found. Please manually configure:"
        echo "  1. Enable GitHub Pages in repository settings"
        echo "  2. Set Pages source to 'GitHub Actions'"
        echo "  3. Optionally enable branch protection for main branch"
    fi
}

# Initialize secrets baseline for detect-secrets
setup_secrets_baseline() {
    print_status "Setting up secrets detection baseline..."
    
    if ! command -v detect-secrets &> /dev/null; then
        print_status "Installing detect-secrets..."
        pip install detect-secrets || {
            print_warning "Could not install detect-secrets. Please install manually: pip install detect-secrets"
            return
        }
    fi
    
    # Create secrets baseline
    detect-secrets scan --baseline .secrets.baseline --exclude-files '\.git/.*|\.secrets\.baseline|go\.sum|go\.mod|docs/.*\.md'
    print_success "Secrets baseline created at .secrets.baseline"
}

# Setup pre-commit hooks
setup_precommit() {
    print_status "Setting up pre-commit hooks..."
    
    # Run the documentation setup script first
    if [ -f "scripts/setup-docs.sh" ]; then
        print_status "Running documentation setup..."
        chmod +x scripts/setup-docs.sh
        ./scripts/setup-docs.sh
    else
        print_warning "Documentation setup script not found. Setting up pre-commit manually..."
        
        # Install pre-commit if not available
        if ! command -v pre-commit &> /dev/null; then
            print_status "Installing pre-commit..."
            pip install pre-commit || {
                print_error "Could not install pre-commit. Please install manually: pip install pre-commit"
                exit 1
            }
        fi
        
        # Install hooks
        pre-commit install
        pre-commit install --hook-type commit-msg
    fi
    
    print_success "Pre-commit hooks installed!"
}

# Test CI workflows locally (if act is available)
test_workflows() {
    print_status "Testing workflows..."
    
    if command -v act &> /dev/null; then
        print_status "act detected. Running workflow tests..."
        
        # Test pre-commit workflow
        act -W .github/workflows/pre-commit.yml --dryrun || print_warning "Pre-commit workflow test failed"
        
        # Test documentation workflow
        act -W .github/workflows/docs.yml --dryrun || print_warning "Documentation workflow test failed"
        
        print_success "Workflow tests completed!"
    else
        print_warning "act not found. Install act to test workflows locally: https://github.com/nektos/act"
        print_status "Workflows will be tested when pushed to GitHub"
    fi
}

# Generate GitHub repository configuration
generate_repo_config() {
    print_status "Generating repository configuration..."
    
    cat > .github/settings.yml << 'EOF'
# Repository settings for GoLangGraph
repository:
  # Repository basics
  name: GoLangGraph
  description: A powerful Go framework for building AI agent workflows with graph-based execution
  homepage: https://golanggraph.dev
  topics:
    - golang
    - ai
    - agents
    - graph
    - workflow
    - llm
    - rag
    - framework
  
  # Repository features
  has_issues: true
  has_projects: true
  has_wiki: true
  has_pages: true
  has_downloads: true
  
  # Repository settings
  default_branch: main
  allow_squash_merge: true
  allow_merge_commit: true
  allow_rebase_merge: true
  delete_branch_on_merge: true
  
  # Security settings
  enable_automated_security_fixes: true
  enable_vulnerability_alerts: true

# Branch protection rules
branches:
  - name: main
    protection:
      required_status_checks:
        strict: true
        contexts:
          - "CI"
          - "Pre-commit"
          - "Documentation"
      enforce_admins: true
      required_pull_request_reviews:
        required_approving_review_count: 1
        dismiss_stale_reviews: true
        require_code_owner_reviews: true
      restrictions: null

# Labels for issues and PRs
labels:
  - name: "bug"
    color: "d73a4a"
    description: "Something isn't working"
  
  - name: "enhancement"
    color: "a2eeef"
    description: "New feature or request"
  
  - name: "documentation"
    color: "0075ca"
    description: "Improvements or additions to documentation"
  
  - name: "good first issue"
    color: "7057ff"
    description: "Good for newcomers"
  
  - name: "help wanted"
    color: "008672"
    description: "Extra attention is needed"
  
  - name: "dependencies"
    color: "0366d6"
    description: "Pull requests that update a dependency file"
  
  - name: "ci/cd"
    color: "f9d0c4"
    description: "Related to CI/CD pipeline"
EOF
    
    print_success "Repository configuration generated at .github/settings.yml"
}

# Create development documentation
create_dev_docs() {
    print_status "Creating development documentation..."
    
    mkdir -p docs/development
    
    cat > docs/development/ci-cd.md << 'EOF'
# CI/CD Pipeline

This document describes the CI/CD pipeline for GoLangGraph.

## Overview

The CI/CD pipeline consists of several workflows:

### 1. CI Workflow (`.github/workflows/ci.yml`)
- **Trigger**: Push/PR to main/develop branches
- **Jobs**: Test, Lint, Security, Build, Integration Tests, Benchmarks
- **Services**: PostgreSQL, Redis for integration tests

### 2. Pre-commit Workflow (`.github/workflows/pre-commit.yml`)
- **Trigger**: Push/PR to main/develop branches
- **Jobs**: Pre-commit hooks, Security scan, Dependency check, Code quality
- **Purpose**: Enforce code quality standards

### 3. Documentation Workflow (`.github/workflows/docs.yml`)
- **Trigger**: Changes to docs, Go files, or workflow
- **Jobs**: Build MkDocs, Deploy to GitHub Pages, Test links, Generate Go docs
- **Output**: Documentation site at GitHub Pages

### 4. Release Workflow (`.github/workflows/release.yml`)
- **Trigger**: Git tags or manual dispatch
- **Jobs**: Validate, Build binaries, Docker image, Documentation, Release
- **Artifacts**: Multi-platform binaries, Docker images, Documentation

## Local Development

### Setup
```bash
# Install all development dependencies
make dev-setup

# Install pre-commit hooks
make pre-commit-install

# Start documentation server
make docs-serve
```

### Testing
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run pre-commit checks
make pre-commit-run
```

### Documentation
```bash
# Serve documentation locally
make docs-serve

# Build documentation
make docs-build

# Generate Go documentation
make godoc
```

## Branch Strategy

- **main**: Production-ready code, protected branch
- **develop**: Development branch for feature integration
- **feature/***: Feature branches, merged to develop
- **hotfix/***: Urgent fixes, merged to main and develop

## Release Process

1. Create release tag: `git tag v1.0.0`
2. Push tag: `git push origin v1.0.0`
3. GitHub Actions automatically:
   - Runs all tests and validations
   - Builds multi-platform binaries
   - Creates Docker images
   - Deploys documentation
   - Creates GitHub release

## Security

- **Gosec**: Security vulnerability scanning
- **Trivy**: Container and filesystem scanning
- **detect-secrets**: Secret detection in code
- **Dependabot**: Automated dependency updates
- **SARIF**: Security findings uploaded to GitHub Security tab

## Quality Gates

All PRs must pass:
- Unit tests with coverage > 80%
- Integration tests
- Linting (golangci-lint)
- Security scans
- Pre-commit hooks
- Documentation builds

## Monitoring

- **Codecov**: Test coverage reporting
- **GitHub Security**: Vulnerability alerts
- **Dependabot**: Dependency monitoring
- **Workflow status**: CI/CD pipeline health
EOF

    print_success "Development documentation created"
}

# Main function
main() {
    print_status "Setting up CI/CD pipeline for GoLangGraph..."
    
    # Check prerequisites
    check_git_repo
    
    # Setup components
    setup_secrets_baseline
    setup_precommit
    generate_repo_config
    create_dev_docs
    
    # Optional GitHub setup
    read -p "Do you want to configure GitHub repository settings? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        setup_github_settings
    fi
    
    # Test workflows if possible
    test_workflows
    
    print_success "CI/CD pipeline setup complete!"
    print_status ""
    print_status "Next steps:"
    echo "  1. Commit and push the changes to GitHub"
    echo "  2. Check that workflows run successfully"
    echo "  3. Configure GitHub Pages source to 'GitHub Actions'"
    echo "  4. Review and adjust branch protection rules"
    echo "  5. Set up any required secrets (DOCKER_TOKEN, etc.)"
    print_status ""
    print_status "Available commands:"
    echo "  make dev-setup       # Complete development setup"
    echo "  make pre-commit-run  # Run pre-commit hooks"
    echo "  make docs-serve      # Start documentation server"
    echo "  make test            # Run tests"
    echo "  make ci-test         # Run CI tests locally"
}

# Run main function
main "$@" 