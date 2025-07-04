# Streaming Example

This example demonstrates **real-time streaming responses** with GoLangGraph. Instead of waiting for complete responses, you'll see text generated progressively, creating a more interactive and responsive user experience.

## 🎯 What You'll Learn

- **Streaming Architecture**: Real-time response generation
- **Progressive Output**: See responses as they're generated
- **Buffer Management**: Handle streaming data efficiently
- **User Experience**: Create responsive chat interfaces
- **Error Handling**: Manage streaming failures gracefully

## 🌊 Streaming Benefits

- **Immediate Feedback**: Users see responses starting immediately
- **Better UX**: No waiting for long responses to complete
- **Perceived Performance**: Feels faster even with same latency
- **Cancellation**: Stop generation mid-stream if needed
- **Real-time Interaction**: More natural conversation flow

## 🚀 Features

- **Token-by-Token Streaming**: See individual tokens as generated
- **Chunk-based Streaming**: Receive responses in meaningful chunks
- **Multiple Stream Types**: Support for different streaming modes
- **Stream Cancellation**: Stop generation at any time
- **Progress Indicators**: Visual feedback during generation
- **Buffering Strategies**: Optimize for different use cases

## 📋 Prerequisites

1. **Ollama Installation**:
   ```bash
   # Install Ollama
   curl -fsSL https://ollama.com/install.sh | sh
   
   # Pull streaming-optimized models
   ollama pull gemma3:1b          # Fast, efficient model
   ollama pull llama3.2:3b        # Higher quality responses
   ```

2. **Terminal Support**: Modern terminal with real-time display capabilities

## 🔧 Streaming Modes

### 1. Token Streaming
```go
// Stream individual tokens
stream := agent.StreamTokens(ctx, prompt)
for token := range stream {
    fmt.Print(token.Text)
}
```

### 2. Chunk Streaming
```go
// Stream meaningful chunks
stream := agent.StreamChunks(ctx, prompt)
for chunk := range stream {
    fmt.Print(chunk.Content)
}
```

### 3. Sentence Streaming
```go
// Stream complete sentences
stream := agent.StreamSentences(ctx, prompt)
for sentence := range stream {
    fmt.Println(sentence.Text)
}
```

## 💻 Usage

### Basic Streaming
```bash
cd examples/05-streaming
go run main.go

# Interactive streaming chat
> Tell me about artificial intelligence
[Streaming response appears in real-time...]
```

### Advanced Options
```bash
# Different streaming modes
go run main.go --mode token      # Token-by-token
go run main.go --mode chunk      # Chunk-based
go run main.go --mode sentence   # Sentence-based

# Streaming speed control
go run main.go --delay 50ms      # Add artificial delay
go run main.go --buffer-size 10  # Adjust buffer size

# Visual enhancements
go run main.go --typing-effect   # Typewriter effect
go run main.go --colors          # Colored output
```

## 📁 Project Structure

```
05-streaming/
├── main.go              # Main streaming application
├── stream_handler.go    # Stream processing logic
├── buffer_manager.go    # Buffering and flow control
├── display_manager.go   # Terminal display management
├── config.go           # Streaming configuration
└── README.md           # This file
```

## 🎬 Example Interactions

### Real-time Chat Experience
```
You: Explain quantum computing in simple terms

🤖 Assistant: [Streaming response]
Quantum computing is a revolutionary approach to processing information that 
leverages the strange properties of quantum mechanics. Unlike classical 
computers that use bits (0 or 1), quantum computers use quantum bits or 
"qubits" that can exist in multiple states simultaneously...

[Response continues streaming in real-time]
⏱️ Stream completed in 3.2s
```

### Code Generation Streaming
```
You: Write a Python function to calculate fibonacci numbers

🤖 Assistant: [Streaming response]
```python
def fibonacci(n):
    """
    Calculate the nth Fibonacci number using dynamic programming.
    
    Args:
        n (int): The position in the Fibonacci sequence
        
    Returns:
        int: The nth Fibonacci number
    """
    if n <= 1:
        return n
    
    a, b = 0, 1
    for _ in range(2, n + 1):
        a, b = b, a + b
    
    return b
```

[Code appears progressively with syntax highlighting]
```

## ⚙️ Configuration Options

### Streaming Settings
```go
type StreamConfig struct {
    Mode           StreamMode    `json:"mode"`
    BufferSize     int          `json:"buffer_size"`
    FlushInterval  time.Duration `json:"flush_interval"`
    TypingEffect   bool         `json:"typing_effect"`
    ShowProgress   bool         `json:"show_progress"`
    EnableColors   bool         `json:"enable_colors"`
}
```

### Performance Tuning
```go
// Optimize for responsiveness
config := &StreamConfig{
    BufferSize:    1,           // Immediate output
    FlushInterval: 10*time.Millisecond,
    TypingEffect:  false,       // No artificial delay
}

// Optimize for readability
config := &StreamConfig{
    BufferSize:    10,          // Chunk output
    FlushInterval: 100*time.Millisecond,
    TypingEffect:  true,        // Typewriter effect
}
```

## 🔄 Stream Types

### 1. Token Stream
- **Granularity**: Individual tokens/words
- **Latency**: Lowest possible
- **Use Case**: Maximum responsiveness
- **Buffering**: Minimal

### 2. Chunk Stream
- **Granularity**: Meaningful text chunks
- **Latency**: Balanced
- **Use Case**: Good UX with readability
- **Buffering**: Moderate

### 3. Sentence Stream
- **Granularity**: Complete sentences
- **Latency**: Higher but acceptable
- **Use Case**: Natural reading experience
- **Buffering**: Sentence boundaries

### 4. Paragraph Stream
- **Granularity**: Complete paragraphs
- **Latency**: Highest
- **Use Case**: Document generation
- **Buffering**: Paragraph boundaries

## 🎨 Visual Enhancements

### Typing Effect
```go
// Simulate human typing
func typewriterEffect(text string, delay time.Duration) {
    for _, char := range text {
        fmt.Print(string(char))
        time.Sleep(delay)
    }
}
```

### Progress Indicators
```go
// Show streaming progress
func showProgress(current, total int) {
    percent := float64(current) / float64(total) * 100
    fmt.Printf("\r[%.1f%%] Generating response...", percent)
}
```

### Color Coding
```go
// Color different message types
const (
    ColorUser      = "\033[36m"  // Cyan
    ColorAssistant = "\033[32m"  // Green
    ColorSystem    = "\033[33m"  // Yellow
    ColorReset     = "\033[0m"   // Reset
)
```

## 🔧 Advanced Features

### Stream Cancellation
```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    // Cancel on user input
    if userWantsToCancel() {
        cancel()
    }
}()

stream := agent.StreamResponse(ctx, prompt)
```

### Buffering Strategies
```go
// Time-based buffering
buffer := NewTimeBuffer(100 * time.Millisecond)

// Size-based buffering
buffer := NewSizeBuffer(10) // 10 tokens

// Adaptive buffering
buffer := NewAdaptiveBuffer()
```

### Error Recovery
```go
stream := agent.StreamResponse(ctx, prompt)
for chunk := range stream {
    if chunk.Error != nil {
        fmt.Printf("Stream error: %v\n", chunk.Error)
        // Attempt recovery or graceful degradation
        continue
    }
    fmt.Print(chunk.Content)
}
```

## 📊 Performance Metrics

The system tracks streaming performance:

- **Time to First Token (TTFT)**: Latency before first output
- **Tokens per Second**: Generation throughput
- **Stream Latency**: Delay between generation and display
- **Buffer Efficiency**: Buffering overhead
- **User Perceived Performance**: Subjective responsiveness

## 🛠️ Implementation Details

### Stream Processing Pipeline
```
LLM Generation → Buffer → Flow Control → Display → User
```

### Buffer Management
```go
type StreamBuffer struct {
    tokens    []string
    maxSize   int
    flushTime time.Duration
    timer     *time.Timer
}

func (sb *StreamBuffer) Add(token string) {
    sb.tokens = append(sb.tokens, token)
    if len(sb.tokens) >= sb.maxSize {
        sb.Flush()
    }
}
```

### Display Coordination
```go
type DisplayManager struct {
    output   io.Writer
    cursor   CursorManager
    colors   ColorScheme
    effects  VisualEffects
}
```

## 🐛 Troubleshooting

### Common Issues

1. **Choppy Streaming**
   ```
   Issue: Irregular token delivery
   Solution: Adjust buffer size and flush intervals
   ```

2. **High Latency**
   ```
   Issue: Slow time to first token
   Solution: Use faster models or optimize network
   ```

3. **Terminal Issues**
   ```
   Issue: Display artifacts or formatting problems
   Solution: Ensure terminal supports ANSI escape codes
   ```

4. **Memory Usage**
   ```
   Issue: High memory consumption during streaming
   Solution: Implement proper buffer cleanup
   ```

### Performance Optimization

1. **Model Selection**: Choose models optimized for streaming
2. **Buffer Tuning**: Optimize buffer size for your use case
3. **Network Optimization**: Minimize network overhead
4. **Terminal Optimization**: Use efficient display methods

## 🔗 Integration Examples

### Web Streaming
```go
// Server-Sent Events (SSE)
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")

stream := agent.StreamResponse(ctx, prompt)
for chunk := range stream {
    fmt.Fprintf(w, "data: %s\n\n", chunk.Content)
    w.(http.Flusher).Flush()
}
```

### WebSocket Streaming
```go
// WebSocket real-time streaming
conn.WriteMessage(websocket.TextMessage, []byte("stream_start"))

stream := agent.StreamResponse(ctx, prompt)
for chunk := range stream {
    conn.WriteMessage(websocket.TextMessage, []byte(chunk.Content))
}
```

### gRPC Streaming
```go
// gRPC server streaming
stream := agent.StreamResponse(ctx, prompt)
for chunk := range stream {
    grpcStream.Send(&StreamResponse{
        Content: chunk.Content,
        Done:    chunk.Final,
    })
}
```

## 📚 Learning Resources

- **Streaming Protocols**: Understanding real-time communication
- **Buffer Management**: Optimizing data flow
- **Terminal Programming**: Advanced console applications
- **User Experience**: Designing responsive interfaces
- **Performance Optimization**: Minimizing latency

## 🚀 Next Steps

After mastering streaming:
1. Explore **06-persistence** for stateful streaming sessions
2. Try **07-tools-integration** for streaming tool responses
3. Check **08-production-ready** for production streaming setups
4. Build web applications with streaming APIs

## 🤝 Contributing

Enhance this example by:
- Adding new streaming modes
- Improving visual effects
- Contributing performance optimizations
- Sharing integration patterns

---

**Happy Streaming!** 🌊

This streaming example demonstrates how to create responsive, real-time applications with GoLangGraph that provide immediate feedback and engaging user experiences. 