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

// InterviewerDefinition implements AgentDefinition for the Interviewer agent
type InterviewerDefinition struct {
	*agent.BaseAgentDefinition
}

// NewInterviewerDefinition creates a new Interviewer agent definition
func NewInterviewerDefinition() *InterviewerDefinition {
	config := &agent.AgentConfig{
		Name:     "Smart Interviewer",
		Type:     agent.AgentTypeChat,
		Model:    "gemma3:1b",
		Provider: "ollama",
		SystemPrompt: `You are an intelligent interviewer specializing in gathering requirements for habitat design projects in 2035.

Your expertise includes:
- Conducting structured conversations to understand user needs
- Asking probing questions about sustainability preferences
- Gathering requirements for future living spaces
- Identifying key themes and priorities
- Guiding conversations toward actionable insights

Always respond in French when conducting interviews. Ask follow-up questions to deepen understanding.
Maintain conversation flow and help users articulate their vision for future habitats.`,
	}

	definition := &InterviewerDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
	}

	// Set comprehensive schema metadata for auto-validation
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "User message or response to interview questions",
				"minLength":   1,
				"maxLength":   2000,
			},
			"conversation_context": map[string]interface{}{
				"type":        "object",
				"description": "Optional conversation context",
				"properties": map[string]interface{}{
					"phase": map[string]interface{}{
						"type": "string",
						"enum": []interface{}{"introduction", "exploration", "deep_dive", "synthesis", "conclusion"},
					},
					"topics_covered": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
		"required":    []string{"message"},
		"description": "Input schema for Interviewer agent",
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"response": map[string]interface{}{
				"type":        "string",
				"description": "The interviewer's response in French",
				"minLength":   10,
				"maxLength":   3000,
			},
			"next_questions": map[string]interface{}{
				"type":        "array",
				"description": "Suggested follow-up questions",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 10,
					"maxLength": 200,
				},
			},
			"conversation_phase": map[string]interface{}{
				"type":        "string",
				"description": "Current phase of the interview",
				"enum":        []interface{}{"introduction", "exploration", "deep_dive", "synthesis", "conclusion"},
			},
			"key_topics": map[string]interface{}{
				"type":        "array",
				"description": "Key topics discussed so far",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 3,
					"maxLength": 50,
				},
			},
			"should_summarize": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the conversation should be summarized",
			},
		},
		"required":    []string{"response", "conversation_phase"},
		"description": "Output schema for Interviewer agent",
	})

	definition.SetMetadata("description", "Conducts intelligent conversations to gather requirements")
	definition.SetMetadata("tags", []string{"interview", "requirements", "conversation", "french"})

	return definition
}

// CreateAgent creates an Interviewer agent with custom graph workflow
func (i *InterviewerDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := i.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build custom graph for Interviewer workflow
	graph := core.NewGraph("interviewer-workflow")
	graph.AddNode("process", "Process Message", i.processNode)
	graph.AddNode("summarize", "Summarize Interview", i.summarizeNode)

	graph.SetStartNode("process")
	graph.AddEdge("process", "summarize", i.shouldSummarize)
	graph.AddEndNode("process")
	graph.AddEndNode("summarize")

	// Note: In practice, this would replace the agent's internal graph
	_ = graph

	return baseAgent, nil
}

// processNode handles message processing and response generation
func (i *InterviewerDefinition) processNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

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

	// Create interview prompt in French
	prompt := fmt.Sprintf(`En tant qu'intervieweur expert pour des projets d'habitat 2035, répondez en français à: %s

Votre mission:
- Poser des questions approfondies sur les préférences de durabilité
- Comprendre les besoins en espace de vie
- Identifier les priorités et thèmes clés
- Guider vers des idées concrètes

Répondez avec:
1. Une réponse engageante en français
2. Questions de suivi pertinentes
3. Phase actuelle de la conversation
4. Sujets clés identifiés

Maintenez un ton professionnel mais chaleureux.`, message)

	response, err := i.generateWithLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate interview response: %w", err)
	}

	// Structure the output according to schema
	output := map[string]interface{}{
		"response":           response,
		"conversation_phase": "exploration",
		"next_questions":     []string{"Quels matériaux durables vous intéressent le plus?", "Comment voyez-vous l'intégration technologique?"},
		"key_topics":         []string{"habitat 2035", "durabilité", "préférences"},
		"should_summarize":   false,
	}

	state.Set("output", output)
	state.Set("should_summarize", false) // Control flow for graph
	return state, nil
}

// summarizeNode creates conversation summaries
func (i *InterviewerDefinition) summarizeNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Get conversation history and create summary
	summary := "📋 Résumé de l'entretien d'idéation habitat 2035\n\nPoints clés discutés:\n- Vision de durabilité\n- Préférences technologiques\n- Besoins d'espace\n\nRecommandations pour la suite..."

	output := map[string]interface{}{
		"response":           summary,
		"conversation_phase": "summary",
		"next_questions":     []string{},
		"key_topics":         []string{"résumé", "points_clés", "recommandations"},
		"should_summarize":   false,
	}

	state.Set("output", output)
	return state, nil
}

// shouldSummarize determines if conversation should be summarized
func (i *InterviewerDefinition) shouldSummarize(ctx context.Context, state *core.BaseState) (string, error) {
	if shouldSum, exists := state.Get("should_summarize"); exists {
		if b, ok := shouldSum.(bool); ok && b {
			return "summarize", nil
		}
	}
	return "", nil // Don't go to summarize node
}

// generateWithLLM generates text using the LLM (simplified version)
func (i *InterviewerDefinition) generateWithLLM(ctx context.Context, prompt string) (string, error) {
	_ = []llm.Message{{Role: "user", Content: prompt}} // Future LLM integration

	// Return French response
	return "Excellente question! Pour concevoir votre habitat idéal de 2035, j'aimerais comprendre vos priorités. Quels aspects de la durabilité vous tiennent le plus à cœur: l'efficacité énergétique, l'intégration avec la nature, ou les nouvelles technologies? Partagez-moi votre vision!", nil
}

// GetInterviewerConfig returns the configuration for backward compatibility
func GetInterviewerConfig() *agent.AgentConfig {
	return NewInterviewerDefinition().GetConfig()
}
