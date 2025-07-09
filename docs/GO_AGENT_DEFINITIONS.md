# Go-Based Agent Definitions

## Overview

The GoLangGraph multi-agent system now supports defining agents programmatically using Go code in addition to YAML/JSON configuration files. This provides developers with the flexibility to create sophisticated agents with custom logic, tools, and workflows while maintaining compatibility with the existing configuration-based approach.

## Key Features

### 🔧 **Programmatic Agent Definition**
- Define agents using Go structs and interfaces
- Custom initialization logic and validation
- Advanced agent behaviors with custom graphs
- Integration with existing configuration system

### 🚀 **Flexible Loading Mechanisms**
- Load agents from Go source files
- Plugin-based agent loading (.so files)
- Factory pattern support for dynamic agent creation
- Global agent registry for easy management

### 🔄 **Seamless Integration**
- Works alongside YAML/JSON configuration
- CLI commands for loading and managing Go-based agents
- Automatic discovery and registration
- Mixed deployment scenarios (config + code)

### 📊 **Enhanced Management**
- Agent metadata and versioning
- Source tracking (config, definition, factory, plugin)
- Comprehensive validation and error handling
- Real-time agent information and status

## Architecture

### Core Components

#### 1. **AgentDefinition Interface**
```go
type AgentDefinition interface {
    GetConfig() *AgentConfig
    Initialize(llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) error
    CreateAgent() (*Agent, error)
    GetMetadata() map[string]interface{}
    Validate() error
}
```

#### 2. **AgentRegistry**
- Manages all registered agent definitions
- Supports both definitions and factories
- Plugin loading capabilities
- Thread-safe operations

#### 3. **BaseAgentDefinition**
- Default implementation of AgentDefinition
- Builder pattern support
- Metadata management
- Standard validation

#### 4. **AdvancedAgentDefinition**
- Custom graph building
- Custom tools and middleware
- Advanced workflows
- Complex agent behaviors

## Usage Examples

### Basic Agent Definition

```go
// Create a simple chat agent
func NewChatAgent() *agent.BaseAgentDefinition {
    config := &agent.AgentConfig{
        Name:         "chat-agent",
        Type:         agent.AgentTypeChat,
        Model:        "gpt-3.5-turbo",
        Provider:     "openai",
        SystemPrompt: "You are a helpful assistant",
        Temperature:  0.7,
        MaxTokens:    1000,
        Tools:        []string{"web_search", "calculator"},
    }
    
    definition := agent.NewBaseAgentDefinition(config)
    definition.SetMetadata("version", "1.0.0")
    definition.SetMetadata("author", "Your Name")
    
    return definition
}

// Register the agent globally
func init() {
    agent.RegisterAgent("chat-agent", NewChatAgent())
}
```

### Advanced Agent with Custom Graph

```go
type ReasoningAgent struct {
    *agent.AdvancedAgentDefinition
    reasoningSteps int
}

func NewReasoningAgent() *ReasoningAgent {
    config := &agent.AgentConfig{
        Name:         "reasoning-agent",
        Type:         agent.AgentTypeReAct,
        Model:        "gpt-4",
        Provider:     "openai",
        SystemPrompt: "You are an advanced reasoning agent",
        Temperature:  0.3,
        MaxTokens:    2000,
        Tools:        []string{"web_search", "calculator", "logic_tool"},
    }
    
    return &ReasoningAgent{
        AdvancedAgentDefinition: agent.NewAdvancedAgentDefinition(config),
        reasoningSteps:          0,
    }
}

func (ra *ReasoningAgent) BuildGraph() (*core.Graph, error) {
    graph := core.NewGraph("reasoning-graph")
    
    // Add custom nodes
    graph.AddNode("plan", "Plan", ra.planningNode)
    graph.AddNode("reason", "Reason", ra.reasoningNode)
    graph.AddNode("validate", "Validate", ra.validationNode)
    graph.AddNode("finalize", "Finalize", ra.finalizationNode)
    
    // Define flow
    graph.AddEdge("plan", "reason", nil)
    graph.AddEdge("reason", "validate", nil)
    graph.AddEdge("validate", "reason", ra.needsMoreReasoning)
    graph.AddEdge("validate", "finalize", ra.reasoningComplete)
    
    graph.SetStartNode("plan")
    graph.AddEndNode("finalize")
    
    return graph, nil
}
```

### Factory Pattern

```go
// Register using factory for dynamic creation
func init() {
    factory := func() agent.AgentDefinition {
        return NewReasoningAgent()
    }
    agent.RegisterAgentFactory("reasoning-agent", factory)
}
```

### Builder Pattern

```go
// Fluent interface for agent creation
agent := agent.NewAgentDefinitionBuilder().
    WithName("custom-agent").
    WithType(agent.AgentTypeReAct).
    WithModel("gpt-4").
    WithProvider("openai").
    WithSystemPrompt("Custom system prompt").
    WithTemperature(0.8).
    WithMaxTokens(1500).
    WithTools("tool1", "tool2").
    WithMetadata("version", "2.0").
    Build()

agent.RegisterAgent("custom-agent", agent)
```

## CLI Commands

### Load Agent Definitions

```bash
# Load from plugin file
golanggraph multi-agent load ./agents.so

# Load from directory (planned)
golanggraph multi-agent load ./agents/ --recursive

# Load with validation
golanggraph multi-agent load ./agents.so --validate --verbose
```

### List Registered Agents

```bash
# Table format (default)
golanggraph multi-agent list

# JSON format
golanggraph multi-agent list --format json

# With metadata
golanggraph multi-agent list --show-metadata

# Filter by name
golanggraph multi-agent list --filter "chat"
```

### Example Output

```
Agent Definitions (4 total):

ID                   Source       Type            Model
--                   ------       ----            -----
chat-agent          definition   chat            gpt-3.5-turbo
reasoning-agent     factory      react           gpt-4
config-agent        config       tool            gpt-3.5-turbo
workflow-agent      plugin       react           gpt-4
```

## Integration with Multi-Agent Manager

The multi-agent manager automatically detects and uses Go-based agent definitions:

```go
// Priority order:
// 1. Go-based definitions (highest priority)
// 2. Go-based factories
// 3. Configuration files (fallback)

config := &agent.MultiAgentConfig{
    Agents: map[string]*agent.AgentConfig{
        "chat-agent": {...},      // Will use Go definition if available
        "config-only": {...},     // Will use config since no Go definition exists
    },
}

manager, _ := agent.NewMultiAgentManager(config, llmManager, toolRegistry)
```

## File Structure

### Recommended Project Layout

```
my-multi-agent-project/
├── agents/
│   ├── chat_agent.go          # Go-based agent definitions
│   ├── reasoning_agent.go
│   └── workflow_agent.go
├── configs/
│   ├── multi-agent.yaml       # Main configuration
│   └── agents/
│       ├── agent-1/config.yaml
│       └── agent-2/config.yaml
├── plugins/
│   ├── custom_agents.so       # Compiled plugins
│   └── third_party.so
├── tools/
│   ├── custom_tools.go        # Custom tool implementations
│   └── specialized_tools.go
├── docker-compose.yml
├── Dockerfile
└── README.md
```

### Example Multi-Agent Configuration

```yaml
name: "hybrid-multi-agent"
version: "1.0"
description: "Multi-agent system with Go and config definitions"

agents:
  # These will use Go definitions if available
  chat-agent:
    id: "chat-agent"
    name: "Chat Agent"
    type: "chat"
    model: "gpt-3.5-turbo"
    provider: "openai"
    
  reasoning-agent:
    id: "reasoning-agent"
    name: "Reasoning Agent"
    type: "react"
    model: "gpt-4"
    provider: "openai"
    
  # This will use config only
  config-only-agent:
    id: "config-only-agent"
    name: "Config Only Agent"
    type: "tool"
    model: "gpt-3.5-turbo"
    provider: "openai"
    tools: ["calculator", "web_search"]

routing:
  type: "path"
  rules:
    - id: "chat-rule"
      pattern: "/chat"
      agent_id: "chat-agent"
      method: "POST"
      priority: 1
    - id: "reason-rule"
      pattern: "/reason"
      agent_id: "reasoning-agent"
      method: "POST"
      priority: 2
    - id: "config-rule"
      pattern: "/config"
      agent_id: "config-only-agent"
      method: "POST"
      priority: 3
```

## Advanced Features

### Custom Tools Integration

```go
type CustomTool struct{}

func (ct *CustomTool) GetName() string {
    return "custom_tool"
}

func (ct *CustomTool) GetDefinition() llm.ToolDefinition {
    return llm.ToolDefinition{
        Type: "function",
        Function: llm.Function{
            Name:        "custom_tool",
            Description: "A custom tool for specialized tasks",
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "input": map[string]interface{}{
                        "type": "string",
                        "description": "Tool input",
                    },
                },
                "required": []string{"input"},
            },
        },
    }
}

func (ct *CustomTool) Execute(ctx context.Context, args string) (string, error) {
    // Custom tool logic
    return "Tool result", nil
}

// Add to advanced agent
func NewAdvancedAgent() *agent.AdvancedAgentDefinition {
    // ... config setup ...
    
    advanced := agent.NewAdvancedAgentDefinition(config)
    advanced.WithCustomTools(&CustomTool{})
    
    return advanced
}
```

### Plugin Development

```go
// plugin.go - compile with: go build -buildmode=plugin
package main

import "github.com/piotrlaczkowski/GoLangGraph/pkg/agent"

func GetAgentDefinitions() map[string]agent.AgentDefinition {
    return map[string]agent.AgentDefinition{
        "plugin-agent-1": NewPluginAgent1(),
        "plugin-agent-2": NewPluginAgent2(),
    }
}

func NewPluginAgent1() agent.AgentDefinition {
    // Agent implementation
}
```

Compile and load:
```bash
go build -buildmode=plugin -o agents.so plugin.go
golanggraph multi-agent load agents.so
```

## Best Practices

### 1. **Naming Conventions**
- Use descriptive, kebab-case names for agent IDs
- Include version information in metadata
- Add author and description metadata

### 2. **Error Handling**
- Implement comprehensive validation
- Provide meaningful error messages
- Handle initialization failures gracefully

### 3. **Performance**
- Use factories for expensive initialization
- Lazy load resources when possible
- Cache agent instances appropriately

### 4. **Testing**
- Unit test agent definitions
- Mock LLM and tool dependencies
- Test custom graph execution flows

### 5. **Deployment**
- Use plugins for production deployments
- Version your agent definitions
- Document agent capabilities and requirements

## Migration Guide

### From Config-Only to Hybrid

1. **Identify agents for Go conversion**
   - Complex logic agents
   - Custom tool requirements
   - Advanced workflow needs

2. **Create Go definitions**
   - Start with BaseAgentDefinition
   - Add custom logic incrementally
   - Preserve existing configurations

3. **Test compatibility**
   - Verify agent behavior matches
   - Test routing and deployment
   - Validate metrics and monitoring

4. **Gradual rollout**
   - Deploy one agent at a time
   - Monitor for issues
   - Keep config fallbacks

## Troubleshooting

### Common Issues

#### Agent Not Found
```bash
# Check registered agents
golanggraph multi-agent list

# Verify loading
golanggraph multi-agent load ./agents.so --verbose
```

#### Validation Errors
```go
// Check validation in agent definition
func (ad *MyAgentDefinition) Validate() error {
    if err := ad.BaseAgentDefinition.Validate(); err != nil {
        return err
    }
    // Custom validation
    return nil
}
```

#### Plugin Loading Issues
- Ensure plugin is compiled with same Go version
- Check function signatures match exactly
- Verify all dependencies are available

## Performance Considerations

- **Memory Usage**: Agent definitions are kept in memory
- **Startup Time**: Plugin loading adds initialization overhead
- **Concurrency**: Registry operations are thread-safe
- **Caching**: Consider caching expensive agent creations

## Security Considerations

- **Plugin Security**: Only load trusted plugins
- **Code Injection**: Validate all agent inputs
- **Access Control**: Implement proper authentication
- **Resource Limits**: Set appropriate timeouts and limits

## Future Enhancements

- **Hot Reloading**: Dynamic agent updates
- **Remote Loading**: Load agents from repositories
- **Visual Builder**: GUI for agent creation
- **Template System**: Reusable agent templates
- **Version Management**: Agent versioning and rollback

## Conclusion

The Go-based agent definition system provides powerful flexibility for creating sophisticated multi-agent systems while maintaining compatibility with existing configuration approaches. This hybrid model allows teams to start simple with YAML configurations and gradually migrate to programmatic definitions as their needs become more complex.

The system supports the full spectrum from simple configuration-driven agents to complex programmatic agents with custom graphs, tools, and workflows, all managed through a unified CLI and deployment system.
