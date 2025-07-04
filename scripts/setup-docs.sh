#!/bin/bash

# GoLangGraph Documentation Setup Script
# This script sets up the documentation environment for GoLangGraph

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

# Check if Python is installed
check_python() {
    if command -v python3 &> /dev/null; then
        PYTHON_CMD="python3"
    elif command -v python &> /dev/null; then
        PYTHON_CMD="python"
    else
        print_error "Python is not installed. Please install Python 3.8 or later."
        exit 1
    fi

    # Check Python version
    PYTHON_VERSION=$($PYTHON_CMD --version 2>&1 | cut -d' ' -f2)
    print_status "Using Python $PYTHON_VERSION"
}

# Check if pip is installed
check_pip() {
    if command -v pip3 &> /dev/null; then
        PIP_CMD="pip3"
    elif command -v pip &> /dev/null; then
        PIP_CMD="pip"
    else
        print_error "pip is not installed. Please install pip."
        exit 1
    fi

    print_status "Using pip from $PIP_CMD"
}

# Install Python dependencies
install_dependencies() {
    print_status "Installing Python dependencies..."

    # Create virtual environment if it doesn't exist
    if [ ! -d "venv" ]; then
        print_status "Creating virtual environment..."
        $PYTHON_CMD -m venv venv
    fi

    # Activate virtual environment
    source venv/bin/activate

    # Upgrade pip
    $PIP_CMD install --upgrade pip

    # Install requirements
    if [ -f "requirements.txt" ]; then
        $PIP_CMD install -r requirements.txt
        print_success "Python dependencies installed!"
    else
        print_warning "requirements.txt not found. Installing basic dependencies..."
        $PIP_CMD install mkdocs-material pre-commit
    fi
}

# Install Go dependencies
install_go_dependencies() {
    print_status "Installing Go dependencies..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Install Go documentation tools
    go install golang.org/x/tools/cmd/godoc@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

    print_success "Go dependencies installed!"
}

# Setup pre-commit hooks
setup_precommit() {
    print_status "Setting up pre-commit hooks..."

    # Activate virtual environment
    source venv/bin/activate

    # Install pre-commit hooks
    pre-commit install
    pre-commit install --hook-type commit-msg

    print_success "Pre-commit hooks installed!"
}

# Generate initial documentation
generate_docs() {
    print_status "Generating initial documentation..."

    # Create docs directory structure
    mkdir -p docs/{getting-started,user-guide,examples,api,deployment,development,community}
    mkdir -p docs/{stylesheets,javascripts,overrides,includes}

    # Generate Go documentation
    mkdir -p docs/api
    go doc -all ./... > docs/api/generated.md

    print_success "Initial documentation generated!"
}

# Test documentation build
test_docs() {
    print_status "Testing documentation build..."

    # Activate virtual environment
    source venv/bin/activate

    # Build documentation
    mkdocs build --strict

    print_success "Documentation build successful!"
}

# Main function
main() {
    print_status "Setting up GoLangGraph documentation environment..."

    # Check prerequisites
    check_python
    check_pip

    # Install dependencies
    install_dependencies
    install_go_dependencies

    # Setup pre-commit
    setup_precommit

    # Generate documentation
    generate_docs

    # Test build
    test_docs

    print_success "Documentation environment setup complete!"
    print_status "You can now run:"
    echo "  make docs-serve    # Start development server"
    echo "  make docs-build    # Build documentation"
    echo "  make docs-deploy   # Deploy to GitHub Pages"
    echo "  make godoc         # Start Go documentation server"
    echo "  make pre-commit-run # Run pre-commit hooks"
}

# Run main function
main "$@"
