# GoLangGraph

A comprehensive Go implementation of the LangGraph Python framework for building stateful, multi-agent conversational AI applications.

## 🚀 Features

### Core Framework
- **Stateful Graph Execution**: Pregel-inspired graph execution engine with state management
- **Multi-Agent Support**: Coordinate multiple AI agents with different capabilities
- **ReAct Agents**: Reasoning and Acting agents with tool integration
- **State Persistence**: Database-backed state persistence with time travel capabilities
- **Visual Debugging**: Real-time graph visualization and execution tracing

### LLM Provider Support
- **OpenAI**: Complete integration with OpenAI API (GPT-3.5, GPT-4)
- **Ollama**: Local LLM support for privacy-focused deployments
- **Google Gemini**: Integration with Google's Gemini API
- **Extensible**: Easy to add new LLM providers

### Tools & Integrations
- **Built-in Tools**: Calculator, Web Search, File Operations, Shell Commands, HTTP Requests
- **Custom Tools**: Extensible tool system for domain-specific functionality
- **Tool Registry**: Centralized tool management and configuration

### Deployment & Operations
- **HTTP API Server**: REST endpoints for agent and graph management
- **WebSocket Streaming**: Real-time execution streaming
- **Database Support**: PostgreSQL and Redis for state persistence
- **CLI Tools**: Command-line interface for deployment and management
- **Health Monitoring**: Built-in health checks and monitoring

## 📁 Project Structure

```
GoLangGraph/
├── pkg/
│   ├── core/           # Core graph execution engine
│   │   ├── graph.go    # Graph structure and execution
│   │   └── state.go    # State management
│   ├── llm/            # LLM provider implementations
│   │   ├── provider.go # Provider interface and manager
│   │   ├── openai.go   # OpenAI integration
│   │   ├── ollama.go   # Ollama integration
│   │   └── gemini.go   # Google Gemini integration
│   ├── agent/          # Agent implementations
│   │   └── agent.go    # ReAct, Chat, and Tool agents
│   ├── tools/          # Tool implementations
│   │   └── tools.go    # Built-in tools and registry
│   ├── persistence/    # State persistence
│   │   ├── checkpointer.go # Memory and file checkpointers
│   │   └── database.go     # PostgreSQL and Redis persistence
│   ├── server/         # HTTP API server
│   │   └── server.go   # REST API and WebSocket endpoints
│   └── debug/          # Debugging and visualization
│       └── visualizer.go # Graph visualization tools
├── cmd/
│   └── golanggraph/    # CLI application
│       └── main.go
├── examples/           # Usage examples
│   └── simple_agent.go
├── go.mod             # Go module definition
└── README.md          # This file
```

## 🛠 Installation

### Prerequisites
- Go 1.19 or later
- PostgreSQL (optional, for database persistence)
- Redis (optional, for fast caching)
- Ollama (optional, for local LLM support)

### Install Dependencies

```bash
go mod download
```

### Environment Variables

```bash
# OpenAI API Key (optional)
export OPENAI_API_KEY="your-openai-api-key"

# Ollama URL (optional, defaults to localhost:11434)
export OLLAMA_URL="http://localhost:11434"

# Google Gemini API Key (optional)
export GEMINI_API_KEY="your-gemini-api-key"

# Database Configuration (optional)
export DATABASE_URL="postgres://user:password@localhost:5432/golanggraph"
export REDIS_URL="redis://localhost:6379"
```

## 🚀 Quick Start

### 1. Basic Agent Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
    // Initialize LLM provider
    llmManager := llm.NewProviderManager()
    
    // Initialize tools
    toolRegistry := tools.NewToolRegistry()
    toolRegistry.RegisterTool(tools.NewCalculatorTool())
    
    // Create agent configuration
    config := &agent.AgentConfig{
        Name:         "helpful-assistant",
        Type:         agent.AgentTypeReAct,
        Model:        "gpt-3.5-turbo",
        Provider:     "openai",
        SystemPrompt: "You are a helpful assistant.",
        Temperature:  0.7,
        MaxTokens:    1000,
        Tools:        []string{"calculator"},
        Timeout:      30 * time.Second,
    }
    
    // Create and execute agent
    agent := agent.NewAgent(config, llmManager, toolRegistry)
    
    ctx := context.Background()
    execution, err := agent.Execute(ctx, "What is 25 * 34?")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %s\n", execution.Output)
}
```

### 2. Start the HTTP Server

```bash
# Start the server
go run cmd/golanggraph/main.go serve --port 8080

# Or with custom configuration
go run cmd/golanggraph/main.go serve \
    --host 0.0.0.0 \
    --port 8080 \
    --static-dir ./static \
    --enable-cors
```

### 3. API Usage

```bash
# Health check
curl http://localhost:8080/api/v1/health

# List providers
curl http://localhost:8080/api/v1/providers

# Create an agent
curl -X POST http://localhost:8080/api/v1/agents \
    -H "Content-Type: application/json" \
    -d '{
        "name": "test-agent",
        "type": "react",
        "model": "gpt-3.5-turbo",
        "provider": "openai",
        "system_prompt": "You are a helpful assistant.",
        "temperature": 0.7,
        "max_tokens": 1000,
        "tools": ["calculator"]
    }'

# Execute agent
curl -X POST http://localhost:8080/api/v1/agents/{agent-id}/execute \
    -H "Content-Type: application/json" \
    -d '{
        "input": "What is 25 * 34?",
        "stream": false
    }'
```

## 🔧 Configuration

### Agent Configuration

```go
type AgentConfig struct {
    ID              string                 `json:"id"`
    Name            string                 `json:"name"`
    Type            AgentType              `json:"type"`           // "react", "chat", "tool"
    Model           string                 `json:"model"`          // "gpt-3.5-turbo", "llama2", etc.
    Provider        string                 `json:"provider"`       // "openai", "ollama", "gemini"
    SystemPrompt    string                 `json:"system_prompt"`
    Temperature     float64                `json:"temperature"`
    MaxTokens       int                    `json:"max_tokens"`
    MaxIterations   int                    `json:"max_iterations"`
    Tools           []string               `json:"tools"`
    EnableStreaming bool                   `json:"enable_streaming"`
    Timeout         time.Duration          `json:"timeout"`
    Metadata        map[string]interface{} `json:"metadata"`
}
```

### Database Configuration

```go
type DatabaseConfig struct {
    Type         string `json:"type"`         // "postgres", "redis"
    Host         string `json:"host"`
    Port         int    `json:"port"`
    Database     string `json:"database"`
    Username     string `json:"username"`
    Password     string `json:"password"`
    SSLMode      string `json:"ssl_mode"`
    MaxOpenConns int    `json:"max_open_conns"`
    MaxIdleConns int    `json:"max_idle_conns"`
    MaxLifetime  string `json:"max_lifetime"`
}
```

## 🔍 Debugging & Visualization

### Graph Visualization

```bash
# Generate Mermaid diagram
go run cmd/golanggraph/main.go debug visualize --format mermaid

# Generate DOT diagram
go run cmd/golanggraph/main.go debug visualize --format dot --output graph.dot

# Save to file
go run cmd/golanggraph/main.go debug visualize --format mermaid --output graph.mmd
```

### Real-time Debugging

Connect to WebSocket endpoints for real-time execution monitoring:

```javascript
// Connect to agent execution stream
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/agents/{agent-id}/stream');

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Execution step:', data);
};

// Send execution request
ws.send(JSON.stringify({
    type: 'execute',
    input: 'What is the weather like today?'
}));
```

## 🧪 Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/core/
go test ./pkg/agent/
```

### Test Agent Configuration

```bash
# Test agent configuration
go run cmd/golanggraph/main.go test
```

## 📊 Database Setup

### PostgreSQL

```bash
# Run migrations
go run cmd/golanggraph/main.go migrate \
    --db-type postgres \
    --db-host localhost \
    --db-port 5432 \
    --db-name golanggraph \
    --db-user postgres \
    --db-password password
```

### Redis

```bash
# Setup Redis
go run cmd/golanggraph/main.go migrate \
    --db-type redis \
    --db-host localhost \
    --db-port 6379 \
    --db-password ""
```

## 🛡 Security Considerations

- **API Keys**: Store API keys securely using environment variables
- **Authentication**: Implement proper authentication for production deployments
- **Rate Limiting**: Consider implementing rate limiting for API endpoints
- **Input Validation**: Validate all user inputs to prevent injection attacks
- **Network Security**: Use HTTPS in production environments

## 🚀 Deployment

### Docker Deployment

```dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o golanggraph cmd/golanggraph/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/golanggraph .
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./golanggraph", "serve"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: golanggraph
spec:
  replicas: 3
  selector:
    matchLabels:
      app: golanggraph
  template:
    metadata:
      labels:
        app: golanggraph
    spec:
      containers:
      - name: golanggraph
        image: golanggraph:latest
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: golanggraph-secrets
              key: openai-api-key
---
apiVersion: v1
kind: Service
metadata:
  name: golanggraph-service
spec:
  selector:
    app: golanggraph
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by the [LangGraph](https://github.com/langchain-ai/langgraph) Python framework
- Built with Go's excellent concurrency primitives
- Uses the Pregel model for distributed graph computation

## 📚 Documentation

For more detailed documentation, see:

- [API Reference](docs/api.md)
- [Agent Development Guide](docs/agents.md)
- [Tool Development Guide](docs/tools.md)
- [Deployment Guide](docs/deployment.md)
- [Examples](examples/)

## 🐛 Issues & Support

- [GitHub Issues](https://github.com/piotrlaczkowski/GoLangGraph/issues)
- [Discussions](https://github.com/piotrlaczkowski/GoLangGraph/discussions)

---

**GoLangGraph** - Building the future of stateful AI applications in Go! 🚀 