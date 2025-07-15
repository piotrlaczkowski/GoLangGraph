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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

// InterviewerDefinition implements AgentDefinition for the Interviewer agent
type InterviewerDefinition struct {
	*agent.BaseAgentDefinition
}

// NewInterviewerDefinition creates a new Interviewer agent definition
func NewInterviewerDefinition() *InterviewerDefinition {
	config := &agent.AgentConfig{
		Name:        "Smart Interviewer",
		Type:        agent.AgentTypeChat,
		Model:       "gemma3:1b",
		Provider:    "ollama",
		Temperature: 0.7,
		MaxTokens:   800, // Increased for complete responses
		SystemPrompt: `You are a friendly French-speaking interviewer for sustainable habitat design projects in 2035.

Your job:
1. Ask engaging questions about sustainable living preferences
2. Identify conversation phases: exploration, energy_focus, materials_focus, technology_focus, synthesis
3. Always respond in French with emojis
4. Always return valid JSON in this format:

{
  "response": "Your French response with emojis",
  "conversation_phase": "exploration",
  "key_topics": ["topic1", "topic2"],
  "next_questions": ["question1", "question2"],
  "should_summarize": false
}

Be conversational and engaging!`,
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
				"enum":        []interface{}{"exploration", "energy_focus", "materials_focus", "technology_focus", "synthesis"},
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

	definition.SetMetadata("description", "Conducts intelligent AI-powered conversations to gather requirements")
	definition.SetMetadata("tags", []string{"interview", "requirements", "conversation", "french", "ai-powered"})

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

	// Set the custom graph on the agent
	baseAgent.SetGraph(graph)

	return baseAgent, nil
}

// processNode handles message processing with simplified prompt
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

	// Create a simple, direct user prompt
	userPrompt := fmt.Sprintf(`User said: "%s"

Respond in French as a friendly interviewer. Return JSON only.`, message)

	// Use the base agent directly without modifying system prompts
	baseAgent, err := i.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, fmt.Errorf("failed to create base agent: %w", err)
	}

	// Execute with the user prompt (system prompt is already configured)
	execution, err := baseAgent.Execute(ctx, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM execution failed: %w", err)
	}

	response := execution.Output
	var structuredOutput map[string]interface{}

	// Clean up response by removing markdown code blocks if present
	cleanedResponse := strings.TrimSpace(response)
	if strings.HasPrefix(cleanedResponse, "```json") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
	}
	if strings.HasPrefix(cleanedResponse, "```") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```")
	}
	if strings.HasSuffix(cleanedResponse, "```") {
		cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
	}
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(cleanedResponse), &structuredOutput); err != nil {
		// If JSON parsing fails, create a fallback response
		phase := i.determinePhase(message)
		topics := i.extractTopics(message)

		structuredOutput = map[string]interface{}{
			"response":           fmt.Sprintf("Merci pour votre message ! Pouvez-vous me parler plus de vos intérêts en matière d'habitat durable ? 🏡✨"),
			"conversation_phase": phase,
			"key_topics":         topics,
			"next_questions":     i.generateFallbackQuestions(phase),
			"should_summarize":   false,
		}
	}

	// Store both for backward compatibility
	state.Set("output", structuredOutput)
	state.Set("structured_output", structuredOutput)

	return state, nil
}

// summarizeNode creates AI-generated conversation summaries
func (i *InterviewerDefinition) summarizeNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	summary := `📋 **Synthèse - Entretien Habitat 2035**

**Points clés identifiés :**
- Vision durabilité personnalisée
- Préférences technologiques définies
- Besoins spatiaux contextualisés
- Approche intégrée nature-innovation

**Recommandations :**
Développement concept sur mesure avec experts complémentaires.`

	output := map[string]interface{}{
		"response":           summary,
		"conversation_phase": "summary",
		"next_questions":     []string{},
		"key_topics":         []string{"synthèse", "recommandations"},
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
	return "", nil
}

// Helper functions
func (i *InterviewerDefinition) determinePhase(message string) string {
	lowerMessage := strings.ToLower(message)

	if strings.Contains(lowerMessage, "energy") || strings.Contains(lowerMessage, "énergie") ||
		strings.Contains(lowerMessage, "solar") || strings.Contains(lowerMessage, "solaire") {
		return "energy_focus"
	}
	if strings.Contains(lowerMessage, "material") || strings.Contains(lowerMessage, "matériau") ||
		strings.Contains(lowerMessage, "construction") {
		return "materials_focus"
	}
	if strings.Contains(lowerMessage, "tech") || strings.Contains(lowerMessage, "smart") ||
		strings.Contains(lowerMessage, "iot") || strings.Contains(lowerMessage, "ai") ||
		strings.Contains(lowerMessage, "robot") || strings.Contains(lowerMessage, "automation") ||
		strings.Contains(lowerMessage, "domotique") {
		return "technology_focus"
	}
	if len(message) > 50 {
		return "synthesis"
	}
	return "exploration"
}

func (i *InterviewerDefinition) extractTopics(message string) []string {
	topics := []string{}
	lowerMessage := strings.ToLower(message)

	if strings.Contains(lowerMessage, "energy") || strings.Contains(lowerMessage, "énergie") {
		topics = append(topics, "énergie")
	}
	if strings.Contains(lowerMessage, "material") || strings.Contains(lowerMessage, "matériau") {
		topics = append(topics, "matériaux")
	}
	if strings.Contains(lowerMessage, "tech") || strings.Contains(lowerMessage, "robot") {
		topics = append(topics, "technologie")
	}
	if len(topics) == 0 {
		topics = append(topics, "habitat 2035")
	}

	return topics
}

// generateFallbackQuestions creates default questions when JSON parsing fails
func (i *InterviewerDefinition) generateFallbackQuestions(phase string) []string {
	switch phase {
	case "energy_focus":
		return []string{"Quel type d'énergie renouvelable vous intéresse le plus ?", "Préférez-vous l'autonomie énergétique ou l'intégration au réseau ?"}
	case "materials_focus":
		return []string{"Préférez-vous les matériaux naturels ou high-tech ?", "L'approvisionnement local est-il important pour vous ?"}
	case "technology_focus":
		return []string{"Quel niveau d'automatisation souhaitez-vous ?", "Quelle est votre priorité : commodité ou confidentialité ?"}
	default:
		return []string{"Quel aspect de l'habitat durable vous passionne le plus ?", "Êtes-vous plus intéressé par la technologie ou les solutions basées sur la nature ?"}
	}
}

// GetInterviewerConfig returns the configuration for backward compatibility
func GetInterviewerConfig() *agent.AgentConfig {
	return NewInterviewerDefinition().GetConfig()
}
