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
	"context"
	"fmt"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

// DesignerDefinition implements AgentDefinition for the Designer agent
// This replaces 500+ lines of custom production code with GoLangGraph auto-server
type DesignerDefinition struct {
	*agent.BaseAgentDefinition
}

// NewDesignerDefinition creates a new Designer agent definition
func NewDesignerDefinition() *DesignerDefinition {
	config := &agent.AgentConfig{
		Name:     "Visual Designer",
		Type:     agent.AgentTypeChat,
		Model:    "gemma3:1b",
		Provider: "ollama",
		SystemPrompt: `You are a visionary creative designer specializing in sustainable architecture and habitat design for the year 2035.

Your expertise includes:
- Eco-friendly materials and construction techniques
- Energy-efficient design principles
- Integration with natural environments
- Smart home technology integration
- Sustainable living solutions

Always respond in a structured format with design descriptions, materials, and style categorization.
Provide detailed, imaginative descriptions that inspire and inform.`,
	}

	definition := &DesignerDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
	}

	// Set comprehensive schema metadata for auto-validation
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Design request or description",
				"minLength":   1,
				"maxLength":   1000,
				"pattern":     "^.+$",
			},
		},
		"required":    []string{"message"},
		"description": "Input schema for Designer agent",
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Design description",
				"minLength":   10,
			},
			"image_data": map[string]interface{}{
				"type":        "string",
				"description": "Base64 encoded image data",
				"pattern":     "^data:image/(png|jpeg|jpg|gif|webp);base64,[A-Za-z0-9+/]+={0,2}$",
			},
			"materials": map[string]interface{}{
				"type":        "array",
				"description": "List of materials used",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"style": map[string]interface{}{
				"type":        "string",
				"description": "Design style category",
				"enum":        []interface{}{"modern", "minimalist", "futuristic", "organic", "industrial"},
			},
		},
		"required":    []string{"description", "image_data", "style"},
		"description": "Output schema for Designer agent",
	})

	definition.SetMetadata("description", "Creates visual designs and concepts for habitat spaces")
	definition.SetMetadata("tags", []string{"design", "architecture", "sustainability", "creativity"})

	return definition
}

// CreateAgent creates a Designer agent with custom graph workflow
func (d *DesignerDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := d.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build custom graph for Designer workflow
	graph := core.NewGraph("designer-workflow")
	graph.AddNode("design", "Generate Design", d.designNode)
	graph.SetStartNode("design")
	graph.AddEndNode("design")

	// Note: In practice, this would replace the agent's internal graph
	// For now, the graph serves as documentation of the intended workflow
	_ = graph

	return baseAgent, nil
}

// designNode handles the design generation logic with custom workflow
func (d *DesignerDefinition) designNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Get input from state
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

	// Extract message from input
	var message string
	switch v := inputData.(type) {
	case string:
		message = v
	case map[string]interface{}:
		if msg, ok := v["message"].(string); ok {
			message = msg
		} else {
			return nil, fmt.Errorf("no message field in input")
		}
	default:
		return nil, fmt.Errorf("invalid input type")
	}

	// Create comprehensive design prompt
	prompt := fmt.Sprintf(`Create a detailed sustainable habitat design for: %s

Design for the year 2035 with focus on:
- Eco-friendly materials (bamboo, recycled composites, bio-concrete)
- Energy systems (solar, wind, geothermal integration)
- Water management (rainwater harvesting, greywater recycling)
- Food systems (vertical gardens, hydroponic systems)
- Smart technology integration
- Natural lighting and ventilation
- Waste reduction systems

Provide:
1. Detailed architectural description
2. Key sustainable materials
3. Design style classification
4. Visual composition details

Respond in a structured format with descriptions, materials list, and style categorization.`, message)

	// Generate response using LLM
	response, err := d.generateWithLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate design: %w", err)
	}

	// Create structured output with placeholder image
	output := map[string]interface{}{
		"description": response,
		"image_data":  d.createPlaceholderImage(),
		"materials":   []string{"bamboo", "recycled steel", "bio-concrete", "solar panels"},
		"style":       "sustainable-futuristic",
	}

	state.Set("output", output)
	return state, nil
}

// generateWithLLM generates text using the LLM (simplified version)
func (d *DesignerDefinition) generateWithLLM(ctx context.Context, prompt string) (string, error) {
	// This is a simplified implementation
	// In practice, this would use the agent's configured LLM provider
	_ = []llm.Message{{Role: "user", Content: prompt}} // Future LLM integration

	// For demo purposes, return a structured response
	return fmt.Sprintf("🏗️ Sustainable Habitat Design 2035\n\n%s\n\nThis design features cutting-edge sustainable architecture with integrated renewable energy systems, advanced water recycling, and smart automation for optimal living comfort while minimizing environmental impact.", prompt), nil
}

// createPlaceholderImage creates a base64 placeholder image
func (d *DesignerDefinition) createPlaceholderImage() string {
	// Simple 1x1 transparent PNG placeholder
	return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="
}

// GetDesignerConfig returns the configuration for backward compatibility
func GetDesignerConfig() *agent.AgentConfig {
	return NewDesignerDefinition().GetConfig()
}
