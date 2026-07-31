// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Integration test for ToolNode with command-based control flow
package agent

import (
	"context"
	"testing"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
	"github.com/stretchr/testify/assert"
)

// TestToolNode_BasicExecution tests basic tool execution through ToolNode
func TestToolNode_BasicExecution(t *testing.T) {
	registry := tools.NewToolRegistry()
	toolNode := NewToolNode(registry)

	ctx := context.Background()
	state := core.NewBaseState()

	// Create a tool call
	toolCalls := []llm.ToolCall{
		{
			ID: "call_1",
			Function: llm.FunctionCall{
				Name:      "calculator",
				Arguments: `{"expression": "2 + 2"}`,
			},
		},
	}

	runtime := &ToolRuntime{
		State:      state,
		ThreadID:   "test-thread",
		ToolCallID: "call_1",
	}

	// Execute tools
	messages, cmd, err := toolNode.ExecuteTools(ctx, toolCalls, runtime)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "tool", messages[0].Role)
	assert.Contains(t, messages[0].Content, "4") // Should contain the result
	assert.Nil(t, cmd)                           // Calculator returns string, not command
}

// TestToolNode_CommandReturn tests tool returning a Command
func TestToolNode_CommandReturn(t *testing.T) {
	registry := tools.NewToolRegistry()

	// Register a tool that returns a Command
	commandTool := tools.NewGenericTool(
		"command_tool",
		"A tool that returns a Command",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			// Return a Command instead of a string
			return &Command{
				Update: map[string]interface{}{
					"status": "completed",
					"result": "success",
				},
				Goto: "next_node",
			}, nil
		},
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	)
	registry.RegisterTool(commandTool)

	toolNode := NewToolNode(registry)
	ctx := context.Background()
	state := core.NewBaseState()

	toolCalls := []llm.ToolCall{
		{
			ID: "call_2",
			Function: llm.FunctionCall{
				Name:      "command_tool",
				Arguments: `{}`,
			},
		},
	}

	runtime := &ToolRuntime{
		State:      state,
		ThreadID:   "test-thread",
		ToolCallID: "call_2",
	}

	// Execute tool
	messages, cmd, err := toolNode.ExecuteTools(ctx, toolCalls, runtime)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.NotNil(t, cmd, "Expected a Command to be returned")

	// Verify Command contents
	assert.NotNil(t, cmd.Update)
	assert.Equal(t, "completed", cmd.Update["status"])
	assert.Equal(t, "success", cmd.Update["result"])
	assert.Equal(t, "next_node", cmd.Goto)
}

// TestToolNode_ParallelExecution tests parallel tool execution
func TestToolNode_ParallelExecution(t *testing.T) {
	registry := tools.NewToolRegistry()
	toolNode := NewToolNode(registry)
	toolNode.SetParallel(true) // Enable parallel execution

	ctx := context.Background()
	state := core.NewBaseState()

	// Create multiple tool calls
	toolCalls := []llm.ToolCall{
		{
			ID: "call_1",
			Function: llm.FunctionCall{
				Name:      "calculator",
				Arguments: `{"expression": "2 + 2"}`,
			},
		},
		{
			ID: "call_2",
			Function: llm.FunctionCall{
				Name:      "time",
				Arguments: `{"format": "RFC3339"}`,
			},
		},
	}

	runtime := &ToolRuntime{
		State:    state,
		ThreadID: "test-thread",
	}

	// Execute tools in parallel
	messages, cmd, err := toolNode.ExecuteTools(ctx, toolCalls, runtime)

	assert.NoError(t, err)
	assert.Len(t, messages, 2, "Should execute both tools")
	assert.Nil(t, cmd)

	// Verify messages are in correct order (even though executed in parallel)
	assert.Equal(t, "call_1", messages[0].ToolCallID)
	assert.Equal(t, "call_2", messages[1].ToolCallID)
}

// TestToolNode_ErrorHandling tests custom error handling
func TestToolNode_ErrorHandling(t *testing.T) {
	registry := tools.NewToolRegistry()
	toolNode := NewToolNode(registry)

	// Set custom error handler
	customErrorHandler := func(err error) string {
		return "CUSTOM_ERROR: " + err.Error()
	}
	toolNode.SetErrorHandler(customErrorHandler)

	ctx := context.Background()
	state := core.NewBaseState()

	// Create a tool call that will fail (invalid JSON)
	toolCalls := []llm.ToolCall{
		{
			ID: "call_fail",
			Function: llm.FunctionCall{
				Name:      "calculator",
				Arguments: `invalid json`,
			},
		},
	}

	runtime := &ToolRuntime{
		State:    state,
		ThreadID: "test-thread",
	}

	// Execute tool (should handle error gracefully)
	messages, _, err := toolNode.ExecuteTools(ctx, toolCalls, runtime)

	assert.NoError(t, err, "Error handler should prevent error propagation")
	assert.Len(t, messages, 1)
	assert.Contains(t, messages[0].Content, "CUSTOM_ERROR")
}

// TestAgentWithToolNode tests full agent integration with ToolNode
func TestAgentWithToolNode(t *testing.T) {
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	config := DefaultAgentConfig()
	config.Name = "TestAgent"
	config.Type = AgentTypeReAct
	config.Model = "test-model"
	config.Provider = "test"
	config.Tools = []string{"calculator"}

	agent := NewAgent(config, llmManager, toolRegistry)

	// Verify ToolNode is initialized
	assert.NotNil(t, agent.toolNode, "ToolNode should be initialized")

	// Verify it's connected to the same registry
	tool, exists := agent.toolRegistry.GetTool("calculator")
	assert.True(t, exists)
	assert.Equal(t, "calculator", tool.GetName())
}
