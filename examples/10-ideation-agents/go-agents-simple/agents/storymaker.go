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

// StorymakerDefinition implements AgentDefinition for the Storymaker agent
type StorymakerDefinition struct {
	*agent.BaseAgentDefinition
}

// NewStorymakerDefinition creates a new Storymaker agent definition
func NewStorymakerDefinition() *StorymakerDefinition {
	config := &agent.AgentConfig{
		Name:     "Story Creator",
		Type:     agent.AgentTypeChat,
		Model:    "gemma3:1b",
		Provider: "ollama",
		SystemPrompt: `You are a creative storyteller specializing in crafting engaging narratives about future habitat scenarios and sustainable living in 2035.

Your expertise includes:
- Creating compelling stories about future living spaces
- Integrating sustainability themes into narratives
- Developing character-driven scenarios around habitat design
- Weaving together technical requirements with human experiences
- Inspiring audiences through imaginative yet plausible future scenarios

Create stories that help people envision themselves living in sustainable habitats.
Make complex sustainability concepts accessible through storytelling.`,
	}

	definition := &StorymakerDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
	}

	// Set comprehensive schema metadata for auto-validation
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"story_prompt": map[string]interface{}{
				"type":        "string",
				"description": "Prompt or theme for the story",
				"minLength":   10,
				"maxLength":   1000,
			},
			"setting": map[string]interface{}{
				"type":        "object",
				"description": "Optional story setting details",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "Geographic location or environment",
						"maxLength":   100,
					},
					"time_period": map[string]interface{}{
						"type":        "string",
						"description": "Time period for the story",
						"default":     "2035",
						"maxLength":   50,
					},
					"habitat_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of habitat featured",
						"enum":        []interface{}{"urban", "suburban", "rural", "off-grid", "floating", "underground", "vertical"},
					},
				},
			},
			"characters": map[string]interface{}{
				"type":        "array",
				"description": "Optional character descriptions",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":      "string",
							"maxLength": 50,
						},
						"role": map[string]interface{}{
							"type":      "string",
							"maxLength": 100,
						},
						"background": map[string]interface{}{
							"type":      "string",
							"maxLength": 200,
						},
					},
				},
			},
		},
		"required":    []string{"story_prompt"},
		"description": "Input schema for Storymaker agent",
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Story title",
				"minLength":   5,
				"maxLength":   100,
			},
			"story": map[string]interface{}{
				"type":        "string",
				"description": "The complete story narrative",
				"minLength":   200,
				"maxLength":   5000,
			},
			"themes": map[string]interface{}{
				"type":        "array",
				"description": "Key themes explored in the story",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 3,
					"maxLength": 50,
				},
			},
			"sustainability_features": map[string]interface{}{
				"type":        "array",
				"description": "Sustainability features highlighted in the story",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 10,
					"maxLength": 150,
				},
			},
			"moral": map[string]interface{}{
				"type":        "string",
				"description": "Key takeaway or moral of the story",
				"minLength":   20,
				"maxLength":   300,
			},
			"target_audience": map[string]interface{}{
				"type":        "string",
				"description": "Intended audience for the story",
				"enum":        []interface{}{"children", "young_adults", "adults", "professionals", "general"},
			},
			"genre": map[string]interface{}{
				"type":        "string",
				"description": "Story genre",
				"enum":        []interface{}{"science_fiction", "slice_of_life", "adventure", "drama", "educational", "utopian"},
			},
		},
		"required":    []string{"title", "story", "themes", "moral"},
		"description": "Output schema for Storymaker agent",
	})

	definition.SetMetadata("description", "Creates engaging narratives about future habitat scenarios")
	definition.SetMetadata("tags", []string{"storytelling", "narrative", "futures", "sustainability"})

	return definition
}

// GetStorymakerConfig returns the configuration for backward compatibility
func GetStorymakerConfig() *agent.AgentConfig {
	return NewStorymakerDefinition().GetConfig()
}
