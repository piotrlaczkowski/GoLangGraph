# Multi-Agent Deployment Guide

This guide explains how to create, configure, and deploy multiple independent agents using the GoLangGraph multi-agent system. This feature allows you to run several agents simultaneously with different configurations, routing rules, and deployment settings.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [CLI Commands](#cli-commands)
- [Deployment Options](#deployment-options)
- [Routing and Load Balancing](#routing-and-load-balancing)
- [Monitoring and Management](#monitoring-and-management)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

The multi-agent deployment feature enables you to:

- **Deploy Multiple Agents**: Run several agents with different configurations simultaneously
- **Flexible Routing**: Route requests to different agents based on path, headers, queries, or hosts
- **Independent Configuration**: Each agent can have its own model, tools, and behavior
- **Scalable Deployment**: Support for Docker, Kubernetes, and serverless deployments
- **Centralized Management**: Unified monitoring, logging, and health checking
- **Schema Validation**: Individual input/output validation for each agent

## Quick Start

### 1. Initialize a Multi-Agent Project

```bash
# Create a new multi-agent project with 3 agents
golanggraph multi-agent init my-multi-agent --agents 3 --template basic

# Or with different templates
golanggraph multi-agent init ecommerce-agents --template microservices --agents 5
golanggraph multi-agent init rag-system --template rag --agents 3
golanggraph multi-agent init workflow-system --template workflow --agents 5
```

### 2. Validate Configuration

```bash
# Validate the generated configuration
golanggraph multi-agent validate configs/multi-agent.yaml

# Strict validation with schema checking
golanggraph multi-agent validate configs/multi-agent.yaml --strict --check-schemas
```

### 3. Start the Multi-Agent Server

```bash
# Start the server locally
golanggraph multi-agent serve configs/multi-agent.yaml

# Start on a specific host and port
golanggraph multi-agent serve configs/multi-agent.yaml --host 0.0.0.0 --port 8080
```

### 4. Test the Agents

```bash
# Test different agents via routing
curl -X POST http://localhost:8080/chat -H "Content-Type: application/json" -d '{"input": "Hello!"}'
curl -X POST http://localhost:8080/reason -H "Content-Type: application/json" -d '{"input": "Solve: 2+2*3"}'
curl -X POST http://localhost:8080/tools -H "Content-Type: application/json" -d '{"input": "Read file: README.md"}'
```

## Configuration

### Multi-Agent Configuration Structure

```yaml
name: "my-multi-agent-system"
version: "1.0.0"
description: "Multi-agent system with specialized roles"

# Define multiple agents
agents:
  chat-agent:
    id: "chat-agent"
    name: "Chat Assistant"
    type: "chat"
    model: "gpt-3.5-turbo"
    provider: "openai"
    system_prompt: "You are a helpful chat assistant."
    tools: ["calculator", "web_search"]
    
  reasoning-agent:
    id: "reasoning-agent"
    name: "Reasoning Agent"
    type: "react"
    model: "gpt-4"
    provider: "openai"
    system_prompt: "You are a step-by-step reasoning agent."
    tools: ["calculator", "web_search", "file_read"]

# Configure routing between agents
routing:
  type: "path"  # or "host", "header", "query"
  default_agent: "chat-agent"
  rules:
    - id: "chat-route"
      pattern: "/chat"
      agent_id: "chat-agent"
      method: "POST"
      priority: 100
    - id: "reason-route"
      pattern: "/reason"
      agent_id: "reasoning-agent"
      method: "POST"
      priority: 90

# Deployment configuration
deployment:
  type: "docker"  # or "kubernetes", "serverless"
  environment: "production"
  replicas: 3
  
  resources:
    cpu: "1000m"
    memory: "1Gi"
  
  networking:
    type: "LoadBalancer"
    ports:
      - name: "http"
        port: 8080
        target_port: 8080

# Shared configuration for all agents
shared:
  database:
    type: "postgres"
    host: "postgres"
    port: 5432
    database: "golanggraph"
  
  llm_providers:
    openai:
      type: "openai"
      api_key: "${OPENAI_API_KEY}"
      endpoint: "https://api.openai.com/v1"
  
  monitoring:
    enabled: true
    metrics:
      enabled: true
      path: "/metrics"
      port: 9090
```

### Agent-Specific Configuration

Each agent directory contains its own configuration:

```
my-multi-agent/
├── agents/
│   ├── agent-1/
│   │   ├── config.yaml
│   │   └── schema.json
│   ├── agent-2/
│   │   ├── config.yaml
│   │   └── schema.json
│   └── agent-3/
│       ├── config.yaml
│       └── schema.json
├── configs/
│   └── multi-agent.yaml
├── deploy/
│   ├── docker-compose.yml
│   └── Dockerfile
└── k8s/
    ├── deployment.yaml
    ├── service.yaml
    └── ingress.yaml
```

## CLI Commands

### `golanggraph multi-agent init`

Initialize a new multi-agent project.

```bash
golanggraph multi-agent init [project-name] [flags]

Flags:
  -t, --template string    Project template (basic, microservices, rag, workflow) (default "basic")
  -a, --agents int         Number of agents to create (default 3)
  -f, --format string      Configuration format (yaml, json) (default "yaml")
  -r, --routing string     Routing type (path, host, header, query) (default "path")
```

### `golanggraph multi-agent validate`

Validate multi-agent configuration.

```bash
golanggraph multi-agent validate [config-file] [flags]

Flags:
  -s, --strict             Enable strict validation
      --check-schemas      Validate input/output schemas (default true)
```

### `golanggraph multi-agent deploy`

Deploy multiple agents.

```bash
golanggraph multi-agent deploy [config-file] [flags]

Flags:
  -t, --type string           Deployment type (docker, kubernetes, serverless) (default "docker")
  -e, --environment string    Deployment environment (default "development")
      --dry-run               Show what would be deployed without actually deploying
      --parallel              Deploy agents in parallel (default true)
```

### `golanggraph multi-agent serve`

Start multi-agent server.

```bash
golanggraph multi-agent serve [config-file] [flags]

Flags:
  -H, --host string    Host to bind to (default "0.0.0.0")
  -p, --port int       Port to bind to (default 8080)
```

### `golanggraph multi-agent status`

Check status of deployed agents.

```bash
golanggraph multi-agent status [config-file] [flags]

Flags:
  -f, --format string    Output format (table, json, yaml) (default "table")
  -w, --watch            Watch for status changes
```

### Generate Commands

Generate deployment artifacts:

```bash
# Generate Docker files
golanggraph multi-agent generate docker [config-file] --output ./deploy

# Generate Kubernetes manifests
golanggraph multi-agent generate k8s [config-file] --output ./k8s --namespace golanggraph
```

## Deployment Options

### Docker Deployment

```bash
# Generate Docker Compose file
golanggraph multi-agent generate docker configs/multi-agent.yaml

# Deploy with Docker Compose
docker-compose up -d

# Deploy specific environment
golanggraph multi-agent deploy configs/multi-agent.yaml --type docker --environment production
```

### Kubernetes Deployment

```bash
# Generate Kubernetes manifests
golanggraph multi-agent generate k8s configs/multi-agent.yaml

# Apply to cluster
kubectl apply -f k8s/

# Deploy with CLI
golanggraph multi-agent deploy configs/multi-agent.yaml --type kubernetes --environment production
```

### Serverless Deployment

```bash
# Deploy to serverless platform
golanggraph multi-agent deploy configs/multi-agent.yaml --type serverless --environment production
```

## Routing and Load Balancing

### Path-Based Routing

Route requests based on URL path:

```yaml
routing:
  type: "path"
  rules:
    - pattern: "/chat"
      agent_id: "chat-agent"
    - pattern: "/api/v1/reasoning"
      agent_id: "reasoning-agent"
    - pattern: "/tools"
      agent_id: "tool-agent"
```

### Host-Based Routing

Route requests based on hostname:

```yaml
routing:
  type: "host"
  rules:
    - pattern: "chat.example.com"
      agent_id: "chat-agent"
    - pattern: "api.example.com"
      agent_id: "api-agent"
```

### Header-Based Routing

Route requests based on HTTP headers:

```yaml
routing:
  type: "header"
  rules:
    - pattern: "X-Agent-Type:chat"
      agent_id: "chat-agent"
    - pattern: "X-Agent-Type:reasoning"
      agent_id: "reasoning-agent"
```

### Query-Based Routing

Route requests based on query parameters:

```yaml
routing:
  type: "query"
  rules:
    - pattern: "agent=chat"
      agent_id: "chat-agent"
    - pattern: "agent=reasoning"
      agent_id: "reasoning-agent"
```

### Load Balancing

Configure load balancing for high availability:

```yaml
deployment:
  replicas: 3
  scaling:
    enabled: true
    min_replicas: 2
    max_replicas: 10
    target_cpu_percent: 70
    target_memory_percent: 80
```

## Monitoring and Management

### Health Checks

Monitor agent health:

```bash
# Overall health
curl http://localhost:8080/health

# Specific agent health
curl http://localhost:8080/health/chat-agent
```

### Metrics

Access system metrics:

```bash
# Prometheus metrics
curl http://localhost:8080/metrics

# Agent-specific metrics
curl http://localhost:8080/agents/chat-agent/metrics
```

### Management APIs

```bash
# List all agents
curl http://localhost:8080/agents

# Get agent details
curl http://localhost:8080/agents/chat-agent

# Get agent status
curl http://localhost:8080/agents/chat-agent/status

# Get deployment status
curl http://localhost:8080/deployment/status
```

## Examples

### Example 1: E-commerce Microservices

```yaml
name: "ecommerce-multi-agent"
agents:
  user-service:
    type: "chat"
    model: "gpt-3.5-turbo"
    system_prompt: "Handle user authentication and profile management"
    tools: ["user_db", "auth"]
  
  product-service:
    type: "react"
    model: "gpt-4"
    system_prompt: "Manage product catalog and recommendations"
    tools: ["product_db", "search", "recommendations"]
  
  order-service:
    type: "tool"
    model: "gpt-3.5-turbo"
    system_prompt: "Process orders and payments"
    tools: ["payment", "inventory", "shipping"]

routing:
  type: "path"
  rules:
    - pattern: "/api/users"
      agent_id: "user-service"
    - pattern: "/api/products"
      agent_id: "product-service"
    - pattern: "/api/orders"
      agent_id: "order-service"
```

### Example 2: RAG System

```yaml
name: "rag-multi-agent"
agents:
  document-processor:
    type: "tool"
    model: "gpt-4"
    system_prompt: "Process and index documents"
    tools: ["document_loader", "text_splitter", "embeddings"]
  
  knowledge-retriever:
    type: "react"
    model: "gpt-3.5-turbo"
    system_prompt: "Retrieve relevant knowledge"
    tools: ["vector_search", "reranker"]
  
  answer-generator:
    type: "chat"
    model: "gpt-4"
    system_prompt: "Generate answers from retrieved context"
    tools: ["summarizer"]

routing:
  type: "path"
  rules:
    - pattern: "/ingest"
      agent_id: "document-processor"
    - pattern: "/search"
      agent_id: "knowledge-retriever"
    - pattern: "/answer"
      agent_id: "answer-generator"
```

### Example 3: Workflow System

```yaml
name: "workflow-multi-agent"
agents:
  input-validator:
    type: "tool"
    system_prompt: "Validate and preprocess input data"
    tools: ["validator", "preprocessor"]
  
  task-planner:
    type: "react"
    system_prompt: "Plan task execution workflow"
    tools: ["planner", "scheduler"]
  
  executor:
    type: "react"
    system_prompt: "Execute planned tasks"
    tools: ["executor", "monitor"]
  
  result-aggregator:
    type: "tool"
    system_prompt: "Aggregate and format results"
    tools: ["aggregator", "formatter"]

routing:
  type: "path"
  rules:
    - pattern: "/validate"
      agent_id: "input-validator"
    - pattern: "/plan"
      agent_id: "task-planner"
    - pattern: "/execute"
      agent_id: "executor"
    - pattern: "/results"
      agent_id: "result-aggregator"
```

## Best Practices

### Configuration Management

1. **Use Environment Variables**: Store sensitive data in environment variables
2. **Version Control**: Keep configurations in version control
3. **Validation**: Always validate configurations before deployment
4. **Documentation**: Document agent purposes and routing rules

### Deployment

1. **Gradual Rollout**: Deploy new versions gradually
2. **Health Checks**: Configure proper health checks for all agents
3. **Resource Limits**: Set appropriate CPU and memory limits
4. **Monitoring**: Enable comprehensive monitoring and alerting

### Security

1. **Authentication**: Enable API key authentication for production
2. **HTTPS**: Use HTTPS in production deployments
3. **Rate Limiting**: Configure rate limiting to prevent abuse
4. **Network Policies**: Use network policies in Kubernetes

### Performance

1. **Caching**: Enable caching for frequently accessed data
2. **Connection Pooling**: Use connection pooling for databases
3. **Load Balancing**: Distribute load across multiple replicas
4. **Auto-scaling**: Configure auto-scaling based on metrics

## Troubleshooting

### Common Issues

#### Agent Not Responding

```bash
# Check agent status
golanggraph multi-agent status configs/multi-agent.yaml

# Check logs
kubectl logs deployment/golanggraph-multi-agent

# Check health
curl http://localhost:8080/health/agent-id
```

#### Routing Issues

```bash
# Verify routing configuration
golanggraph multi-agent validate configs/multi-agent.yaml --strict

# Check routing rules
curl http://localhost:8080/routing
```

#### Performance Issues

```bash
# Check metrics
curl http://localhost:8080/metrics

# Monitor resource usage
kubectl top pods

# Check scaling configuration
kubectl describe hpa golanggraph-multi-agent
```

### Debug Mode

Enable debug mode for detailed logging:

```yaml
shared:
  logging:
    level: "debug"
    structured: true
```

### Validation Errors

Common validation errors and solutions:

1. **Missing Agent ID**: Ensure each agent has a unique ID
2. **Invalid Routing Rules**: Check that agent IDs in routing rules exist
3. **Port Conflicts**: Ensure no port conflicts in networking configuration
4. **Resource Constraints**: Verify resource requests and limits are reasonable

## API Reference

### Agent Execution

```bash
POST /{agent-path}
Content-Type: application/json

{
  "input": "Your message here"
}
```

### Management Endpoints

- `GET /health` - Overall system health
- `GET /health/{agent-id}` - Specific agent health
- `GET /metrics` - Prometheus metrics
- `GET /agents` - List all agents
- `GET /agents/{agent-id}` - Get agent details
- `GET /agents/{agent-id}/status` - Get agent status
- `GET /config` - Get configuration
- `GET /routing` - Get routing configuration
- `GET /deployment/status` - Get deployment status
- `POST /deployment/restart` - Restart deployment

### Response Format

```json
{
  "agent_id": "chat-agent",
  "execution": {
    "id": "exec-12345",
    "timestamp": "2024-01-01T12:00:00Z",
    "input": "Hello!",
    "output": "Hello! How can I help you?",
    "duration": "0.5s",
    "success": true
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Migration Guide

### From Single Agent to Multi-Agent

1. **Backup Configuration**: Save your current single-agent configuration
2. **Initialize Multi-Agent**: Create a new multi-agent project
3. **Migrate Configuration**: Move single-agent config to multi-agent format
4. **Update Routing**: Configure routing rules for the migrated agent
5. **Test**: Validate and test the new configuration
6. **Deploy**: Deploy the multi-agent system

### Configuration Migration

```bash
# Convert single agent config to multi-agent
golanggraph migrate single-to-multi agent-config.yaml multi-agent-config.yaml
```

For more detailed information, see the [API Documentation](./API_REFERENCE.md) and [Examples](../examples/).
