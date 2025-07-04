#!/bin/bash

# Test script for Ollama Demo with Gemma 3:1B
# This script validates the end-to-end functionality of GoLangGraph with Ollama

set -e

echo "🚀 GoLangGraph Ollama Demo Test Script"
echo "======================================"

# Color codes for output
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

# Check if Ollama is installed
check_ollama_installed() {
    print_status "Checking if Ollama is installed..."
    if command -v ollama &> /dev/null; then
        print_success "Ollama is installed"
        return 0
    else
        print_error "Ollama is not installed"
        echo "Please install Ollama from: https://ollama.ai/download"
        echo "Or run: curl -fsSL https://ollama.ai/install.sh | sh"
        return 1
    fi
}

# Check if Ollama is running
check_ollama_running() {
    print_status "Checking if Ollama is running..."
    if curl -s http://localhost:11434/api/tags &> /dev/null; then
        print_success "Ollama is running"
        return 0
    else
        print_warning "Ollama is not running, attempting to start..."
        ollama serve &
        OLLAMA_PID=$!
        sleep 5

        if curl -s http://localhost:11434/api/tags &> /dev/null; then
            print_success "Ollama started successfully"
            return 0
        else
            print_error "Failed to start Ollama"
            return 1
        fi
    fi
}

# Check if Gemma 3:1B model is available
check_gemma_model() {
    print_status "Checking if Gemma 3:1B model is available..."
    if ollama list | grep -q "gemma3:1b"; then
        print_success "Gemma 3:1B model is available"
        return 0
    else
        print_warning "Gemma 3:1B model not found, pulling model..."
        print_status "This may take a few minutes..."

        if ollama pull gemma3:1b; then
            print_success "Gemma 3:1B model pulled successfully"
            return 0
        else
            print_error "Failed to pull Gemma 3:1B model"
            return 1
        fi
    fi
}

# Test basic Ollama functionality
test_ollama_basic() {
    print_status "Testing basic Ollama functionality..."

    response=$(ollama run gemma3:1b "Say 'Hello World' and nothing else" 2>/dev/null | head -1)
    if [[ "$response" == *"Hello"* ]]; then
        print_success "Basic Ollama test passed"
        return 0
    else
        print_error "Basic Ollama test failed"
        print_error "Response: $response"
        return 1
    fi
}

# Build the demo
build_demo() {
    print_status "Building GoLangGraph demo..."

    cd "$(dirname "$0")/.."

    if go build -o bin/ollama-demo ./cmd/examples; then
        print_success "Demo built successfully"
        return 0
    else
        print_error "Failed to build demo"
        return 1
    fi
}

# Run the demo
run_demo() {
    print_status "Running GoLangGraph Ollama demo..."

    # Set timeout for demo execution
    timeout 300 ./bin/ollama-demo 2>&1 | tee demo_output.log

    if [[ ${PIPESTATUS[0]} -eq 0 ]]; then
        print_success "Demo completed successfully"
        return 0
    else
        print_error "Demo failed or timed out"
        return 1
    fi
}

# Validate demo output
validate_demo_output() {
    print_status "Validating demo output..."

    if [[ ! -f "demo_output.log" ]]; then
        print_error "Demo output log not found"
        return 1
    fi

    # Check for expected patterns in output
    local tests_passed=0
    local total_tests=6

    if grep -q "✅ Basic chat test passed!" demo_output.log; then
        print_success "Basic chat test validation passed"
        ((tests_passed++))
    else
        print_error "Basic chat test validation failed"
    fi

    if grep -q "✅ ReAct agent test passed!" demo_output.log; then
        print_success "ReAct agent test validation passed"
        ((tests_passed++))
    else
        print_error "ReAct agent test validation failed"
    fi

    if grep -q "✅ Multi-agent test passed!" demo_output.log; then
        print_success "Multi-agent test validation passed"
        ((tests_passed++))
    else
        print_error "Multi-agent test validation failed"
    fi

    if grep -q "✅ Quick builder test passed!" demo_output.log; then
        print_success "Quick builder test validation passed"
        ((tests_passed++))
    else
        print_error "Quick builder test validation failed"
    fi

    if grep -q "✅ Graph execution test passed!" demo_output.log; then
        print_success "Graph execution test validation passed"
        ((tests_passed++))
    else
        print_error "Graph execution test validation failed"
    fi

    if grep -q "✅ Streaming test passed!" demo_output.log; then
        print_success "Streaming test validation passed"
        ((tests_passed++))
    else
        print_error "Streaming test validation failed"
    fi

    print_status "Test Results: $tests_passed/$total_tests tests passed"

    if [[ $tests_passed -eq $total_tests ]]; then
        print_success "All demo tests passed!"
        return 0
    else
        print_error "Some demo tests failed"
        return 1
    fi
}

# Cleanup function
cleanup() {
    print_status "Cleaning up..."

    # Remove demo output log
    if [[ -f "demo_output.log" ]]; then
        rm demo_output.log
    fi

    # Kill Ollama if we started it
    if [[ -n "$OLLAMA_PID" ]]; then
        print_status "Stopping Ollama process..."
        kill $OLLAMA_PID 2>/dev/null || true
    fi

    print_success "Cleanup completed"
}

# Main execution
main() {
    print_status "Starting GoLangGraph Ollama Demo Test"

    # Set up trap for cleanup
    trap cleanup EXIT

    # Run all checks and tests
    if ! check_ollama_installed; then
        exit 1
    fi

    if ! check_ollama_running; then
        exit 1
    fi

    if ! check_gemma_model; then
        exit 1
    fi

    if ! test_ollama_basic; then
        exit 1
    fi

    if ! build_demo; then
        exit 1
    fi

    if ! run_demo; then
        exit 1
    fi

    if ! validate_demo_output; then
        exit 1
    fi

    print_success "🎉 All tests passed! GoLangGraph is working correctly with Ollama and Gemma 3:1B"

    # Show demo output summary
    echo ""
    echo "Demo Output Summary:"
    echo "==================="
    grep "✅\|❌" demo_output.log || echo "No test results found in output"

    return 0
}

# Allow script to be run with different options
case "${1:-}" in
    "check-only")
        check_ollama_installed && check_ollama_running && check_gemma_model
        ;;
    "build-only")
        build_demo
        ;;
    "run-only")
        run_demo
        ;;
    *)
        main
        ;;
esac
