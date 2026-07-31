// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

// SubAgentExecutor handles parallel execution of multiple subagents
type SubAgentExecutor struct {
	registry *AgentRegistry
	parallel bool
	timeout  time.Duration
}

// NewSubAgentExecutor creates a new subagent executor
func NewSubAgentExecutor(registry *AgentRegistry) *SubAgentExecutor {
	return &SubAgentExecutor{
		registry: registry,
		parallel: true,
		timeout:  30 * time.Second,
	}
}

// SetParallel configures whether to execute subagents in parallel
func (se *SubAgentExecutor) SetParallel(parallel bool) {
	se.parallel = parallel
}

// SetTimeout sets the default timeout for subagent execution
func (se *SubAgentExecutor) SetTimeout(timeout time.Duration) {
	se.timeout = timeout
}

// ExecuteSubAgents executes multiple subagents with optional parallel execution
func (se *SubAgentExecutor) ExecuteSubAgents(
	ctx context.Context,
	requests []SubAgentRequest,
	sharedState *SharedState,
) ([]SubAgentResult, error) {
	if len(requests) == 0 {
		return []SubAgentResult{}, nil
	}

	if se.parallel && len(requests) > 1 {
		return se.executeParallel(ctx, requests, sharedState)
	}
	return se.executeSequential(ctx, requests, sharedState)
}

// executeParallel executes subagents concurrently
func (se *SubAgentExecutor) executeParallel(
	ctx context.Context,
	requests []SubAgentRequest,
	sharedState *SharedState,
) ([]SubAgentResult, error) {
	results := make([]SubAgentResult, len(requests))
	var wg sync.WaitGroup
	errChan := make(chan error, len(requests))

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request SubAgentRequest) {
			defer wg.Done()

			result := se.executeSingle(ctx, request, sharedState)
			results[index] = result

			if result.Error != nil {
				errChan <- result.Error
			}
		}(i, req)
	}

	wg.Wait()
	close(errChan)

	// Collect first error if any
	var firstError error
	for err := range errChan {
		if firstError == nil {
			firstError = err
		}
	}

	return results, firstError
}

// executeSequential executes subagents one by one
func (se *SubAgentExecutor) executeSequential(
	ctx context.Context,
	requests []SubAgentRequest,
	sharedState *SharedState,
) ([]SubAgentResult, error) {
	results := make([]SubAgentResult, len(requests))

	for i, req := range requests {
		result := se.executeSingle(ctx, req, sharedState)
		results[i] = result

		if result.Error != nil {
			return results, result.Error
		}
	}

	return results, nil
}

// executeSingle executes a single subagent request
func (se *SubAgentExecutor) executeSingle(
	ctx context.Context,
	request SubAgentRequest,
	sharedState *SharedState,
) SubAgentResult {
	start := time.Now()

	// Get agent from registry
	agentDef, exists := se.registry.GetDefinition(request.AgentID)
	if !exists {
		return SubAgentResult{
			AgentID:  request.AgentID,
			Error:    fmt.Errorf("subagent '%s' not found", request.AgentID),
			Duration: time.Since(start),
		}
	}

	// Create agent instance
	subAgent, err := agentDef.CreateAgent()
	if err != nil {
		return SubAgentResult{
			AgentID:  request.AgentID,
			Error:    fmt.Errorf("failed to create subagent: %w", err),
			Duration: time.Since(start),
			TaskID:   request.TaskID,
		}
	}

	// Seed ReAct history for mid-run HITL resume.
	if len(request.Messages) > 0 || request.Resume || len(request.PendingToolCalls) > 0 {
		if seeder, ok := subAgent.(interface {
			SeedResumeState([]llm.Message, int, []llm.ToolCall)
		}); ok {
			seeder.SeedResumeState(request.Messages, request.Iteration, request.PendingToolCalls)
		} else {
			subAgent.SeedConversation(request.Messages)
		}
	}

	// Setup timeout context
	timeout := request.Timeout
	if timeout == 0 {
		timeout = se.timeout
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Setup state
	var localState *SharedState
	if request.ShareState && sharedState != nil {
		localState = sharedState
	} else {
		localState = NewSharedState()
	}

	// Get local state for this subagent
	agentState := localState.GetLocalState(request.AgentID)

	// On resume with history, prefer empty/continue input so Execute doesn't
	// duplicate the original user turn (already in Messages).
	input := request.Input
	if request.Resume && len(request.Messages) > 0 && strings.TrimSpace(input) == "" {
		input = "Continue from the interrupted step. Complete pending tool work if any, then finish the task."
	}

	// Execute subagent
	execution, err := subAgent.Execute(execCtx, input)

	result := SubAgentResult{
		AgentID:  request.AgentID,
		Duration: time.Since(start),
		State:    agentState,
		Error:    err,
		TaskID:   request.TaskID,
	}

	if execution != nil {
		result.Output = execution.Output
		if execution.Metadata != nil {
			if u, ok := execution.Metadata["usage"].(llm.Usage); ok {
				result.Usage = u
			}
			if est, ok := execution.Metadata["usage_estimated"].(bool); ok {
				result.UsageEstimated = est
			}
		}
	}

	// Always capture conversation for checkpointing (including cancel/interrupt).
	result.Messages = subAgent.GetConversation()
	if cfg := subAgent.GetConfig(); cfg != nil {
		result.Provider = cfg.Provider
		result.Model = cfg.Model
	}
	if iter, pending := extractResumeHints(agentState, execution, result.Messages); true {
		result.Iteration = iter
		result.PendingToolCalls = pending
	}
	if result.Iteration == 0 && request.Iteration > 0 && err != nil {
		result.Iteration = request.Iteration
	}
	if len(result.PendingToolCalls) == 0 && len(request.PendingToolCalls) > 0 && err != nil {
		result.PendingToolCalls = request.PendingToolCalls
	}

	return result
}

func extractResumeHints(state *core.BaseState, execution *AgentExecution, msgs []llm.Message) (int, []llm.ToolCall) {
	iter := 0
	var pending []llm.ToolCall
	if state != nil {
		if v, ok := state.Get("iteration"); ok {
			switch n := v.(type) {
			case int:
				iter = n
			case int64:
				iter = int(n)
			case float64:
				iter = int(n)
			}
		}
		if v, ok := state.Get("pending_tool_calls"); ok {
			switch tc := v.(type) {
			case []llm.ToolCall:
				pending = tc
			}
		}
	}
	if execution != nil && len(execution.ToolCalls) > 0 && len(pending) == 0 {
		// Last assistant tool calls may still be in-flight when canceled mid-act.
		pending = execution.ToolCalls
	}
	if len(pending) == 0 && len(msgs) > 0 {
		last := msgs[len(msgs)-1]
		if last.Role == "assistant" && len(last.ToolCalls) > 0 {
			pending = last.ToolCalls
		}
	}
	return iter, pending
}

// AggregateResults combines results from multiple subagents
func (se *SubAgentExecutor) AggregateResults(results []SubAgentResult) (string, error) {
	var outputs []string
	var errors []error

	for _, result := range results {
		if result.Error != nil {
			errors = append(errors, result.Error)
		} else if result.Output != nil {
			outputs = append(outputs, fmt.Sprintf("%v", result.Output))
		}
	}

	if len(errors) > 0 {
		return "", fmt.Errorf("subagent errors: %v", errors)
	}

	aggregated := ""
	for i, output := range outputs {
		aggregated += fmt.Sprintf("[SubAgent %d]: %s\n", i+1, output)
	}

	return aggregated, nil
}
