#!/bin/bash

# GoLangGraph Stateful Ideation Agents - System Test Suite
# Tests all migrated agents with comprehensive validation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
API_BASE="$BASE_URL/api"
TIMEOUT=30
SESSION_ID="test_session_$(date +%s)"
USER_ID="test_user_123"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
    ((PASSED_TESTS++))
}

error() {
    echo -e "${RED}❌ $1${NC}"
    ((FAILED_TESTS++))
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

test_endpoint() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="$3"
    local expected_status="${4:-200}"
    local description="$5"

    ((TOTAL_TESTS++))

    log "Testing: $description"
    log "Endpoint: $method $endpoint"

    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            --max-time $TIMEOUT \
            "$endpoint" 2>/dev/null || echo "HTTPSTATUS:000")
    else
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            --max-time $TIMEOUT \
            "$endpoint" 2>/dev/null || echo "HTTPSTATUS:000")
    fi

    http_code=$(echo "$response" | sed -n 's/.*HTTPSTATUS:\([0-9]*\)$/\1/p')
    body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*$//')

    if [ "$http_code" = "$expected_status" ]; then
        success "$description - Status: $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        error "$description - Expected: $expected_status, Got: $http_code"
        echo "Response: $body"
        return 1
    fi
}

# Wait for system to be ready
wait_for_system() {
    log "Waiting for system to be ready..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s "$BASE_URL/health" >/dev/null 2>&1; then
            success "System is ready!"
            return 0
        fi

        log "Attempt $attempt/$max_attempts - System not ready yet..."
        sleep 2
        ((attempt++))
    done

    error "System failed to become ready within timeout"
    return 1
}

# Test system health
test_system_health() {
    log "🏥 Testing System Health"

    test_endpoint "$BASE_URL/health" "GET" "" "200" "Health Check"
    test_endpoint "$BASE_URL/metrics" "GET" "" "200" "Metrics Endpoint"
    test_endpoint "$BASE_URL/agents" "GET" "" "200" "Agent List"
}

# Test Designer Agent (Stateful)
test_designer_agent() {
    log "🎨 Testing Designer Agent (Stateful)"

    local design_request='{
        "message": "Design a sustainable tiny house for $50k",
        "session_id": "'$SESSION_ID'",
        "user_id": "'$USER_ID'",
        "context": {
            "project_type": "residential",
            "budget_range": 50000,
            "sustainability_priority": 9,
            "location": "Pacific Northwest"
        }
    }'

    test_endpoint "$API_BASE/designer" "POST" "$design_request" "200" "Designer Agent - Initial Design Request"

    # Test session continuity
    local follow_up_request='{
        "message": "Can you make it more modern and add solar panels?",
        "session_id": "'$SESSION_ID'",
        "user_id": "'$USER_ID'"
    }'

    test_endpoint "$API_BASE/designer" "POST" "$follow_up_request" "200" "Designer Agent - Follow-up with Session Continuity"

    # Test session retrieval
    test_endpoint "$API_BASE/designer/session/$SESSION_ID" "GET" "" "200" "Designer Agent - Session Retrieval"
}

# Test Interviewer Agent (Stateful)
test_interviewer_agent() {
    log "🎤 Testing Interviewer Agent (Stateful)"

    local interview_session="interview_$SESSION_ID"

    local interview_request='{
        "message": "Je veux construire une maison écologique",
        "session_id": "'$interview_session'",
        "user_id": "'$USER_ID'"
    }'

    test_endpoint "$API_BASE/interviewer" "POST" "$interview_request" "200" "Interviewer Agent - French Interview Start"

    # Test follow-up question
    local follow_up='{
        "message": "Je préfère les matériaux naturels et l'\''énergie solaire",
        "session_id": "'$interview_session'",
        "user_id": "'$USER_ID'"
    }'

    test_endpoint "$API_BASE/interviewer" "POST" "$follow_up" "200" "Interviewer Agent - Follow-up Response"
}

# Test Highlighter Agent (Stateful)
test_highlighter_agent() {
    log "🔍 Testing Highlighter Agent (Stateful)"

    local analysis_session="analysis_$SESSION_ID"

    local conversation_history='[
        {
            "role": "user",
            "content": "I want to build a sustainable house with solar panels and rainwater collection",
            "timestamp": "2024-01-01T10:00:00Z"
        },
        {
            "role": "assistant",
            "content": "Great! Let me help you design an eco-friendly home with renewable energy systems",
            "timestamp": "2024-01-01T10:01:00Z"
        },
        {
            "role": "user",
            "content": "Budget is around $80k and I prefer natural materials like bamboo",
            "timestamp": "2024-01-01T10:02:00Z"
        }
    ]'

    local analysis_request='{
        "conversation_history": '$conversation_history',
        "session_id": "'$analysis_session'",
        "analysis_focus": ["sustainability", "budget", "materials"],
        "depth_level": "detailed"
    }'

    test_endpoint "$API_BASE/highlighter" "POST" "$analysis_request" "200" "Highlighter Agent - Conversation Analysis"
}

# Test Storymaker Agent (Stateful)
test_storymaker_agent() {
    log "📚 Testing Storymaker Agent (Stateful)"

    local story_session="story_$SESSION_ID"

    local story_request='{
        "story_prompt": "A family living in a floating eco-village in 2035",
        "session_id": "'$story_session'",
        "target_audience": "general",
        "genre": "science_fiction",
        "story_length": "medium",
        "setting": {
            "location": "Floating Islands",
            "time_period": "2035",
            "habitat_type": "floating",
            "environment": "coastal"
        },
        "characters": [
            {
                "name": "Marina",
                "role": "Marine Biologist",
                "background": "Expert in sustainable ocean farming",
                "motivation": "Create harmony between human habitation and ocean ecosystems"
            }
        ]
    }'

    test_endpoint "$API_BASE/storymaker" "POST" "$story_request" "200" "Storymaker Agent - Story Generation"

    # Test story continuation
    local follow_up_story='{
        "story_prompt": "Continue the story with Marina discovering a new way to grow food underwater",
        "session_id": "'$story_session'",
        "genre": "science_fiction",
        "story_length": "short"
    }'

    test_endpoint "$API_BASE/storymaker" "POST" "$follow_up_story" "200" "Storymaker Agent - Story Continuation"
}

# Test streaming endpoints
test_streaming() {
    log "🌊 Testing Streaming Endpoints"

    local stream_request='{
        "message": "Design a quick sustainable cabin",
        "session_id": "'$SESSION_ID'_stream",
        "user_id": "'$USER_ID'"
    }'

    # Test streaming for designer (simplified test)
    local stream_response=$(curl -s --max-time 10 \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$stream_request" \
        "$API_BASE/designer/stream" || echo "FAILED")

    if [ "$stream_response" != "FAILED" ]; then
        success "Designer Agent - Streaming Response"
    else
        error "Designer Agent - Streaming Response Failed"
    fi
}

# Test agent status endpoints
test_agent_status() {
    log "📊 Testing Agent Status Endpoints"

    test_endpoint "$API_BASE/designer/status" "GET" "" "200" "Designer Agent Status"
    test_endpoint "$API_BASE/interviewer/status" "GET" "" "200" "Interviewer Agent Status"
    test_endpoint "$API_BASE/highlighter/status" "GET" "" "200" "Highlighter Agent Status"
    test_endpoint "$API_BASE/storymaker/status" "GET" "" "200" "Storymaker Agent Status"
}

# Test error handling
test_error_handling() {
    log "🚫 Testing Error Handling"

    # Test invalid agent
    test_endpoint "$API_BASE/nonexistent" "POST" '{"message":"test"}' "404" "Invalid Agent Endpoint"

    # Test malformed JSON
    test_endpoint "$API_BASE/designer" "POST" '{"invalid json":}' "400" "Malformed JSON Handling"

    # Test missing required fields
    test_endpoint "$API_BASE/designer" "POST" '{}' "400" "Missing Required Fields"
}

# Test session persistence
test_session_persistence() {
    log "💾 Testing Session Persistence"

    local persistence_session="persist_$SESSION_ID"

    # Create initial session
    local initial_request='{
        "message": "Start a new sustainable house project",
        "session_id": "'$persistence_session'",
        "user_id": "'$USER_ID'",
        "context": {"project_name": "EcoHome"}
    }'

    test_endpoint "$API_BASE/designer" "POST" "$initial_request" "200" "Session Creation"

    # Test session retrieval
    test_endpoint "$API_BASE/designer/session/$persistence_session" "GET" "" "200" "Session Persistence Check"

    # Continue session after delay
    sleep 2
    local continue_request='{
        "message": "Add more details to the project",
        "session_id": "'$persistence_session'",
        "user_id": "'$USER_ID'"
    }'

    test_endpoint "$API_BASE/designer" "POST" "$continue_request" "200" "Session Continuity After Delay"
}

# Performance test
test_performance() {
    log "⚡ Testing Performance"

    local perf_session="perf_$SESSION_ID"
    local start_time=$(date +%s)

    local perf_request='{
        "message": "Quick design for performance test",
        "session_id": "'$perf_session'",
        "user_id": "'$USER_ID'"
    }'

    if test_endpoint "$API_BASE/designer" "POST" "$perf_request" "200" "Performance Test Request"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))

        if [ $duration -le 10 ]; then
            success "Performance Test - Response time: ${duration}s (Good)"
        elif [ $duration -le 20 ]; then
            warning "Performance Test - Response time: ${duration}s (Acceptable)"
        else
            error "Performance Test - Response time: ${duration}s (Too Slow)"
        fi
    fi
}

# Main test execution
main() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════════════════════╗"
    echo "║              GoLangGraph Stateful Agents Test Suite                   ║"
    echo "║                    Testing All Migrated Agents                        ║"
    echo "╚════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    log "Starting comprehensive test suite..."
    log "Base URL: $BASE_URL"
    log "Session ID: $SESSION_ID"
    log "User ID: $USER_ID"

    # Wait for system
    if ! wait_for_system; then
        error "System is not available. Please ensure the application is running."
        exit 1
    fi

    # Run all tests
    test_system_health
    test_designer_agent
    test_interviewer_agent
    test_highlighter_agent
    test_storymaker_agent
    test_streaming
    test_agent_status
    test_error_handling
    test_session_persistence
    test_performance

    # Summary
    echo
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}                              TEST SUMMARY                                   ${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════════${NC}"

    echo -e "Total Tests:  ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"

    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo -e "Success Rate: ${BLUE}$success_rate%${NC}"

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed! The stateful agent system is working correctly.${NC}"
        exit 0
    else
        echo -e "\n${RED}💥 Some tests failed. Please check the logs above for details.${NC}"
        exit 1
    fi
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    warning "jq is not installed. JSON responses will not be formatted."
fi

# Run main function
main "$@"
