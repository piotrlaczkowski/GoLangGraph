# GoLangGraph CLI Enhancements

This document describes the enhanced CLI capabilities for GoLangGraph, including Docker packaging, development mode, and comprehensive testing features.

## 🚀 New CLI Commands

### Project Initialization

Initialize a new GoLangGraph project with predefined templates:

```bash
# Initialize with basic template
golanggraph init my-agent --template=basic

# Initialize with advanced multi-agent template
golanggraph init my-agent --template=advanced

# Initialize with RAG-enabled template
golanggraph init my-agent --template=rag
```

### Docker Packaging

Package your agents into production-ready Docker containers:

```bash
# Build regular Docker container
golanggraph docker build --tag=my-agent:latest

# Build distroless Docker container for enhanced security
golanggraph docker build --distroless --tag=my-agent:distroless

# Build for multiple platforms
golanggraph docker build --platform=linux/amd64,linux/arm64 --tag=my-agent:latest

# Use custom Dockerfile
golanggraph docker build --dockerfile=custom.Dockerfile --tag=my-agent:custom
```

### Development Mode

Start a development server with debugging capabilities:

```bash
# Start dev server with hot-reload
golanggraph dev --host=localhost --port=8080

# Start with specific agent configuration
golanggraph dev --agent-config=configs/my-agent.yaml

# Start with debug logging
golanggraph dev --log-level=debug --debug=true
```

### Configuration Validation

Validate your agent configurations:

```bash
# Validate configuration file
golanggraph validate configs/agent-config.yaml

# Strict validation mode
golanggraph validate configs/agent-config.yaml --strict
```

### Deployment

Deploy agents to production environments:

```bash
# Deploy using Docker
golanggraph deploy docker configs/agent-config.yaml

# Deploy to Kubernetes (coming soon)
golanggraph deploy k8s configs/agent-config.yaml
```

## 🛠️ Development Mode Features

When running `golanggraph dev`, you get access to:

### Debug Dashboard
- **URL**: `http://localhost:8080/debug`
- **Features**:
  - Real-time system status
  - Agent configuration overview
  - Performance metrics
  - Log viewer
  - Configuration reload

### Agent Playground
- **URL**: `http://localhost:8080/playground`  
- **Features**:
  - Interactive agent testing
  - Input/output debugging
  - Agent performance analysis
  - WebSocket streaming support

### Hot-Reload
- Automatic restart on configuration changes
- Real-time code updates (Go files)
- Dynamic agent reconfiguration

## 🐳 Docker Container Variants

### Regular Container
```dockerfile
FROM golang:1.21-alpine AS builder
# ... build steps ...
FROM alpine:latest
# Production runtime with full OS
```

**Features**:
- Full Alpine Linux base
- Health checks
- Shell access for debugging
- Package manager available

### Distroless Container  
```dockerfile
FROM golang:1.21-alpine AS builder
# ... build steps ...
FROM gcr.io/distroless/static:nonroot
# Minimal runtime, no shell
```

**Features**:
- Minimal attack surface
- No shell or package manager
- Smaller image size
- Enhanced security

## 📊 Configuration Templates

### Basic Template
```yaml
name: "basic-agent"
type: "chat"
model: "gpt-3.5-turbo"
provider: "openai"
system_prompt: "You are a helpful assistant."
temperature: 0.7
max_tokens: 1000

tools:
  - name: "calculator"
    enabled: true
  - name: "web_search"
    enabled: false

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"
```

### Advanced Template
```yaml
name: "advanced-agent"
type: "multi-agent"
model: "gpt-4"
provider: "openai"
system_prompt: "You are an advanced AI assistant with multiple capabilities."
temperature: 0.7
max_tokens: 2000

agents:
  - name: "research-agent"
    type: "research"
    tools: ["web_search", "document_reader"]
  - name: "analysis-agent"
    type: "analysis"
    tools: ["calculator", "data_analyzer"]
  - name: "synthesis-agent"
    type: "synthesis"
    tools: ["summarizer", "report_generator"]

workflow:
  start_node: "research-agent"
  edges:
    - from: "research-agent"
      to: "analysis-agent"
    - from: "analysis-agent"
      to: "synthesis-agent"
  end_node: "synthesis-agent"
```

### RAG Template
```yaml
name: "rag-agent"
type: "rag"
model: "gpt-4"
provider: "openai"
system_prompt: "You are a RAG-enabled AI assistant."
temperature: 0.7
max_tokens: 2000

rag:
  enabled: true
  chunk_size: 1000
  chunk_overlap: 200
  similarity_threshold: 0.7
  max_chunks: 5
  embedding_model: "text-embedding-ada-002"

vector_store:
  type: "pgvector"
  host: "localhost"
  port: 5432
  database: "vectordb"
  username: "postgres"
  password: "password"
  dimensions: 1536
  collection_name: "documents"
```

## 🧪 Enhanced Testing

### CLI Testing
```bash
# Run CLI-specific tests
make test-cli

# Run enhanced test suite
make test-enhanced

# Test all CLI commands
make cli-test-all
```

### Docker Testing
```bash
# Test Docker container builds
make docker-build-agent
make docker-build-distroless

# Test container functionality
make docker-agent-complete
```

### Integration Testing
```bash
# Run full integration tests
make test-integration

# Test with local services
make test-local
```

## 🔄 Workflows

### Complete Development Workflow
```bash
# 1. Initialize project
golanggraph init my-agent --template=advanced

# 2. Navigate to project
cd my-agent

# 3. Start development server
golanggraph dev

# 4. Test configuration
golanggraph validate configs/agent-config.yaml

# 5. Build Docker container
golanggraph docker build --tag=my-agent:latest

# 6. Deploy to production
golanggraph deploy docker configs/agent-config.yaml
```

### Production Deployment Workflow
```bash
# 1. Validate configuration
golanggraph validate configs/production-config.yaml --strict

# 2. Build optimized container
golanggraph docker build --distroless --tag=my-agent:v1.0.0

# 3. Test container locally
docker run -p 8080:8080 my-agent:v1.0.0

# 4. Deploy to production
golanggraph deploy docker configs/production-config.yaml
```

## 📈 Performance Optimizations

### Container Optimizations
- **Multi-stage builds**: Minimize final image size
- **Layer caching**: Optimize build times
- **Security scanning**: Automated vulnerability detection
- **Resource limits**: Memory and CPU constraints

### Development Optimizations
- **Hot-reload**: Sub-second restart times
- **Incremental builds**: Only rebuild changed components
- **Debug profiling**: Built-in performance monitoring
- **Memory optimization**: Efficient resource usage

## 🔐 Security Features

### Container Security
- **Non-root users**: All containers run as non-root
- **Read-only filesystems**: Immutable runtime environments
- **Minimal dependencies**: Reduced attack surface
- **Security scanning**: Automated vulnerability detection

### Configuration Security
- **Secret management**: Environment variable injection
- **Configuration validation**: Prevent misconfigurations
- **Access control**: Role-based permissions
- **Audit logging**: Complete operation tracking

## 🚀 Quick Start Examples

### Basic Agent
```bash
# Create and run a basic chat agent
golanggraph init chat-agent --template=basic
cd chat-agent
golanggraph dev
# Visit http://localhost:8080/playground
```

### RAG Agent
```bash
# Create and run a RAG-enabled agent
golanggraph init rag-agent --template=rag
cd rag-agent
docker-compose up -d  # Start databases
golanggraph dev
# Visit http://localhost:8080/playground
```

### Production Deployment
```bash
# Deploy to production
golanggraph init prod-agent --template=advanced
cd prod-agent
golanggraph validate configs/advanced-config.yaml --strict
golanggraph docker build --distroless --tag=prod-agent:v1.0.0
golanggraph deploy docker configs/advanced-config.yaml
```

## 🔧 Advanced Configuration

### Environment Variables
```bash
# LLM Provider Configuration
export OPENAI_API_KEY="your-api-key"
export OLLAMA_URL="http://localhost:11434"

# Database Configuration
export POSTGRES_HOST="localhost"
export POSTGRES_PORT="5432"
export POSTGRES_DB="golanggraph"
export POSTGRES_USER="postgres"
export POSTGRES_PASSWORD="password"

# Redis Configuration
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""
```

### Custom Dockerfiles
```dockerfile
# Custom production Dockerfile
FROM golanggraph-base:latest
COPY configs/ /app/configs/
COPY custom-tools/ /app/tools/
EXPOSE 8080
CMD ["serve", "--config", "/app/configs/production.yaml"]
```

## 📚 Additional Resources

- **API Documentation**: `/api/v1/docs`
- **Debug Dashboard**: `/debug`
- **Agent Playground**: `/playground`
- **Health Checks**: `/api/v1/health`
- **Metrics**: `/debug/metrics`

## 🤝 Contributing

To contribute to the CLI enhancements:

1. **Fork the repository**
2. **Create a feature branch**
3. **Add tests for new functionality**
4. **Update documentation**
5. **Submit a pull request**

### Development Setup
```bash
# Clone and setup
git clone https://github.com/piotrlaczkowski/GoLangGraph.git
cd GoLangGraph

# Install dependencies
make install

# Run tests
make test-enhanced

# Build CLI
make build

# Test CLI commands
make cli-test-all
```

## 🐛 Troubleshooting

### Common Issues

1. **Docker build fails**
   ```bash
   # Check Docker daemon
   docker info
   
   # Clean build cache
   docker system prune -a
   ```

2. **Dev server won't start**
   ```bash
   # Check port availability
   netstat -tuln | grep 8080
   
   # Use different port
   golanggraph dev --port=8081
   ```

3. **Configuration validation fails**
   ```bash
   # Check file permissions
   ls -la configs/
   
   # Validate YAML syntax
   yamllint configs/agent-config.yaml
   ```

### Debug Commands
```bash
# Check system status
golanggraph health

# View detailed logs
golanggraph dev --log-level=debug

# Test configuration
golanggraph validate --strict configs/agent-config.yaml
```

This enhanced CLI makes GoLangGraph production-ready with enterprise-grade features for development, testing, and deployment of AI agents.