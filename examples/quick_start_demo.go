// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package examples

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// This file demonstrates minimal code examples for creating agents
// Run with: go run quick_start_demo.go

func QuickStartDemo() {
	fmt.Println("🚀 GoLangGraph Minimal Code Examples")
	fmt.Println("====================================")
	fmt.Println()

	ctx := context.Background()

	// Example 1: Simple Chat Agent (3 lines of code)
	fmt.Println("📝 Example 1: Simple Chat Agent (3 lines)")
	chatAgent := CreateSimpleChatAgent()
	response1, _ := chatAgent.Execute(ctx, "Hello! Tell me about Go programming.")
	fmt.Printf("Response: %s\n\n", response1.Output)

	// Example 2: ReAct Agent with Tools (4 lines of code)
	fmt.Println("🔧 Example 2: ReAct Agent with Tools (4 lines)")
	reactAgent := CreateReActAgent()
	response2, _ := reactAgent.Execute(ctx, "Calculate the square root of 144")
	fmt.Printf("Response: %s\n\n", response2.Output)

	// Example 3: Multi-Agent Collaboration (5 lines of code)
	fmt.Println("👥 Example 3: Multi-Agent Collaboration (5 lines)")
	coordinator := CreateMultiAgentSystem()
	responses, _ := coordinator.ExecuteSequential(ctx, []string{"researcher", "writer"}, "Research Go benefits")
	for i, resp := range responses {
		fmt.Printf("Agent %d Response: %s\n", i+1, resp.Output)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✨ Ultra-Minimal Examples (1-2 lines each)")
	fmt.Println(strings.Repeat("=", 50))

	// Ultra-minimal examples
	quickAgent := QuickChat()
	resp, _ := quickAgent.Execute(ctx, "Hello!")
	fmt.Printf("One-liner Chat Agent: %s\n", resp.Output)

	quickReact := QuickReAct()
	resp2, _ := quickReact.Execute(ctx, "Calculate 15 + 27")
	fmt.Printf("One-liner ReAct Agent: %s\n", resp2.Output)

	fmt.Println("\n✨ All examples completed successfully!")
	fmt.Println("\n💡 Key Takeaways:")
	fmt.Println("• Create agents with minimal code (1-5 lines)")
	fmt.Println("• Support for multiple agent types (Chat, ReAct, Tool)")
	fmt.Println("• Built-in tool integration")
	fmt.Println("• Multi-agent coordination")
	fmt.Println("• Production-ready with persistence and streaming")
}

// Example 1: Simple Chat Agent - Just 3 lines!
func CreateSimpleChatAgent() *agent.Agent {
	config := &agent.AgentConfig{Name: "ChatBot", Type: agent.AgentTypeChat, SystemPrompt: "You are a helpful AI assistant specialized in Go programming."}
	llmManager := createMockLLMManager()
	return agent.NewAgent(config, llmManager, tools.NewToolRegistry())
}

// Example 2: ReAct Agent with Tools - Just 4 lines!
func CreateReActAgent() *agent.Agent {
	config := &agent.AgentConfig{Name: "ReActAgent", Type: agent.AgentTypeReAct, SystemPrompt: "You are a helpful assistant that can reason and use tools.", Tools: []string{"calculator"}}
	llmManager := createMockLLMManager()
	toolRegistry := createToolRegistry()
	return agent.NewAgent(config, llmManager, toolRegistry)
}

// Example 3: Multi-Agent System - Just 5 lines!
func CreateMultiAgentSystem() *agent.MultiAgentCoordinator {
	coordinator := agent.NewMultiAgentCoordinator()
	researcher := agent.NewAgent(&agent.AgentConfig{Name: "Researcher", Type: agent.AgentTypeReAct, SystemPrompt: "You are a research specialist.", Tools: []string{"web_search"}}, createMockLLMManager(), createToolRegistry())
	writer := agent.NewAgent(&agent.AgentConfig{Name: "Writer", Type: agent.AgentTypeChat, SystemPrompt: "You are a technical writer."}, createMockLLMManager(), tools.NewToolRegistry())
	coordinator.AddAgent("researcher", researcher)
	coordinator.AddAgent("writer", writer)
	return coordinator
}

// One-liner functions for ultra-minimal agent creation

// OneLiner: Chat Agent
func QuickChat() *agent.Agent {
	return agent.NewAgent(&agent.AgentConfig{Name: "QuickChat", Type: agent.AgentTypeChat}, createMockLLMManager(), tools.NewToolRegistry())
}

// OneLiner: ReAct Agent
func QuickReAct() *agent.Agent {
	return agent.NewAgent(&agent.AgentConfig{Name: "QuickReAct", Type: agent.AgentTypeReAct, Tools: []string{"calculator"}}, createMockLLMManager(), createToolRegistry())
}

// Helper functions for mock examples

func createMockLLMManager() *llm.ProviderManager {
	manager := llm.NewProviderManager()
	mockProvider := &MockProvider{}
	manager.RegisterProvider("mock", mockProvider)
	return manager
}

func createToolRegistry() *tools.ToolRegistry {
	registry := tools.NewToolRegistry()
	registry.RegisterTool(&tools.CalculatorTool{})
	registry.RegisterTool(&tools.WebSearchTool{})
	return registry
}

// MockProvider for demonstration
type MockProvider struct{}

func (m *MockProvider) GetName() string { return "mock" }
func (m *MockProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"mock-model"}, nil
}

func (m *MockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	lastMessage := req.Messages[len(req.Messages)-1].Content

	var response string
	switch {
	case contains(lastMessage, "Hello"):
		response = "Hello! I'm here to help you with Go programming and any other questions you have."
	case contains(lastMessage, "square root") || contains(lastMessage, "144"):
		response = "The square root of 144 is 12. This is because 12 × 12 = 144."
	case contains(lastMessage, "Go programming") || contains(lastMessage, "Go benefits"):
		response = "Go is an excellent programming language known for its simplicity, performance, and excellent concurrency support with goroutines."
	case contains(lastMessage, "15 + 27") || contains(lastMessage, "15+27"):
		response = "15 + 27 = 42"
	case contains(lastMessage, "Research"):
		response = "I've researched Go programming benefits: excellent performance, simple syntax, built-in concurrency, strong standard library, and great tooling."
	default:
		response = "I understand your request and I'm here to help!"
	}

	return &llm.CompletionResponse{
		ID:      "mock-response",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mock-model",
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: response,
				},
				FinishReason: "stop",
			},
		},
		Usage: llm.Usage{
			PromptTokens:     len(lastMessage) / 4,
			CompletionTokens: len(response) / 4,
			TotalTokens:      (len(lastMessage) + len(response)) / 4,
		},
	}, nil
}

func (m *MockProvider) CompleteStream(ctx context.Context, req llm.CompletionRequest, callback llm.StreamCallback) error {
	return nil
}

func (m *MockProvider) IsHealthy(ctx context.Context) error           { return nil }
func (m *MockProvider) GetConfig() map[string]interface{}             { return map[string]interface{}{} }
func (m *MockProvider) SetConfig(config map[string]interface{}) error { return nil }
func (m *MockProvider) Close() error                                  { return nil }

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
