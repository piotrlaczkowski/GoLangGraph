# Quick Start

Get up and running with GoLangGraph in just a few minutes! This guide will walk you through creating your first AI agent workflow.

## Prerequisites

- Go 1.21 or later
- Basic familiarity with Go programming
- API key for an LLM provider (OpenAI, Ollama, etc.)

## Installation

### Option 1: Go Module (Recommended)

```bash
# Initialize your Go module
go mod init my-agent-project

# Add GoLangGraph dependency
go get github.com/piotrlaczkowski/GoLangGraph
```

### Option 2: Clone and Build

```bash
git clone https://github.com/piotrlaczkowski/GoLangGraph.git
cd GoLangGraph
go build ./cmd/golanggraph
```

## Your First Agent

Let's create a simple chat agent that can respond to user messages:

### 1. Create the Basic Agent

```go title="main.go"
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/piotrlaczkowski/GoLangGraph/pkg/builder"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

func main() {
    // Create a simple chat agent using the builder
    agent := builder.OneLineChat("MyFirstAgent")
    
    // Configure with OpenAI (replace with your API key)
    provider, err := llm.NewOpenAIProvider(llm.OpenAIConfig{
        APIKey: "your-openai-api-key",
        Model:  "gpt-3.5-turbo",
    })
    if err != nil {
        log.Fatal("Failed to create OpenAI provider:", err)
    }
    
    // Set the LLM provider for the agent
    agent.SetLLMProvider(provider)
    
    // Execute the agent with a simple message
    ctx := context.Background()
    response, err := agent.Execute(ctx, "Hello! Tell me a joke.")
    if err != nil {
        log.Fatal("Agent execution failed:", err)
    }
    
    fmt.Printf("🤖 Agent Response: %s\n", response.Content)
}
```

### 2. Run Your Agent

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-actual-api-key"

# Run the agent
go run main.go
```

!!! success "Expected Output"
    ```
    🤖 Agent Response: Why don't scientists trust atoms? Because they make up everything!
    ```

## Using Local LLM (Ollama)

If you prefer to use a local LLM, you can use Ollama:

### 1. Install Ollama

```bash
# On macOS
brew install ollama

# On Linux
curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama service
ollama serve
```

### 2. Pull a Model

```bash
# Pull a lightweight model
ollama pull llama3.2:1b

# Or a more capable model
ollama pull llama3.2:3b
```

### 3. Update Your Code

```go title="main.go" hl_lines="18-22"
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/piotrlaczkowski/GoLangGraph/pkg/builder"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

func main() {
    // Create a simple chat agent
    agent := builder.OneLineChat("MyFirstAgent")
    
    // Configure with Ollama (local LLM)
    provider, err := llm.NewOllamaProvider(llm.OllamaConfig{
        BaseURL: "http://localhost:11434",
        Model:   "llama3.2:1b",
    })
    if err != nil {
        log.Fatal("Failed to create Ollama provider:", err)
    }
    
    agent.SetLLMProvider(provider)
    
    // Execute the agent
    ctx := context.Background()
    response, err := agent.Execute(ctx, "Hello! Tell me about Go programming.")
    if err != nil {
        log.Fatal("Agent execution failed:", err)
    }
    
    fmt.Printf("🤖 Agent Response: %s\n", response.Content)
}
```

## Building a More Complex Workflow

Let's create a more sophisticated agent that uses tools and state management:

```go title="advanced_agent.go"
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/core"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
    // Create a new agent with configuration
    config := agent.Config{
        Name:        "AdvancedAgent",
        Type:        "tool",
        Description: "An agent that can perform calculations and web searches",
        MaxSteps:    10,
        Temperature: 0.7,
    }
    
    agent, err := agent.NewAgent(config)
    if err != nil {
        log.Fatal("Failed to create agent:", err)
    }
    
    // Set up LLM provider
    provider, err := llm.NewOpenAIProvider(llm.OpenAIConfig{
        APIKey: "your-openai-api-key",
        Model:  "gpt-4",
    })
    if err != nil {
        log.Fatal("Failed to create LLM provider:", err)
    }
    agent.SetLLMProvider(provider)
    
    // Add tools to the agent
    toolRegistry := tools.NewToolRegistry()
    
    // Add calculator tool
    calculator := tools.NewCalculatorTool()
    toolRegistry.Register("calculator", calculator)
    
    // Add time tool
    timeTool := tools.NewTimeTool()
    toolRegistry.Register("time", timeTool)
    
    agent.SetToolRegistry(toolRegistry)
    
    // Create initial state
    state := core.NewBaseState()
    state.Set("task", "Calculate the square root of 144 and tell me the current time")
    
    // Execute the agent
    ctx := context.Background()
    result, err := agent.Execute(ctx, state)
    if err != nil {
        log.Fatal("Agent execution failed:", err)
    }
    
    fmt.Printf("🤖 Agent Result: %s\n", result.Get("response"))
}
```

## Next Steps

Congratulations! You've successfully created your first GoLangGraph agent. Here's what you can explore next:

<div class="grid cards" markdown>

-   :material-graph-outline:{ .lg .middle } **Graph Workflows**

    ---

    Learn how to create complex workflows with multiple nodes and conditional logic.

    [:octicons-arrow-right-24: Graph Workflows](../user-guide/graph-workflows.md)

-   :material-tools:{ .lg .middle } **Tools & Extensions**

    ---

    Discover built-in tools and learn how to create custom tools for your agents.

    [:octicons-arrow-right-24: Tools Guide](../user-guide/tools-extensions.md)

-   :material-database-outline:{ .lg .middle } **Persistence**

    ---

    Add database persistence and build RAG applications with vector databases.

    [:octicons-arrow-right-24: Persistence Guide](../user-guide/persistence.md)

-   :material-code-braces:{ .lg .middle } **Examples**

    ---

    Explore comprehensive examples including multi-agent systems and RAG implementations.

    [:octicons-arrow-right-24: Browse Examples](../examples/simple-chatbot.md)

</div>

## Troubleshooting

### Common Issues

!!! warning "API Key Not Set"
    If you get authentication errors, make sure your API key is correctly set:
    ```bash
    export OPENAI_API_KEY="your-actual-api-key"
    ```

!!! warning "Ollama Connection Failed"
    If Ollama connection fails, ensure the service is running:
    ```bash
    ollama serve
    # In another terminal
    ollama list  # Check available models
    ```

!!! warning "Module Not Found"
    If you get module import errors, ensure you're in a Go module directory:
    ```bash
    go mod init my-project
    go mod tidy
    ```

### Getting Help

- 📚 [Browse the full documentation](../user-guide/core-concepts.md)
- 💬 [Join our Discord community](https://discord.gg/golanggraph)
- 🐛 [Report issues on GitHub](https://github.com/piotrlaczkowski/GoLangGraph/issues)
- 📖 [Check out more examples](../examples/simple-chatbot.md) 
