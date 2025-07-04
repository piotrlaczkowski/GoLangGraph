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
# Run the complete Ollama demo
make example-ollama

# Run comprehensive integration tests
make test-ollama

# Test only the Ollama setup (no demo execution)
make test-ollama-setup
```

### Manual Execution

```bash
# Build the demo
go build -o bin/ollama-demo ./cmd/examples

# Run the demo
./bin/ollama-demo
```

## Demo Features

The Ollama demo (`examples/ollama_demo.go`) demonstrates six key capabilities:

### 1. Basic Chat Agent
```go
config := &agent.AgentConfig{
    Name:        "demo-chat",
    Type:        agent.AgentTypeChat,
    Provider:    "ollama",
    Model:       "gemma3:1b",
    Temperature: 0.1,
    MaxTokens:   100,
}

chatAgent := agent.NewAgent(config, llmManager, toolRegistry)
execution, err := chatAgent.Execute(ctx, "Hello! Please say 'Hello from Gemma 3:1B!'")
```

### 2. ReAct Agent with Tools
```go
config := &agent.AgentConfig{
    Name:          "demo-react",
    Type:          agent.AgentTypeReAct,
    Provider:      "ollama",
    Model:         "gemma3:1b",
    Tools:         []string{"calculator"},
    MaxIterations: 3,
}

reactAgent := agent.NewAgent(config, llmManager, toolRegistry)
execution, err := reactAgent.Execute(ctx, "What is 25 + 17? Please calculate this.")
```

### 3. Multi-Agent Coordination
```go
// Create researcher and writer agents
researcher := agent.NewAgent(researcherConfig, llmManager, toolRegistry)
writer := agent.NewAgent(writerConfig, llmManager, toolRegistry)

// Coordinate sequential execution
coordinator := agent.NewMultiAgentCoordinator()
coordinator.AddAgent("researcher", researcher)
coordinator.AddAgent("writer", writer)

results, err := coordinator.ExecuteSequential(ctx, 
    []string{"researcher", "writer"}, 
    "Research and summarize: What is machine learning?")
```

### 4. Quick Builder Pattern
```go
quick := builder.Quick().WithConfig(&builder.QuickConfig{
    DefaultModel:   "gemma3:1b",
    OllamaURL:      "http://localhost:11434",
    Temperature:    0.1,
    MaxTokens:      100,
    EnableAllTools: true,
})

chatAgent := quick.Chat("quick-demo")
execution, err := chatAgent.Execute(ctx, "Say 'Quick builder works!'")
```

### 5. Custom Graph Execution
```go
graph := core.NewGraph("demo-graph")

// Add processing nodes
graph.AddNode("input", "Input Processing", inputFunc)
graph.AddNode("llm", "LLM Processing", llmFunc)
graph.AddNode("output", "Output Processing", outputFunc)

// Connect nodes
graph.AddEdge("input", "llm", nil)
graph.AddEdge("llm", "output", nil)

// Execute graph
result, err := graph.Execute(ctx, initialState)
```

### 6. Streaming Response
```go
request := llm.CompletionRequest{
    Messages: []llm.Message{
        {Role: "user", Content: "Count from 1 to 5"},
    },
    Model:       "gemma3:1b",
    Stream:      true,
}

callback := func(chunk llm.CompletionResponse) error {
    // Process streaming chunks
    return nil
}

err := llmManager.CompleteStream(ctx, "ollama", request, callback)
```

## Configuration Options

### LLM Provider Configuration

```go
config := &llm.ProviderConfig{
    Type:        "ollama",
    Endpoint:    "http://localhost:11434",  // Ollama server URL
    Model:       "gemma3:1b",               // Model name
    Temperature: 0.1,                       // Creativity (0.0-1.0)
    MaxTokens:   200,                       // Response length limit
    Timeout:     60 * time.Second,          // Request timeout
}
```

### Agent Configuration

```go
config := &agent.AgentConfig{
    Name:          "my-agent",
    Type:          agent.AgentTypeChat,     // Chat, ReAct, or Custom
    Provider:      "ollama",
    Model:         "gemma3:1b",
    Temperature:   0.1,
    MaxTokens:     100,
    MaxIterations: 3,                       // For ReAct agents
    Tools:         []string{"calculator"},  // Available tools
    SystemPrompt:  "You are a helpful AI assistant.",
}
```

## Available Models

Ollama supports many open-source models. Popular choices for GoLangGraph:

### Small Models (Good for Development)
- **gemma3:1b** - Google's Gemma 3 1B parameters
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
import "github.com/sirupsen/logrus"

// Set log level to debug
logrus.SetLevel(logrus.DebugLevel)

// Enable detailed LLM logging
config.Debug = true
```

## Integration Testing

The test script (`scripts/test-ollama-demo.sh`) provides comprehensive validation:

### Test Components
1. **Ollama Installation Check**
2. **Service Health Verification**
3. **Model Availability Validation**
4. **Basic Functionality Test**
5. **Demo Execution**
6. **Output Validation**

### Running Tests

```bash
# Full integration test
./scripts/test-ollama-demo.sh

# Check setup only
./scripts/test-ollama-demo.sh check-only

# Build only
./scripts/test-ollama-demo.sh build-only

# Run demo only
./scripts/test-ollama-demo.sh run-only
```

### Test Output
The script validates that all six demo components pass:
- ✅ Basic chat test passed!
- ✅ ReAct agent test passed!
- ✅ Multi-agent test passed!
- ✅ Quick builder test passed!
- ✅ Graph execution test passed!
- ✅ Streaming test passed!

## Production Considerations

### Security
- Ollama runs locally, keeping data private
- No API keys or external connections required
- Consider firewall rules if exposing Ollama externally

### Scalability
- Single Ollama instance serves multiple agents
- Consider load balancing for high-throughput applications
- Monitor resource usage and scale horizontally if needed

### Monitoring
```go
// Add metrics collection
import "github.com/prometheus/client_golang/prometheus"

// Track response times, error rates, etc.
responseTime := prometheus.NewHistogramVec(...)
errorRate := prometheus.NewCounterVec(...)
```

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
      - OLLAMA_HOST=http://ollama:11434

volumes:
  ollama_data:
```

## Next Steps

1. **Explore More Models**: Try different models for your use case
2. **Custom Tools**: Implement domain-specific tools for your agents
3. **Advanced Workflows**: Build complex multi-agent systems
4. **Performance Tuning**: Optimize for your specific requirements
5. **Production Deployment**: Scale and monitor your agent applications

## Resources

- [Ollama Documentation](https://ollama.ai/docs)
- [Gemma Model Card](https://ai.google.dev/gemma)
- [GoLangGraph Examples](../examples/)
- [Agent Configuration Guide](../user-guide/agents.md)
- [Tool Development Guide](../user-guide/tools.md) 
