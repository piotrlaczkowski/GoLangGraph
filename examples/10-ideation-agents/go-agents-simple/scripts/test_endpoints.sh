#!/bin/bash
# GoLangGraph Go-Agents-Simple Endpoint Testing Script
# Tests all auto-generated endpoints for functionality

set -e

# Configuration
BASE_URL="http://localhost:8080"
TIMEOUT=30
VERBOSE=false
FAILED_TESTS=0
TOTAL_TESTS=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "SUCCESS") echo -e "${GREEN}✅ $message${NC}" ;;
        "FAIL") echo -e "${RED}❌ $message${NC}" ;;
        "INFO") echo -e "${BLUE}ℹ️  $message${NC}" ;;
        "WARN") echo -e "${YELLOW}⚠️  $message${NC}" ;;
    esac
}

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    local data=$5

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$VERBOSE" = true ]; then
        print_status "INFO" "Testing: $method $endpoint - $description"
    fi

    # Build curl command
    local curl_cmd="curl -s -w '%{http_code}' --connect-timeout $TIMEOUT"

    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        curl_cmd="$curl_cmd -X POST -H 'Content-Type: application/json' -d '$data'"
    elif [ "$method" = "POST" ]; then
        curl_cmd="$curl_cmd -X POST -H 'Content-Type: application/json'"
    fi

    curl_cmd="$curl_cmd $BASE_URL$endpoint"

    # Execute request
    local response
    response=$(eval $curl_cmd 2>/dev/null)

    if [ $? -ne 0 ]; then
        print_status "FAIL" "$description - Connection failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi

    # Extract status code (last 3 characters)
    local status_code=${response: -3}
    local body=${response%???}

    # Check status code
    if [ "$status_code" = "$expected_status" ]; then
        print_status "SUCCESS" "$description ($status_code)"
        if [ "$VERBOSE" = true ] && [ -n "$body" ]; then
            echo "Response: $body" | head -c 200
            echo
        fi
        return 0
    else
        print_status "FAIL" "$description - Expected $expected_status, got $status_code"
        if [ -n "$body" ]; then
            echo "Response: $body" | head -c 200
            echo
        fi
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Function to wait for service
wait_for_service() {
    local max_attempts=30
    local attempt=1

    print_status "INFO" "Waiting for service to be ready..."

    while [ $attempt -le $max_attempts ]; do
        if curl -s --connect-timeout 5 "$BASE_URL/health" > /dev/null 2>&1; then
            print_status "SUCCESS" "Service is ready!"
            return 0
        fi

        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done

    print_status "FAIL" "Service did not become ready within $((max_attempts * 2)) seconds"
    return 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -u|--url)
            BASE_URL="$2"
            shift 2
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -v, --verbose     Verbose output"
            echo "  -u, --url URL     Base URL (default: http://localhost:8080)"
            echo "  -t, --timeout N   Timeout in seconds (default: 30)"
            echo "  -h, --help        Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "🧪 GoLangGraph Go-Agents-Simple Endpoint Testing"
echo "================================================"
echo "Base URL: $BASE_URL"
echo "Timeout: ${TIMEOUT}s"
echo "Verbose: $VERBOSE"
echo

# Wait for service to be ready
if ! wait_for_service; then
    exit 1
fi

echo
echo "🧪 Testing System Endpoints..."
echo "------------------------------"

# Test system endpoints
test_endpoint "GET" "/health" "200" "Health check endpoint"
test_endpoint "GET" "/capabilities" "200" "System capabilities endpoint"
test_endpoint "GET" "/agents" "200" "List agents endpoint"

echo
echo "🌐 Testing Web Interface Endpoints..."
echo "------------------------------------"

# Test web interface endpoints
test_endpoint "GET" "/chat" "200" "Chat interface endpoint"
test_endpoint "GET" "/playground" "200" "API playground endpoint"
test_endpoint "GET" "/debug" "200" "Debug interface endpoint"

echo
echo "📄 Testing Schema API Endpoints..."
echo "---------------------------------"

# Test schema endpoints
test_endpoint "GET" "/schemas" "200" "All schemas endpoint"
test_endpoint "GET" "/schemas/designer" "200" "Designer schema endpoint"
test_endpoint "GET" "/schemas/interviewer" "200" "Interviewer schema endpoint"
test_endpoint "GET" "/schemas/highlighter" "200" "Highlighter schema endpoint"
test_endpoint "GET" "/schemas/storymaker" "200" "Storymaker schema endpoint"

echo
echo "📊 Testing Metrics Endpoints..."
echo "------------------------------"

# Test metrics endpoints
test_endpoint "GET" "/metrics" "200" "System metrics endpoint"
test_endpoint "GET" "/metrics/designer" "200" "Designer metrics endpoint"

echo
echo "🤖 Testing Agent Endpoints..."
echo "----------------------------"

# Test agent information endpoints
test_endpoint "GET" "/agents/designer" "200" "Designer agent info"
test_endpoint "GET" "/agents/interviewer" "200" "Interviewer agent info"
test_endpoint "GET" "/agents/highlighter" "200" "Highlighter agent info"
test_endpoint "GET" "/agents/storymaker" "200" "Storymaker agent info"

# Test agent status endpoints
test_endpoint "GET" "/api/designer/status" "200" "Designer status endpoint"
test_endpoint "GET" "/api/interviewer/status" "200" "Interviewer status endpoint"
test_endpoint "GET" "/api/highlighter/status" "200" "Highlighter status endpoint"
test_endpoint "GET" "/api/storymaker/status" "200" "Storymaker status endpoint"

echo
echo "🔍 Testing Agent Execution Endpoints..."
echo "--------------------------------------"

# Test agent execution with sample data
test_endpoint "POST" "/api/designer" "200" "Designer execution" '{"message": "Design a sustainable treehouse"}'
test_endpoint "POST" "/api/interviewer" "200" "Interviewer execution" '{"message": "Hello, I want to design a sustainable home"}'
test_endpoint "POST" "/api/highlighter" "200" "Highlighter execution" '{"conversation_history": [{"role": "user", "content": "I want eco-friendly materials"}]}'
test_endpoint "POST" "/api/storymaker" "200" "Storymaker execution" '{"story_prompt": "A family living in a sustainable habitat in 2035"}'

echo
echo "📋 Testing Conversation Endpoints..."
echo "-----------------------------------"

# Test conversation management
test_endpoint "GET" "/api/designer/conversation" "200" "Designer conversation history"
test_endpoint "POST" "/api/designer/conversation" "200" "Designer add to conversation" '{"message": "Add this to conversation", "role": "user"}'
test_endpoint "DELETE" "/api/designer/conversation" "200" "Designer clear conversation"

echo
echo "❌ Testing Error Handling..."
echo "---------------------------"

# Test error scenarios
test_endpoint "GET" "/api/nonexistent" "404" "Non-existent agent endpoint"
test_endpoint "POST" "/api/designer" "400" "Designer with invalid data" '{"invalid": "data"}'
test_endpoint "GET" "/schemas/nonexistent" "404" "Non-existent schema"

echo
echo "📊 Test Results Summary"
echo "======================"

PASSED_TESTS=$((TOTAL_TESTS - FAILED_TESTS))
SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))

echo "Total Tests: $TOTAL_TESTS"
echo "Passed: $PASSED_TESTS"
echo "Failed: $FAILED_TESTS"
echo "Success Rate: ${SUCCESS_RATE}%"

if [ $FAILED_TESTS -eq 0 ]; then
    print_status "SUCCESS" "All tests passed! 🎉"
    echo
    echo "🚀 GoLangGraph Go-Agents-Simple is fully functional!"
    echo "   🌐 Web UI: $BASE_URL/"
    echo "   🛝 API Playground: $BASE_URL/playground"
    echo "   📋 Health: $BASE_URL/health"
    echo "   📊 Metrics: $BASE_URL/metrics"
    exit 0
else
    print_status "FAIL" "$FAILED_TESTS tests failed"
    echo
    echo "❌ Some endpoints are not working correctly."
    echo "   Check the service logs and ensure all dependencies are running."
    exit 1
fi
