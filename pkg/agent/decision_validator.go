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
	"strings"
)

// ValidateDecision validates a decision against allowed actions
func ValidateDecision(decision Decision, allowedActions []string) error {
	// Check if decision type is valid
	validTypes := []string{"approve", "edit", "reject", "response"}
	if !contains(validTypes, decision.Type) {
		return fmt.Errorf("invalid decision type: %s (must be one of: %s)",
			decision.Type, strings.Join(validTypes, ", "))
	}

	// Check if decision type is allowed
	if !contains(allowedActions, decision.Type) {
		return fmt.Errorf("decision type %s not allowed (allowed: %s)",
			decision.Type, strings.Join(allowedActions, ", "))
	}

	// Validate decision-specific requirements
	switch decision.Type {
	case "edit":
		// Edit requires 'action' and 'args' in Args
		if decision.Args == nil {
			return fmt.Errorf("edit decision requires Args")
		}
		if _, ok := decision.Args["action"]; !ok {
			return fmt.Errorf("edit decision requires 'action' in Args")
		}
		if _, ok := decision.Args["args"]; !ok {
			return fmt.Errorf("edit decision requires 'args' in Args")
		}

	case "response":
		// Response requires 'response' text in Args
		if decision.Args == nil {
			return fmt.Errorf("response decision requires Args")
		}
		if _, ok := decision.Args["response"]; !ok {
			return fmt.Errorf("response decision requires 'response' in Args")
		}

	case "approve", "reject":
		// These don't require Args (can be nil)
	}

	return nil
}

// ValidateResumeCommand validates a complete resume command against an interrupt
func ValidateResumeCommand(resume *ResumeCommand, interrupt *InterruptData) error {
	if resume == nil {
		return fmt.Errorf("resume command cannot be nil")
	}

	if interrupt == nil {
		return fmt.Errorf("interrupt data cannot be nil")
	}

	// Check decision count matches tool call count
	if len(resume.Decisions) != len(interrupt.ToolCalls) {
		return fmt.Errorf(
			"decision count mismatch: got %d decisions for %d tool calls",
			len(resume.Decisions),
			len(interrupt.ToolCalls),
		)
	}

	// Validate each decision
	for i, decision := range resume.Decisions {
		toolName := interrupt.ToolCalls[i].Function.Name

		// Get allowed actions for this tool
		allowedActions, exists := interrupt.AllowedActions[toolName]
		if !exists {
			// Default to approve/reject if not specified
			allowedActions = []string{"approve", "reject"}
		}

		// Validate the decision
		if err := ValidateDecision(decision, allowedActions); err != nil {
			return fmt.Errorf("invalid decision for tool %s (index %d): %w", toolName, i, err)
		}
	}

	return nil
}

// contains is a helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
