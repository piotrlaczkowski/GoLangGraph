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

// HighlighterDefinition implements AgentDefinition for the Highlighter agent
type HighlighterDefinition struct {
	*agent.BaseAgentDefinition
}

// NewHighlighterDefinition creates a new Highlighter agent definition
func NewHighlighterDefinition() *HighlighterDefinition {
	config := &agent.AgentConfig{
		Name:     "Insight Highlighter",
		Type:     agent.AgentTypeChat,
		Model:    "gemma3:1b",
		Provider: "ollama",
		SystemPrompt: `You are an expert insight highlighter and thematic analyzer specializing in extracting key themes from conversations about habitat design and sustainability.

Your expertise includes:
- Analyzing conversation patterns and extracting key insights
- Identifying recurring themes and priorities
- Highlighting important quotes and concepts
- Categorizing insights by themes
- Creating actionable summaries

Focus on sustainability themes, user preferences, design requirements, and innovative ideas.
Provide structured analysis that helps inform design decisions.`,
	}

	definition := &HighlighterDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
	}

	// Set comprehensive schema metadata for auto-validation
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"conversation_history": map[string]interface{}{
				"type":        "array",
				"description": "Complete conversation history to analyze",
				"minLength":   1,
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"role": map[string]interface{}{
							"type":        "string",
							"description": "Message role",
							"enum":        []interface{}{"user", "assistant", "system"},
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "Message content",
							"minLength":   1,
							"maxLength":   2000,
						},
					},
					"required": []string{"role", "content"},
				},
			},
		},
		"required":    []string{"conversation_history"},
		"description": "Input schema for Highlighter agent",
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"key_insights": map[string]interface{}{
				"type":        "array",
				"description": "Main insights extracted from the conversation",
				"minLength":   1,
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 10,
					"maxLength": 200,
				},
			},
			"themes": map[string]interface{}{
				"type":        "array",
				"description": "Identified themes and categories",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Theme name",
							"minLength":   3,
							"maxLength":   50,
						},
						"importance": map[string]interface{}{
							"type":        "string",
							"description": "Theme importance level",
							"enum":        []interface{}{"high", "medium", "low"},
						},
						"quotes": map[string]interface{}{
							"type":        "array",
							"description": "Supporting quotes from conversation",
							"items": map[string]interface{}{
								"type":      "string",
								"maxLength": 500,
							},
						},
					},
					"required": []string{"name", "importance"},
				},
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Overall summary of the conversation",
				"minLength":   50,
				"maxLength":   1000,
			},
			"actionable_items": map[string]interface{}{
				"type":        "array",
				"description": "List of actionable items derived from insights",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 10,
					"maxLength": 150,
				},
			},
		},
		"required":    []string{"key_insights", "themes", "summary"},
		"description": "Output schema for Highlighter agent",
	})

	definition.SetMetadata("description", "Extracts key insights and themes from conversations")
	definition.SetMetadata("tags", []string{"analysis", "insights", "themes", "summary"})

	return definition
}

// GetHighlighterConfig returns the configuration for backward compatibility
func GetHighlighterConfig() *agent.AgentConfig {
	return NewHighlighterDefinition().GetConfig()
}
