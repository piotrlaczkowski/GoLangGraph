# Quick Start Guide

Get started with GoLangGraph in minutes! This guide will walk you through setting up your first AI agent.

## Prerequisites

- Go 1.21 or later
- Ollama (for local LLM inference) - optional
- Docker (for containerized deployment) - optional

## Installation

### 1. Install GoLangGraph

```bash
go mod init my-agent-app
go get github.com/piotrlaczkowski/GoLangGraph
```

### 2. Set up Ollama (Recommended)

Install Ollama and pull a model:

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a lightweight model
ollama pull gemma3:1b

# Start Ollama server
ollama serve
```

## Your First Agent

Create a simple chat agent:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
    // Create LLM manager
    llmManager := llm.NewLLMManager()
    
    // Add Ollama provider
    err := llmManager.AddProvider("ollama", &llm.OllamaConfig{
        BaseURL: "http://localhost:11434",
    })
    if err != nil {
        log.Fatal("Failed to add Ollama provider:", err)
    }
    
    // Create tool registry
    toolRegistry := tools.NewToolRegistry()
    
    // Create agent configuration
    config := &agent.AgentConfig{
        Name:         "my-chat-agent",
        Type:         agent.AgentTypeChat,
        Model:        "gemma3:1b",
        Provider:     "ollama",
        SystemPrompt: "You are a helpful assistant.",
    }
    
    // Create the agent
    chatAgent := agent.NewAgent(config, llmManager, toolRegistry)
    
    // Execute a simple chat
    execution, err := chatAgent.Execute(context.Background(), "Hello! What can you help me with?")
    if err != nil {
        log.Fatal("Failed to execute agent:", err)
    }
    
    fmt.Printf("Agent response: %s\n", execution.Response)
}
```

Run your agent:

```bash
go run main.go
```

## Adding Tools

Enhance your agent with built-in tools:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
    // Create LLM manager
    llmManager := llm.NewLLMManager()
    
    // Add Ollama provider
    err := llmManager.AddProvider("ollama", &llm.OllamaConfig{
        BaseURL: "http://localhost:11434",
    })
    if err != nil {
        log.Fatal("Failed to add Ollama provider:", err)
    }
    
    // Create tool registry and add tools
    toolRegistry := tools.NewToolRegistry()
    
    // Add calculator tool
    calculator := tools.NewCalculatorTool()
    toolRegistry.RegisterTool("calculator", calculator)
    
    // Add web search tool (optional - requires API key)
    // webSearch := tools.NewWebSearchTool("your-api-key")
    // toolRegistry.RegisterTool("web_search", webSearch)
    
    // Create ReAct agent (can use tools)
    config := &agent.AgentConfig{
        Name:         "my-react-agent",
        Type:         agent.AgentTypeReAct,
        Model:        "gemma3:1b",
        Provider:     "ollama",
        SystemPrompt: "You are a helpful assistant with access to tools. Use tools when needed to provide accurate information.",
    }
    
    // Create the agent
    reactAgent := agent.NewAgent(config, llmManager, toolRegistry)
    
    // Execute with a calculation request
    execution, err := reactAgent.Execute(context.Background(), "What is 15 * 24 + 137?")
    if err != nil {
        log.Fatal("Failed to execute agent:", err)
    }
    
    fmt.Printf("Agent response: %s\n", execution.Response)
}
```

## Building a Workflow Graph

Create a multi-step workflow:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/piotrlaczkowski/GoLangGraph/pkg/core"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
    // Create LLM manager
    llmManager := llm.NewLLMManager()
    
    // Add Ollama provider
    err := llmManager.AddProvider("ollama", &llm.OllamaConfig{
        BaseURL: "http://localhost:11434",
    })
    if err != nil {
        log.Fatal("Failed to add Ollama provider:", err)
    }
    
    // Create tool registry
    toolRegistry := tools.NewToolRegistry()
    
    // Create graph
    graph := core.NewGraph()
    
    // Add nodes (agents)
    analyzerAgent := agent.NewAgent(&agent.AgentConfig{
        Name:         "analyzer",
        Type:         agent.AgentTypeChat,
        Model:        "gemma3:1b",
        Provider:     "ollama",
        SystemPrompt: "You analyze user requests and extract key information.",
    }, llmManager, toolRegistry)
    
    processorAgent := agent.NewAgent(&agent.AgentConfig{
        Name:         "processor",
        Type:         agent.AgentTypeChat,
        Model:        "gemma3:1b",
        Provider:     "ollama",
        SystemPrompt: "You process information and provide detailed responses.",
    }, llmManager, toolRegistry)
    
    // Add nodes to graph
    graph.AddNode("analyzer", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        input, _ := state.Get("input")
        execution, err := analyzerAgent.Execute(ctx, fmt.Sprintf("Analyze this request: %s", input))
        if err != nil {
            return nil, err
        }
        state.Set("analysis", execution.Response)
        return state, nil
    })
    
    graph.AddNode("processor", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        analysis, _ := state.Get("analysis")
        execution, err := processorAgent.Execute(ctx, fmt.Sprintf("Process this analysis: %s", analysis))
        if err != nil {
            return nil, err
        }
        state.Set("result", execution.Response)
        return state, nil
    })
    
    // Add edges
    graph.AddEdge("analyzer", "processor")
    
    // Set entry point
    graph.SetEntryPoint("analyzer")
    
    // Execute the graph
    initialState := core.NewBaseState()
    initialState.Set("input", "I need help planning a vacation to Japan")
    
    finalState, err := graph.Execute(context.Background(), initialState)
    if err != nil {
        log.Fatal("Failed to execute graph:", err)
    }
    
    result, _ := finalState.Get("result")
    fmt.Printf("Final result: %s\n", result)
}
```

## Adding Persistence

Save conversation state:

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
    "github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

func main() {
    // Create LLM manager
    llmManager := llm.NewLLMManager()
    
    // Add Ollama provider
    err := llmManager.AddProvider("ollama", &llm.OllamaConfig{
        BaseURL: "http://localhost:11434",
    })
    if err != nil {
        log.Fatal("Failed to add Ollama provider:", err)
    }
    
    // Create tool registry
    toolRegistry := tools.NewToolRegistry()
    
    // Create agent
    config := &agent.AgentConfig{
        Name:         "persistent-agent",
        Type:         agent.AgentTypeChat,
        Model:        "gemma3:1b",
        Provider:     "ollama",
        SystemPrompt: "You are a helpful assistant with memory.",
    }
    
    chatAgent := agent.NewAgent(config, llmManager, toolRegistry)
    
    // Set up persistence
    checkpointer := persistence.NewMemoryCheckpointer()
    sessionManager := persistence.NewSessionManager(checkpointer)
    
    // Create session
    session, err := sessionManager.CreateSession(context.Background(), &persistence.SessionConfig{
        UserID: "user_123",
        AgentID: config.Name,
    })
    if err != nil {
        log.Fatal("Failed to create session:", err)
    }
    
    // Execute with session context
    ctx := context.WithValue(context.Background(), "session_id", session.ID)
    execution, err := chatAgent.Execute(ctx, "Hello! Remember that I like Go programming.")
    if err != nil {
        log.Fatal("Failed to execute agent:", err)
    }
    
    fmt.Printf("Agent response: %s\n", execution.Response)
    
    // Save checkpoint
    checkpoint := &persistence.Checkpoint{
        ID:       fmt.Sprintf("checkpoint_%d", time.Now().Unix()),
        ThreadID: session.ID,
        State:    execution.State,
        CreatedAt: time.Now(),
    }
    
    err = checkpointer.Save(ctx, checkpoint)
    if err != nil {
        log.Printf("Failed to save checkpoint: %v", err)
    }
    
    fmt.Println("Conversation state saved!")
}
```

## Running Examples

The repository includes working examples:

```bash
# Clone the repository
git clone https://github.com/piotrlaczkowski/GoLangGraph.git
cd GoLangGraph

# Run basic chat example
go run examples/01-basic-chat/main.go

# Run ReAct agent example
go run examples/02-react-agent/main.go

# Run multi-agent example
go run examples/03-multi-agent/main.go

# See all examples
ls examples/
```

## Configuration

### Environment Variables

```bash
# Ollama configuration
export OLLAMA_BASE_URL=http://localhost:11434

# OpenAI configuration (optional)
export OPENAI_API_KEY=your-api-key

# Database configuration (optional)
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DB=golanggraph
export POSTGRES_USER=user
export POSTGRES_PASSWORD=password
```

### Configuration File

Create a `config.yaml` file:

```yaml
llm:
  providers:
    ollama:
      base_url: "http://localhost:11434"
    openai:
      api_key: "your-api-key"

agents:
  default_model: "gemma3:1b"
  default_provider: "ollama"

persistence:
  type: "memory"  # or "postgres"
  
tools:
  calculator:
    enabled: true
  web_search:
    enabled: false
    api_key: "your-api-key"
```

## Next Steps

1. **Explore Examples**: Check out the `/examples` directory for more complex use cases
2. **Read Documentation**: Browse the `/docs` directory for detailed guides
3. **Add Custom Tools**: Learn how to create custom tools for your agents
4. **Deploy to Production**: Use the server package for HTTP API deployment
5. **Integrate Databases**: Set up PostgreSQL or Redis for production persistence

## Common Issues

### Ollama Connection Error

```bash
# Make sure Ollama is running
ollama serve

# Check if the model is available
ollama list

# Pull the model if needed
ollama pull gemma3:1b
```

### Memory Issues

For large conversations, consider:
- Using database persistence instead of memory
- Implementing conversation summarization
- Setting conversation length limits

### Performance

- Use lighter models like `gemma3:1b` for development
- Consider GPU acceleration for production
- Implement connection pooling for databases

## Support

- **Examples**: `/examples` directory
- **Documentation**: `/docs` directory
- **Issues**: GitHub Issues for bug reports and feature requests

Happy building with GoLangGraph! 🚀 
