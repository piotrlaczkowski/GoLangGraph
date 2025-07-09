// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Example Agent Definitions

package main

import (
	"context"
	"fmt"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// ChatAgentDefinition demonstrates a programmatic chat agent definition
type ChatAgentDefinition struct {
	*agent.BaseAgentDefinition
	conversationMemory []string
}

// NewChatAgentDefinition creates a new chat agent definition
func NewChatAgentDefinition() *ChatAgentDefinition {
	config := &agent.AgentConfig{
		Name:         "chat-agent",
		Type:         agent.AgentTypeChat,
		Model:        "gpt-3.5-turbo",
		Provider:     "openai",
		SystemPrompt: "You are a helpful chat assistant. Be friendly and concise.",
		Temperature:  0.7,
		MaxTokens:    1000,
		Tools:        []string{"web_search", "calculator"},
	}

	chatAgent := &ChatAgentDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
		conversationMemory:  make([]string, 0),
	}

	// Add custom metadata
	chatAgent.SetMetadata("version", "1.0.0")
	chatAgent.SetMetadata("author", "GoLangGraph Team")
	chatAgent.SetMetadata("description", "A conversational chat agent with memory")
	chatAgent.SetMetadata("capabilities", []string{"chat", "memory", "web_search", "calculations"})

	return chatAgent
}

// Initialize customizes the agent setup
func (cad *ChatAgentDefinition) Initialize(llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) error {
	// Call base initialization
	if err := cad.BaseAgentDefinition.Initialize(llmManager, toolRegistry); err != nil {
		return err
	}

	// Custom initialization logic
	cad.SetMetadata("initialized_at", "startup")
	cad.SetMetadata("memory_enabled", true)

	return nil
}

// CreateAgent creates a specialized chat agent with memory
func (cad *ChatAgentDefinition) CreateAgent() (*agent.Agent, error) {
	// Create base agent
	baseAgent, err := cad.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Add memory functionality metadata
	cad.SetMetadata("memory_size", len(cad.conversationMemory))
	cad.SetMetadata("features", []string{"persistent_memory", "context_awareness"})

	return baseAgent, nil
}

// Validate performs custom validation
func (cad *ChatAgentDefinition) Validate() error {
	// Call base validation
	if err := cad.BaseAgentDefinition.Validate(); err != nil {
		return err
	}

	// Custom validation
	config := cad.GetConfig()
	if config.Temperature < 0.1 || config.Temperature > 1.0 {
		return fmt.Errorf("temperature must be between 0.1 and 1.0")
	}

	if config.MaxTokens < 100 {
		return fmt.Errorf("max_tokens must be at least 100 for chat agents")
	}

	return nil
}

// AddMemory adds to conversation memory
func (cad *ChatAgentDefinition) AddMemory(memory string) {
	cad.conversationMemory = append(cad.conversationMemory, memory)

	// Keep only last 10 memories
	if len(cad.conversationMemory) > 10 {
		cad.conversationMemory = cad.conversationMemory[len(cad.conversationMemory)-10:]
	}

	cad.SetMetadata("memory_size", len(cad.conversationMemory))
}

// GetMemory returns conversation memory
func (cad *ChatAgentDefinition) GetMemory() []string {
	return cad.conversationMemory
}

// ReasoningAgentDefinition demonstrates an advanced reasoning agent
type ReasoningAgentDefinition struct {
	*agent.AdvancedAgentDefinition
	reasoningSteps int
}

// NewReasoningAgentDefinition creates a new reasoning agent definition
func NewReasoningAgentDefinition() *ReasoningAgentDefinition {
	config := &agent.AgentConfig{
		Name:          "reasoning-agent",
		Type:          agent.AgentTypeReAct,
		Model:         "gpt-4",
		Provider:      "openai",
		SystemPrompt:  "You are an advanced reasoning agent. Think step by step and use tools when needed.",
		Temperature:   0.3,
		MaxTokens:     2000,
		MaxIterations: 5,
		Tools:         []string{"web_search", "calculator", "file_read"},
	}

	reasoningAgent := &ReasoningAgentDefinition{
		AdvancedAgentDefinition: agent.NewAdvancedAgentDefinition(config),
		reasoningSteps:          0,
	}

	// Add custom tools
	reasoningAgent.WithCustomTools(&LogicTool{})

	// Add custom metadata
	reasoningAgent.SetMetadata("version", "2.0.0")
	reasoningAgent.SetMetadata("reasoning_type", "step_by_step")
	reasoningAgent.SetMetadata("max_reasoning_depth", 10)

	return reasoningAgent
}

// BuildGraph builds a custom reasoning graph
func (rad *ReasoningAgentDefinition) BuildGraph() (*core.Graph, error) {
	// Create custom graph for reasoning
	graph := core.NewGraph("reasoning-graph")

	// Add custom reasoning nodes
	_ = graph.AddNode("plan", "Plan", rad.planningNode)
	_ = graph.AddNode("reason", "Reason", rad.reasoningNode)
	_ = graph.AddNode("validate", "Validate", rad.validationNode)
	_ = graph.AddNode("finalize", "Finalize", rad.finalizationNode)

	// Connect nodes
	graph.AddEdge("plan", "reason", nil)
	graph.AddEdge("reason", "validate", nil)
	graph.AddEdge("validate", "reason", rad.needsMoreReasoning)
	graph.AddEdge("validate", "finalize", rad.reasoningComplete)

	// Set start and end nodes
	graph.SetStartNode("plan")
	graph.AddEndNode("finalize")

	return graph, nil
}

// Custom node implementations
func (rad *ReasoningAgentDefinition) planningNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Planning logic
	state.Set("plan", "Created reasoning plan")
	return state, nil
}

func (rad *ReasoningAgentDefinition) reasoningNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Reasoning logic
	rad.reasoningSteps++
	state.Set("reasoning_step", rad.reasoningSteps)
	state.Set("reasoning", fmt.Sprintf("Reasoning step %d", rad.reasoningSteps))
	return state, nil
}

func (rad *ReasoningAgentDefinition) validationNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Validation logic
	state.Set("validation", "Reasoning validated")
	return state, nil
}

func (rad *ReasoningAgentDefinition) finalizationNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Finalization logic
	state.Set("final_reasoning", fmt.Sprintf("Completed after %d steps", rad.reasoningSteps))
	return state, nil
}

// Condition functions
func (rad *ReasoningAgentDefinition) needsMoreReasoning(ctx context.Context, state *core.BaseState) (string, error) {
	if rad.reasoningSteps < 3 {
		return "reason", nil
	}
	return "finalize", nil
}

func (rad *ReasoningAgentDefinition) reasoningComplete(ctx context.Context, state *core.BaseState) (string, error) {
	return "finalize", nil
}

// LogicTool is a custom tool for logical reasoning
type LogicTool struct{}

func (lt *LogicTool) GetName() string {
	return "logic_tool"
}

func (lt *LogicTool) GetDescription() string {
	return "Performs logical reasoning operations"
}

func (lt *LogicTool) GetDefinition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.Function{
			Name:        lt.GetName(),
			Description: lt.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "Logic operation to perform",
						"enum":        []string{"and", "or", "not", "implies"},
					},
					"premises": map[string]interface{}{
						"type":        "array",
						"description": "List of logical premises",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"operation", "premises"},
			},
		},
	}
}

func (lt *LogicTool) Execute(ctx context.Context, args string) (string, error) {
	// Simple logic tool implementation
	return fmt.Sprintf("Logical operation performed on: %s", args), nil
}

func (lt *LogicTool) Validate(args string) error {
	// Validation logic
	return nil
}

func (lt *LogicTool) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"type": "logic",
	}
}

func (lt *LogicTool) SetConfig(config map[string]interface{}) error {
	return nil
}

// Register agents with the global registry
func init() {
	// Register the chat agent
	chatAgent := NewChatAgentDefinition()
	if err := agent.RegisterAgent("chat-agent", chatAgent); err != nil {
		panic(fmt.Sprintf("Failed to register chat agent: %v", err))
	}

	// Register the reasoning agent using a factory
	reasoningFactory := func() agent.AgentDefinition {
		return NewReasoningAgentDefinition()
	}
	if err := agent.RegisterAgentFactory("reasoning-agent", reasoningFactory); err != nil {
		panic(fmt.Sprintf("Failed to register reasoning agent factory: %v", err))
	}
}

// GetAgentDefinitions is required for plugin loading
func GetAgentDefinitions() map[string]agent.AgentDefinition {
	return map[string]agent.AgentDefinition{
		"chat-agent":      NewChatAgentDefinition(),
		"reasoning-agent": NewReasoningAgentDefinition(),
	}
}

// Example main function for testing
func main() {
	fmt.Println("Agent definitions loaded successfully!")

	// Display registered agents
	registry := agent.GetGlobalRegistry()

	fmt.Println("Registered definitions:", registry.ListDefinitions())
	fmt.Println("Registered factories:", registry.ListFactories())

	// Get agent info
	infos := registry.GetAgentInfo()
	for _, info := range infos {
		fmt.Printf("Agent: %s, Source: %s, Type: %s\n",
			info.ID, info.Source, info.Config.Type)
	}
}
