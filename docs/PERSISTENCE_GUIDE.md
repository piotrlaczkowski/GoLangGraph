# Persistence Package Documentation

The `pkg/persistence` package provides comprehensive database integration and state persistence capabilities for GoLangGraph, including support for traditional databases, vector databases, and checkpointing systems.

## Overview

The persistence package enables:
- **Database Connections**: Support for PostgreSQL, Redis, and vector databases
- **State Checkpointing**: Save and restore workflow states
- **Vector Storage**: RAG (Retrieval-Augmented Generation) capabilities
- **Document Management**: Store and search documents with embeddings
- **Session Management**: Thread-safe session and conversation handling

## Supported Databases

### PostgreSQL
Full-featured relational database support with advanced features:
- JSON/JSONB support for flexible data storage
- Connection pooling and transaction management
- Schema migrations and versioning
- Full-text search capabilities

### Redis
High-performance in-memory data store:
- Key-value storage with expiration
- Pub/Sub messaging for real-time updates
- Caching layer for improved performance
- Session storage and management

### pgvector
Vector database capabilities for AI applications:
- High-dimensional vector storage
- Similarity search with multiple distance metrics
- Embedding storage and retrieval
- RAG (Retrieval-Augmented Generation) support

## Database Configuration

### PostgreSQL Configuration

```go
import "github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"

// Basic PostgreSQL configuration
pgConfig := persistence.PostgreSQLConfig{
    Host:     "localhost",
    Port:     5432,
    Database: "golanggraph",
    Username: "user",
    Password: "password",
    SSLMode:  "disable",
}

// Advanced configuration with connection pooling
pgConfig = persistence.PostgreSQLConfig{
    Host:            "localhost",
    Port:            5432,
    Database:        "golanggraph",
    Username:        "user",
    Password:        "password",
    SSLMode:         "require",
    MaxConnections:  25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 30 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
}

// Validate configuration
if err := pgConfig.Validate(); err != nil {
    log.Fatal("Invalid PostgreSQL config:", err)
}
```

### Redis Configuration

```go
// Basic Redis configuration
redisConfig := persistence.RedisConfig{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    Database: 0,
}

// Advanced configuration with clustering
redisConfig = persistence.RedisConfig{
    Host:           "localhost",
    Port:           6379,
    Password:       "secure_password",
    Database:       0,
    PoolSize:       10,
    MinIdleConns:   3,
    MaxRetries:     3,
    DialTimeout:    5 * time.Second,
    ReadTimeout:    3 * time.Second,
    WriteTimeout:   3 * time.Second,
    PoolTimeout:    4 * time.Second,
    IdleTimeout:    5 * time.Minute,
}

// Validate configuration
if err := redisConfig.Validate(); err != nil {
    log.Fatal("Invalid Redis config:", err)
}
```

### pgvector Configuration

```go
// pgvector configuration for RAG applications
pgvectorConfig := persistence.PgVectorConfig{
    Host:       "localhost",
    Port:       5432,
    Database:   "vectordb",
    Username:   "vector_user",
    Password:   "vector_password",
    SSLMode:    "disable",
    Dimensions: 1536, // OpenAI embedding dimensions
    
    // Vector-specific settings
    IndexType:    "ivfflat",
    IndexOptions: map[string]interface{}{
        "lists": 100,
    },
    DistanceMetric: "cosine",
}

// Validate configuration
if err := pgvectorConfig.Validate(); err != nil {
    log.Fatal("Invalid pgvector config:", err)
}
```

## Database Manager

The `DatabaseManager` provides centralized database connection management:

```go
// Create database manager
dbManager := persistence.NewDatabaseManager()

// Add database connections
err := dbManager.AddPostgreSQL("main", pgConfig)
if err != nil {
    log.Fatal("Failed to add PostgreSQL:", err)
}

err = dbManager.AddRedis("cache", redisConfig)
if err != nil {
    log.Fatal("Failed to add Redis:", err)
}

err = dbManager.AddPgVector("vectors", pgvectorConfig)
if err != nil {
    log.Fatal("Failed to add pgvector:", err)
}

// Get connections
pgConn, err := dbManager.GetPostgreSQL("main")
if err != nil {
    log.Fatal("Failed to get PostgreSQL connection:", err)
}

redisConn, err := dbManager.GetRedis("cache")
if err != nil {
    log.Fatal("Failed to get Redis connection:", err)
}

vectorConn, err := dbManager.GetPgVector("vectors")
if err != nil {
    log.Fatal("Failed to get pgvector connection:", err)
}

// Health checks
healthStatus := dbManager.HealthCheck()
for name, healthy := range healthStatus {
    if healthy {
        fmt.Printf("Database %s: healthy\n", name)
    } else {
        fmt.Printf("Database %s: unhealthy\n", name)
    }
}

// Close all connections
defer dbManager.Close()
```

## Checkpointing System

The checkpointing system allows you to save and restore workflow states:

### Database Checkpointer

```go
// Create database checkpointer
checkpointer, err := persistence.NewDatabaseCheckpointer(dbManager, "main")
if err != nil {
    log.Fatal("Failed to create checkpointer:", err)
}

// Save a checkpoint
checkpoint := &persistence.Checkpoint{
    ThreadID:    "conversation-123",
    State:       state,
    Metadata:    map[string]interface{}{
        "user_id":    "user123",
        "session_id": "session456",
        "step":       "processing",
    },
    Timestamp:   time.Now(),
    Version:     1,
}

err = checkpointer.SaveCheckpoint(checkpoint)
if err != nil {
    log.Fatal("Failed to save checkpoint:", err)
}

// Load a checkpoint
loadedCheckpoint, err := checkpointer.LoadCheckpoint("conversation-123")
if err != nil {
    log.Fatal("Failed to load checkpoint:", err)
}

// List checkpoints for a thread
checkpoints, err := checkpointer.ListCheckpoints("conversation-123")
if err != nil {
    log.Fatal("Failed to list checkpoints:", err)
}

// Delete old checkpoints
err = checkpointer.DeleteCheckpoint("conversation-123", 1)
if err != nil {
    log.Fatal("Failed to delete checkpoint:", err)
}
```

### Memory Checkpointer

```go
// Create in-memory checkpointer (for testing/development)
memCheckpointer := persistence.NewMemoryCheckpointer()

// Use the same interface as database checkpointer
err = memCheckpointer.SaveCheckpoint(checkpoint)
if err != nil {
    log.Fatal("Failed to save checkpoint:", err)
}

loadedCheckpoint, err := memCheckpointer.LoadCheckpoint("conversation-123")
if err != nil {
    log.Fatal("Failed to load checkpoint:", err)
}
```

## Vector Database & RAG Support

### Document Storage

```go
// Create vector store
vectorStore, err := persistence.NewPgVectorStore(pgvectorConfig)
if err != nil {
    log.Fatal("Failed to create vector store:", err)
}

// Define documents
documents := []persistence.Document{
    {
        ID:      "doc1",
        Content: "GoLangGraph is a powerful framework for building AI agent workflows.",
        Metadata: map[string]interface{}{
            "source":    "documentation",
            "category":  "framework",
            "timestamp": time.Now(),
        },
        Embedding: []float32{0.1, 0.2, 0.3, /* ... 1536 dimensions */},
    },
    {
        ID:      "doc2",
        Content: "The persistence package provides database integration capabilities.",
        Metadata: map[string]interface{}{
            "source":    "documentation",
            "category":  "persistence",
            "timestamp": time.Now(),
        },
        Embedding: []float32{0.2, 0.3, 0.4, /* ... 1536 dimensions */},
    },
}

// Store documents
err = vectorStore.StoreDocuments(documents)
if err != nil {
    log.Fatal("Failed to store documents:", err)
}

// Update a document
updatedDoc := persistence.Document{
    ID:      "doc1",
    Content: "GoLangGraph is an advanced framework for building AI agent workflows with persistence.",
    Metadata: map[string]interface{}{
        "source":    "documentation",
        "category":  "framework",
        "updated":   time.Now(),
    },
    Embedding: []float32{0.15, 0.25, 0.35, /* ... 1536 dimensions */},
}

err = vectorStore.UpdateDocument(updatedDoc)
if err != nil {
    log.Fatal("Failed to update document:", err)
}
```

### Similarity Search

```go
// Search by embedding vector
queryEmbedding := []float32{0.1, 0.2, 0.3, /* ... 1536 dimensions */}
results, err := vectorStore.SimilaritySearchByVector(queryEmbedding, 5)
if err != nil {
    log.Fatal("Failed to search by vector:", err)
}

// Search by text (requires embedding generation)
textResults, err := vectorStore.SimilaritySearch("AI agent workflows", 5)
if err != nil {
    log.Fatal("Failed to search by text:", err)
}

// Process results
for _, result := range results {
    fmt.Printf("Document ID: %s\n", result.ID)
    fmt.Printf("Content: %s\n", result.Content)
    fmt.Printf("Similarity Score: %.4f\n", result.Score)
    fmt.Printf("Metadata: %+v\n", result.Metadata)
    fmt.Println("---")
}
```

### Advanced Search with Filters

```go
// Search with metadata filters
filters := map[string]interface{}{
    "category": "framework",
    "source":   "documentation",
}

filteredResults, err := vectorStore.SimilaritySearchWithFilters(
    queryEmbedding, 
    5, 
    filters,
)
if err != nil {
    log.Fatal("Failed to search with filters:", err)
}

// Search with score threshold
thresholdResults, err := vectorStore.SimilaritySearchWithThreshold(
    queryEmbedding,
    5,
    0.8, // Minimum similarity score
)
if err != nil {
    log.Fatal("Failed to search with threshold:", err)
}
```

## Session and Thread Management

### Session Management

```go
// Create session manager
sessionManager := persistence.NewSessionManager(dbManager, "main")

// Create a new session
session := &persistence.Session{
    ID:        "session-123",
    UserID:    "user-456",
    Metadata:  map[string]interface{}{
        "app_version": "1.0.0",
        "user_agent":  "GoLangGraph-Client/1.0",
    },
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err = sessionManager.CreateSession(session)
if err != nil {
    log.Fatal("Failed to create session:", err)
}

// Get session
retrievedSession, err := sessionManager.GetSession("session-123")
if err != nil {
    log.Fatal("Failed to get session:", err)
}

// Update session
retrievedSession.Metadata["last_activity"] = time.Now()
err = sessionManager.UpdateSession(retrievedSession)
if err != nil {
    log.Fatal("Failed to update session:", err)
}

// List user sessions
userSessions, err := sessionManager.ListUserSessions("user-456")
if err != nil {
    log.Fatal("Failed to list user sessions:", err)
}

// Delete session
err = sessionManager.DeleteSession("session-123")
if err != nil {
    log.Fatal("Failed to delete session:", err)
}
```

### Thread Management

```go
// Create thread manager
threadManager := persistence.NewThreadManager(dbManager, "main")

// Create a new thread
thread := &persistence.Thread{
    ID:        "thread-789",
    SessionID: "session-123",
    UserID:    "user-456",
    Title:     "AI Workflow Discussion",
    Metadata:  map[string]interface{}{
        "topic":    "workflow_design",
        "priority": "high",
    },
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err = threadManager.CreateThread(thread)
if err != nil {
    log.Fatal("Failed to create thread:", err)
}

// Get thread
retrievedThread, err := threadManager.GetThread("thread-789")
if err != nil {
    log.Fatal("Failed to get thread:", err)
}

// List session threads
sessionThreads, err := threadManager.ListSessionThreads("session-123")
if err != nil {
    log.Fatal("Failed to list session threads:", err)
}

// Update thread
retrievedThread.Title = "Updated AI Workflow Discussion"
err = threadManager.UpdateThread(retrievedThread)
if err != nil {
    log.Fatal("Failed to update thread:", err)
}

// Delete thread
err = threadManager.DeleteThread("thread-789")
if err != nil {
    log.Fatal("Failed to delete thread:", err)
}
```

## Advanced Features

### Connection Pooling

```go
// Configure connection pooling for PostgreSQL
pgConfig := persistence.PostgreSQLConfig{
    // ... basic config
    MaxConnections:  25,              // Maximum number of connections
    MaxIdleConns:    5,               // Maximum idle connections
    ConnMaxLifetime: 30 * time.Minute, // Connection lifetime
    ConnMaxIdleTime: 5 * time.Minute,  // Idle connection timeout
}

// Configure connection pooling for Redis
redisConfig := persistence.RedisConfig{
    // ... basic config
    PoolSize:     10,               // Connection pool size
    MinIdleConns: 3,                // Minimum idle connections
    PoolTimeout:  4 * time.Second,  // Pool timeout
    IdleTimeout:  5 * time.Minute,  // Idle connection timeout
}
```

### Transaction Management

```go
// Begin transaction
tx, err := pgConn.BeginTx(context.Background(), nil)
if err != nil {
    log.Fatal("Failed to begin transaction:", err)
}
defer tx.Rollback() // Rollback if not committed

// Perform operations within transaction
_, err = tx.ExecContext(context.Background(), 
    "INSERT INTO checkpoints (thread_id, state_data) VALUES ($1, $2)",
    "thread-123", stateData)
if err != nil {
    log.Fatal("Failed to insert checkpoint:", err)
}

_, err = tx.ExecContext(context.Background(),
    "UPDATE sessions SET updated_at = $1 WHERE id = $2",
    time.Now(), "session-123")
if err != nil {
    log.Fatal("Failed to update session:", err)
}

// Commit transaction
err = tx.Commit()
if err != nil {
    log.Fatal("Failed to commit transaction:", err)
}
```

### Batch Operations

```go
// Batch insert documents
batchSize := 100
documents := make([]persistence.Document, 1000)
// ... populate documents

for i := 0; i < len(documents); i += batchSize {
    end := i + batchSize
    if end > len(documents) {
        end = len(documents)
    }
    
    batch := documents[i:end]
    err := vectorStore.StoreDocuments(batch)
    if err != nil {
        log.Printf("Failed to store batch %d-%d: %v", i, end, err)
        continue
    }
    
    fmt.Printf("Stored batch %d-%d\n", i, end)
}
```

### Monitoring and Metrics

```go
// Database health monitoring
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            health := dbManager.HealthCheck()
            for name, healthy := range health {
                if !healthy {
                    log.Printf("Database %s is unhealthy", name)
                    // Trigger alerts or recovery procedures
                }
            }
        }
    }
}()

// Connection pool monitoring
stats := pgConn.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

## Testing

The persistence package includes comprehensive tests:

```bash
# Run all persistence tests
go test ./pkg/persistence -v

# Run specific test
go test ./pkg/persistence -v -run TestDatabaseManager

# Run integration tests (requires running databases)
go test ./pkg/persistence -v -tags=integration
```

### Example Test

```go
func TestDatabaseCheckpointer(t *testing.T) {
    // Setup test database
    dbManager := setupTestDatabase(t)
    defer dbManager.Close()
    
    // Create checkpointer
    checkpointer, err := persistence.NewDatabaseCheckpointer(dbManager, "test")
    require.NoError(t, err)
    
    // Create test checkpoint
    state := core.NewBaseState()
    state.Set("test_key", "test_value")
    
    checkpoint := &persistence.Checkpoint{
        ThreadID:  "test-thread",
        State:     state,
        Metadata:  map[string]interface{}{"test": true},
        Timestamp: time.Now(),
        Version:   1,
    }
    
    // Save checkpoint
    err = checkpointer.SaveCheckpoint(checkpoint)
    require.NoError(t, err)
    
    // Load checkpoint
    loaded, err := checkpointer.LoadCheckpoint("test-thread")
    require.NoError(t, err)
    require.Equal(t, checkpoint.ThreadID, loaded.ThreadID)
    
    // Verify state
    value, exists := loaded.State.Get("test_key")
    require.True(t, exists)
    require.Equal(t, "test_value", value)
}
```

## Best Practices

### 1. Connection Management

```go
// ✅ Good: Use connection pooling
dbManager := persistence.NewDatabaseManager()
defer dbManager.Close() // Always close connections

// ❌ Bad: Creating new connections for each operation
// This leads to connection leaks and poor performance
```

### 2. Error Handling

```go
// ✅ Good: Handle specific database errors
err := checkpointer.SaveCheckpoint(checkpoint)
if err != nil {
    var pgErr *pq.Error
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // Unique violation
            return fmt.Errorf("checkpoint already exists: %w", err)
        case "23503": // Foreign key violation
            return fmt.Errorf("invalid thread reference: %w", err)
        default:
            return fmt.Errorf("database error: %w", err)
        }
    }
    return fmt.Errorf("failed to save checkpoint: %w", err)
}
```

### 3. Resource Cleanup

```go
// ✅ Good: Always clean up resources
func processWithTransaction(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback if not committed
    
    // ... perform operations
    
    return tx.Commit()
}
```

### 4. Configuration Validation

```go
// ✅ Good: Validate configuration before use
if err := config.Validate(); err != nil {
    return fmt.Errorf("invalid configuration: %w", err)
}

// ❌ Bad: Using configuration without validation
// This can lead to runtime errors
```

## Performance Optimization

### 1. Connection Pooling

```go
// Optimize connection pool settings based on your workload
pgConfig := persistence.PostgreSQLConfig{
    MaxConnections:  25,              // Based on database limits
    MaxIdleConns:    5,               // Keep some connections ready
    ConnMaxLifetime: 30 * time.Minute, // Prevent stale connections
    ConnMaxIdleTime: 5 * time.Minute,  // Clean up idle connections
}
```

### 2. Batch Operations

```go
// Process documents in batches for better performance
const batchSize = 100
for i := 0; i < len(documents); i += batchSize {
    batch := documents[i:min(i+batchSize, len(documents))]
    if err := vectorStore.StoreDocuments(batch); err != nil {
        log.Printf("Batch failed: %v", err)
    }
}
```

### 3. Indexing

```go
// Create appropriate indexes for your queries
queries := []string{
    "CREATE INDEX IF NOT EXISTS idx_checkpoints_thread_id ON checkpoints(thread_id)",
    "CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id)",
    "CREATE INDEX IF NOT EXISTS idx_documents_metadata ON documents USING GIN(metadata)",
}

for _, query := range queries {
    if _, err := db.Exec(query); err != nil {
        log.Printf("Failed to create index: %v", err)
    }
}
```

## Conclusion

The persistence package provides a comprehensive solution for data storage and retrieval in GoLangGraph applications. With support for multiple database types, advanced features like vector search, and robust error handling, it enables building production-ready AI applications with reliable data persistence. 