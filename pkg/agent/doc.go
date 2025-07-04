// Package agent provides intelligent agent implementations for building AI-powered workflows.
//
// The agent package implements various types of AI agents that can interact with Large Language Models (LLMs),
// use tools, and execute complex reasoning patterns. This package is designed to support multiple agent
// architectures including chat agents, ReAct (Reasoning and Acting) agents, and tool-using agents.
//
// # Agent Types
//
// The package supports several agent types:
//
//   - Chat Agent: Simple conversational agent for basic interactions
//   - ReAct Agent: Implements the ReAct pattern for reasoning and acting
//   - Tool Agent: Specialized agent that can use external tools
//   - Custom Agent: Extensible agent type for custom implementations
//
// # Basic Usage
//
// Creating a simple chat agent:
//
//	config := agent.Config{
//		Name:        "ChatBot",
//		Type:        "chat",
//		Description: "A helpful chatbot",
//		MaxSteps:    5,
//		Temperature: 0.7,
//	}
//
//	agent, err := agent.NewAgent(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Set LLM provider
//	provider, _ := llm.NewOpenAIProvider(llm.OpenAIConfig{
//		APIKey: "your-api-key",
//		Model:  "gpt-3.5-turbo",
//	})
//	agent.SetLLMProvider(provider)
//
//	// Execute the agent
//	ctx := context.Background()
//	state := core.NewBaseState()
//	state.Set("input", "Hello, how are you?")
//
//	result, err := agent.Execute(ctx, state)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println(result.Get("response"))
//
// # ReAct Pattern
//
// The ReAct (Reasoning and Acting) pattern allows agents to alternate between reasoning about
// a problem and taking actions to solve it:
//
//	config := agent.Config{
//		Name:        "ReActAgent",
//		Type:        "react",
//		Description: "An agent that can reason and act",
//		MaxSteps:    10,
//		Temperature: 0.3,
//	}
//
//	agent, err := agent.NewAgent(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Add tools for the agent to use
//	toolRegistry := tools.NewToolRegistry()
//	toolRegistry.Register("calculator", tools.NewCalculatorTool())
//	toolRegistry.Register("search", tools.NewWebSearchTool())
//	agent.SetToolRegistry(toolRegistry)
//
// # Tool Integration
//
// Agents can use external tools to extend their capabilities:
//
//	// Create a tool-using agent
//	config := agent.Config{
//		Name: "ToolAgent",
//		Type: "tool",
//	}
//
//	agent, err := agent.NewAgent(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Register tools
//	toolRegistry := tools.NewToolRegistry()
//	toolRegistry.Register("file_read", tools.NewFileReadTool())
//	toolRegistry.Register("http_request", tools.NewHTTPTool())
//	agent.SetToolRegistry(toolRegistry)
//
// # Multi-Agent Coordination
//
// The package supports multi-agent systems where agents can coordinate and collaborate:
//
//	coordinator := agent.NewMultiAgentCoordinator()
//
//	// Add agents to the coordinator
//	coordinator.AddAgent("researcher", researchAgent)
//	coordinator.AddAgent("writer", writerAgent)
//	coordinator.AddAgent("reviewer", reviewAgent)
//
//	// Execute coordinated workflow
//	result, err := coordinator.Execute(ctx, task)
//
// # Configuration Options
//
// Agents can be configured with various options:
//
//   - Name: Human-readable name for the agent
//   - Type: Agent type (chat, react, tool, custom)
//   - Description: Description of the agent's purpose
//   - MaxSteps: Maximum number of execution steps
//   - Temperature: LLM temperature for response generation
//   - SystemPrompt: System prompt for the agent
//   - Tools: List of available tools
//   - Memory: Memory configuration for conversation history
//
// # Error Handling
//
// The package provides comprehensive error handling:
//
//	result, err := agent.Execute(ctx, state)
//	if err != nil {
//		switch {
//		case errors.Is(err, agent.ErrMaxStepsExceeded):
//			// Handle max steps exceeded
//		case errors.Is(err, agent.ErrInvalidInput):
//			// Handle invalid input
//		case errors.Is(err, agent.ErrLLMProviderNotSet):
//			// Handle missing LLM provider
//		default:
//			// Handle other errors
//		}
//	}
//
// # Performance Considerations
//
// For optimal performance:
//
//   - Reuse agent instances when possible
//   - Set appropriate MaxSteps to prevent infinite loops
//   - Use connection pooling for LLM providers
//   - Implement proper context cancellation
//   - Monitor memory usage in long-running agents
//
// # Thread Safety
//
// Agent instances are thread-safe and can be used concurrently. However, individual
// execution contexts should not be shared between goroutines.
//
// # Integration with Core Package
//
// The agent package is tightly integrated with the core package for graph-based workflows:
//
//	// Use agent as a node in a graph
//	graph := core.NewGraph()
//	graph.AddNode("agent_node", func(ctx context.Context, state core.State) (core.State, error) {
//		return agent.Execute(ctx, state)
//	})
//
// For more examples and detailed usage, see the examples directory and the comprehensive
// test suite in agent_test.go.
package agent
