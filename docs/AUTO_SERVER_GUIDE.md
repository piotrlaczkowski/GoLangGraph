# GoLangGraph Auto Server & Monitoring Guide

## Overview

The GoLangGraph Auto Server automatically generates REST APIs for your AI agents, complete with web interfaces, monitoring, and production-ready features.

## Quick Start

### Basic Auto Server

```go
package main

import (
    "context"
    "log"
    
    "github.com/piotrlaczkowski/GoLangGraph/pkg/server"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
)

func main() {
    // Create auto server with default config
    config := server.DefaultAutoServerConfig()
    config.Port = 8080
    config.EnableWebUI = true
    config.EnablePlayground = true
    config.EnableMetricsAPI = true
    
    autoServer := server.NewAutoServer(config)
    
    // Register agents
    chatConfig := &agent.AgentConfig{
        Name:     "ChatAgent",
        Type:     agent.AgentTypeChat,
        Model:    "gemma3:1b",
        Provider: "ollama",
    }
    chatDefinition := agent.NewBaseAgentDefinition(chatConfig)
    autoServer.RegisterAgent("chat", chatDefinition)
    
    // Generate endpoints
    if err := autoServer.GenerateEndpoints(); err != nil {
        log.Fatal(err)
    }
    
    // Start server
    ctx := context.Background()
    autoServer.Start(ctx)
}
```

## Generated Endpoints

### Core System Endpoints

- `GET /health` - Health check
- `GET /capabilities` - Server capabilities
- `GET /agents` - List all agents
- `GET /agents/{id}` - Agent information
- `GET /metrics` - System metrics

### Agent Endpoints (per agent)

- `POST /api/{agent-id}` - Execute agent
- `POST /api/{agent-id}/stream` - Streaming execution
- `GET /api/{agent-id}/conversation` - Get conversation history
- `POST /api/{agent-id}/conversation` - Continue conversation
- `DELETE /api/{agent-id}/conversation` - Clear conversation
- `GET /api/{agent-id}/status` - Agent status

### Web Interfaces

- `GET /` - Main chat interface
- `GET /chat` - Chat interface (alias)
- `GET /playground` - API testing playground
- `GET /debug` - Debug interface

### Schema & Validation

- `GET /schemas` - All agent schemas
- `GET /schemas/{agent-id}` - Specific agent schema
- `POST /validate/{agent-id}` - Validate request schema

## Configuration Options

```go
type AutoServerConfig struct {
    Host             string                 `yaml:"host" json:"host"`
    Port             int                    `yaml:"port" json:"port"`
    BasePath         string                 `yaml:"base_path" json:"base_path"`
    EnableWebUI      bool                   `yaml:"enable_web_ui" json:"enable_web_ui"`
    EnablePlayground bool                   `yaml:"enable_playground" json:"enable_playground"`
    EnableSchemaAPI  bool                   `yaml:"enable_schema_api" json:"enable_schema_api"`
    EnableMetricsAPI bool                   `yaml:"enable_metrics_api" json:"enable_metrics_api"`
    EnableCORS       bool                   `yaml:"enable_cors" json:"enable_cors"`
    SchemaValidation bool                   `yaml:"schema_validation" json:"schema_validation"`
    OllamaEndpoint   string                 `yaml:"ollama_endpoint" json:"ollama_endpoint"`
    ServerTimeout    time.Duration          `yaml:"server_timeout" json:"server_timeout"`
    MaxRequestSize   int64                  `yaml:"max_request_size" json:"max_request_size"`
    Middleware       []string               `yaml:"middleware" json:"middleware"`
}
```

## Monitoring & Observability

### Prometheus Metrics

The auto server exposes Prometheus metrics at `/metrics`:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `agent_executions_total` - Agent execution count
- `agent_execution_duration_seconds` - Agent execution time
- `system_uptime_seconds` - Server uptime

### Example Integration with Grafana

For the comprehensive monitoring setup, see the `examples/10-ideation-agents/go-agents-simple` example which includes:

- **Prometheus Configuration** - Metrics collection
- **Grafana Dashboards** - Visual monitoring
- **Alerting Rules** - Automated alerts
- **Docker Compose** - Complete monitoring stack

### Available Dashboards

1. **System Overview** - Overall system health and performance
2. **Agent Performance** - Individual agent metrics and success rates
3. **Infrastructure** - System resources and container metrics

### Sample Monitoring Setup

```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    image: go-agents-simple
    ports:
      - "8080:8080"
    environment:
      - OLLAMA_ENDPOINT=http://ollama:11434
    depends_on:
      - ollama
      - prometheus

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./monitoring/alerting.yml:/etc/prometheus/alerting.yml

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - ./monitoring/dashboards:/var/lib/grafana/dashboards
      - ./monitoring/provisioning:/etc/grafana/provisioning
```

## Web Interface Features

### Chat Interface

The built-in chat interface provides:

- 💬 **Real-time Chat** - Interactive conversation with agents
- 🎨 **Modern UI** - Clean, responsive design
- 📱 **Mobile Friendly** - Works on all devices
- 🔄 **Agent Switching** - Choose between different agents
- 📝 **Conversation History** - Persistent chat history

### API Playground

The playground interface offers:

- 🧪 **Interactive Testing** - Test all endpoints
- 📋 **Request/Response** - See full API communication
- 🔧 **Parameter Editing** - Modify request parameters
- 📚 **API Documentation** - Built-in endpoint documentation

## Production Deployment

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Health Checks

```yaml
# docker-compose.yml
services:
  app:
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
```

### Environment Configuration

```bash
# Required environment variables
OLLAMA_ENDPOINT=http://localhost:11434
SERVER_PORT=8080
ENABLE_MONITORING=true

# Optional configurations
LOG_LEVEL=info
MAX_REQUEST_SIZE=10485760  # 10MB
SERVER_TIMEOUT=30s
```

## Security Considerations

### CORS Configuration

```go
config.EnableCORS = true
config.Middleware = []string{"cors", "logging", "recovery"}
```

### Request Validation

```go
config.SchemaValidation = true
config.MaxRequestSize = 10 * 1024 * 1024  // 10MB
```

### Rate Limiting

For production deployment, consider adding rate limiting middleware:

```go
// Custom middleware example
func rateLimitMiddleware() func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(10, 100) // 10 requests per second, burst 100
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

## Best Practices

1. **Always enable health checks** for container orchestration
2. **Use structured logging** for better observability
3. **Implement graceful shutdown** for production deployments
4. **Monitor resource usage** with the provided dashboards
5. **Set appropriate timeouts** based on your use case
6. **Enable CORS** only when necessary for security
7. **Validate all inputs** to prevent injection attacks
8. **Use HTTPS** in production environments

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   Error: listen tcp :8080: bind: address already in use
   ```
   Solution: Change the port in configuration or kill the process using the port.

2. **Ollama connection failed**
   ```bash
   Error: failed to connect to Ollama at http://localhost:11434
   ```
   Solution: Ensure Ollama is running and the endpoint is correct.

3. **Agent registration failed**
   ```bash
   Error: failed to register agent: invalid configuration
   ```
   Solution: Check agent configuration and ensure all required fields are set.

### Debug Mode

Enable debug logging:

```go
config.LogLevel = "debug"
```

Check debug interface:
```bash
curl http://localhost:8080/debug
```

## Examples

See the complete examples in:
- `examples/10-ideation-agents/go-agents-simple/` - Production-ready auto server with monitoring
- `examples/08-production-ready/` - Production deployment patterns

## API Reference

For complete API documentation, visit the playground interface at `http://localhost:8080/playground` when your server is running. 
