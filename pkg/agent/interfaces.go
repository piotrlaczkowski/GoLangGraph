// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"context"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

// Agent is the core interface for all AI agents in the system.
// It defines the standard behavior that any agent (Base, Deep, or Custom) must implement.
type Agent interface {
	// Core execution
	Execute(ctx context.Context, input string) (*AgentExecution, error)

	// Configuration
	GetConfig() *AgentConfig
	UpdateConfig(config *AgentConfig)

	// State management
	GetGraph() *core.Graph
	// IsRunning returns true if the agent is currently executing
	IsRunning() bool

	// Conversation management
	GetConversation() []llm.Message
	ClearConversation()
	// SeedConversation restores prior ReAct messages (HITL mid-run resume).
	SeedConversation(messages []llm.Message)

	// History management
	GetExecutionHistory() []AgentExecution
	ClearHistory()

	// Metadata
	Name() string
}
