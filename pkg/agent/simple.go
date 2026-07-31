// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
package agent

import (
	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent/backends"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// CreateDeepAgent creates a fully-configured Deep Agent with all middleware enabled.
// This provides a "batteries-included" experience similar to Python's create_deep_agent.
//
// Features included:
// - TodoListMiddleware (Planning)
// - FilesystemMiddleware (Context/Memory)
// - SubAgentMiddleware (Delegation)
// - SummarizationMiddleware (Long context handling)
func CreateDeepAgent(config *AgentConfig, llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) Agent {
	// 1. Ensure config has necessary defaults
	if config.Type == "" {
		config.Type = AgentTypeChat // Deep agents typically use chat/react loop
	}

	// 2. Initialize Middleware slice if empty
	if config.Middleware == nil {
		config.Middleware = make([]Middleware, 0)
	}

	// 3. Add TodoList Middleware (Planning)
	// Adds 'write_todos' tool and prompts agent to plan complex tasks
	config.Middleware = append(config.Middleware, NewTodoListMiddleware())

	// 4. Add Filesystem Middleware (Context)
	// Adds 'read_file', 'write_file', 'ls', etc.
	// Uses in-memory state backend by default for ephemeral context
	fsMiddleware := NewFilesystemMiddleware(backends.NewStateBackend(func() map[string]interface{} {
		return make(map[string]interface{}) // Placeholder, actual state injection happens at runtime
	}))
	config.Middleware = append(config.Middleware, fsMiddleware)

	// 5. Add SubAgent Middleware (Delegation)
	// Adds 'delegate_task' tool
	subAgentMiddleware := NewSubAgentMiddleware()

	// Register a general-purpose sub-agent clone
	// This allows the agent to offload tasks to a copy of itself
	// We create a new config for the sub-agent based on the parent
	subConfig := *config
	subConfig.Name = config.Name + "-sub"
	subConfig.Middleware = []Middleware{} // Sub-agents don't inherit middleware by default to avoid infinite recursion depth issues
	// But for a "Deep" sub-agent, maybe we want some? For now, keep it simple.

	subAgent := NewAgent(&subConfig, llmManager, toolRegistry)
	subAgentMiddleware.RegisterSubAgent("general-purpose", subAgent)

	config.Middleware = append(config.Middleware, subAgentMiddleware)

	// 6. Add Summarization Middleware (Memory Management)
	// Automatically summarizes conversation when it gets too long
	config.Middleware = append(config.Middleware, NewSummarizationMiddleware(20)) // Default to 20 messages

	// 7. Create the Agent
	agent := NewAgent(config, llmManager, toolRegistry)

	return agent
}

// CreateReActAgent creates a standard ReAct agent without deep middleware.
// Useful for simpler tasks or as a base for custom agents.
func CreateReActAgent(config *AgentConfig, llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) Agent {
	config.Type = AgentTypeReAct
	return NewAgent(config, llmManager, toolRegistry)
}

// NewSimpleAgent creates a simple chat agent with minimal configuration
func NewSimpleAgent(name, model, provider string) (Agent, error) {
	config := &AgentConfig{
		Name:          name,
		Type:          AgentTypeChat,
		Model:         model,
		Provider:      provider,
		Temperature:   0.7,
		MaxTokens:     2000,
		MaxIterations: 5,
	}

	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	agent := NewAgent(config, llmManager, toolRegistry)
	return agent, nil
}

// NewQuickAgent creates a QuickStart agent with default Ollama LLM
func NewQuickAgent(name string) (Agent, error) {
	config := DefaultAgentConfig()
	config.Name = name

	// Use Ollama by default
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	return NewAgent(config, llmManager, toolRegistry), nil
}
