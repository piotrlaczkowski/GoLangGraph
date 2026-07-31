// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
package agent

import (
	"context"
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
	"github.com/stretchr/testify/assert"
)

func TestSubAgentExecutor_Sequential(t *testing.T) {
	// Setup
	registry := NewAgentRegistry()
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Create test subagents
	config1 := DefaultAgentConfig()
	config1.Name = "agent1"
	config1.Model = "test"
	config1.Provider = "test"

	def1 := NewBaseAgentDefinition(config1)
	def1.Initialize(llmManager, toolRegistry)
	registry.RegisterDefinition("agent1", def1)

	// Create executor
	executor := NewSubAgentExecutor(registry)
	executor.SetParallel(false) // Sequential

	// Create requests
	requests := []SubAgentRequest{
		{AgentID: "agent1", Input: "test1", Timeout: 5 * time.Second},
	}

	sharedState := NewSharedState()

	// Execute
	results, err := executor.ExecuteSubAgents(context.Background(), requests, sharedState)

	// We expect an error because the agent execution will fail without a real LLM provider
	assert.Error(t, err)

	// Non-existent agent will error, but executor should handle it
	assert.Len(t, results, 1)
	// Error expected since we don't have a real LLM
	assert.NotNil(t, results[0])
}

func TestSubAgentExecutor_Parallel(t *testing.T) {
	registry := NewAgentRegistry()
	executor := NewSubAgentExecutor(registry)
	executor.SetParallel(true)
	executor.SetTimeout(1 * time.Second)

	// Empty requests
	results, err := executor.ExecuteSubAgents(context.Background(), []SubAgentRequest{}, nil)

	assert.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestSubAgentExecutor_NonExistentAgent(t *testing.T) {
	registry := NewAgentRegistry()
	executor := NewSubAgentExecutor(registry)

	requests := []SubAgentRequest{
		{AgentID: "nonexistent", Input: "test"},
	}

	results, err := executor.ExecuteSubAgents(context.Background(), requests, nil)

	assert.Error(t, err)
	assert.Len(t, results, 1)
	assert.Error(t, results[0].Error)
	assert.Contains(t, results[0].Error.Error(), "not found")
}

func TestSubAgentExecutor_AggregateResults(t *testing.T) {
	executor := NewSubAgentExecutor(nil)

	results := []SubAgentResult{
		{AgentID: "agent1", Output: "result1", Error: nil},
		{AgentID: "agent2", Output: "result2", Error: nil},
	}

	aggregated, err := executor.AggregateResults(results)

	assert.NoError(t, err)
	assert.Contains(t, aggregated, "result1")
	assert.Contains(t, aggregated, "result2")
}

func TestSubAgentExecutor_Timeout(t *testing.T) {
	registry := NewAgentRegistry()
	executor := NewSubAgentExecutor(registry)
	executor.SetTimeout(1 * time.Millisecond) // Very short timeout

	requests := []SubAgentRequest{
		{AgentID: "slow-agent", Input: "test"},
	}

	// Should timeout quickly
	start := time.Now()
	_, _ = executor.ExecuteSubAgents(context.Background(), requests, nil)
	duration := time.Since(start)

	// Should complete fast (timeout + overhead)
	assert.Less(t, duration, 100*time.Millisecond)
}
