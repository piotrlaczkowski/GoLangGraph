// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package builder

import (
	"context"
	"testing"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
)

func TestNewQuickBuilder(t *testing.T) {
	builder := NewQuickBuilder()

	if builder == nil {
		t.Fatal("NewQuickBuilder returned nil")
	}

	if builder.llmManager == nil {
		t.Error("LLM manager should not be nil")
	}

	if builder.toolRegistry == nil {
		t.Error("Tool registry should not be nil")
	}

	if builder.config == nil {
		t.Error("Config should not be nil")
	}

	if builder.checkpointer == nil {
		t.Error("Checkpointer should not be nil")
	}
}

func TestDefaultQuickConfig(t *testing.T) {
	config := DefaultQuickConfig()

	if config == nil {
		t.Fatal("DefaultQuickConfig returned nil")
	}

	if config.DefaultModel == "" {
		t.Error("DefaultModel should not be empty")
	}

	if config.Temperature == 0 {
		t.Error("Temperature should not be zero")
	}

	if config.MaxTokens == 0 {
		t.Error("MaxTokens should not be zero")
	}

	if config.MaxIterations == 0 {
		t.Error("MaxIterations should not be zero")
	}

	if !config.UseMemory {
		t.Error("UseMemory should be true by default")
	}

	if !config.EnableAllTools {
		t.Error("EnableAllTools should be true by default")
	}
}

func TestQuickBuilder_WithConfig(t *testing.T) {
	builder := NewQuickBuilder()
	customConfig := &QuickConfig{
		DefaultModel:   "custom-model",
		Temperature:    0.5,
		MaxTokens:      2000,
		MaxIterations:  20,
		UseMemory:      false,
		EnableAllTools: false,
	}

	result := builder.WithConfig(customConfig)

	if result != builder {
		t.Error("WithConfig should return the same builder instance")
	}

	if builder.config.DefaultModel != "custom-model" {
		t.Error("DefaultModel should be updated")
	}

	if builder.config.Temperature != 0.5 {
		t.Error("Temperature should be updated")
	}

	if builder.config.MaxTokens != 2000 {
		t.Error("MaxTokens should be updated")
	}

	if builder.config.MaxIterations != 20 {
		t.Error("MaxIterations should be updated")
	}

	if builder.config.UseMemory {
		t.Error("UseMemory should be false")
	}

	if builder.config.EnableAllTools {
		t.Error("EnableAllTools should be false")
	}
}

func TestQuickBuilder_Chat(t *testing.T) {
	builder := NewQuickBuilder()

	// Test with default name
	chatAgent := builder.Chat()
	if chatAgent == nil {
		t.Fatal("Chat agent should not be nil")
	}

	config := chatAgent.GetConfig()
	if config.Type != agent.AgentTypeChat {
		t.Error("Agent type should be Chat")
	}

	if config.Name != "ChatAgent" {
		t.Error("Default name should be ChatAgent")
	}

	// Test with custom name
	customChatAgent := builder.Chat("CustomChat")
	customConfig := customChatAgent.GetConfig()
	if customConfig.Name != "CustomChat" {
		t.Error("Custom name should be set")
	}
}

func TestQuickBuilder_ReAct(t *testing.T) {
	builder := NewQuickBuilder()

	reactAgent := builder.ReAct("TestReAct")
	if reactAgent == nil {
		t.Fatal("ReAct agent should not be nil")
	}

	config := reactAgent.GetConfig()
	if config.Type != agent.AgentTypeReAct {
		t.Error("Agent type should be ReAct")
	}

	if config.Name != "TestReAct" {
		t.Error("Name should be TestReAct")
	}

	if len(config.Tools) == 0 {
		t.Error("ReAct agent should have tools")
	}
}

func TestQuickBuilder_Tool(t *testing.T) {
	builder := NewQuickBuilder()

	toolAgent := builder.Tool("TestTool")
	if toolAgent == nil {
		t.Fatal("Tool agent should not be nil")
	}

	config := toolAgent.GetConfig()
	if config.Type != agent.AgentTypeTool {
		t.Error("Agent type should be Tool")
	}

	if config.Name != "TestTool" {
		t.Error("Name should be TestTool")
	}

	if len(config.Tools) == 0 {
		t.Error("Tool agent should have tools")
	}
}

func TestQuickBuilder_RAG(t *testing.T) {
	builder := NewQuickBuilder()

	ragAgent := builder.RAG("TestRAG")
	if ragAgent == nil {
		t.Fatal("RAG agent should not be nil")
	}

	config := ragAgent.GetConfig()
	if config.Type != agent.AgentTypeChat {
		t.Error("Agent type should be Chat for RAG")
	}

	if config.Name != "TestRAG" {
		t.Error("Name should be TestRAG")
	}

	if len(config.Tools) == 0 {
		t.Error("RAG agent should have tools")
	}
}

func TestQuickBuilder_Specialized(t *testing.T) {
	builder := NewQuickBuilder()

	tests := []struct {
		name   string
		create func() *agent.Agent
	}{
		{"Researcher", func() *agent.Agent { return builder.Researcher("TestResearcher") }},
		{"Writer", func() *agent.Agent { return builder.Writer("TestWriter") }},
		{"Analyst", func() *agent.Agent { return builder.Analyst("TestAnalyst") }},
		{"Coder", func() *agent.Agent { return builder.Coder("TestCoder") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := tt.create()
			if agent == nil {
				t.Fatalf("%s agent should not be nil", tt.name)
			}

			config := agent.GetConfig()
			if config.Name != "Test"+tt.name {
				t.Errorf("Name should be Test%s", tt.name)
			}
		})
	}
}

func TestQuickBuilder_Multi(t *testing.T) {
	builder := NewQuickBuilder()

	multi := builder.Multi()
	if multi == nil {
		t.Fatal("Multi coordinator should not be nil")
	}

	// Test adding agents
	chatAgent := builder.Chat("TestChat")
	multi.AddAgent("chat", chatAgent)

	agents := multi.ListAgents()
	if len(agents) != 1 {
		t.Error("Should have 1 agent")
	}

	if agents[0] != "chat" {
		t.Error("Agent ID should be 'chat'")
	}

	// Test getting agent
	retrievedAgent, exists := multi.GetAgent("chat")
	if !exists {
		t.Error("Agent should exist")
	}

	if retrievedAgent != chatAgent {
		t.Error("Retrieved agent should be the same as added agent")
	}
}

func TestQuickBuilder_Pipeline(t *testing.T) {
	builder := NewQuickBuilder()

	agent1 := builder.Chat("Agent1")
	agent2 := builder.Chat("Agent2")

	pipeline := builder.Pipeline(agent1, agent2)
	if pipeline == nil {
		t.Fatal("Pipeline should not be nil")
	}

	if len(pipeline.agents) != 2 {
		t.Error("Pipeline should have 2 agents")
	}

	if pipeline.agents[0] != agent1 {
		t.Error("First agent should be agent1")
	}

	if pipeline.agents[1] != agent2 {
		t.Error("Second agent should be agent2")
	}
}

func TestQuickBuilder_Swarm(t *testing.T) {
	builder := NewQuickBuilder()

	agent1 := builder.Chat("Agent1")
	agent2 := builder.Chat("Agent2")

	swarm := builder.Swarm(agent1, agent2)
	if swarm == nil {
		t.Fatal("Swarm should not be nil")
	}

	if len(swarm.agents) != 2 {
		t.Error("Swarm should have 2 agents")
	}

	if swarm.agents[0] != agent1 {
		t.Error("First agent should be agent1")
	}

	if swarm.agents[1] != agent2 {
		t.Error("Second agent should be agent2")
	}
}

func TestAgentPipeline_Execute(t *testing.T) {
	// Skip this test as it requires actual LLM providers
	t.Skip("Skipping pipeline execution test - requires actual LLM providers")
}

func TestAgentSwarm_Execute(t *testing.T) {
	// Skip this test as it requires actual LLM providers
	t.Skip("Skipping swarm execution test - requires actual LLM providers")
}

func TestGlobalQuickFunctions(t *testing.T) {
	// Test global convenience functions
	t.Run("OneLineChat", func(t *testing.T) {
		agent := OneLineChat("TestChat")
		if agent == nil {
			t.Fatal("OneLineChat should not return nil")
		}

		config := agent.GetConfig()
		if config.Name != "TestChat" {
			t.Error("Agent name should be TestChat")
		}
	})

	t.Run("OneLineReAct", func(t *testing.T) {
		agent := OneLineReAct("TestReAct")
		if agent == nil {
			t.Fatal("OneLineReAct should not return nil")
		}

		config := agent.GetConfig()
		if config.Type != "react" {
			t.Error("Agent type should be ReAct")
		}
	})

	t.Run("OneLineTool", func(t *testing.T) {
		agent := OneLineTool("TestTool")
		if agent == nil {
			t.Fatal("OneLineTool should not return nil")
		}

		config := agent.GetConfig()
		if config.Type != "tool" {
			t.Error("Agent type should be Tool")
		}
	})

	t.Run("OneLineRAG", func(t *testing.T) {
		agent := OneLineRAG("TestRAG")
		if agent == nil {
			t.Fatal("OneLineRAG should not return nil")
		}

		config := agent.GetConfig()
		if config.Name != "TestRAG" {
			t.Error("Agent name should be TestRAG")
		}
	})

	t.Run("OneLinePipeline", func(t *testing.T) {
		agent1 := OneLineChat("Agent1")
		agent2 := OneLineChat("Agent2")
		pipeline := OneLinePipeline(agent1, agent2)

		if pipeline == nil {
			t.Fatal("OneLinePipeline should not return nil")
		}

		if len(pipeline.agents) != 2 {
			t.Error("Pipeline should have 2 agents")
		}
	})

	t.Run("OneLineSwarm", func(t *testing.T) {
		agent1 := OneLineChat("Agent1")
		agent2 := OneLineChat("Agent2")
		swarm := OneLineSwarm(agent1, agent2)

		if swarm == nil {
			t.Fatal("OneLineSwarm should not return nil")
		}

		if len(swarm.agents) != 2 {
			t.Error("Swarm should have 2 agents")
		}
	})
}

func TestQuickBuilder_getBestProvider(t *testing.T) {
	builder := NewQuickBuilder()

	// This is a private method, so we test it indirectly
	// by creating an agent and checking its provider
	agent := builder.Chat("TestAgent")
	config := agent.GetConfig()

	// The provider should be set to something
	if config.Provider == "" {
		t.Error("Provider should not be empty")
	}
}

// Benchmark tests
func BenchmarkNewQuickBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewQuickBuilder()
	}
}

func BenchmarkQuickBuilder_Chat(b *testing.B) {
	builder := NewQuickBuilder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Chat("TestAgent")
	}
}

func BenchmarkQuickBuilder_ReAct(b *testing.B) {
	builder := NewQuickBuilder()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.ReAct("TestAgent")
	}
}

func BenchmarkOneLineChat(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		OneLineChat("TestAgent")
	}
}

func BenchmarkPipelineExecution(b *testing.B) {
	agent1 := OneLineChat("Agent1")
	agent2 := OneLineChat("Agent2")
	pipeline := OneLinePipeline(agent1, agent2)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.Execute(ctx, "test input")
	}
}
