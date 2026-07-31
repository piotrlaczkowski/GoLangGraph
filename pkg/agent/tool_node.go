// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package agent provides the core agent implementation

package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// ErrorHandlerFunc defines how tool errors are handled
type ErrorHandlerFunc func(error) string

// ToolNode handles tool execution with all LangGraph features
// Supports parallel execution, error handling, and Command control flow
type ToolNode struct {
	registry       *tools.ToolRegistry
	handleErrors   ErrorHandlerFunc
	parallel       bool
	stateInjection bool
}

// NewToolNode creates a new tool node
func NewToolNode(registry *tools.ToolRegistry) *ToolNode {
	return &ToolNode{
		registry: registry,
		handleErrors: func(err error) string {
			return fmt.Sprintf("Error: %v", err)
		},
		parallel:       true, // Default to parallel execution (v1 mode)
		stateInjection: true,
	}
}

// SetErrorHandler sets a custom error handler
func (tn *ToolNode) SetErrorHandler(handler ErrorHandlerFunc) {
	tn.handleErrors = handler
}

// SetParallel sets whether tools should be executed in parallel
func (tn *ToolNode) SetParallel(parallel bool) {
	tn.parallel = parallel
}

// ExecuteTools handles tool execution with parallel support and error handling
func (tn *ToolNode) ExecuteTools(
	ctx context.Context,
	toolCalls []llm.ToolCall,
	runtime *ToolRuntime,
) ([]llm.Message, *Command, error) {
	if len(toolCalls) == 0 {
		return nil, nil, nil
	}

	if tn.parallel && len(toolCalls) > 1 {
		return tn.executeParallel(ctx, toolCalls, runtime)
	}
	return tn.executeSequential(ctx, toolCalls, runtime)
}

// executeSequential executes tools one by one
func (tn *ToolNode) executeSequential(
	ctx context.Context,
	toolCalls []llm.ToolCall,
	runtime *ToolRuntime,
) ([]llm.Message, *Command, error) {
	var messages []llm.Message
	var lastCommand *Command

	for _, call := range toolCalls {
		msg, cmd, err := tn.executeSingleTool(ctx, call, runtime)
		if err != nil {
			// If error handler is set, use it. Otherwise return error.
			if tn.handleErrors != nil {
				msg = llm.Message{
					Role:       "tool",
					Content:    tn.handleErrors(err),
					ToolCallID: call.ID,
				}
			} else {
				return nil, nil, err
			}
		}

		messages = append(messages, msg)
		if cmd != nil {
			lastCommand = cmd
		}
	}

	return messages, lastCommand, nil
}

// executeParallel executes tools concurrently
func (tn *ToolNode) executeParallel(
	ctx context.Context,
	toolCalls []llm.ToolCall,
	runtime *ToolRuntime,
) ([]llm.Message, *Command, error) {
	type result struct {
		idx int
		msg llm.Message
		cmd *Command
		err error
	}

	results := make(chan result, len(toolCalls))
	var wg sync.WaitGroup

	for i, call := range toolCalls {
		wg.Add(1)
		go func(idx int, c llm.ToolCall) {
			defer wg.Done()

			// Create a runtime copy for this goroutine if needed
			// For now, we share the runtime as it's mostly read-only or thread-safe
			// But we update the ToolCallID
			localRuntime := *runtime
			localRuntime.ToolCallID = c.ID

			msg, cmd, err := tn.executeSingleTool(ctx, c, &localRuntime)
			results <- result{idx: idx, msg: msg, cmd: cmd, err: err}
		}(i, call)
	}

	wg.Wait()
	close(results)

	// Collect results in order
	messages := make([]llm.Message, len(toolCalls))
	var lastCommand *Command
	var firstErr error

	for r := range results {
		if r.err != nil {
			if tn.handleErrors != nil {
				messages[r.idx] = llm.Message{
					Role:       "tool",
					Content:    tn.handleErrors(r.err),
					ToolCallID: toolCalls[r.idx].ID,
				}
			} else if firstErr == nil {
				firstErr = r.err
			}
		} else {
			messages[r.idx] = r.msg
			if r.cmd != nil {
				lastCommand = r.cmd
			}
		}
	}

	if firstErr != nil {
		return nil, nil, firstErr
	}

	return messages, lastCommand, nil
}

// executeSingleTool executes a single tool call
func (tn *ToolNode) executeSingleTool(
	ctx context.Context,
	call llm.ToolCall,
	runtime *ToolRuntime,
) (llm.Message, *Command, error) {
	tool, exists := tn.registry.GetTool(call.Function.Name)
	if !exists {
		return llm.Message{}, nil, fmt.Errorf("tool not found: %s", call.Function.Name)
	}

	// Inject state if supported
	// In a real implementation, we would check if the tool accepts injected state
	// For now, we rely on the tool implementation to use the context or args

	// Execute tool
	result, err := tool.Execute(ctx, call.Function.Arguments)
	if err != nil {
		return llm.Message{}, nil, err
	}

	return tn.handleToolResult(result, call)
}

// handleToolResult processes the raw tool result into a message and optional command
func (tn *ToolNode) handleToolResult(result interface{}, toolCall llm.ToolCall) (llm.Message, *Command, error) {
	switch v := result.(type) {
	case *Command:
		// Tool returned a Command for control flow
		// We still return a tool message to acknowledge the call
		return llm.Message{
			Role:       "tool",
			Content:    "Command executed",
			ToolCallID: toolCall.ID,
		}, v, nil
	case string:
		return llm.Message{
			Role:       "tool",
			Content:    v,
			ToolCallID: toolCall.ID,
		}, nil, nil
	default:
		return llm.Message{
			Role:       "tool",
			Content:    fmt.Sprintf("%v", v),
			ToolCallID: toolCall.ID,
		}, nil, nil
	}
}
