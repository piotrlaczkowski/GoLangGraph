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
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

// InterruptManager handles interrupt lifecycle for HITL workflows
type InterruptManager struct {
	mu          sync.RWMutex
	interrupts  map[string]*InterruptData
	onInterrupt func(*InterruptData) // Optional callback for notifications
}

// NewInterruptManager creates a new interrupt manager
func NewInterruptManager() *InterruptManager {
	return &InterruptManager{
		interrupts: make(map[string]*InterruptData),
	}
}

// SetInterruptCallback sets a callback function called when interrupts occur
func (im *InterruptManager) SetInterruptCallback(callback func(*InterruptData)) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.onInterrupt = callback
}

// Interrupt creates a new interrupt for the given tool calls
// Returns the interrupt ID and the InterruptData
func (im *InterruptManager) Interrupt(
	toolCalls []llm.ToolCall,
	allowedActions map[string][]string,
) (string, *InterruptData, error) {
	if len(toolCalls) == 0 {
		return "", nil, fmt.Errorf("cannot interrupt with zero tool calls")
	}

	// Generate unique interrupt ID
	interruptID := uuid.New().String()

	// Create interrupt data
	interrupt := &InterruptData{
		ToolCalls:      toolCalls,
		AllowedActions: allowedActions,
	}

	// Store interrupt
	im.mu.Lock()
	im.interrupts[interruptID] = interrupt
	im.mu.Unlock()

	// Notify callback if set
	if im.onInterrupt != nil {
		go im.onInterrupt(interrupt)
	}

	return interruptID, interrupt, nil
}

// GetPending retrieves a pending interrupt by ID
func (im *InterruptManager) GetPending(interruptID string) (*InterruptData, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	interrupt, exists := im.interrupts[interruptID]
	if !exists {
		return nil, fmt.Errorf("interrupt not found: %s", interruptID)
	}

	return interrupt, nil
}

// Resume processes a resume command for an interrupted execution
// Validates decisions and removes the interrupt from pending list
func (im *InterruptManager) Resume(
	interruptID string,
	decisions []Decision,
) (*ResumeCommand, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Get interrupt
	interrupt, exists := im.interrupts[interruptID]
	if !exists {
		return nil, fmt.Errorf("interrupt not found: %s", interruptID)
	}

	// Validate number of decisions
	if len(decisions) != len(interrupt.ToolCalls) {
		return nil, fmt.Errorf(
			"decision count mismatch: got %d decisions for %d tool calls",
			len(decisions),
			len(interrupt.ToolCalls),
		)
	}

	// Validate each decision
	for i, decision := range decisions {
		toolName := interrupt.ToolCalls[i].Function.Name
		allowedActions, exists := interrupt.AllowedActions[toolName]

		if !exists {
			// If no specific allowed actions, default to approve/reject
			allowedActions = []string{"approve", "reject"}
		}

		if err := ValidateDecision(decision, allowedActions); err != nil {
			return nil, fmt.Errorf("invalid decision for tool %s: %w", toolName, err)
		}
	}

	// Create resume command
	resumeCmd := &ResumeCommand{
		Decisions: decisions,
	}

	// Remove interrupt from pending (it's now being processed)
	delete(im.interrupts, interruptID)

	return resumeCmd, nil
}

// ListPending returns all pending interrupt IDs
func (im *InterruptManager) ListPending() []string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	ids := make([]string, 0, len(im.interrupts))
	for id := range im.interrupts {
		ids = append(ids, id)
	}
	return ids
}

// Clear removes all pending interrupts
func (im *InterruptManager) Clear() {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.interrupts = make(map[string]*InterruptData)
}
