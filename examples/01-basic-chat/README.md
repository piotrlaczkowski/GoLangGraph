# 🤖 Basic Chat Agent Example

This example demonstrates the simplest way to create and use a chat agent with GoLangGraph and Ollama.

## 📋 Prerequisites

1. **Ollama installed and running**:
   ```bash
   # Install Ollama (if not already installed)
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Start Ollama service
   ollama serve
   ```

2. **Pull the Gemma3 1B model**:
   ```bash
   # Pull the basic model
   ollama pull gemma3:1b
   
   # Or pull the tool-enabled version for better performance
   ollama pull orieg/gemma3-tools:1b
   ```

## 🚀 Running the Example

```bash
# From the project root
cd examples/01-basic-chat
go run main.go
```

## 🎯 What This Example Demonstrates

- ✅ **Basic agent creation** with minimal configuration
- ✅ **Ollama integration** with Gemma3 1B model
- ✅ **Simple conversation** handling
- ✅ **Error handling** and graceful degradation
- ✅ **Response formatting** and display

## 🔧 Key Features

### Simple Configuration
```go
config := &agent.AgentConfig{
    Name:         "BasicChat",
    Type:         agent.AgentTypeChat,
    Model:        "gemma3:1b",
    Temperature:  0.7,
    MaxTokens:    500,
    SystemPrompt: "You are a helpful AI assistant.",
}
```

### Ollama Provider Setup
```go
provider, err := llm.NewOllamaProvider(&llm.ProviderConfig{
    Endpoint: "http://localhost:11434",
    Model:    "gemma3:1b",
    Timeout:  30 * time.Second,
})
```

### Interactive Chat Loop
The example includes an interactive chat session where you can:
- Ask questions and get responses
- See response times and token usage
- Exit gracefully with `/quit`

## 📊 Expected Output

```
🤖 GoLangGraph Basic Chat Agent Example
=======================================

✅ Ollama provider initialized
✅ Chat agent created: BasicChat
✅ Agent ready for conversation

💬 Chat Session Started (type '/quit' to exit)
─────────────────────────────────────────────

You: Hello! Can you tell me about Go programming?

🤖 BasicChat: Hello! Go is a powerful programming language...
⏱️  Response time: 1.2s | Tokens: 45

You: /quit

👋 Goodbye! Chat session ended.
```

## 🛠️ Customization Options

You can modify the example to:

1. **Change the model**:
   ```go
   Model: "orieg/gemma3-tools:1b", // Tool-enabled version
   ```

2. **Adjust temperature** for different response styles:
   ```go
   Temperature: 0.1, // More focused responses
   Temperature: 0.9, // More creative responses
   ```

3. **Modify the system prompt**:
   ```go
   SystemPrompt: "You are a Go programming expert assistant.",
   ```

4. **Add conversation history**:
   ```go
   // The example shows how to maintain conversation context
   ```

## 🔍 Code Structure

```
01-basic-chat/
├── README.md          # This documentation
├── main.go           # Main example code
├── config.go         # Configuration helpers
└── utils.go          # Utility functions
```

## 🎓 Learning Objectives

After running this example, you'll understand:

1. **Basic GoLangGraph setup** with Ollama
2. **Agent configuration** and initialization
3. **Simple conversation handling**
4. **Error handling** best practices
5. **Response processing** and display

## 🔗 Next Steps

Once you're comfortable with this basic example, try:

- **[02-react-agent](../02-react-agent/)** - Agent with reasoning and tools
- **[03-multi-agent](../03-multi-agent/)** - Multiple agents working together
- **[04-rag-system](../04-rag-system/)** - Retrieval-Augmented Generation

## 🐛 Troubleshooting

### Common Issues

1. **Ollama not running**:
   ```bash
   # Check if Ollama is running
   curl http://localhost:11434/api/tags
   
   # If not, start it
   ollama serve
   ```

2. **Model not found**:
   ```bash
   # Pull the required model
   ollama pull gemma3:1b
   ```

3. **Connection timeout**:
   - Increase the timeout in the configuration
   - Check your network connection
   - Ensure Ollama is accessible on localhost:11434

### Performance Tips

- Use `orieg/gemma3-tools:1b` for better tool integration
- Adjust `MaxTokens` based on your needs (lower = faster)
- Set `Temperature` to 0.1 for more consistent responses

## 📚 Additional Resources

- [GoLangGraph Documentation](../../docs/)
- [Ollama Documentation](https://ollama.ai/docs)
- [Gemma3 Model Information](https://ollama.com/library/gemma3) 