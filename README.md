# GoLangGraph

[![CI](https://github.com/piotrlaczkowski/GoLangGraph/actions/workflows/ci.yml/badge.svg)](https://github.com/piotrlaczkowski/GoLangGraph/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/piotrlaczkowski/GoLangGraph/branch/main/graph/badge.svg)](https://codecov.io/gh/piotrlaczkowski/GoLangGraph)
[![Go Report Card](https://goreportcard.com/badge/github.com/piotrlaczkowski/GoLangGraph)](https://goreportcard.com/report/github.com/piotrlaczkowski/GoLangGraph)
[![GoDoc](https://godoc.org/github.com/piotrlaczkowski/GoLangGraph?status.svg)](https://godoc.org/github.com/piotrlaczkowski/GoLangGraph)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A powerful and flexible Go library for building AI agent workflows using graph-based execution patterns. GoLangGraph provides a robust framework for creating, managing, and executing complex AI agent workflows with built-in persistence, streaming, and observability features.

## 🚀 Features

- **Graph-Based Execution**: Build complex workflows using nodes and edges
- **Conditional Routing**: Dynamic execution paths based on runtime conditions
- **State Management**: Persistent state across workflow executions
- **Database Integration**: Support for PostgreSQL, Redis, and vector databases
- **RAG Support**: Built-in Retrieval-Augmented Generation capabilities
- **LLM Integration**: Support for OpenAI, Ollama, and Gemini providers
- **Streaming Execution**: Real-time workflow execution monitoring
- **Checkpointing**: Save and restore workflow states
- **Observability**: Comprehensive logging and metrics
- **Production Ready**: Docker support, CI/CD, and monitoring

## 📦 Installation

```bash
go get github.com/piotrlaczkowski/GoLangGraph
```

## 🏃 Quick Start

### Basic Graph Execution

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

func main() {
    // Create a new graph
    graph := core.NewGraph("my_workflow")

    // Define node functions
    node1 := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        state.Set("step1_completed", true)
        state.Set("message", "Hello from Node 1")
        return state, nil
    }

    node2 := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        msg, _ := state.Get("message")
        state.Set("final_message", fmt.Sprintf("%s -> Node 2", msg))
        return state, nil
    }

    // Build the graph
    graph.AddNode("node1", "First Node", node1)
    graph.AddNode("node2", "Second Node", node2)
    graph.AddEdge("node1", "node2", nil)
    graph.SetStartNode("node1")
    graph.AddEndNode("node2")

    // Execute the workflow
    initialState := core.NewBaseState()
    result, err := graph.Execute(context.Background(), initialState)
    if err != nil {
        log.Fatal(err)
    }

    // Get the final result
    finalMsg, _ := result.Get("final_message")
    fmt.Printf("Final result: %s\n", finalMsg)
}
```

### Conditional Routing

```go
// Define a conditional edge function
condition := func(ctx context.Context, state *core.BaseState) (string, error) {
    value, _ := state.Get("decision")
    if value == "path_a" {
        return "nodeA", nil
    }
    return "nodeB", nil
}

// Add conditional edges
graph.AddEdge("decision_node", "nodeA", condition)
graph.AddEdge("decision_node", "nodeB", condition)
```

### AI Agent with LLM Integration

```go
package main

import (
    "context"
    "log"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

func main() {
    // Create OpenAI provider
    provider, err := llm.NewOpenAIProvider(llm.OpenAIConfig{
        APIKey: "your-api-key",
        Model:  "gpt-4",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create agent
    agent := agent.NewAgent("my_agent", provider)

    // Add tools and behaviors
    agent.AddTool("search", searchTool)
    agent.AddTool("calculator", calculatorTool)

    // Execute agent workflow
    response, err := agent.Execute(context.Background(), "Analyze the market trends")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Agent response: %s\n", response)
}
```

## 🏗️ Architecture

GoLangGraph is built with a modular architecture:

```
├── pkg/
│   ├── core/           # Core graph execution engine
│   ├── agent/          # AI agent framework
│   ├── llm/            # LLM provider integrations
│   ├── persistence/    # Database and storage
│   ├── tools/          # Built-in tools and utilities
│   ├── server/         # HTTP server and API
│   └── debug/          # Debugging and visualization
```

### Core Components

- **Graph Engine**: Manages workflow execution, state transitions, and routing
- **State Management**: Handles persistent state across executions
- **LLM Providers**: Abstractions for different AI model providers
- **Persistence Layer**: Database connections and checkpointing
- **Agent Framework**: High-level AI agent abstractions

## 🔧 Configuration

### Database Configuration

```go
import "github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"

// PostgreSQL configuration
pgConfig := persistence.PostgreSQLConfig{
    Host:     "localhost",
    Port:     5432,
    Database: "golanggraph",
    Username: "user",
    Password: "password",
}

// Redis configuration
redisConfig := persistence.RedisConfig{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    Database: 0,
}

// Create database manager
dbManager := persistence.NewDatabaseManager()
dbManager.AddPostgreSQL("main", pgConfig)
dbManager.AddRedis("cache", redisConfig)
```

### LLM Provider Configuration

```go
import "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"

// OpenAI
openaiProvider, err := llm.NewOpenAIProvider(llm.OpenAIConfig{
    APIKey:      "your-api-key",
    Model:       "gpt-4",
    Temperature: 0.7,
    MaxTokens:   1000,
})

// Ollama (local)
ollamaProvider, err := llm.NewOllamaProvider(llm.OllamaConfig{
    BaseURL: "http://localhost:11434",
    Model:   "llama2",
})

// Gemini
geminiProvider, err := llm.NewGeminiProvider(llm.GeminiConfig{
    APIKey: "your-gemini-api-key",
    Model:  "gemini-pro",
})
```

## 📊 Persistence & RAG

### Vector Database Integration

```go
import "github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"

// Configure pgvector for RAG
pgvectorConfig := persistence.PgVectorConfig{
    Host:       "localhost",
    Port:       5432,
    Database:   "vectordb",
    Username:   "user",
    Password:   "password",
    Dimensions: 1536, // OpenAI embedding dimensions
}

// Create vector store
vectorStore, err := persistence.NewPgVectorStore(pgvectorConfig)
if err != nil {
    log.Fatal(err)
}

// Store documents
documents := []persistence.Document{
    {
        ID:      "doc1",
        Content: "This is important information about AI",
        Metadata: map[string]interface{}{
            "source": "manual",
            "type":   "knowledge",
        },
    },
}

err = vectorStore.StoreDocuments(documents)
if err != nil {
    log.Fatal(err)
}

// Search similar documents
results, err := vectorStore.SimilaritySearch("AI information", 5)
if err != nil {
    log.Fatal(err)
}
```

### Checkpointing

```go
import "github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"

// Create checkpointer
checkpointer, err := persistence.NewDatabaseCheckpointer(dbManager, "main")
if err != nil {
    log.Fatal(err)
}

// Save checkpoint
checkpoint := &persistence.Checkpoint{
    ThreadID:  "thread-123",
    State:     state,
    Metadata:  metadata,
    Timestamp: time.Now(),
}

err = checkpointer.SaveCheckpoint(checkpoint)
if err != nil {
    log.Fatal(err)
}

// Load checkpoint
loadedCheckpoint, err := checkpointer.LoadCheckpoint("thread-123")
if err != nil {
    log.Fatal(err)
}
```

## 🔄 Streaming & Real-time Execution

```go
// Enable streaming in graph configuration
graph.Config.EnableStreaming = true

// Get streaming channel
streamChan := graph.Stream()

// Execute in background
go func() {
    _, err := graph.Execute(context.Background(), initialState)
    if err != nil {
        log.Printf("Execution error: %v", err)
    }
}()

// Listen for execution updates
for result := range streamChan {
    fmt.Printf("Node %s completed in %v\n", result.NodeID, result.Duration)
    if result.Error != nil {
        fmt.Printf("Error: %v\n", result.Error)
    }
}
```

## 🛠️ Development

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL 14+ (for persistence)
- Redis 6+ (for caching)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/piotrlaczkowski/GoLangGraph.git
cd GoLangGraph

# Install dependencies
go mod tidy

# Start development services
make dev-up

# Run tests
make test

# Run with coverage
make test-coverage

# Lint code
make lint

# Format code
make fmt
```

### Running Examples

```bash
# Quick start demo
make run-example EXAMPLE=quick_start_demo

# Database persistence demo
make run-example EXAMPLE=database_persistence_demo

# Simple agent demo
make run-example EXAMPLE=simple_agent

# Ultimate minimal demo
make run-example EXAMPLE=ultimate_minimal_demo
```

### Docker Development

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## 🧪 Testing

The project includes comprehensive tests across all packages:

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/core -v
go test ./pkg/llm -v
go test ./pkg/persistence -v

# Run integration tests
make test-integration

# Run benchmarks
make benchmark

# Generate coverage report
make test-coverage
```

## 📈 Performance

GoLangGraph is designed for high performance:

- **Concurrent Execution**: Parallel node execution where possible
- **Efficient State Management**: Optimized state copying and merging
- **Connection Pooling**: Database connection reuse
- **Streaming**: Real-time execution monitoring without blocking
- **Memory Optimization**: Efficient memory usage patterns

### Benchmarks

```bash
# Run performance benchmarks
make benchmark

# Example results:
BenchmarkGraph_Execute-8           1000000    1.2 ms/op    512 B/op    8 allocs/op
BenchmarkState_Set-8              10000000    120 ns/op     48 B/op    1 allocs/op
BenchmarkLLM_Complete-8              1000    1.5 s/op    1024 B/op   12 allocs/op
```

## 🔐 Security

- **Input Validation**: All inputs are validated and sanitized
- **SQL Injection Prevention**: Parameterized queries throughout
- **API Key Management**: Secure credential handling
- **Access Control**: Role-based permissions (coming soon)
- **Audit Logging**: Comprehensive execution logging

## 📚 Examples

### Advanced Agent with Tools

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
    // Create LLM provider
    provider, err := llm.NewOpenAIProvider(llm.OpenAIConfig{
        APIKey: "your-api-key",
        Model:  "gpt-4",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create agent with tools
    agent := agent.NewAgent("research_agent", provider)
    
    // Add web search tool
    searchTool := tools.NewWebSearchTool("your-search-api-key")
    agent.AddTool("web_search", searchTool)
    
    // Add calculator tool
    calcTool := tools.NewCalculatorTool()
    agent.AddTool("calculator", calcTool)
    
    // Add custom tool
    weatherTool := &tools.CustomTool{
        Name:        "weather",
        Description: "Get current weather for a location",
        Function: func(args map[string]interface{}) (interface{}, error) {
            location := args["location"].(string)
            // Implement weather API call
            return getWeather(location), nil
        },
    }
    agent.AddTool("weather", weatherTool)

    // Execute complex task
    task := `Research the current market trends for renewable energy, 
             calculate the projected growth rate, and provide a summary 
             including weather patterns that might affect solar energy.`
    
    response, err := agent.Execute(context.Background(), task)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Agent Response:\n%s\n", response)
}

func getWeather(location string) map[string]interface{} {
    // Mock weather data
    return map[string]interface{}{
        "location":    location,
        "temperature": 22,
        "condition":   "sunny",
        "humidity":    65,
    }
}
```

### Multi-Agent Collaboration

```go
package main

import (
    "context"
    "log"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/core"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

func main() {
    // Create providers
    provider, err := llm.NewOpenAIProvider(llm.OpenAIConfig{
        APIKey: "your-api-key",
        Model:  "gpt-4",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create specialized agents
    researchAgent := agent.NewAgent("researcher", provider)
    researchAgent.SetSystemPrompt("You are a research specialist. Focus on gathering and analyzing information.")

    writerAgent := agent.NewAgent("writer", provider)
    writerAgent.SetSystemPrompt("You are a technical writer. Create clear, well-structured documents.")

    reviewerAgent := agent.NewAgent("reviewer", provider)
    reviewerAgent.SetSystemPrompt("You are a quality reviewer. Ensure accuracy and completeness.")

    // Create collaboration graph
    graph := core.NewGraph("multi_agent_workflow")

    // Research phase
    researchNode := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        topic, _ := state.Get("topic")
        research, err := researchAgent.Execute(ctx, fmt.Sprintf("Research: %s", topic))
        if err != nil {
            return nil, err
        }
        state.Set("research_results", research)
        return state, nil
    }

    // Writing phase
    writeNode := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        research, _ := state.Get("research_results")
        document, err := writerAgent.Execute(ctx, fmt.Sprintf("Write a document based on: %s", research))
        if err != nil {
            return nil, err
        }
        state.Set("draft_document", document)
        return state, nil
    }

    // Review phase
    reviewNode := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        draft, _ := state.Get("draft_document")
        review, err := reviewerAgent.Execute(ctx, fmt.Sprintf("Review and improve: %s", draft))
        if err != nil {
            return nil, err
        }
        state.Set("final_document", review)
        return state, nil
    }

    // Build workflow
    graph.AddNode("research", "Research Phase", researchNode)
    graph.AddNode("write", "Writing Phase", writeNode)
    graph.AddNode("review", "Review Phase", reviewNode)
    
    graph.AddEdge("research", "write", nil)
    graph.AddEdge("write", "review", nil)
    
    graph.SetStartNode("research")
    graph.AddEndNode("review")

    // Execute workflow
    initialState := core.NewBaseState()
    initialState.Set("topic", "Artificial Intelligence in Healthcare")

    result, err := graph.Execute(context.Background(), initialState)
    if err != nil {
        log.Fatal(err)
    }

    finalDoc, _ := result.Get("final_document")
    fmt.Printf("Final Document:\n%s\n", finalDoc)
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run tests: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Code Style

- Follow Go best practices and idioms
- Use `gofmt` for formatting
- Write comprehensive tests
- Add documentation for public APIs
- Use meaningful variable and function names

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [GoDoc](https://godoc.org/github.com/piotrlaczkowski/GoLangGraph)
- **Issues**: [GitHub Issues](https://github.com/piotrlaczkowski/GoLangGraph/issues)
- **Discussions**: [GitHub Discussions](https://github.com/piotrlaczkowski/GoLangGraph/discussions)
- **Email**: support@golanggraph.dev

## 🙏 Acknowledgments

- Inspired by LangGraph and similar workflow engines
- Built with the excellent Go ecosystem
- Special thanks to all contributors

## 🗺️ Roadmap

- [ ] **v1.1**: Enhanced RAG capabilities
- [ ] **v1.2**: Multi-modal support (images, audio)
- [ ] **v1.3**: Distributed execution
- [ ] **v1.4**: Visual workflow editor
- [ ] **v1.5**: Advanced monitoring and analytics
- [ ] **v2.0**: Cloud-native deployment options

---

**GoLangGraph** - Building the future of AI agent workflows in Go! 🚀 