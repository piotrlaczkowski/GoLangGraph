// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
package agent

import (
	"testing"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/stretchr/testify/assert"
)

func TestInterruptManager_Basic(t *testing.T) {
	im := NewInterruptManager()

	// Create test tool calls
	toolCalls := []llm.ToolCall{
		{
			ID: "call_1",
			Function: llm.FunctionCall{
				Name:      "delete_file",
				Arguments: `{"path": "/tmp/test.txt"}`,
			},
		},
	}

	allowedActions := map[string][]string{
		"delete_file": {"approve", "reject"},
	}

	// Create interrupt
	id, interrupt, err := im.Interrupt(toolCalls, allowedActions)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotNil(t, interrupt)
	assert.Equal(t, 1, len(interrupt.ToolCalls))
}

func TestInterruptManager_Resume(t *testing.T) {
	im := NewInterruptManager()

	toolCalls := []llm.ToolCall{
		{ID: "call_1", Function: llm.FunctionCall{Name: "tool1", Arguments: "{}"}},
	}
	allowedActions := map[string][]string{
		"tool1": {"approve", "reject"},
	}

	// Create interrupt
	id, _, err := im.Interrupt(toolCalls, allowedActions)
	assert.NoError(t, err)

	// Resume with approval
	decisions := []Decision{
		{Type: "approve"},
	}

	resumeCmd, err := im.Resume(id, decisions)
	assert.NoError(t, err)
	assert.NotNil(t, resumeCmd)
	assert.Equal(t, 1, len(resumeCmd.Decisions))

	// Interrupt should be removed after resume
	_, err = im.GetPending(id)
	assert.Error(t, err)
}

func TestDecisionValidator_Approve(t *testing.T) {
	decision := Decision{Type: "approve"}
	allowed := []string{"approve", "reject"}

	err := ValidateDecision(decision, allowed)
	assert.NoError(t, err)
}

func TestDecisionValidator_Edit(t *testing.T) {
	decision := Decision{
		Type: "edit",
		Args: map[string]interface{}{
			"action": "tool_name",
			"args":   map[string]interface{}{"param": "value"},
		},
	}
	allowed := []string{"approve", "edit", "reject"}

	err := ValidateDecision(decision, allowed)
	assert.NoError(t, err)
}

func TestDecisionValidator_EditMissingArgs(t *testing.T) {
	decision := Decision{Type: "edit"}
	allowed := []string{"approve", "edit", "reject"}

	err := ValidateDecision(decision, allowed)
	assert.Error(t, err)
}

func TestDecisionValidator_Response(t *testing.T) {
	decision := Decision{
		Type: "response",
		Args: map[string]interface{}{
			"response": "Custom response text",
		},
	}
	allowed := []string{"response", "reject"}

	err := ValidateDecision(decision, allowed)
	assert.NoError(t, err)
}

func TestDecisionValidator_NotAllowed(t *testing.T) {
	decision := Decision{Type: "edit"}
	allowed := []string{"approve", "reject"} // edit not allowed

	err := ValidateDecision(decision, allowed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateResumeCommand(t *testing.T) {
	interrupt := &InterruptData{
		ToolCalls: []llm.ToolCall{
			{ID: "call_1", Function: llm.FunctionCall{Name: "tool1"}},
		},
		AllowedActions: map[string][]string{
			"tool1": {"approve", "reject"},
		},
	}

	resume := &ResumeCommand{
		Decisions: []Decision{
			{Type: "approve"},
		},
	}

	err := ValidateResumeCommand(resume, interrupt)
	assert.NoError(t, err)
}

func TestValidateResumeCommand_CountMismatch(t *testing.T) {
	interrupt := &InterruptData{
		ToolCalls: []llm.ToolCall{
			{ID: "call_1", Function: llm.FunctionCall{Name: "tool1"}},
		},
		AllowedActions: map[string][]string{
			"tool1": {"approve", "reject"},
		},
	}

	// Two decisions for one tool call
	resume := &ResumeCommand{
		Decisions: []Decision{
			{Type: "approve"},
			{Type: "reject"},
		},
	}

	err := ValidateResumeCommand(resume, interrupt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatch")
}
