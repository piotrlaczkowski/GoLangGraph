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
	"strings"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

// HighlighterDefinition implements AgentDefinition for the Highlighter agent
type HighlighterDefinition struct {
	*agent.BaseAgentDefinition
}

// NewHighlighterDefinition creates a new Highlighter agent definition
func NewHighlighterDefinition() *HighlighterDefinition {
	config := &agent.AgentConfig{
		Name:        "Insight Highlighter",
		Type:        agent.AgentTypeChat,
		Model:       "gemma3:1b",
		Provider:    "ollama",
		Temperature: 0.7,
		MaxTokens:   500,
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

// CreateAgent creates a Highlighter agent with custom graph workflow
func (h *HighlighterDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := h.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build custom graph for Highlighter workflow
	graph := core.NewGraph("highlighter-workflow")
	graph.AddNode("analyze_conversation", "Analyze Conversation", h.analyzeConversationNode)
	graph.AddNode("extract_insights", "Extract Insights", h.extractInsightsNode)

	graph.SetStartNode("analyze_conversation")
	graph.AddEdge("analyze_conversation", "extract_insights", nil)
	graph.AddEndNode("extract_insights")

	// Set the custom graph on the agent
	baseAgent.SetGraph(graph)

	return baseAgent, nil
}

// analyzeConversationNode analyzes the conversation patterns
func (h *HighlighterDefinition) analyzeConversationNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

	var conversationHistory []map[string]interface{}
	var fullConversationText string

	switch v := inputData.(type) {
	case string:
		// If just a string, treat as a single message
		conversationHistory = []map[string]interface{}{
			{"role": "user", "content": v},
		}
		fullConversationText = v
	case map[string]interface{}:
		if hist, ok := v["conversation_history"].([]interface{}); ok {
			for _, item := range hist {
				if msg, ok := item.(map[string]interface{}); ok {
					conversationHistory = append(conversationHistory, msg)
					if content, ok := msg["content"].(string); ok {
						fullConversationText += content + " "
					}
				}
			}
		} else if content, ok := v["message"].(string); ok {
			// Fallback for simple message format
			conversationHistory = []map[string]interface{}{
				{"role": "user", "content": content},
			}
			fullConversationText = content
		} else {
			return nil, fmt.Errorf("no conversation_history or message field in input")
		}
	default:
		return nil, fmt.Errorf("invalid input type")
	}

	// Analyze conversation content to determine focus areas
	lowerText := strings.ToLower(fullConversationText)
	var analysisContext string

	if strings.Contains(lowerText, "energy") || strings.Contains(lowerText, "solar") || strings.Contains(lowerText, "power") {
		analysisContext = "energy_sustainability"
	} else if strings.Contains(lowerText, "water") || strings.Contains(lowerText, "rain") || strings.Contains(lowerText, "hydro") {
		analysisContext = "water_management"
	} else if strings.Contains(lowerText, "urban") || strings.Contains(lowerText, "city") || strings.Contains(lowerText, "vertical") {
		analysisContext = "urban_design"
	} else if strings.Contains(lowerText, "natural") || strings.Contains(lowerText, "bio") || strings.Contains(lowerText, "eco") {
		analysisContext = "biophilic_design"
	} else if strings.Contains(lowerText, "technology") || strings.Contains(lowerText, "smart") || strings.Contains(lowerText, "ai") {
		analysisContext = "technology_integration"
	} else {
		analysisContext = "general_sustainability"
	}

	analysis := fmt.Sprintf(`🔍 **Conversation Analysis - Habitat 2035 Focus: %s**

**Messages Analyzed:** %d entries
**Primary Context:** %s discussion
**Analysis Scope:** Sustainable habitat design requirements and preferences
**Extraction Focus:** User needs, design priorities, and innovation opportunities`, analysisContext, len(conversationHistory), analysisContext)

	state.Set("conversation_analysis", analysis)
	state.Set("conversation_history", conversationHistory)
	state.Set("analysis_context", analysisContext)
	state.Set("full_text", fullConversationText)
	return state, nil
}

// extractInsightsNode extracts key insights and themes
func (h *HighlighterDefinition) extractInsightsNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	analysisContext, _ := state.Get("analysis_context")

	var keyInsights []string
	var themes []map[string]interface{}
	var actionableItems []string
	var summary string

	// Generate contextual insights based on conversation content
	switch analysisContext {
	case "energy_sustainability":
		keyInsights = []string{
			"🔋 Strong emphasis on energy independence and renewable power generation",
			"⚡ Interest in cutting-edge energy storage and distribution technologies",
			"📊 Focus on monitoring and optimizing energy consumption patterns",
			"🔄 Desire for integration with smart grid and community energy sharing",
			"🌞 Preference for building-integrated photovoltaics over traditional panels",
		}
		themes = []map[string]interface{}{
			{
				"name":       "Energy Independence",
				"importance": "high",
				"quotes":     []string{"renewable energy", "energy positive", "self-sufficient power"},
			},
			{
				"name":       "Smart Energy Management",
				"importance": "high",
				"quotes":     []string{"intelligent distribution", "consumption optimization", "AI energy systems"},
			},
			{
				"name":       "Grid Integration",
				"importance": "medium",
				"quotes":     []string{"community sharing", "smart grid", "energy trading"},
			},
		}
		actionableItems = []string{
			"Design integrated photovoltaic systems with architectural aesthetics",
			"Implement AI-driven energy optimization and predictive management",
			"Create community energy sharing platforms and grid integration",
			"Develop advanced battery storage solutions using recycled materials",
		}
		summary = "This conversation reveals a sophisticated understanding of energy systems with emphasis on both independence and community integration. Users seek advanced renewable energy solutions that go beyond basic sustainability to create energy-positive habitats."

	case "water_management":
		keyInsights = []string{
			"💧 Comprehensive water cycle integration from collection to reuse",
			"🌊 Aesthetic appreciation for water features as functional design elements",
			"🔄 Strong interest in closed-loop water systems with zero waste",
			"🌱 Integration of water management with food production systems",
			"📈 Focus on water quality monitoring and purification technologies",
		}
		themes = []map[string]interface{}{
			{
				"name":       "Water Cycle Integration",
				"importance": "high",
				"quotes":     []string{"rainwater collection", "greywater recycling", "atmospheric extraction"},
			},
			{
				"name":       "Aesthetic Functionality",
				"importance": "medium",
				"quotes":     []string{"water features", "living walls", "sculptural elements"},
			},
			{
				"name":       "Quality and Purity",
				"importance": "high",
				"quotes":     []string{"filtration systems", "water quality", "purification technology"},
			},
		}
		actionableItems = []string{
			"Design beautiful water collection and display systems",
			"Integrate aquaponics and hydroponic food production",
			"Implement advanced filtration using natural biological processes",
			"Create transparent water flow systems for educational awareness",
		}
		summary = "The discussion demonstrates a holistic approach to water management that values both functionality and beauty. Users want water systems that serve as both practical infrastructure and aesthetic features of their habitat."

	case "urban_design":
		keyInsights = []string{
			"🏙️ Recognition that urban density requires innovative spatial solutions",
			"🌿 Strong desire to bring nature into dense urban environments",
			"🔗 Interest in community connectivity and shared urban resources",
			"📱 Appreciation for technology that enables flexible urban living",
			"🚁 Forward-thinking about urban logistics and transportation integration",
		}
		themes = []map[string]interface{}{
			{
				"name":       "Vertical Living Optimization",
				"importance": "high",
				"quotes":     []string{"space efficiency", "vertical expansion", "modular design"},
			},
			{
				"name":       "Urban Nature Integration",
				"importance": "high",
				"quotes":     []string{"green corridors", "living facades", "air filtration"},
			},
			{
				"name":       "Community Connectivity",
				"importance": "medium",
				"quotes":     []string{"sky bridges", "shared spaces", "social interaction"},
			},
		}
		actionableItems = []string{
			"Develop modular housing systems for urban vertical expansion",
			"Create green infrastructure that connects urban nature networks",
			"Design community spaces that foster social interaction",
			"Integrate urban logistics with drone and autonomous vehicle systems",
		}
		summary = "Urban habitat discussions focus on solving density challenges while maintaining quality of life. Users seek innovative solutions that maximize space efficiency while preserving human connection to nature and community."

	case "biophilic_design":
		keyInsights = []string{
			"🌱 Deep understanding of human need for connection with natural systems",
			"🔬 Interest in biomimicry and learning from natural processes",
			"🌺 Appreciation for living buildings that grow and adapt over time",
			"🐝 Recognition of habitat role in supporting local ecosystems",
			"🌍 Holistic view of human habitation as part of larger environmental systems",
		}
		themes = []map[string]interface{}{
			{
				"name":       "Living Architecture",
				"importance": "high",
				"quotes":     []string{"mycelium structures", "self-healing materials", "biological processes"},
			},
			{
				"name":       "Ecosystem Integration",
				"importance": "high",
				"quotes":     []string{"biodiversity support", "natural habitats", "symbiotic relationships"},
			},
			{
				"name":       "Biomimetic Innovation",
				"importance": "medium",
				"quotes":     []string{"natural patterns", "organic forms", "evolutionary solutions"},
			},
		}
		actionableItems = []string{
			"Research and develop self-healing bio-materials for construction",
			"Design habitats that actively support local biodiversity",
			"Integrate biological air and water purification systems",
			"Create adaptive structures that respond to environmental changes",
		}
		summary = "This conversation reveals a sophisticated appreciation for biophilic design principles that go beyond adding plants to truly integrating human habitation with natural ecosystems. Users seek living buildings that benefit both humans and the environment."

	case "technology_integration":
		keyInsights = []string{
			"🤖 Strong interest in AI systems that learn and adapt to user preferences",
			"📱 Desire for seamless technology integration without visual clutter",
			"🔮 Appreciation for predictive systems that anticipate needs",
			"🔒 Awareness of privacy and security concerns with smart home systems",
			"🌐 Interest in connectivity with broader smart city infrastructure",
		}
		themes = []map[string]interface{}{
			{
				"name":       "Intelligent Automation",
				"importance": "high",
				"quotes":     []string{"AI optimization", "predictive systems", "adaptive technology"},
			},
			{
				"name":       "Seamless Integration",
				"importance": "high",
				"quotes":     []string{"invisible technology", "user experience", "natural interfaces"},
			},
			{
				"name":       "Privacy and Security",
				"importance": "medium",
				"quotes":     []string{"data protection", "secure systems", "user control"},
			},
		}
		actionableItems = []string{
			"Develop AI systems with transparent learning and user control",
			"Design technology integration that enhances rather than dominates spaces",
			"Implement robust privacy protections for smart home data",
			"Create intuitive interfaces for complex home automation systems",
		}
		summary = "Technology discussions emphasize the importance of smart systems that enhance daily life without compromising privacy or aesthetic appeal. Users want intelligent homes that feel natural and intuitive to interact with."

	default: // general_sustainability
		keyInsights = []string{
			"🌍 Comprehensive understanding of sustainability as a multifaceted challenge",
			"⚖️ Appreciation for balance between environmental impact and human comfort",
			"🔄 Interest in circular economy principles applied to housing",
			"👥 Recognition of social sustainability alongside environmental concerns",
			"🔮 Forward-thinking about long-term adaptability and resilience",
		}
		themes = []map[string]interface{}{
			{
				"name":       "Holistic Sustainability",
				"importance": "high",
				"quotes":     []string{"environmental impact", "social responsibility", "economic viability"},
			},
			{
				"name":       "Future Adaptability",
				"importance": "medium",
				"quotes":     []string{"climate resilience", "technology upgrades", "changing needs"},
			},
			{
				"name":       "Quality of Life",
				"importance": "high",
				"quotes":     []string{"human comfort", "health optimization", "well-being"},
			},
		}
		actionableItems = []string{
			"Develop integrated sustainability frameworks for habitat design",
			"Create adaptive systems that can evolve with changing technologies",
			"Balance environmental goals with human comfort and health",
			"Implement circular economy principles in building materials and systems",
		}
		summary = "The conversation demonstrates a mature understanding of sustainability that considers environmental, social, and economic factors. Users seek comprehensive solutions that create positive impacts while maintaining high quality of life."
	}

	// Structure the output according to schema
	output := map[string]interface{}{
		"key_insights":     keyInsights,
		"themes":           themes,
		"summary":          summary,
		"actionable_items": actionableItems,
	}

	state.Set("output", output)
	return state, nil
}
