// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

// AgentExecution tracks the execution state of an agent
type AgentExecution struct {
	ID               string
	Input            string
	Output           interface{}
	Success          bool
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
	Status           string // "running", "completed", "failed", "interrupted"
	Steps            []AgentStep
	ToolCalls        []llm.ToolCall
	Error            error
	Metadata         map[string]interface{}
	StructuredOutput interface{}
	ExecutionPath    []string
}

// AgentStep represents a single step in the agent's execution
type AgentStep struct {
	NodeID    string
	Timestamp time.Time
	Input     map[string]interface{}
	Output    map[string]interface{}
	Error     error
}

// Command represents a control flow instruction for LangGraph-style operations
// This enables tools to return state updates, navigation commands, or resume instructions
type Command struct {
	// Update contains state changes to apply
	Update map[string]interface{}
	// Resume contains instructions to resume from an interrupt
	Resume *ResumeCommand
	// Goto specifies a specific node to navigate to
	Goto string
}

// ResumeCommand represents instructions to resume from a human-in-the-loop interrupt
type ResumeCommand struct {
	// Decisions contains one decision per interrupted action, in order
	Decisions []Decision
}

// Decision represents a human decision on an interrupted tool action
type Decision struct {
	// Type is one of: "approve", "edit", "reject", "response"
	Type string
	// Args contains additional arguments for edit/response types
	// For "edit": {"action": "tool_name", "args": {...}}
	// For "response": {"response": "text to return as tool result"}
	Args map[string]interface{}
}

// ToolRuntime provides context and state access for tool execution
// This matches LangGraph's InjectedState/InjectedStore pattern
type ToolRuntime struct {
	// State is the current agent state
	State *core.BaseState
	// Store provides persistent storage across sessions
	Store persistence.Store
	// ToolCallID is the unique identifier for this tool invocation
	ToolCallID string
	// ThreadID is the conversation thread identifier
	ThreadID string
}

// InterruptData holds information about an interrupted execution
// Used for Human-in-the-Loop (HITL) workflows
type InterruptData struct {
	// ToolCalls are the pending tool calls awaiting approval
	ToolCalls []llm.ToolCall
	// AllowedActions maps tool names to allowed decision types
	// e.g., {"delete_file": ["approve", "reject"], "send_email": ["approve", "edit", "reject"]}
	AllowedActions map[string][]string
}

// StatusUpdate represents a status notification from the agent
type StatusUpdate struct {
	// Type is the type of update: "progress", "tool_call", "completion", "error"
	Type string
	// Message is a human-readable status message
	Message string
	// Metadata contains additional context
	Metadata map[string]interface{}
}

// SubAgentRequest represents a request to execute a subagent
type SubAgentRequest struct {
	// AgentID is the identifier of the subagent to execute
	AgentID string
	// Input is the input to pass to the subagent
	Input string
	// Timeout is the maximum duration for subagent execution
	Timeout time.Duration
	// ShareState determines if global state should be shared
	ShareState bool
	// Messages seeds ReAct conversation history on resume (HITL mid-run).
	// When non-empty, the agent continues from this history instead of a cold start.
	Messages []llm.Message
	// Iteration restores the ReAct loop counter on resume.
	Iteration int
	// PendingToolCalls restores in-flight tool calls interrupted mid-act.
	PendingToolCalls []llm.ToolCall
	// Resume skips re-adding Input as a fresh user turn when Messages already end
	// with an equivalent user/assistant/tool exchange.
	Resume bool
	// TaskID is an optional caller tag (persisted in result metadata for checkpoints).
	TaskID string
}

// SubAgentResult contains the result of a subagent execution
type SubAgentResult struct {
	// AgentID is the identifier of the executed subagent
	AgentID string
	// Output is the final output from the subagent
	Output interface{}
	// State is the final state after execution
	State *core.BaseState
	// Duration is how long the execution took
	Duration time.Duration
	// Error contains any error that occurred
	Error error
	// Messages is the ReAct conversation at exit (including on cancel/interrupt).
	Messages []llm.Message
	// Iteration is the ReAct loop counter at exit.
	Iteration int
	// PendingToolCalls are tool calls awaiting observation (mid-tool-call interrupt).
	PendingToolCalls []llm.ToolCall
	// Provider / Model identify the LLM backend used (for checkpoint restore).
	Provider string
	Model    string
	// TaskID echoes the request tag when set.
	TaskID string
	// Usage aggregates token counts from LLM calls (estimated when providers omit).
	Usage llm.Usage
	// UsageEstimated is true when Usage was filled by heuristics (e.g. early_exit).
	UsageEstimated bool
}
