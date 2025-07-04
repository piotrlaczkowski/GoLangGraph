# 🚀 GoLangGraph Examples

This directory contains comprehensive examples showcasing the full functionality of GoLangGraph with different complexity levels. Each example is designed to be runnable locally using Ollama with the Gemma3 1B model.

## 🚀 Quick Start

Each example is completely independent with its own `go.mod` file and simple structure:

```
examples/
├── 01-basic-chat/
│   ├── main.go          # Complete chat agent implementation
│   ├── go.mod           # Independent module
│   └── README.md        # Detailed documentation
├── 02-react-agent/
│   ├── main.go          # Entry point and session management
│   ├── agent.go         # ReAct agent implementation
│   ├── tools.go         # Tool implementations
│   ├── go.mod           # Independent module
│   └── README.md
```

### Prerequisites

1. **Install Ollama**: Download from [ollama.com](https://ollama.com)
2. **Start Ollama**: `ollama serve`
3. **Pull Models**:
   ```bash
   ollama pull gemma3:1b                    # For basic examples
   ollama pull orieg/gemma3-tools:1b        # For tool-enabled examples (recommended)
   ```

### Running Examples

Navigate to any example directory and run:

```bash
cd examples/01-basic-chat
go run main.go
```

Or compile and run:

```bash
go build -o example-name main.go  # For single-file examples
go build -o example-name *.go     # For multi-file examples
./example-name
```

## 🎯 Examples Overview

### 🤖 [01-basic-chat](./01-basic-chat/) - Basic Chat Agent
**Complexity: Beginner** | **Runtime: ~2 minutes**

Learn the fundamentals of creating a simple chat agent.

**Features:**
- ✅ Basic agent creation and configuration
- ✅ Simple conversation handling
- ✅ Ollama integration
- ✅ Performance monitoring
- ✅ Interactive commands

**What you'll learn:**
- Agent initialization and configuration
- LLM provider integration
- Basic conversation flow
- Error handling and monitoring

```bash
cd examples/01-basic-chat
go run *.go
```

---

### 🧠 [02-react-agent](./02-react-agent/) - ReAct Agent with Tools
**Complexity: Intermediate** | **Runtime: ~5 minutes**

Explore the ReAct (Reasoning and Acting) pattern with tool integration.

**Features:**
- ✅ ReAct pattern implementation
- ✅ Tool integration (calculator, web search, file ops)
- ✅ Multi-step problem solving
- ✅ Advanced mathematical functions
- ✅ Statistical analysis tools

**What you'll learn:**
- ReAct reasoning pattern
- Tool creation and integration
- Complex problem decomposition
- Advanced agent capabilities

```bash
cd examples/02-react-agent
go run *.go
```

---

### 👥 [03-multi-agent](./03-multi-agent/) - Multi-Agent System
**Complexity: Advanced** | **Runtime: ~8 minutes**

Discover how multiple specialized agents work together.

**Features:**
- ✅ Multiple specialized agents
- ✅ Task coordination and workflow
- ✅ Agent communication patterns
- ✅ Parallel and sequential execution
- ✅ Workflow orchestration

**What you'll learn:**
- Multi-agent architecture
- Task decomposition and delegation
- Agent coordination strategies
- Workflow design patterns

```bash
cd examples/03-multi-agent
go run *.go
```

---

### 📚 [04-rag-system](./04-rag-system/) - RAG Implementation
**Complexity: Advanced** | **Runtime: ~10 minutes**

Build a Retrieval-Augmented Generation system.

**Features:**
- ✅ Document ingestion and vectorization
- ✅ Semantic search and retrieval
- ✅ Context-aware generation
- ✅ Knowledge base management
- ✅ Vector database integration

**What you'll learn:**
- RAG architecture and implementation
- Document processing and embedding
- Vector search and similarity
- Context management strategies

```bash
cd examples/04-rag-system
go run *.go
```

---

### 🌊 [05-streaming](./05-streaming/) - Real-time Streaming
**Complexity: Intermediate** | **Runtime: ~5 minutes**

Implement real-time streaming responses.

**Features:**
- ✅ Real-time response streaming
- ✅ WebSocket integration
- ✅ Progressive output display
- ✅ Cancellation and timeout handling
- ✅ Performance optimization

**What you'll learn:**
- Streaming response implementation
- Real-time communication patterns
- Performance optimization techniques
- User experience enhancement

```bash
cd examples/05-streaming
go run *.go
```

---

### 💾 [06-persistence](./06-persistence/) - Data Persistence
**Complexity: Advanced** | **Runtime: ~10 minutes**

Explore data persistence and memory management.

**Features:**
- ✅ Conversation history storage
- ✅ Agent memory management
- ✅ Database integration
- ✅ Session management
- ✅ Data retrieval and search

**What you'll learn:**
- Persistence strategies
- Database integration patterns
- Memory management techniques
- Session handling

```bash
cd examples/06-persistence
# Install SQLite dependency (first time only)
go mod init persistence-example
go get github.com/mattn/go-sqlite3
# Run the example
go run *.go
```

---

### 🔧 [07-tools-integration](./07-tools-integration/) - Advanced Tools
**Complexity: Advanced** | **Runtime: ~8 minutes**

Master advanced tool integration and custom tool development.

**Features:**
- ✅ Custom tool development
- ✅ External API integration
- ✅ Tool chaining and composition
- ✅ Security and validation
- ✅ Performance optimization

**What you'll learn:**
- Advanced tool development
- API integration patterns
- Security best practices
- Tool ecosystem design

```bash
cd examples/07-tools-integration
go run main.go
```

---

### 🏭 [08-production-ready](./08-production-ready/) - Production Deployment
**Complexity: Expert** | **Runtime: ~15 minutes**

Build production-ready applications with full enterprise features.

**Features:**
- ✅ Production configuration
- ✅ Monitoring and observability
- ✅ Error handling and recovery
- ✅ Security and authentication
- ✅ Scalability patterns
- ✅ Docker deployment

**What you'll learn:**
- Production deployment strategies
- Monitoring and observability
- Security implementation
- Scalability patterns

```bash
cd examples/08-production-ready
go run main.go
# Or run with config file:
# GOLANGGRAPH_SERVER_PORT=9090 go run main.go
```

---

### 🔄 [09-workflow-graph](./09-workflow-graph/) - Complex Workflow Graph
**Complexity: Expert** | **Runtime: ~12 minutes**

Master advanced workflow orchestration with graph-based architecture, nodes, edges, and ReAct agent integration.

**Features:**
- ✅ Graph-based workflow architecture
- ✅ Multi-node workflows with conditional edges
- ✅ ReAct agent integration with tools
- ✅ Dynamic routing and state management
- ✅ Parallel execution paths
- ✅ Result aggregation and synthesis

**What you'll learn:**
- Graph-based workflow design
- ReAct pattern implementation
- Conditional routing and state management
- Advanced workflow orchestration
- Tool integration within workflows

```bash
cd examples/09-workflow-graph
go run main.go
```

## 🎓 Learning Path

### Beginner Path
1. **[01-basic-chat](./01-basic-chat/)** - Start here to understand fundamentals
2. **[05-streaming](./05-streaming/)** - Add real-time capabilities
3. **[06-persistence](./06-persistence/)** - Learn data management

### Intermediate Path
1. **[02-react-agent](./02-react-agent/)** - Master tool integration
2. **[04-rag-system](./04-rag-system/)** - Implement knowledge systems
3. **[07-tools-integration](./07-tools-integration/)** - Advanced tool development

### Advanced Path
1. **[03-multi-agent](./03-multi-agent/)** - Multi-agent coordination
2. **[09-workflow-graph](./09-workflow-graph/)** - Complex workflow orchestration
3. **[08-production-ready](./08-production-ready/)** - Production deployment
4. **Custom Implementation** - Build your own system

## 🛠️ Common Configuration

All examples use consistent configuration patterns:

### Model Configuration
```go
// Standard configuration
ollamaConfig := &llm.ProviderConfig{
    Type:        "ollama",
    Endpoint:    "http://localhost:11434",
    Model:       "gemma3:1b",
    Temperature: 0.7,
    MaxTokens:   500,
    Timeout:     30 * time.Second,
}

// Tool-enabled configuration (recommended)
ollamaConfig := &llm.ProviderConfig{
    Type:        "ollama",
    Endpoint:    "http://localhost:11434",
    Model:       "orieg/gemma3-tools:1b",  // Better tool integration
    Temperature: 0.7,
    MaxTokens:   500,
    Timeout:     30 * time.Second,
}
```

### Agent Types
- **Chat Agent**: Simple conversational agents
- **ReAct Agent**: Reasoning and acting with tools
- **Custom Agent**: Specialized implementations

### Tool Integration
All examples demonstrate different aspects of tool integration:
- Built-in tools (calculator, file operations, web search)
- Custom tools (domain-specific functionality)
- Tool chaining and composition

## 🔧 Troubleshooting

### Common Issues

1. **Ollama not running**:
   ```bash
   ollama serve
   ```

2. **Model not found**:
   ```bash
   ollama pull gemma3:1b
   ollama pull orieg/gemma3-tools:1b
   ```

3. **Port conflicts**:
   ```bash
   # Check if Ollama is running on port 11434
   curl http://localhost:11434/api/tags
   ```

4. **Memory issues**:
   - Use smaller models for limited resources
   - Reduce MaxTokens in configuration
   - Implement proper timeout handling

5. **Compilation errors**:
   ```bash
   # For basic-chat example with multiple files
   go run *.go
   
   # For persistence example requiring SQLite
   go mod init persistence-example
   go get github.com/mattn/go-sqlite3
   go run *.go
   ```

6. **Missing dependencies**:
   ```bash
   # If you get "no required module provides package" errors
   go mod init example-name
   go get [missing-package]
   ```

### Performance Tips

- Use `orieg/gemma3-tools:1b` for better tool integration
- Set appropriate timeouts for your use case
- Monitor memory usage with multiple agents
- Implement proper error handling and retries

## 📊 Performance Benchmarks

| Example | Avg Response Time | Memory Usage | Complexity |
|---------|------------------|--------------|------------|
| 01-basic-chat | 2-4s | ~100MB | ⭐ |
| 02-react-agent | 3-8s | ~150MB | ⭐⭐ |
| 03-multi-agent | 5-15s | ~300MB | ⭐⭐⭐ |
| 04-rag-system | 4-12s | ~200MB | ⭐⭐⭐ |
| 05-streaming | 1-3s | ~120MB | ⭐⭐ |
| 06-persistence | 3-8s | ~180MB | ⭐⭐⭐ |
| 07-tools-integration | 4-10s | ~160MB | ⭐⭐⭐ |
| 08-production-ready | 3-12s | ~250MB | ⭐⭐⭐⭐ |
| 09-workflow-graph | 5-15s | ~350MB | ⭐⭐⭐⭐ |

*Benchmarks based on Gemma3 1B model on standard hardware*

## 🤝 Contributing

Want to contribute more examples? Please:

1. Follow the established structure and patterns
2. Include comprehensive documentation
3. Ensure examples are runnable with Ollama
4. Add appropriate error handling
5. Include performance considerations

## 📚 Additional Resources

- **[GoLangGraph Documentation](../docs/)** - Complete framework documentation
- **[Ollama Documentation](https://ollama.ai/docs)** - Ollama setup and configuration
- **[Gemma Models](https://ollama.ai/library/gemma3)** - Available Gemma models
- **[Tool Development Guide](../docs/tools.md)** - Creating custom tools

## 🎯 Next Steps

After completing these examples, you'll be ready to:

1. **Build Custom Applications** - Create your own AI agent systems
2. **Integrate with Existing Systems** - Add AI capabilities to current projects
3. **Scale to Production** - Deploy robust, production-ready solutions
4. **Contribute to GoLangGraph** - Help improve the framework

Happy coding! 🚀 