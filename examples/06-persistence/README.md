# Persistence Example

This example demonstrates **persistent conversation management** with GoLangGraph. Learn how to save conversation history, maintain agent state, and create stateful applications that remember previous interactions.

## 🎯 What You'll Learn

- **Conversation Persistence**: Save and restore chat history
- **State Management**: Maintain agent state across sessions
- **Database Integration**: Store data in SQLite, PostgreSQL, or memory
- **Session Management**: Handle multiple concurrent conversations
- **Data Serialization**: Efficiently store complex agent states

## 💾 Persistence Benefits

- **Conversation Continuity**: Resume conversations from where you left off
- **Context Preservation**: Maintain long-term conversation context
- **User Experience**: Personalized interactions based on history
- **Analytics**: Track conversation patterns and user behavior
- **Backup & Recovery**: Protect against data loss

## 🚀 Features

- **Multiple Storage Backends**: SQLite, PostgreSQL, in-memory
- **Conversation History**: Full chat history with timestamps
- **Agent State Persistence**: Save agent configuration and memory
- **Session Management**: Handle multiple users and conversations
- **Data Compression**: Efficient storage of large conversations
- **Export/Import**: Backup and restore conversation data

## 📋 Prerequisites

1. **Ollama Installation**:
   ```bash
   # Install Ollama
   curl -fsSL https://ollama.com/install.sh | sh
   
   # Pull required models
   ollama pull gemma3:1b
   ```

2. **Database Setup** (Optional):
   ```bash
   # For PostgreSQL support
   sudo apt-get install postgresql postgresql-contrib
   
   # For SQLite (included in Go standard library)
   # No additional installation needed
   ```

## 🗄️ Storage Options

### 1. SQLite (Default)
```go
// Local file-based database
config := &PersistenceConfig{
    Backend: "sqlite",
    Database: "./conversations.db",
}
```

### 2. PostgreSQL
```go
// Production-ready database
config := &PersistenceConfig{
    Backend: "postgresql",
    Database: "postgres://user:pass@localhost/chatdb",
}
```

### 3. In-Memory
```go
// Temporary storage for testing
config := &PersistenceConfig{
    Backend: "memory",
}
```

## 💻 Usage

### Basic Persistence
```bash
cd examples/06-persistence
go run main.go

# Conversation automatically saved
> Hello, I'm working on a Go project
> [Agent responds and conversation is saved]

# Exit and restart
> exit
go run main.go

# Previous conversation restored
> What was I working on?
> [Agent remembers the Go project from previous session]
```

### Advanced Options
```bash
# Specify database backend
go run main.go --backend sqlite --db ./my_conversations.db

# PostgreSQL backend
go run main.go --backend postgresql --db "postgres://user:pass@localhost/chatdb"

# Session management
go run main.go --session user123 --conversation project-discussion

# Export conversations
go run main.go --export conversations.json

# Import conversations
go run main.go --import conversations.json
```

## 📁 Project Structure

```
06-persistence/
├── main.go              # Main persistence application
├── persistence.go       # Persistence layer implementation
├── sqlite_store.go      # SQLite storage backend
├── postgres_store.go    # PostgreSQL storage backend
├── memory_store.go      # In-memory storage backend
├── models.go           # Data models and structures
├── migrations/         # Database schema migrations
│   ├── 001_initial.sql
│   └── 002_add_metadata.sql
└── README.md           # This file
```

## 🔍 Example Interactions

### Session Continuity
```
# First session
You: I'm learning about machine learning algorithms

🤖 Assistant: That's great! Machine learning is a fascinating field. 
What specific algorithms are you interested in learning about?

You: I want to start with supervised learning
[Session saved and closed]

# Second session (later)
You: Can you remind me what we were discussing?

🤖 Assistant: We were talking about your interest in learning machine 
learning algorithms, specifically supervised learning. Would you like me to 
explain some common supervised learning algorithms like linear regression, 
decision trees, or neural networks?

[Previous context fully restored]
```

### Multi-User Support
```
# User A's session
go run main.go --session userA
You: I'm working on a React application

# User B's session  
go run main.go --session userB
You: I need help with Python data analysis

# Each user maintains separate conversation history
```

## ⚙️ Configuration

### Persistence Settings
```go
type PersistenceConfig struct {
    Backend          string        `json:"backend"`
    Database         string        `json:"database"`
    MaxHistory       int           `json:"max_history"`
    CompressionLevel int           `json:"compression_level"`
    AutoSave         bool          `json:"auto_save"`
    SaveInterval     time.Duration `json:"save_interval"`
}
```

### Database Schema
```sql
-- Conversations table
CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSON
);

-- Messages table
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    role TEXT NOT NULL, -- 'user' or 'assistant'
    content TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSON,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);

-- Agent states table
CREATE TABLE agent_states (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    state_data JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id)
);
```

## 🔄 Data Models

### Conversation Model
```go
type Conversation struct {
    ID          string            `json:"id"`
    UserID      string            `json:"user_id"`
    Title       string            `json:"title"`
    Messages    []Message         `json:"messages"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
    Metadata    map[string]string `json:"metadata"`
}
```

### Message Model
```go
type Message struct {
    ID             string            `json:"id"`
    ConversationID string            `json:"conversation_id"`
    Role           string            `json:"role"` // "user" or "assistant"
    Content        string            `json:"content"`
    Timestamp      time.Time         `json:"timestamp"`
    Metadata       map[string]string `json:"metadata"`
}
```

### Agent State Model
```go
type AgentState struct {
    ID             string                 `json:"id"`
    ConversationID string                 `json:"conversation_id"`
    AgentConfig    *agent.AgentConfig     `json:"agent_config"`
    Memory         map[string]interface{} `json:"memory"`
    Context        []string               `json:"context"`
    CreatedAt      time.Time              `json:"created_at"`
}
```

## 🛠️ Storage Operations

### Save Conversation
```go
// Save entire conversation
err := store.SaveConversation(ctx, conversation)

// Save individual message
err := store.SaveMessage(ctx, message)

// Save agent state
err := store.SaveAgentState(ctx, agentState)
```

### Load Conversation
```go
// Load conversation by ID
conversation, err := store.LoadConversation(ctx, conversationID)

// Load user's conversations
conversations, err := store.LoadUserConversations(ctx, userID)

// Load recent messages
messages, err := store.LoadRecentMessages(ctx, conversationID, limit)
```

### Search and Filter
```go
// Search conversations by content
results, err := store.SearchConversations(ctx, query)

// Filter by date range
conversations, err := store.FilterConversations(ctx, startDate, endDate)

// Get conversation statistics
stats, err := store.GetConversationStats(ctx, userID)
```

## 📊 Performance Optimization

### Indexing Strategy
```sql
-- Optimize conversation lookups
CREATE INDEX idx_conversations_user_id ON conversations(user_id);
CREATE INDEX idx_conversations_created_at ON conversations(created_at);

-- Optimize message queries
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);

-- Full-text search
CREATE INDEX idx_messages_content_fts ON messages USING gin(to_tsvector('english', content));
```

### Caching Layer
```go
// Redis cache for frequently accessed conversations
cache := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Cache conversation data
err := cache.Set(ctx, conversationID, conversationData, time.Hour)

// Retrieve from cache
data, err := cache.Get(ctx, conversationID).Result()
```

### Batch Operations
```go
// Batch save messages
err := store.BatchSaveMessages(ctx, messages)

// Bulk conversation export
conversations, err := store.ExportConversations(ctx, userID)
```

## 🔐 Security Considerations

### Data Encryption
```go
// Encrypt sensitive conversation data
encryptedContent, err := encrypt(message.Content, encryptionKey)
message.Content = encryptedContent
```

### Access Control
```go
// Verify user access to conversation
if !store.CanAccessConversation(ctx, userID, conversationID) {
    return ErrUnauthorized
}
```

### Data Retention
```go
// Automatic cleanup of old conversations
err := store.CleanupOldConversations(ctx, retentionPeriod)
```

## 📈 Analytics and Monitoring

### Conversation Metrics
```go
type ConversationStats struct {
    TotalConversations int           `json:"total_conversations"`
    TotalMessages      int           `json:"total_messages"`
    AverageLength      float64       `json:"average_length"`
    MostActiveHours    []int         `json:"most_active_hours"`
    TopTopics          []string      `json:"top_topics"`
    UserEngagement     time.Duration `json:"user_engagement"`
}
```

### Usage Tracking
```go
// Track conversation patterns
err := analytics.TrackConversationStart(ctx, userID)
err := analytics.TrackMessageSent(ctx, conversationID, messageType)
err := analytics.TrackConversationEnd(ctx, conversationID, duration)
```

## 🔧 Advanced Features

### Conversation Branching
```go
// Create conversation branches for different topics
branch, err := store.CreateConversationBranch(ctx, parentID, branchPoint)
```

### Conversation Merging
```go
// Merge multiple conversations
merged, err := store.MergeConversations(ctx, conversationIDs)
```

### Export Formats
```go
// Export to different formats
err := store.ExportToJSON(ctx, conversationID, "backup.json")
err := store.ExportToCSV(ctx, conversationID, "data.csv")
err := store.ExportToMarkdown(ctx, conversationID, "conversation.md")
```

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Errors**
   ```
   Error: Failed to connect to database
   Solution: Check database credentials and connectivity
   ```

2. **Large Conversation Performance**
   ```
   Issue: Slow loading of conversations with many messages
   Solution: Implement pagination and message limits
   ```

3. **Storage Space Issues**
   ```
   Issue: Database growing too large
   Solution: Implement data compression and archiving
   ```

4. **Concurrent Access Issues**
   ```
   Issue: Race conditions in multi-user scenarios
   Solution: Use proper database transactions and locking
   ```

## 🔗 Integration Examples

### Web Application
```go
// HTTP handlers for conversation management
http.HandleFunc("/conversations", handleConversations)
http.HandleFunc("/conversations/{id}/messages", handleMessages)
http.HandleFunc("/conversations/{id}/export", handleExport)
```

### gRPC Service
```go
// gRPC service for conversation persistence
service ConversationService {
    rpc SaveConversation(SaveConversationRequest) returns (SaveConversationResponse);
    rpc LoadConversation(LoadConversationRequest) returns (LoadConversationResponse);
    rpc SearchConversations(SearchRequest) returns (SearchResponse);
}
```

### Message Queue Integration
```go
// Async processing of conversation events
producer.Publish("conversation.created", conversationEvent)
producer.Publish("message.sent", messageEvent)
```

## 📚 Learning Resources

- **Database Design**: Principles of conversation storage
- **Data Modeling**: Structuring chat data effectively
- **Performance Optimization**: Scaling conversation storage
- **Security**: Protecting user conversation data
- **Analytics**: Understanding conversation patterns

## 🚀 Next Steps

After mastering persistence:
1. Explore **07-tools-integration** for persistent tool states
2. Try **08-production-ready** for production persistence setups
3. Build web applications with conversation history
4. Implement advanced analytics and insights

## 🤝 Contributing

Enhance this example by:
- Adding new storage backends
- Implementing advanced search features
- Contributing performance optimizations
- Sharing security best practices

---

**Happy Persisting!** 💾

This persistence example provides a solid foundation for building stateful applications with GoLangGraph that maintain conversation history and user context across sessions. 