// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Simplified Ideation Agents

package agents

import (
	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
)

// RegisterAllAgents registers all ideation agents with the provided registry
// This replaces hundreds of lines of manual endpoint creation and validation code
func RegisterAllAgents(registry *agent.AgentRegistry) error {
	// Register Designer agent
	if err := registry.RegisterDefinition("designer", NewDesignerDefinition()); err != nil {
		return err
	}

	// Register Interviewer agent
	if err := registry.RegisterDefinition("interviewer", NewInterviewerDefinition()); err != nil {
		return err
	}

	// Register Highlighter agent
	if err := registry.RegisterDefinition("highlighter", NewHighlighterDefinition()); err != nil {
		return err
	}

	// Register Storymaker agent
	if err := registry.RegisterDefinition("storymaker", NewStorymakerDefinition()); err != nil {
		return err
	}

	return nil
}

// GetAllAgentConfigs returns all agent configurations for backward compatibility
func GetAllAgentConfigs() map[string]*agent.AgentConfig {
	return map[string]*agent.AgentConfig{
		"designer":    GetDesignerConfig(),
		"interviewer": GetInterviewerConfig(),
		"highlighter": GetHighlighterConfig(),
		"storymaker":  GetStorymakerConfig(),
	}
}
