# GoLangGraph Persistence Guide

## Overview

The GoLangGraph persistence package provides comprehensive database connectivity and state management capabilities for building production-ready AI agents. It supports multiple database types, RAG (Retrieval-Augmented Generation) functionality, and advanced checkpointing mechanisms.

## Supported Database Types

### 1. PostgreSQL
- **Type**: `DatabaseTypePostgres`
- **Use Case**: Primary persistence for production applications
- **Features**: ACID transactions, complex queries, robust data integrity
- **Configuration**: `NewPostgresConfig(host, port, database, username, password)`

### 2. PostgreSQL with pgvector
- **Type**: `DatabaseTypePgVector`
- **Use Case**: RAG applications with vector similarity search
- **Features**: Vector embeddings, similarity search, document storage
- **Configuration**: `NewPgVectorConfig(host, port, database, username, password, vectorDim)`

### 3. Redis
- **Type**: `DatabaseTypeRedis`
- **Use Case**: Fast caching and session management
- **Features**: In-memory storage, TTL support, high performance
- **Configuration**: `NewRedisConfig(host, port, password)`

### 4. OpenSearch (Future Support)
- **Type**: `DatabaseTypeOpenSearch`
- **Use Case**: Advanced search and analytics
- **Features**: Full-text search, vector search, real-time analytics
- **Status**: Planned for future implementation

### 5. Elasticsearch (Future Support)
- **Type**: `DatabaseTypeElastic`
- **Use Case**: Enterprise search and analytics
- **Features**: Distributed search, machine learning, observability
- **Status**: Planned for future implementation

### 6. MongoDB (Future Support)
- **Type**: `DatabaseTypeMongoDB`
- **Use Case**: Document-based storage
- **Features**: Flexible schema, horizontal scaling
- **Status**: Planned for future implementation

## Quick Start

### Basic PostgreSQL Setup

```go
package main

import (
    "context"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

func main() {
    // Create PostgreSQL configuration
    config := persistence.NewPostgresConfig("localhost", 5432, "golanggraph", "postgres", "password")
    
    // Create checkpointer
    checkpointer, err := persistence.NewPostgresCheckpointer(config)
    if err != nil {
        panic(err)
    }
    defer checkpointer.Close()
    
    // Create and save a checkpoint
    state := core.NewBaseState()
    state.Set("step", 1)
    state.Set("message", "Hello World")
    
    checkpoint := &persistence.Checkpoint{
        ID:        "checkpoint-1",
        ThreadID:  "thread-123",
        State:     state,
        Metadata:  map[string]interface{}{"agent": "demo"},
        CreatedAt: time.Now(),
        NodeID:    "start_node",
        StepID:    1,
    }
    
    ctx := context.Background()
    err = checkpointer.Save(ctx, checkpoint)
    if err != nil {
        panic(err)
    }
    
    // Load the checkpoint
    loaded, err := checkpointer.Load(ctx, "thread-123", "checkpoint-1")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Loaded checkpoint: %s\n", loaded.ID)
}
```

### RAG with PostgreSQL + pgvector

```go
// Create pgvector configuration
config := persistence.NewPgVectorConfig("localhost", 5432, "golanggraph_rag", "postgres", "password", 1536)

// Create checkpointer with RAG support
checkpointer, err := persistence.NewPostgresCheckpointer(config)
if err != nil {
    panic(err)
}
defer checkpointer.Close()

// Save a document with embeddings
doc := &persistence.Document{
    ID:        "doc-1",
    ThreadID:  "thread-456",
    Content:   "This is a sample document for RAG.",
    Metadata:  map[string]interface{}{"source": "demo"},
    Embedding: []float64{0.1, 0.2, 0.3, ...}, // Your embedding vector
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

ctx := context.Background()
err = checkpointer.SaveDocument(ctx, doc)
if err != nil {
    panic(err)
}

// Search for similar documents
queryEmbedding := []float64{0.1, 0.2, 0.3, ...} // Your query embedding
results, err := checkpointer.SearchDocuments(ctx, "thread-456", queryEmbedding, 5)
if err != nil {
    panic(err)
}

fmt.Printf("Found %d similar documents\n", len(results))
```

### Redis for Fast Caching

```go
// Create Redis configuration
config := persistence.NewRedisConfig("localhost", 6379, "")

// Create Redis checkpointer
checkpointer, err := persistence.NewRedisCheckpointer(config)
if err != nil {
    panic(err)
}
defer checkpointer.Close()

// Save checkpoint to Redis (with TTL)
checkpoint := &persistence.Checkpoint{
    ID:        "cache-1",
    ThreadID:  "thread-789",
    State:     state,
    Metadata:  map[string]interface{}{"cache": "redis"},
    CreatedAt: time.Now(),
    NodeID:    "cache_node",
    StepID:    1,
}

ctx := context.Background()
err = checkpointer.Save(ctx, checkpoint)
if err != nil {
    panic(err)
}
```

## Database Connection Manager

For applications requiring multiple database connections:

```go
// Create connection manager
manager := persistence.NewDatabaseConnectionManager()
defer manager.CloseAll()

// Add PostgreSQL connection
postgresConfig := persistence.NewPostgresConfig("localhost", 5432, "golanggraph", "postgres", "password")
err := manager.AddConnection("postgres-main", postgresConfig)
if err != nil {
    panic(err)
}

// Add pgvector connection
pgvectorConfig := persistence.NewPgVectorConfig("localhost", 5432, "golanggraph_rag", "postgres", "password", 1536)
err = manager.AddConnection("postgres-rag", pgvectorConfig)
if err != nil {
    panic(err)
}

// Get connection
conn, err := manager.GetConnection("postgres-main")
if err != nil {
    panic(err)
}

// Use connection
fmt.Printf("Connected to: %s\n", conn.GetType())
```

## Configuration Options

### PostgreSQL Configuration

```go
config := &persistence.DatabaseConfig{
    Type:         persistence.DatabaseTypePostgres,
    Host:         "localhost",
    Port:         5432,
    Database:     "golanggraph",
    Username:     "postgres",
    Password:     "password",
    SSLMode:      "require",
    MaxOpenConns: 50,
    MaxIdleConns: 10,
    MaxLifetime:  "10m",
    ConnectionParams: map[string]string{
        "application_name": "golanggraph",
        "connect_timeout":  "10",
    },
}
```

### pgvector Configuration

```go
config := &persistence.DatabaseConfig{
    Type:               persistence.DatabaseTypePgVector,
    Host:               "localhost",
    Port:               5432,
    Database:           "golanggraph_rag",
    Username:           "postgres",
    Password:           "password",
    SSLMode:            "require",
    VectorDimension:    1536,
    VectorMetric:       "cosine",
    EnableRAG:          true,
    EmbeddingModel:     "text-embedding-ada-002",
    EmbeddingDimension: 1536,
    SimilarityThreshold: 0.7,
}
```

### Redis Configuration

```go
config := &persistence.DatabaseConfig{
    Type:     persistence.DatabaseTypeRedis,
    Host:     "localhost",
    Port:     6379,
    Password: "your-redis-password",
    ConnectionParams: map[string]string{
        "max_retries":      "3",
        "retry_delay":      "100ms",
        "dial_timeout":     "5s",
        "read_timeout":     "3s",
        "write_timeout":    "3s",
        "pool_size":        "10",
        "pool_timeout":     "4s",
        "idle_timeout":     "5m",
        "idle_check_freq":  "1m",
    },
}
```

## Database Setup Instructions

### PostgreSQL Setup

1. **Install PostgreSQL**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql postgresql-contrib
   
   # macOS
   brew install postgresql
   
   # Docker
   docker run --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres
   ```

2. **Create Database**:
   ```sql
   CREATE DATABASE golanggraph;
   CREATE USER golanggraph_user WITH PASSWORD 'password';
   GRANT ALL PRIVILEGES ON DATABASE golanggraph TO golanggraph_user;
   ```

### PostgreSQL with pgvector Setup

1. **Install pgvector**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install postgresql-15-pgvector
   
   # macOS
   brew install pgvector
   
   # Docker
   docker run --name postgres-pgvector -e POSTGRES_PASSWORD=password -p 5432:5432 -d pgvector/pgvector:pg15
   ```

2. **Enable Extension**:
   ```sql
   CREATE EXTENSION vector;
   ```

3. **Create RAG Database**:
   ```sql
   CREATE DATABASE golanggraph_rag;
   ```

### Redis Setup

1. **Install Redis**:
   ```bash
   # Ubuntu/Debian
   sudo apt-get install redis-server
   
   # macOS
   brew install redis
   
   # Docker
   docker run --name redis -p 6379:6379 -d redis:alpine
   ```

2. **Start Redis**:
   ```bash
   redis-server
   ```

## RAG (Retrieval-Augmented Generation) Support

### Document Storage

The persistence package provides comprehensive RAG support with vector embeddings:

```go
// Save documents with embeddings
doc := &persistence.Document{
    ID:        "doc-1",
    ThreadID:  "thread-123",
    Content:   "Your document content here",
    Metadata:  map[string]interface{}{
        "source": "documentation",
        "category": "api",
        "tags": []string{"important", "reference"},
    },
    Embedding: embedding, // []float64 from your embedding model
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

err := checkpointer.SaveDocument(ctx, doc)
```

### Vector Similarity Search

```go
// Search for similar documents
queryEmbedding := getEmbedding("user query text")
results, err := checkpointer.SearchDocuments(ctx, threadID, queryEmbedding, 5)
if err != nil {
    panic(err)
}

for _, doc := range results {
    fmt.Printf("Similar document: %s\n", doc.Content)
    fmt.Printf("Metadata: %v\n", doc.Metadata)
}
```

### Database Schema

The persistence package automatically creates the following tables:

#### Basic Tables
- `threads`: Conversation threads
- `checkpoints`: State checkpoints
- `sessions`: User sessions

#### RAG Tables (when enabled)
- `documents`: Document storage with optional vector embeddings
- `memory`: Conversational memory with embeddings

#### Vector Indexes (pgvector)
- `idx_documents_embedding`: IVFFlat index for document embeddings
- `idx_memory_embedding`: IVFFlat index for memory embeddings

## Production Considerations

### Connection Pooling

```go
config := &persistence.DatabaseConfig{
    Type:         persistence.DatabaseTypePostgres,
    Host:         "localhost",
    Port:         5432,
    Database:     "golanggraph",
    Username:     "postgres",
    Password:     "password",
    MaxOpenConns: 50,  // Maximum open connections
    MaxIdleConns: 10,  // Maximum idle connections
    MaxLifetime:  "10m", // Connection lifetime
}
```

### SSL Configuration

```go
config := &persistence.DatabaseConfig{
    Type:    persistence.DatabaseTypePostgres,
    SSLMode: "require", // or "verify-full" for production
    ConnectionParams: map[string]string{
        "sslcert":     "/path/to/client-cert.pem",
        "sslkey":      "/path/to/client-key.pem",
        "sslrootcert": "/path/to/ca-cert.pem",
    },
}
```

### Monitoring and Logging

The persistence package includes comprehensive logging:

```go
// Logs are automatically generated for:
// - Connection establishment
// - Query execution
// - Error conditions
// - Performance metrics
```

### Error Handling

```go
checkpointer, err := persistence.NewPostgresCheckpointer(config)
if err != nil {
    // Handle connection errors
    log.Printf("Failed to connect to database: %v", err)
    return
}

err = checkpointer.Save(ctx, checkpoint)
if err != nil {
    // Handle save errors
    log.Printf("Failed to save checkpoint: %v", err)
    return
}
```

## Best Practices

### 1. Connection Management
- Use connection pooling for production
- Close connections properly
- Handle connection timeouts

### 2. RAG Implementation
- Use appropriate embedding dimensions
- Implement proper similarity thresholds
- Consider document chunking strategies

### 3. Performance Optimization
- Create proper indexes
- Use connection pooling
- Implement caching strategies

### 4. Security
- Use SSL/TLS connections
- Implement proper authentication
- Validate input data

### 5. Monitoring
- Monitor connection pool usage
- Track query performance
- Implement health checks

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Check if database server is running
   - Verify host and port configuration
   - Check firewall settings

2. **Authentication Failed**
   - Verify username and password
   - Check database permissions
   - Ensure user has necessary privileges

3. **Database Not Found**
   - Create the database first
   - Check database name spelling
   - Verify user has access to database

4. **pgvector Extension Missing**
   - Install pgvector extension
   - Enable extension in database
   - Check PostgreSQL version compatibility

### Debug Mode

Enable debug logging:

```go
import "github.com/sirupsen/logrus"

// Set log level to debug
logrus.SetLevel(logrus.DebugLevel)
```

## Migration Guide

### From Memory to Database

1. **Export existing data**:
   ```go
   // Export from memory checkpointer
   checkpoints, err := memoryCheckpointer.List(ctx, threadID)
   ```

2. **Import to database**:
   ```go
   // Import to database checkpointer
   for _, checkpoint := range checkpoints {
       err := dbCheckpointer.Save(ctx, checkpoint)
   }
   ```

### Database Schema Updates

The persistence package automatically handles schema migrations when initializing connections.

## Examples

See `examples/database_persistence_demo.go` for comprehensive usage examples covering:
- Basic PostgreSQL operations
- RAG with pgvector
- Redis caching
- Connection management
- Error handling

## API Reference

### Core Interfaces

- `Checkpointer`: Main persistence interface
- `DatabaseConnection`: Database connection interface
- `DatabaseConnectionManager`: Multi-database management

### Configuration Types

- `DatabaseConfig`: Database configuration
- `DatabaseType`: Supported database types
- `Document`: RAG document structure
- `Checkpoint`: State checkpoint structure

### Helper Functions

- `NewPostgresConfig()`: PostgreSQL configuration
- `NewPgVectorConfig()`: pgvector configuration
- `NewRedisConfig()`: Redis configuration
- `CreateCheckpointer()`: Factory function for checkpointers

This comprehensive persistence layer enables GoLangGraph to work with any production database setup while providing advanced features like RAG support and vector similarity search. 