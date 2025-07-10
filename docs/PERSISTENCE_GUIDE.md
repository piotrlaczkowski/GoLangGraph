# Persistence Guide

The `pkg/persistence` package provides database integration and state persistence capabilities for GoLangGraph.

## Overview

The persistence package enables:
- **State Checkpointing**: Save and restore workflow states
- **Database Integration**: Support for PostgreSQL and Redis
- **Session Management**: Thread-safe session handling
- **Memory Storage**: In-memory persistence for development

## Checkpointing

### Memory Checkpointer

For development and testing, use the in-memory checkpointer:

```go
import "github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"

// Create memory checkpointer
checkpointer := persistence.NewMemoryCheckpointer()

// Save a checkpoint
checkpoint := &persistence.Checkpoint{
    ID:       "checkpoint_1",
    ThreadID: "thread_1",
    State:    state, // *core.BaseState
    Metadata: map[string]interface{}{
        "step": "processing",
    },
    CreatedAt: time.Now(),
    NodeID:    "current_node",
    StepID:    1,
}

err := checkpointer.Save(context.Background(), checkpoint)
if err != nil {
    log.Printf("Failed to save checkpoint: %v", err)
}

// Load a checkpoint
loaded, err := checkpointer.Load(context.Background(), "thread_1", "checkpoint_1")
if err != nil {
    log.Printf("Failed to load checkpoint: %v", err)
}

// List checkpoints for a thread
checkpoints, err := checkpointer.List(context.Background(), "thread_1")
if err != nil {
    log.Printf("Failed to list checkpoints: %v", err)
}

for _, meta := range checkpoints {
    fmt.Printf("Checkpoint: %s at %v\n", meta.ID, meta.CreatedAt)
}

// Delete a checkpoint
err = checkpointer.Delete(context.Background(), "thread_1", "checkpoint_1")
if err != nil {
    log.Printf("Failed to delete checkpoint: %v", err)
}
```

### Database Integration

For production use, integrate with databases:

```go
// Database configuration
config := &persistence.DatabaseConfig{
    Type:     persistence.DatabaseTypePostgres,
    Host:     "localhost",
    Port:     5432,
    Database: "golanggraph",
    Username: "user",
    Password: "password",
    SSLMode:  "disable",
}

// Create database connection
db, err := persistence.NewDatabaseConnection(config)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}
defer db.Close()

// Create database checkpointer
checkpointer, err := persistence.NewDatabaseCheckpointer(db)
if err != nil {
    log.Fatal("Failed to create database checkpointer:", err)
}

// Use the same API as memory checkpointer
err = checkpointer.Save(context.Background(), checkpoint)
```

## Session Management

Manage conversation threads and sessions:

```go
// Create session manager
sessionManager := persistence.NewSessionManager(checkpointer)

// Create a new session
session, err := sessionManager.CreateSession(context.Background(), &persistence.SessionConfig{
    UserID:   "user_123",
    AgentID:  "agent_456",
    Metadata: map[string]interface{}{
        "type": "chat",
    },
})
if err != nil {
    log.Printf("Failed to create session: %v", err)
}

// Get session
session, err = sessionManager.GetSession(context.Background(), session.ID)
if err != nil {
    log.Printf("Failed to get session: %v", err)
}

// Create a thread within the session
thread, err := sessionManager.CreateThread(context.Background(), session.ID, &persistence.ThreadConfig{
    Name: "Main Conversation",
    Metadata: map[string]interface{}{
        "topic": "AI assistance",
    },
})
if err != nil {
    log.Printf("Failed to create thread: %v", err)
}

// List threads in a session
threads, err := sessionManager.ListThreads(context.Background(), session.ID)
if err != nil {
    log.Printf("Failed to list threads: %v", err)
}
```

## Integration with Agents

Use persistence with agents for stateful conversations:

```go
// Create agent with persistence
config := &agent.AgentConfig{
    Name:         "persistent-agent",
    Type:         agent.AgentTypeChat,
    Model:        "gemma3:1b",
    Provider:     "ollama",
    SystemPrompt: "You are a helpful assistant with memory.",
}

// Create agent
chatAgent := agent.NewAgent(config, llmManager, toolRegistry)

// Set up persistence
checkpointer := persistence.NewMemoryCheckpointer()
sessionManager := persistence.NewSessionManager(checkpointer)

// Create session for user
session, err := sessionManager.CreateSession(context.Background(), &persistence.SessionConfig{
    UserID: "user_123",
    AgentID: config.Name,
})
if err != nil {
    log.Fatal(err)
}

// Execute with session context
ctx := context.WithValue(context.Background(), "session_id", session.ID)
execution, err := chatAgent.Execute(ctx, "Hello! Remember that I like Go programming.")
if err != nil {
    log.Fatal(err)
}

// Save checkpoint after execution
checkpoint := &persistence.Checkpoint{
    ID:       fmt.Sprintf("checkpoint_%d", time.Now().Unix()),
    ThreadID: session.ID,
    State:    execution.State, // Assuming execution returns state
    CreatedAt: time.Now(),
}

err = checkpointer.Save(ctx, checkpoint)
if err != nil {
    log.Printf("Failed to save checkpoint: %v", err)
}
```

## Configuration

### Environment Variables

Configure persistence using environment variables:

```bash
# PostgreSQL
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DB=golanggraph
export POSTGRES_USER=user
export POSTGRES_PASSWORD=password
export POSTGRES_SSLMODE=disable

# Redis
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0
```

### Configuration Struct

```go
type DatabaseConfig struct {
    Type     DatabaseType `json:"type"`
    Host     string       `json:"host"`
    Port     int          `json:"port"`
    Database string       `json:"database"`
    Username string       `json:"username"`
    Password string       `json:"password"`
    SSLMode  string       `json:"ssl_mode"`
}

// Load from environment
config := &persistence.DatabaseConfig{
    Type:     persistence.DatabaseTypePostgres,
    Host:     os.Getenv("POSTGRES_HOST"),
    Port:     5432, // or parse from env
    Database: os.Getenv("POSTGRES_DB"),
    Username: os.Getenv("POSTGRES_USER"),
    Password: os.Getenv("POSTGRES_PASSWORD"),
    SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
}
```

## Best Practices

### 1. Checkpoint Management

- Save checkpoints at logical workflow boundaries
- Use descriptive checkpoint IDs
- Clean up old checkpoints periodically
- Include relevant metadata for debugging

### 2. Session Management

- Create sessions per user or conversation
- Use threads to organize related interactions
- Store user preferences in session metadata
- Implement session expiration policies

### 3. Error Handling

- Always handle database connection errors
- Implement retry logic for transient failures
- Log persistence operations for debugging
- Validate data before saving

### 4. Performance

- Use connection pooling for database connections
- Implement caching for frequently accessed data
- Batch operations when possible
- Monitor database performance

## Examples

### Simple Checkpointing

```go
func saveWorkflowProgress(checkpointer persistence.Checkpointer, threadID string, state *core.BaseState) error {
    checkpoint := &persistence.Checkpoint{
        ID:        fmt.Sprintf("progress_%d", time.Now().Unix()),
        ThreadID:  threadID,
        State:     state,
        CreatedAt: time.Now(),
    }
    
    return checkpointer.Save(context.Background(), checkpoint)
}

func loadWorkflowProgress(checkpointer persistence.Checkpointer, threadID, checkpointID string) (*core.BaseState, error) {
    checkpoint, err := checkpointer.Load(context.Background(), threadID, checkpointID)
    if err != nil {
        return nil, err
    }
    
    return checkpoint.State, nil
}
```

### Session-based Chat

```go
func handleChatMessage(sessionManager *persistence.SessionManager, userID, message string) (string, error) {
    // Get or create session
    sessions, err := sessionManager.ListSessions(context.Background(), userID)
    if err != nil {
        return "", err
    }
    
    var session *persistence.Session
    if len(sessions) == 0 {
        session, err = sessionManager.CreateSession(context.Background(), &persistence.SessionConfig{
            UserID: userID,
        })
        if err != nil {
            return "", err
        }
    } else {
        session = sessions[0] // Use most recent session
    }
    
    // Process message with session context
    ctx := context.WithValue(context.Background(), "session_id", session.ID)
    
    // Execute agent (implementation depends on your agent setup)
    response := processMessage(ctx, message)
    
    return response, nil
}
```

This persistence package provides the foundation for building stateful AI applications with GoLangGraph, enabling you to maintain conversation history, save workflow progress, and integrate with production databases. 
