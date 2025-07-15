#!/bin/bash

# GoLangGraph Stateful Agents - Local Runner Script
# This script runs the application with proper environment configuration

set -e

echo "🚀 Starting GoLangGraph Stateful Agents System..."

# Environment configuration
export OLLAMA_ENDPOINT=http://localhost:11434
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DB=golanggraph_stateful
export POSTGRES_USER=golanggraph
export POSTGRES_PASSWORD=stateful_password_2024
export POSTGRES_SSL_MODE=disable
export LOG_LEVEL=info

# Check if Ollama is running
echo "🔍 Checking Ollama availability..."
if ! curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "❌ Ollama is not running on localhost:11434"
    echo "Please start Ollama first: ollama serve"
    exit 1
fi
echo "✅ Ollama is running"

# Check if PostgreSQL container is running
echo "🔍 Checking PostgreSQL container..."
if ! docker ps | grep golanggraph-postgres > /dev/null; then
    echo "❌ PostgreSQL container not running"
    echo "Please start it with: docker-compose -f docker-compose.local.yml up -d postgres"
    exit 1
fi
echo "✅ PostgreSQL container is running"

# Test database connection using Docker network
echo "🔍 Testing database connection..."
if docker run --rm --network golanggraph-stateful-network postgres:15-alpine \
    psql postgresql://golanggraph:stateful_password_2024@postgres:5432/golanggraph_stateful \ # pragma: allowlist secret
    -c "SELECT 'Database connection successful!' as status;" > /dev/null 2>&1; then
    echo "✅ Database connection successful"
else
    echo "❌ Database connection failed"
    exit 1
fi

# Build the application if needed
if [ ! -f "./stateful-agents" ]; then
    echo "🔨 Building application..."
    go build -o stateful-agents .
fi

# Run the application in a Docker container with network access
echo "🎯 Starting application in Docker container..."
docker run --rm -it \
    --network golanggraph-stateful-network \
    -p 8080:8080 \
    -e OLLAMA_ENDPOINT=http://host.docker.internal:11434 \
    -e POSTGRES_HOST=postgres \
    -e POSTGRES_PORT=5432 \
    -e POSTGRES_DB=golanggraph_stateful \
    -e POSTGRES_USER=golanggraph \
    -e POSTGRES_PASSWORD=stateful_password_2024 \
    -e POSTGRES_SSL_MODE=disable \
    -e LOG_LEVEL=info \
    -v $(pwd)/stateful-agents:/app/stateful-agents \
    -w /app \
    alpine:latest \
    sh -c "apk add --no-cache curl && chmod +x /app/stateful-agents && /app/stateful-agents"
