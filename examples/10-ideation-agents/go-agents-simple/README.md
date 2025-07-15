# GoLangGraph Go-Agents-Simple

> **🚀 From 2000+ lines to ~200 lines: 95% code reduction with full production features!**

A minimal example demonstrating the power of GoLangGraph's auto-server functionality. This example shows how **4 sophisticated AI agents** with **comprehensive schemas** can be deployed as a **full production system** with just a few lines of code.

## 📋 Overview

This example demonstrates GoLangGraph's **auto-server** capability that automatically generates:

✅ **REST API endpoints** with schema validation  
✅ **Web chat interface** with agent switching  
✅ **API playground** with live documentation  
✅ **Health monitoring** and system metrics  
✅ **CORS support** for web integration  
✅ **Conversation management** with history  
✅ **Streaming responses** for real-time interaction  
✅ **Error handling** and recovery  
✅ **Request/response logging**  
✅ **Docker deployment** with health checks  

## 🤖 Available Agents

| Agent | Model | Purpose | Features |
|-------|-------|---------|----------|
| **Designer** | `gemma3:1b` | Visual design and architecture | Custom graph workflow, comprehensive schemas |
| **Interviewer** | `gemma3:1b` | Smart requirement gathering | Multi-node graph, French responses |
| **Highlighter** | `gemma3:1b` | Insight extraction and analysis | Analysis workflows, theme categorization |
| **Storymaker** | `gemma3:1b` | Narrative creation | Two-stage workflow, sustainability focus |

## 🏃 Quick Start

### Prerequisites

- **Go 1.21+** installed
- **Docker** and **Docker Compose** (for containerized deployment)
- **Ollama** running locally with `gemma3:1b` model

```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Pull required model if needed
ollama pull gemma3:1b
```

### 🎮 Play Commands (Easiest!)

```bash
# Build and run with Docker (recommended)
make play-docker

# Quick functionality test
make quick-test

# Build and run locally
make play

# Stop Docker container
make stop-docker
```

### 🔧 Development Commands

```bash
# Clean development cycle
make dev              # Local: clean → build → run
make dev-docker       # Docker: clean → build → run

# Build only
make build            # Build binary
make docker-build     # Build Docker image

# Testing
make test             # Run tests
make test-endpoints   # Test all API endpoints
make health-check     # Check application health
```

### 🚢 Deployment Options

```bash
# Minimal deployment (app only)
make deploy-local

# Full deployment (with Redis and monitoring)
make deploy-full

# Monitoring only (Prometheus + Grafana)
make deploy-monitoring

# Stop all services
make deploy-stop
```

## 📡 API Endpoints

Once running, access these automatically generated endpoints:

### 🌐 Web Interfaces
- **Chat Interface**: http://localhost:8080/
- **API Playground**: http://localhost:8080/playground
- **Debug Interface**: http://localhost:8080/debug

### 📋 System APIs
- **Health Check**: `GET /health`
- **Capabilities**: `GET /capabilities`
- **List Agents**: `GET /agents`
- **Agent Info**: `GET /agents/{agentId}`

### 🤖 Agent APIs
- **Execute Agent**: `POST /api/{agentId}`
- **Stream Response**: `POST /api/{agentId}/stream`
- **Conversation History**: `GET /api/{agentId}/conversation`
- **Add to Conversation**: `POST /api/{agentId}/conversation`
- **Clear Conversation**: `DELETE /api/{agentId}/conversation`
- **Agent Status**: `GET /api/{agentId}/status`

### 📄 Schema APIs
- **All Schemas**: `GET /schemas`
- **Agent Schema**: `GET /schemas/{agentId}`
- **Validate Input**: `POST /validate/{agentId}`

### 📊 Metrics APIs
- **System Metrics**: `GET /metrics`
- **Agent Metrics**: `GET /metrics/{agentId}`

## 💡 Usage Examples

### Test Agent Execution

```bash
# Test Designer agent
curl -X POST http://localhost:8080/api/designer \
  -H "Content-Type: application/json" \
  -d '{"message": "Design a sustainable treehouse"}'

# Test Interviewer agent  
curl -X POST http://localhost:8080/api/interviewer \
  -H "Content-Type: application/json" \
  -d '{"message": "I want to design a sustainable home"}'

# Test Highlighter agent
curl -X POST http://localhost:8080/api/highlighter \
  -H "Content-Type: application/json" \
  -d '{"conversation_history": [{"role": "user", "content": "Eco-friendly materials"}]}'

# Test Storymaker agent
curl -X POST http://localhost:8080/api/storymaker \
  -H "Content-Type: application/json" \
  -d '{"story_prompt": "A family in a sustainable habitat in 2035"}'
```

### Check System Health

```bash
# Health check
curl http://localhost:8080/health

# List available agents
curl http://localhost:8080/agents

# Get agent capabilities
curl http://localhost:8080/capabilities
```

## 🐳 Docker Deployment

### Simple Docker Run

```bash
# Build image
make docker-build

# Run with host Ollama connection
docker run --rm -d --name go-agents-simple \
  -p 8080:8080 \
  -e OLLAMA_ENDPOINT=http://host.docker.internal:11434 \
  --add-host host.docker.internal:host-gateway \
  golanggraph/go-agents-simple:latest
```

### Docker Compose Profiles

```bash
# Minimal deployment (app only)
docker-compose up -d

# With Redis and monitoring
docker-compose --profile full --profile monitoring up -d

# Monitoring only
docker-compose --profile monitoring up -d
```

## 📊 Monitoring & Observability

The system includes comprehensive monitoring and observability with Prometheus, Grafana, and custom dashboards:

### 🎪 Monitoring Stack Components

- **Prometheus**: Metrics collection, alerting, and time-series database
- **Grafana**: Advanced visualization, dashboards, and alerting  
- **Built-in Metrics**: Application and system metrics from Go runtime and HTTP handlers
- **Custom Dashboards**: Three specialized dashboards for complete observability

### 🎯 Dashboard Overview

#### 1. **System Overview Dashboard** (`go-agents-overview`)
- **Request Rate**: Real-time request throughput
- **Service Health**: Application status and uptime  
- **Response Times**: 95th percentile latency metrics
- **Memory Usage**: Go runtime memory allocation
- **Error Tracking**: HTTP error rates and counts
- **Goroutines**: Concurrent execution monitoring
- **GC Performance**: Garbage collection metrics

#### 2. **Agent Performance Dashboard** (`agent-performance`)
- **Individual Agent Metrics**: Performance per agent (Designer, Interviewer, Highlighter, Storymaker)
- **Response Time Analysis**: 50th and 95th percentile latencies per agent
- **Usage Distribution**: Request distribution across agents
- **Success Rates**: Agent-specific success/failure rates
- **Conversation Metrics**: Streaming and conversation activity
- **Error Analysis**: Error counts and patterns per agent

#### 3. **Infrastructure Dashboard** (`infrastructure`)
- **System Resources**: CPU, Memory, Disk usage
- **Network Traffic**: Inbound/outbound network metrics
- **Container Monitoring**: Docker container resource usage
- **Disk I/O**: Read/write operations and throughput
- **Load Average**: System load indicators
- **Service Status**: Container health and availability

### 🚨 Alerting Rules

Comprehensive alerting for:
- **Service Health**: Down services, high error rates
- **Performance**: High response times, resource exhaustion
- **Agent Issues**: Individual agent failures, high latency
- **Infrastructure**: CPU/Memory/Disk thresholds, system load

### 🔗 Access Monitoring

```bash
# Start with monitoring stack
make deploy-monitoring

# Or full stack with Redis + monitoring  
make deploy-full
```

**Monitoring URLs:**
- **Prometheus**: http://localhost:9091/ (metrics collection)
- **Grafana**: http://localhost:3001/ (visualization, login: admin/admin)
- **Application Metrics**: http://localhost:8080/metrics
- **Redis**: localhost:6380 (session management with full deployment)

### 📈 Key Metrics Tracked

**Application Metrics:**
- HTTP request duration and count
- Go runtime metrics (memory, goroutines, GC)
- Agent-specific performance metrics
- Error rates and response codes

**System Metrics:**
- CPU, memory, disk usage
- Network I/O and throughput
- Container resource utilization
- System load and uptime

**Business Metrics:**
- Agent usage patterns
- Conversation activity
- Response quality indicators
- User interaction metrics

## 🏗️ Architecture

### Code Structure

```
go-agents-simple/
├── main.go                 # ~70 lines - imports and uses existing agents
├── agents/                 # Separate agent definitions
│   ├── designer.go         # Visual designer with custom graph
│   ├── interviewer.go      # Smart interviewer with multi-node workflow
│   ├── highlighter.go      # Insight extractor with analysis
│   ├── storymaker.go       # Story creator with sustainability focus
│   └── registry.go         # Agent registration utilities
├── Makefile               # Comprehensive build and deployment commands
├── Dockerfile             # Multi-stage production build
├── docker-compose.yml     # Multi-service deployment with profiles
├── scripts/
│   └── test_endpoints.sh  # Comprehensive endpoint testing
└── monitoring/            # Prometheus and Grafana configuration
```

### Agent Architecture

Each agent is defined in a separate file with:
- **Custom graph workflows** with conditional edges
- **Comprehensive input/output schemas** with validation
- **Specialized system prompts** for domain expertise
- **Proper model configuration** using available Ollama models

### Auto-Server Benefits

The **GoLangGraph auto-server** eliminates the need for:
- ❌ Manual REST endpoint creation (~500 lines)
- ❌ Request/response validation (~300 lines)
- ❌ Web interface development (~800 lines)
- ❌ Health check implementation (~200 lines)
- ❌ Metrics collection setup (~300 lines)
- ❌ CORS and middleware configuration (~200 lines)
- ❌ Documentation generation (~200 lines)

**Result**: ~2500 lines of boilerplate → **~200 lines of business logic**

## 🔧 Configuration

### Environment Variables

- `OLLAMA_ENDPOINT`: Ollama server URL (default: `http://localhost:11434`)
- `PORT`: Server port (default: `8080`)
- `GIN_MODE`: Gin mode (`debug`, `release`)

### Model Configuration

Update model names in agent files to match your available Ollama models:

```go
// In agents/designer.go, etc.
Model: "gemma3:1b",  // Change to your available model
```

## 🧪 Testing

### Automated Testing

```bash
# Run comprehensive endpoint tests
make test-endpoints

# Quick functionality test
make quick-test

# Full test suite with coverage
make full-test
```

### Manual Testing

```bash
# Test individual endpoints
curl http://localhost:8080/health
curl http://localhost:8080/agents
curl -X POST http://localhost:8080/api/designer \
  -H "Content-Type: application/json" \
  -d '{"message": "test"}'
```

## 🚨 Troubleshooting

### Common Issues

**1. Ollama Connection Failed**
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Check model availability
ollama list
```

**2. Port Already in Use**
```bash
# Stop existing containers
make stop-docker
docker-compose down

# Check what's using the port
lsof -i :8080
```

**3. Docker Build Issues**
```bash
# Clean Docker cache
make docker-clean
docker system prune -f
```

**4. Module/Import Issues**
```bash
# Ensure you're in the correct directory
cd examples/10-ideation-agents/go-agents-simple

# Clean and rebuild
make clean build
```

### Health Check Debugging

```bash
# Check container logs
docker logs go-agents-simple

# Test health endpoint
curl -v http://localhost:8080/health

# Check container status
docker ps | grep go-agents
```

## 📚 Learn More

- **GoLangGraph Documentation**: [Link to main docs]
- **Auto-Server Guide**: [Link to auto-server docs]
- **Agent Development**: [Link to agent guide]
- **Schema Validation**: [Link to schema docs]

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 🎉 Summary

This example demonstrates how **GoLangGraph's auto-server** transforms AI agent development:

- **4 sophisticated agents** with custom workflows
- **Full production infrastructure** auto-generated
- **95% code reduction** compared to manual implementation
- **Docker deployment** with monitoring stack
- **Comprehensive testing** and validation
- **Ready for production** with health checks and metrics

**Try it now**: `make play-docker` and visit http://localhost:8080/ 🚀 
