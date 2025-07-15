# GoLangGraph Stateful Ideation Agents

A production-ready stateful AI agent system built with **GoLangGraph**, featuring comprehensive session management, persistent memory, RAG capabilities, and auto-generated APIs.

## 🚀 Overview

This example demonstrates how to build **stateful AI agents** using GoLangGraph's complete infrastructure stack:

- **🧠 Persistent Memory** - Vector embeddings and RAG-powered memory
- **💾 Session Management** - Full conversation persistence across sessions  
- **🗄️ Database Integration** - PostgreSQL, Redis, and vector databases
- **🌐 Auto-Generated APIs** - Dynamic REST endpoints with OpenAPI docs
- **📊 Real-time Monitoring** - Metrics, health checks, and performance tracking
- **🔄 State Persistence** - Checkpoint and resume conversations
- **🎯 User Learning** - Adaptive preferences and design iteration tracking

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Auto-Generated API Layer                    │
├─────────────────────────────────────────────────────────────────┤
│  🤖 Enhanced Designer Agent (Stateful)                        │
│  ├── Session Management (GoLangGraph)                         │
│  ├── Conversation Persistence                                 │
│  ├── Memory & RAG Integration                                 │
│  └── User Preference Learning                                 │
├─────────────────────────────────────────────────────────────────┤
│                    GoLangGraph Core Stack                      │
│  ├── 💾 Persistence Layer (PostgreSQL + Redis + Vector)      │
│  ├── 🧠 Memory Management (Embeddings + RAG)                 │
│  ├── 🔄 State Management (Checkpoints + Time Travel)         │
│  ├── 🌐 Auto-Server (Dynamic API Generation)                 │
│  └── 🎯 Multi-Agent Coordination                             │
└─────────────────────────────────────────────────────────────────┘
```

## ✨ Key Features

### 🧠 **Enhanced Memory Management**
- **Vector Embeddings**: Store conversation context as searchable vectors
- **RAG Integration**: Retrieve relevant past conversations and designs
- **User Learning**: Adapt to user preferences and design patterns
- **Long-term Memory**: Persistent across sessions and conversations

### 💾 **Complete State Persistence**
- **Session Management**: Handle multiple concurrent user sessions
- **Thread Management**: Organize conversations in persistent threads
- **Checkpointing**: Save and restore conversation state at any point
- **Database Integration**: PostgreSQL, Redis, and vector database support

### 🌐 **Auto-Generated API**
- **Dynamic Endpoints**: Automatically generated REST APIs for each agent
- **OpenAPI Documentation**: Complete API documentation with schemas
- **Real-time Streaming**: Server-sent events for live responses
- **Web Interface**: Built-in chat interface and API playground

### 🎯 **Intelligent Design Agent**
- **Sustainable Architecture**: Specialized in eco-friendly habitat design
- **Iterative Design**: Learn and improve through user feedback
- **Comprehensive Output**: Technical specs, cost estimates, timelines
- **Risk Assessment**: Construction planning and quality control

## 🚀 Quick Start

### Prerequisites

1. **Go 1.21+**
2. **PostgreSQL 14+** with pgvector extension
3. **Redis 6+**
4. **Ollama** (for local LLM inference)

```bash
# Install PostgreSQL with pgvector
sudo apt-get install postgresql postgresql-contrib
sudo -u postgres psql -c "CREATE EXTENSION vector;"

# Install Redis
sudo apt-get install redis-server

# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh
ollama pull gemma3:1b
```

### Installation

```bash
# Clone the repository
git clone https://github.com/piotrlaczkowski/GoLangGraph.git
cd GoLangGraph/examples/10-ideation-agents/go-agents-simple-statefull

# Install dependencies
go mod tidy

# Set up environment
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DB=golanggraph_stateful
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=password
export REDIS_HOST=localhost
export REDIS_PORT=6379
export OLLAMA_ENDPOINT=http://localhost:11434

# Run the system
go run main.go
```

### Access the System

```bash
# Web Interface
open http://localhost:8080

# API Playground
open http://localhost:8080/playground

# Direct API Call
curl -X POST http://localhost:8080/api/designer \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Design a sustainable tiny house for $50k",
    "user_id": "user123",
    "context": {
      "project_type": "residential",
      "budget_range": 50000,
      "sustainability_priority": 9
    }
  }'
```

## 📋 API Reference

### Core Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/designer` | Main design endpoint with full state management |
| `POST` | `/api/designer/stream` | Real-time streaming responses |
| `GET`  | `/api/designer/conversation` | Get conversation history |
| `GET`  | `/api/designer/status` | Agent health and status |
| `GET`  | `/api/agents` | List all available agents |
| `GET`  | `/health` | System health check |

### Session Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET`  | `/api/design/session/{session_id}` | Retrieve session state |
| `GET`  | `/api/design/user/{user_id}/preferences` | User preferences |
| `GET`  | `/api/design/user/{user_id}/history` | User design history |

### Request Format

```json
{
  "message": "Design request or feedback",
  "session_id": "optional-session-uuid",
  "user_id": "user-identifier",
  "context": {
    "project_type": "residential|commercial|industrial",
    "budget_range": 50000,
    "timeline": "6 months",
    "location": "California, USA",
    "sustainability_priority": 9,
    "previous_feedback": "More eco-friendly materials"
  }
}
```

### Response Format

```json
{
  "design_response": "Comprehensive design description...",
  "design_iteration": {
    "id": "iteration-uuid",
    "concept": "Modular Bio-Integrated Habitat 2035",
    "sustainability_score": 9.1,
    "estimated_cost": {
      "total_cost": 125000.00,
      "materials_cost": 65000.00,
      "labor_cost": 35000.00
    },
    "materials": [...],
    "construction_phases": [...],
    "timeline": {...},
    "risk_assessment": {...}
  },
  "session_state": {
    "session_id": "session-uuid",
    "thread_id": "thread-uuid",
    "current_phase": "completed",
    "relevant_memories": [...],
    "user_preferences": {...}
  },
  "metadata": {
    "processing_time": 2.5,
    "memory_items_used": 5,
    "iteration_number": 1,
    "confidence_score": 0.95,
    "checkpoint_id": "checkpoint-uuid"
  }
}
```

## 🔧 Configuration

### Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
API_BASE_PATH=/api

# Database Configuration  
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=golanggraph_stateful
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Vector Database
VECTOR_DIMENSION=1536
VECTOR_METRIC=cosine
SIMILARITY_THRESHOLD=0.7

# LLM Configuration
OLLAMA_ENDPOINT=http://localhost:11434
OPENAI_API_KEY=your-openai-key
GEMINI_API_KEY=your-gemini-key
DEFAULT_MODEL=gemma3:1b
DEFAULT_PROVIDER=ollama

# Feature Flags
ENABLE_WEB_UI=true
ENABLE_PLAYGROUND=true
ENABLE_METRICS=true
ENABLE_PERSISTENCE=true
ENABLE_RAG=true
ENABLE_VECTOR_SEARCH=true
ENABLE_SESSION_MGMT=true
```

### Database Schema

The system automatically creates the following tables:

```sql
-- Core GoLangGraph tables
CREATE TABLE threads (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (thread_id) REFERENCES threads(id)
);

CREATE TABLE checkpoints (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    state_data JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    node_id VARCHAR(255),
    step_id INTEGER,
    FOREIGN KEY (thread_id) REFERENCES threads(id)
);

-- RAG and Memory tables
CREATE TABLE documents (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255),
    content TEXT NOT NULL,
    metadata JSONB,
    embedding vector(1536),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id)
);

CREATE TABLE memory (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    content TEXT NOT NULL,
    memory_type VARCHAR(50) DEFAULT 'conversation',
    embedding vector(1536),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id)
);
```

## 🧪 Advanced Usage

### 1. **Session Continuity**

```bash
# Start a conversation
curl -X POST http://localhost:8080/api/designer \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Design a sustainable house",
    "user_id": "user123"
  }'

# Continue the conversation (session_id returned from previous call)
curl -X POST http://localhost:8080/api/designer \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Make it more affordable",
    "user_id": "user123",
    "session_id": "session-uuid-from-previous-response"
  }'
```

### 2. **Streaming Responses**

```javascript
const eventSource = new EventSource(
  'http://localhost:8080/api/designer/stream', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message: "Design a sustainable office building",
      user_id: "user123"
    })
  }
);

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Streaming response:', data);
};
```

### 3. **User Preference Learning**

The system automatically learns from user interactions:

```json
{
  "user_preferences": {
    "design_styles": {
      "sustainable": 0.95,
      "modern": 0.8,
      "minimalist": 0.7
    },
    "material_preferences": {
      "bamboo": 0.9,
      "recycled_steel": 0.85,
      "bio_concrete": 0.8
    },
    "budget_range": {
      "min": 30000,
      "max": 100000,
      "preferred": 60000
    }
  }
}
```

### 4. **Memory and Context Retrieval**

```bash
# Get user's design history
curl -X GET http://localhost:8080/api/design/user/user123/history

# Get user's learned preferences  
curl -X GET http://localhost:8080/api/design/user/user123/preferences

# Get session state
curl -X GET http://localhost:8080/api/design/session/session-uuid
```

## 🏭 Production Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  postgres:
    image: pgvector/pgvector:pg15
    environment:
      POSTGRES_DB: golanggraph_stateful
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    environment:
      - OLLAMA_HOST=0.0.0.0

  agents:
    build: .
    ports:
      - "8080:8080"
    environment:
      - POSTGRES_HOST=postgres
      - REDIS_HOST=redis
      - OLLAMA_ENDPOINT=http://ollama:11434
    depends_on:
      - postgres
      - redis
      - ollama

volumes:
  postgres_data:
  redis_data:
  ollama_data:
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: stateful-agents
spec:
  replicas: 3
  selector:
    matchLabels:
      app: stateful-agents
  template:
    metadata:
      labels:
        app: stateful-agents
    spec:
      containers:
      - name: agents
        image: golanggraph/stateful-agents:latest
        ports:
        - containerPort: 8080
        env:
        - name: POSTGRES_HOST
          value: postgres-service
        - name: REDIS_HOST
          value: redis-service
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## 📊 Monitoring & Metrics

### Built-in Metrics

```bash
# System health
curl http://localhost:8080/health

# Agent metrics
curl http://localhost:8080/metrics

# Agent-specific metrics
curl http://localhost:8080/metrics/designer
```

### Response Format

```json
{
  "system": {
    "status": "healthy",
    "uptime": "2h30m45s",
    "agents": 1,
    "active_sessions": 25,
    "total_requests": 1247
  },
  "database": {
    "primary": "connected",
    "cache": "connected", 
    "vector": "connected"
  },
  "agents": {
    "designer": {
      "status": "healthy",
      "requests": 523,
      "avg_response_time": "2.3s",
      "success_rate": 0.987
    }
  }
}
```

## 🔧 Development

### Adding New Agents

1. **Create Agent Definition**

```go
package agents

import (
    "../database"
    "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
)

type MyAgentDefinition struct {
    *agent.BaseAgentDefinition
    databaseManager *database.DatabaseManager
}

func NewMyAgentDefinition(dbManager *database.DatabaseManager) *MyAgentDefinition {
    config := &agent.AgentConfig{
        Name: "My Custom Agent",
        Type: agent.AgentTypeChat,
        // ... configuration
    }
    
    return &MyAgentDefinition{
        BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
        databaseManager: dbManager,
    }
}
```

2. **Register in main.go**

```go
// Register the new agent
myAgentDef := agents.NewMyAgentDefinition(s.databaseManager)
s.agentRegistry.RegisterDefinition("my-agent", myAgentDef)
s.autoServer.RegisterAgent("my-agent", myAgentDef)
```

3. **Auto-Generated Endpoints**

The system automatically creates:
- `POST /api/my-agent` - Main execution endpoint
- `POST /api/my-agent/stream` - Streaming endpoint
- `GET /api/my-agent/conversation` - Conversation management
- `GET /api/my-agent/status` - Agent status

### Custom Memory Management

```go
// Store custom memory
memoryItem := &database.MemoryItem{
    ID: uuid.New().String(),
    ThreadID: threadID,
    UserID: userID,
    Content: "Custom memory content",
    MemoryType: "custom_type",
    Embedding: generateEmbedding(content),
    Metadata: map[string]interface{}{
        "custom_field": "value",
    },
    Importance: 0.9,
    CreatedAt: time.Now(),
}

err := databaseManager.MemoryManager.StoreMemory(ctx, memoryItem)

// Retrieve relevant memories
memories, err := databaseManager.MemoryManager.RetrieveMemories(
    ctx, threadID, queryEmbedding, 5,
)
```

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Test specific components
go test ./database/...
go test ./agents/...

# Test with coverage
go test -cover ./...
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./tests/...

# Load testing
wrk -t12 -c400 -d30s --script=load_test.lua http://localhost:8080/api/designer
```

### API Testing

```bash
# Test agent endpoint
curl -X POST http://localhost:8080/api/designer \
  -H "Content-Type: application/json" \
  -d @test_request.json

# Test streaming
curl -N -H "Accept: text/event-stream" \
  -X POST http://localhost:8080/api/designer/stream \
  -d @test_request.json
```

## 🚀 GoLangGraph Features Showcased

This example demonstrates the full power of GoLangGraph:

### ✅ **Complete Persistence Stack**
- PostgreSQL with pgvector for structured and vector data
- Redis for caching and session storage
- Automatic schema creation and migration
- Connection pooling and optimization

### ✅ **Advanced Memory Management**
- Vector embeddings for semantic memory search
- RAG (Retrieval-Augmented Generation) integration
- User preference learning and adaptation
- Long-term conversation memory

### ✅ **Auto-Generated APIs**
- Dynamic REST endpoint generation
- OpenAPI documentation
- Real-time streaming support
- Built-in web interface and playground

### ✅ **Production Features**
- Health checks and monitoring
- Metrics collection and reporting
- Graceful shutdown and error handling
- Scalable architecture

### ✅ **State Management**
- Checkpoint and resume capabilities
- Thread and session management
- Time travel and state restoration
- Multi-agent coordination

## 📚 Learn More

- [GoLangGraph Documentation](../../../docs/)
- [Core Package Guide](../../../docs/CORE_PACKAGE.md)
- [Persistence Guide](../../../docs/PERSISTENCE_GUIDE.md)
- [Auto-Server Guide](../../../docs/AUTO_SERVER_GUIDE.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../../CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.

---

**Built with ❤️ using GoLangGraph** - The most comprehensive Go framework for building stateful AI agent systems. 
