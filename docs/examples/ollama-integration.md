# Ollama Integration with GoLangGraph

This document provides comprehensive guidance for using GoLangGraph with Ollama and local language models, specifically demonstrating integration with Google's Gemma 3:1B model.

## Overview

The Ollama integration allows you to run GoLangGraph agents locally using open-source language models without requiring API keys or cloud services. This is perfect for:

- **Local Development**: Test and develop agents without external dependencies
- **Privacy-Focused Applications**: Keep all data processing local
- **Cost-Effective Solutions**: No API costs for development and testing
- **Offline Capabilities**: Run agents without internet connectivity
- **Educational Purposes**: Learn AI agent concepts with accessible models

## Prerequisites

### 1. Install Ollama

**macOS/Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

**Windows:**
Download from [ollama.ai/download](https://ollama.ai/download)

**Verify Installation:**
```bash
ollama --version
```

### 2. Pull Gemma 3:1B Model

```bash
# Pull the model (this may take a few minutes)
ollama pull gemma3:1b

# Verify the model is available
ollama list
```

### 3. Start Ollama Service

```bash
# Start Ollama server
ollama serve

# Test basic functionality
ollama run gemma3:1b "Hello, world!"
```

## Quick Start

### Using Make Commands

The easiest way to test the Ollama integration:

```bash
# Set up Ollama with required models
make ollama-setup

# Run local demo with all services
make demo-local

# Run comprehensive integration tests
make test-local
```

### Manual Execution

```bash
# Run any example with Ollama
cd examples/01-basic-chat
go run main.go

# Or run from root
go run ./examples/01-basic-chat/main.go
```

## Working Examples

The repository includes 9 working examples that use Ollama with Gemma 3:1B:

### 1. Basic Chat Agent (`examples/01-basic-chat/`)
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
    
    // Create chat agent
    config := &agent.AgentConfig{
        Name:         "chat-agent",
        Type:         agent.AgentTypeChat,
        Model:        "gemma3:1b",
        Provider:     "ollama",
        SystemPrompt: "You are a helpful AI assistant.",
    }
    
    chatAgent := agent.NewAgent(config, llmManager, toolRegistry)
    
    // Execute
    execution, err := chatAgent.Execute(context.Background(), "Hello! Please introduce yourself.")
    if err != nil {
        log.Fatal("Failed to execute agent:", err)
    }
    
    fmt.Printf("Agent: %s\n", execution.Response)
}
```

### 2. ReAct Agent with Tools (`examples/02-react-agent/`)
```go
// Create tool registry with calculator
toolRegistry := tools.NewToolRegistry()
calculator := tools.NewCalculatorTool()
toolRegistry.RegisterTool("calculator", calculator)

// Create ReAct agent
config := &agent.AgentConfig{
    Name:          "react-agent",
    Type:          agent.AgentTypeReAct,
    Model:         "gemma3:1b",
    Provider:      "ollama",
    SystemPrompt:  "You are a helpful assistant that can use tools to solve problems.",
    MaxIterations: 3,
}

reactAgent := agent.NewAgent(config, llmManager, toolRegistry)

// Execute with calculation request
execution, err := reactAgent.Execute(context.Background(), "What is 25 * 17?")
if err != nil {
    log.Fatal("Failed to execute agent:", err)
}

fmt.Printf("Agent: %s\n", execution.Response)
```

### 3. Multi-Agent System (`examples/03-multi-agent/`)
```go
// Create multiple agents
analyzerAgent := agent.NewAgent(&agent.AgentConfig{
    Name:         "analyzer",
    Type:         agent.AgentTypeChat,
    Model:        "gemma3:1b",
    Provider:     "ollama",
    SystemPrompt: "You analyze user requests and extract key information.",
}, llmManager, toolRegistry)

writerAgent := agent.NewAgent(&agent.AgentConfig{
    Name:         "writer",
    Type:         agent.AgentTypeChat,
    Model:        "gemma3:1b",
    Provider:     "ollama",
    SystemPrompt: "You write detailed responses based on analysis.",
}, llmManager, toolRegistry)

// Use in a graph workflow
graph := core.NewGraph()
graph.AddNode("analyzer", analyzerFunc)
graph.AddNode("writer", writerFunc)
graph.AddEdge("analyzer", "writer")

// Execute the workflow
result, err := graph.Execute(context.Background(), initialState)
```

### 4. RAG System (`examples/04-rag-system/`)
```go
// Create RAG agent with document retrieval
config := &agent.AgentConfig{
    Name:         "rag-agent",
    Type:         agent.AgentTypeChat,
    Model:        "gemma3:1b",
    Provider:     "ollama",
    SystemPrompt: "You are a helpful assistant that answers questions based on provided context.",
}

ragAgent := agent.NewAgent(config, llmManager, toolRegistry)

// Add document context to the query
context := "GoLangGraph is a framework for building AI agent workflows..."
query := fmt.Sprintf("Context: %s\n\nQuestion: %s", context, userQuestion)

execution, err := ragAgent.Execute(ctx, query)
```

### 5. Streaming Responses (`examples/05-streaming/`)
```go
// Create streaming agent
config := &agent.AgentConfig{
    Name:     "streaming-agent",
    Type:     agent.AgentTypeChat,
    Model:    "gemma3:1b",
    Provider: "ollama",
    Streaming: true,
}

streamingAgent := agent.NewAgent(config, llmManager, toolRegistry)

// Execute with streaming callback
execution, err := streamingAgent.ExecuteWithCallback(ctx, "Tell me a story", func(chunk string) {
    fmt.Print(chunk)
})
```

## Configuration Options

### LLM Provider Configuration

```go
config := &llm.OllamaConfig{
    BaseURL: "http://localhost:11434",  // Ollama server URL
    Timeout: 60 * time.Second,          // Request timeout
}

err := llmManager.AddProvider("ollama", config)
```

### Agent Configuration

```go
config := &agent.AgentConfig{
    Name:          "my-agent",
    Type:          agent.AgentTypeChat,     // Chat, ReAct, or Tool
    Provider:      "ollama",
    Model:         "gemma3:1b",
    Temperature:   0.1,                     // Creativity (0.0-1.0)
    MaxTokens:     200,                     // Response length limit
    MaxIterations: 3,                       // For ReAct agents
    SystemPrompt:  "You are a helpful AI assistant.",
}
```

## Available Models

Ollama supports many open-source models. Popular choices for GoLangGraph:

### Small Models (Good for Development)
- **gemma3:1b** - Google's Gemma 3 1B parameters (recommended)
- **phi3:mini** - Microsoft's Phi-3 Mini
- **llama3.2:1b** - Meta's Llama 3.2 1B

### Medium Models (Better Performance)
- **gemma3:2b** - Google's Gemma 3 2B parameters
- **llama3.2:3b** - Meta's Llama 3.2 3B
- **phi3:medium** - Microsoft's Phi-3 Medium

### Large Models (Best Quality)
- **llama3.1:8b** - Meta's Llama 3.1 8B
- **gemma3:7b** - Google's Gemma 3 7B
- **mistral:7b** - Mistral AI's 7B model

```bash
# Pull different models
ollama pull gemma3:2b
ollama pull llama3.2:3b
ollama pull phi3:mini

# List available models
ollama list
```

## Performance Optimization

### Model Selection
- **Development**: Use 1B-2B parameter models for fast iteration
- **Production**: Use 7B+ parameter models for better quality
- **Resource Constraints**: Smaller models require less RAM and CPU

### Configuration Tuning
```go
// For faster responses (less creative)
Temperature: 0.0

// For more creative responses
Temperature: 0.7

// For shorter responses
MaxTokens: 50

// For detailed responses
MaxTokens: 500
```

### System Resources
- **RAM Requirements**: 
  - 1B models: ~2GB RAM
  - 3B models: ~4GB RAM
  - 7B models: ~8GB RAM
- **CPU**: Multi-core processors recommended
- **GPU**: Optional but significantly faster with NVIDIA GPUs

## Troubleshooting

### Common Issues

**1. Ollama Not Running**
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama if not running
ollama serve
```

**2. Model Not Found**
```bash
# List available models
ollama list

# Pull missing model
ollama pull gemma3:1b
```

**3. Connection Errors**
```bash
# Check Ollama logs
ollama logs

# Verify endpoint
curl http://localhost:11434/api/version
```

**4. Slow Responses**
- Use smaller models for faster responses
- Reduce `MaxTokens` in configuration
- Lower `Temperature` for more deterministic output

### Debug Mode

Enable debug logging in your Go application:

```go
import "log"

// Set log level to debug
log.SetFlags(log.LstdFlags | log.Lshortfile)

// Enable detailed logging in your application
config.Debug = true
```

## Integration Testing

Use the available Make commands for comprehensive testing:

### Test Commands
```bash
# Set up Ollama with required models
make ollama-setup

# Run comprehensive local tests
make test-local

# Run specific example tests
make test-examples

# Run local demo
make demo-local
```

### Manual Testing
```bash
# Test individual examples
cd examples/01-basic-chat && go run main.go
cd examples/02-react-agent && go run main.go
cd examples/03-multi-agent && go run main.go
# ... etc
```

## Production Considerations

### Security
- Ollama runs locally, keeping data private
- No API keys or external connections required
- Consider firewall rules if exposing Ollama externally

### Scalability
- Single Ollama instance serves multiple agents
- Consider load balancing for high-throughput applications
- Monitor resource usage and scale horizontally if needed

### Deployment
```yaml
# Docker Compose example
version: '3.8'
services:
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    environment:
      - OLLAMA_HOST=0.0.0.0
  
  golanggraph:
    build: .
    depends_on:
      - ollama
    environment:
      - OLLAMA_BASE_URL=http://ollama:11434

volumes:
  ollama_data:
```

## Next Steps

1. **Explore Examples**: Run all 9 examples in the `/examples` directory
2. **Try Different Models**: Experiment with different model sizes and capabilities
3. **Custom Tools**: Add domain-specific tools for your agents
4. **Advanced Workflows**: Build complex multi-agent systems
5. **Performance Tuning**: Optimize for your specific requirements

## Resources

- [Ollama Documentation](https://ollama.ai/docs)
- [Gemma Model Card](https://ai.google.dev/gemma)
- [GoLangGraph Examples](../../examples/)
- [Quick Start Guide](../getting-started/quick-start.md)
- [Development Guide](../DEVELOPMENT.md) 
