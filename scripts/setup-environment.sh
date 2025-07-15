#!/bin/bash

# GoLangGraph Stateful Agents Environment Setup Script
# This script sets up the complete environment for testing stateful agents

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
OLLAMA_ENDPOINT="http://localhost:11434"
POSTGRES_HOST="localhost"
POSTGRES_PORT="5432"
POSTGRES_DB="golanggraph_stateful"
POSTGRES_USER="golanggraph"
POSTGRES_PASSWORD="stateful_password_2024" # pragma: allowlist secret

print_header() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║              GoLangGraph Stateful Agents Setup                      ║"
    echo "║                   Environment Initialization                        ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_dependencies() {
    echo -e "${YELLOW}Checking dependencies...${NC}"

    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker not found. Please install Docker first.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Docker found${NC}"

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose not found. Please install Docker Compose first.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Docker Compose found${NC}"

    # Check Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go not found. Please install Go first.${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Go found ($(go version))${NC}"

    # Check jq (optional but helpful)
    if ! command -v jq &> /dev/null; then
        echo -e "${YELLOW}⚠️  jq not found. Installing jq for better JSON handling...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install jq 2>/dev/null || echo -e "${YELLOW}Please install jq manually: brew install jq${NC}"
        else
            echo -e "${YELLOW}Please install jq manually for better JSON handling${NC}"
        fi
    else
        echo -e "${GREEN}✅ jq found${NC}"
    fi

    # Check Ollama
    echo -e "${YELLOW}Checking Ollama connection...${NC}"
    if curl -s "$OLLAMA_ENDPOINT/api/tags" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Ollama is running and accessible${NC}"
        echo -e "${BLUE}Available models:${NC}"
        curl -s "$OLLAMA_ENDPOINT/api/tags" | jq -r '.models[]?.name // empty' 2>/dev/null | head -5 | sed 's/^/  - /'
    else
        echo -e "${RED}❌ Ollama is not running or not accessible at $OLLAMA_ENDPOINT${NC}"
        echo -e "${YELLOW}Please start Ollama first with: ollama serve${NC}"
        exit 1
    fi
}

setup_directories() {
    echo -e "${YELLOW}Setting up directories...${NC}"

    # Ensure scripts directory exists
    mkdir -p scripts
    mkdir -p bin
    mkdir -p logs

    # Set executable permissions
    chmod +x scripts/*.sh 2>/dev/null || true
    chmod +x test-system.sh 2>/dev/null || true

    echo -e "${GREEN}✅ Directories set up${NC}"
}

clean_existing_containers() {
    echo -e "${YELLOW}Cleaning up existing containers...${NC}"

    # Stop and remove containers if they exist
    docker stop golanggraph-postgres golanggraph-redis 2>/dev/null || true
    docker rm golanggraph-postgres golanggraph-redis 2>/dev/null || true

    # Clean up any orphaned containers
    docker container prune -f 2>/dev/null || true

    echo -e "${GREEN}✅ Existing containers cleaned up${NC}"
}

start_services() {
    echo -e "${YELLOW}Starting PostgreSQL and Redis services...${NC}"

    # Start services using docker-compose
    if [ -f "docker-compose.local.yml" ]; then
        docker-compose -f docker-compose.local.yml up -d postgres redis
    else
        echo -e "${RED}❌ docker-compose.local.yml not found${NC}"
        exit 1
    fi

    echo -e "${YELLOW}Waiting for services to initialize...${NC}"
    sleep 15

    # Verify PostgreSQL is ready
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if docker exec golanggraph-postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1;" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ PostgreSQL is ready${NC}"
            break
        fi

        echo -e "${YELLOW}Waiting for PostgreSQL... (attempt $attempt/$max_attempts)${NC}"
        sleep 2
        ((attempt++))
    done

    if [ $attempt -gt $max_attempts ]; then
        echo -e "${RED}❌ PostgreSQL failed to start within timeout${NC}"
        exit 1
    fi

    # Verify Redis is ready
    if docker exec golanggraph-redis redis-cli ping > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Redis is ready${NC}"
    else
        echo -e "${YELLOW}⚠️  Redis connection check failed, but continuing...${NC}"
    fi
}

verify_database_setup() {
    echo -e "${YELLOW}Verifying database setup...${NC}"

    # Check if database exists and has tables
    local table_count
    table_count=$(docker exec golanggraph-postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ' || echo "0")

    if [ "$table_count" -gt 0 ]; then
        echo -e "${GREEN}✅ Database has $table_count tables${NC}"

        # Show key tables
        echo -e "${BLUE}Key tables in database:${NC}"
        docker exec golanggraph-postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;" 2>/dev/null | grep -E '(threads|sessions|checkpoints|user_preferences)' | sed 's/^/  - /' || true
    else
        echo -e "${YELLOW}⚠️  Database exists but no tables found. This is normal for first run.${NC}"
    fi

    # Check pgvector extension
    if docker exec golanggraph-postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT extname FROM pg_extension WHERE extname = 'vector';" 2>/dev/null | grep -q vector; then
        echo -e "${GREEN}✅ pgvector extension is installed${NC}"
    else
        echo -e "${YELLOW}⚠️  pgvector extension not found${NC}"
    fi
}

build_application() {
    echo -e "${YELLOW}Building stateful agents application...${NC}"

    # Clean and build
    GOWORK=off go mod tidy
    GOWORK=off go build -o stateful-agents .

    if [ -f "stateful-agents" ]; then
        echo -e "${GREEN}✅ Application built successfully${NC}"
        ls -lh stateful-agents
    else
        echo -e "${RED}❌ Application build failed${NC}"
        exit 1
    fi
}

show_connection_info() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║                        Connection Information                       ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    echo -e "${YELLOW}Database Connection:${NC}"
    echo "  Host: $POSTGRES_HOST"
    echo "  Port: $POSTGRES_PORT"
    echo "  Database: $POSTGRES_DB"
    echo "  User: $POSTGRES_USER"
    echo "  Password: $POSTGRES_PASSWORD"
    echo

    echo -e "${YELLOW}Ollama Connection:${NC}"
    echo "  Endpoint: $OLLAMA_ENDPOINT"
    echo

    echo -e "${YELLOW}Application URLs:${NC}"
    echo "  Health: http://localhost:8080/health"
    echo "  API: http://localhost:8080/api/"
    echo

    echo -e "${YELLOW}Docker Containers:${NC}"
    docker ps --filter name=golanggraph --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "  No containers running"
}

show_usage() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║                           Usage Instructions                        ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    echo -e "${YELLOW}To start the application:${NC}"
    echo "  make stateful-run"
    echo "  OR"
    echo "  OLLAMA_ENDPOINT=$OLLAMA_ENDPOINT \\"
    echo "  POSTGRES_HOST=$POSTGRES_HOST POSTGRES_PORT=$POSTGRES_PORT \\"
    echo "  POSTGRES_DB=$POSTGRES_DB POSTGRES_USER=$POSTGRES_USER \\"
    echo "  POSTGRES_PASSWORD=$POSTGRES_PASSWORD POSTGRES_SSL_MODE=disable \\"
    echo "  ./stateful-agents"
    echo

    echo -e "${YELLOW}To test the system:${NC}"
    echo "  make stateful-test"
    echo "  OR"
    echo "  ./test-system.sh"
    echo

    echo -e "${YELLOW}To check system status:${NC}"
    echo "  make stateful-status"
    echo

    echo -e "${YELLOW}To clean up:${NC}"
    echo "  make stateful-clean"
}

main() {
    print_header

    echo -e "${BLUE}Starting environment setup...${NC}"
    echo

    check_dependencies
    echo

    setup_directories
    echo

    clean_existing_containers
    echo

    start_services
    echo

    verify_database_setup
    echo

    build_application
    echo

    show_connection_info
    echo

    show_usage
    echo

    echo -e "${GREEN}🎉 Environment setup complete!${NC}"
    echo -e "${BLUE}You can now run 'make stateful-run' to start the application${NC}"
}

# Script entry point
main "$@"
