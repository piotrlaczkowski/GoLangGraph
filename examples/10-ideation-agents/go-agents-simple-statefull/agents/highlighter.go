// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

// AnalysisContext represents the current analysis state
type AnalysisContext struct {
	SessionID        string    `json:"session_id"`
	AnalysisCount    int       `json:"analysis_count"`
	ThemesIdentified []string  `json:"themes_identified"`
	TotalMessages    int       `json:"total_messages"`
	LastAnalysis     time.Time `json:"last_analysis"`
	QualityScore     float64   `json:"quality_score"`
}

// ConversationAnalysis represents the results of conversation analysis
type ConversationAnalysis struct {
	ID               string                 `json:"id"`
	SessionID        string                 `json:"session_id"`
	KeyInsights      []string               `json:"key_insights"`
	Themes           []ThemeAnalysis        `json:"themes"`
	Summary          string                 `json:"summary"`
	ActionableItems  []string               `json:"actionable_items"`
	SentimentScore   float64                `json:"sentiment_score"`
	ComplexityLevel  string                 `json:"complexity_level"`
	RecommendedSteps []string               `json:"recommended_steps"`
	CreatedAt        time.Time              `json:"created_at"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// ThemeAnalysis represents analysis of a specific theme
type ThemeAnalysis struct {
	Name            string   `json:"name"`
	Importance      string   `json:"importance"`
	Quotes          []string `json:"quotes"`
	Frequency       int      `json:"frequency"`
	SentimentScore  float64  `json:"sentiment_score"`
	Recommendations []string `json:"recommendations"`
	Keywords        []string `json:"keywords"`
}

// HighlighterDefinition implements enhanced AgentDefinition for the Highlighter agent
type HighlighterDefinition struct {
	*agent.BaseAgentDefinition
	checkpointer persistence.Checkpointer
}

// NewHighlighterDefinition creates a new enhanced Highlighter agent definition
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
- Categorizing insights by themes and importance levels
- Creating actionable summaries with practical recommendations
- Detecting sentiment and emotional context
- Providing structured analysis for decision-making

Focus on sustainability themes, user preferences, design requirements, innovative ideas, and practical considerations.
Provide comprehensive structured analysis that helps inform design decisions and next steps.`,
	}

	definition := &HighlighterDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
		checkpointer:        persistence.NewMemoryCheckpointer(),
	}

	// Set comprehensive schema metadata
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
							"maxLength":   5000,
						},
						"timestamp": map[string]interface{}{
							"type":        "string",
							"description": "Message timestamp",
							"format":      "date-time",
						},
					},
					"required": []string{"role", "content"},
				},
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session ID for analysis tracking",
				"maxLength":   100,
			},
			"analysis_focus": map[string]interface{}{
				"type":        "array",
				"description": "Specific areas to focus analysis on",
				"items": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"sustainability", "technology", "budget", "timeline", "preferences", "requirements"},
				},
			},
			"depth_level": map[string]interface{}{
				"type":        "string",
				"description": "Depth of analysis required",
				"enum":        []interface{}{"surface", "detailed", "comprehensive"},
				"default":     "detailed",
			},
		},
		"required":    []string{"conversation_history"},
		"description": "Input schema for enhanced Highlighter agent",
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
					"maxLength": 300,
				},
			},
			"themes": map[string]interface{}{
				"type":        "array",
				"description": "Identified themes with detailed analysis",
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
						"frequency": map[string]interface{}{
							"type":        "integer",
							"description": "How often this theme appeared",
							"minimum":     1,
						},
						"sentiment_score": map[string]interface{}{
							"type":        "number",
							"description": "Sentiment score for this theme (-1 to 1)",
							"minimum":     -1,
							"maximum":     1,
						},
						"recommendations": map[string]interface{}{
							"type":        "array",
							"description": "Recommendations based on this theme",
							"items": map[string]interface{}{
								"type":      "string",
								"maxLength": 200,
							},
						},
					},
					"required": []string{"name", "importance"},
				},
			},
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Comprehensive summary of the conversation",
				"minLength":   100,
				"maxLength":   2000,
			},
			"actionable_items": map[string]interface{}{
				"type":        "array",
				"description": "List of actionable items derived from insights",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 15,
					"maxLength": 200,
				},
			},
			"sentiment_score": map[string]interface{}{
				"type":        "number",
				"description": "Overall conversation sentiment score (-1 to 1)",
				"minimum":     -1,
				"maximum":     1,
			},
			"complexity_level": map[string]interface{}{
				"type":        "string",
				"description": "Assessed complexity level of requirements",
				"enum":        []interface{}{"simple", "moderate", "complex", "highly_complex"},
			},
			"recommended_steps": map[string]interface{}{
				"type":        "array",
				"description": "Recommended next steps based on analysis",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 20,
					"maxLength": 300,
				},
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session ID for tracking",
			},
		},
		"required":    []string{"key_insights", "themes", "summary", "session_id"},
		"description": "Output schema for enhanced Highlighter agent",
	})

	definition.SetMetadata("description", "Extracts comprehensive insights and themes from conversations with sentiment analysis")
	definition.SetMetadata("tags", []string{"analysis", "insights", "themes", "sentiment", "recommendations", "stateful"})

	return definition
}

// CreateAgent creates an enhanced Highlighter agent with analysis workflow
func (h *HighlighterDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := h.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build enhanced workflow graph with analysis processing
	graph := core.NewGraph("enhanced-highlighter-workflow")

	// Core workflow nodes
	graph.AddNode("analyze_conversation", "Analyze Conversation", h.analyzeConversationNode)
	graph.SetStartNode("analyze_conversation")
	graph.AddEndNode("analyze_conversation")

	return baseAgent, nil
}

// analyzeConversationNode handles the complete conversation analysis
func (h *HighlighterDefinition) analyzeConversationNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

	input := inputData.(map[string]interface{})
	conversationHistory := input["conversation_history"].([]interface{})

	sessionID, _ := input["session_id"].(string)
	if sessionID == "" {
		sessionID = fmt.Sprintf("analysis_%d", time.Now().Unix())
	}

	analysisFocus, _ := input["analysis_focus"].([]interface{})
	depthLevel, _ := input["depth_level"].(string)
	if depthLevel == "" {
		depthLevel = "detailed"
	}

	// Initialize analysis context
	analysisContext := &AnalysisContext{
		SessionID:        sessionID,
		AnalysisCount:    1,
		ThemesIdentified: []string{},
		TotalMessages:    len(conversationHistory),
		LastAnalysis:     time.Now(),
		QualityScore:     0.0,
	}

	// Try to load existing analysis context
	if h.checkpointer != nil {
		if savedCheckpoint, err := h.checkpointer.Load(ctx, sessionID, sessionID); err == nil && savedCheckpoint != nil {
			if savedContextData, exists := savedCheckpoint.State.Get("analysis_context"); exists {
				if savedContext, ok := savedContextData.(*AnalysisContext); ok {
					analysisContext = savedContext
					analysisContext.AnalysisCount++
					analysisContext.LastAnalysis = time.Now()
				}
			}
		}
	}

	// Perform comprehensive conversation analysis
	_ = h.performConversationAnalysis(conversationHistory, analysisFocus, depthLevel, analysisContext)

	// Extract key insights
	keyInsights := h.extractKeyInsights(conversationHistory, analysisFocus)

	// Identify and analyze themes
	themes := h.identifyThemes(conversationHistory, analysisFocus)

	// Generate comprehensive summary
	summary := h.generateSummary(conversationHistory, themes, keyInsights)

	// Extract actionable items
	actionableItems := h.extractActionableItems(conversationHistory, themes)

	// Calculate sentiment score
	sentimentScore := h.calculateSentimentScore(conversationHistory)

	// Assess complexity level
	complexityLevel := h.assessComplexityLevel(conversationHistory, themes)

	// Generate recommendations
	recommendedSteps := h.generateRecommendations(themes, keyInsights, complexityLevel)

	// Update analysis context
	analysisContext.ThemesIdentified = h.extractThemeNames(themes)
	analysisContext.QualityScore = h.calculateQualityScore(keyInsights, themes, summary)

	// Save analysis context using checkpointer
	if h.checkpointer != nil {
		checkpoint := &persistence.Checkpoint{
			ID:        sessionID,
			ThreadID:  sessionID,
			State:     state,
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
			NodeID:    "analysis",
			StepID:    analysisContext.AnalysisCount,
		}
		state.Set("analysis_context", analysisContext)
		h.checkpointer.Save(ctx, checkpoint)
	}

	// Structure final output
	output := map[string]interface{}{
		"key_insights":      keyInsights,
		"themes":            themes,
		"summary":           summary,
		"actionable_items":  actionableItems,
		"sentiment_score":   sentimentScore,
		"complexity_level":  complexityLevel,
		"recommended_steps": recommendedSteps,
		"session_id":        sessionID,
	}

	state.Set("output", output)
	return state, nil
}

// Analysis helper methods

func (h *HighlighterDefinition) performConversationAnalysis(conversation []interface{}, focus []interface{}, depth string, context *AnalysisContext) *ConversationAnalysis {
	analysis := &ConversationAnalysis{
		ID:        fmt.Sprintf("analysis_%d", time.Now().Unix()),
		SessionID: context.SessionID,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Add analysis metadata
	analysis.Metadata["depth_level"] = depth
	analysis.Metadata["focus_areas"] = focus
	analysis.Metadata["message_count"] = len(conversation)
	analysis.Metadata["analysis_count"] = context.AnalysisCount

	return analysis
}

func (h *HighlighterDefinition) extractKeyInsights(conversation []interface{}, focus []interface{}) []string {
	insights := []string{}

	// Analyze conversation for key insights
	for _, msgInterface := range conversation {
		if msg, ok := msgInterface.(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok && msg["role"] == "user" {
				// Extract insights from user messages
				userInsights := h.analyzeUserMessage(content)
				insights = append(insights, userInsights...)
			}
		}
	}

	if len(insights) == 0 {
		insights = append(insights, "Conversation contains general habitat design discussion")
	}

	return h.deduplicateInsights(insights)
}

func (h *HighlighterDefinition) analyzeUserMessage(content string) []string {
	insights := []string{}
	content = strings.ToLower(content)

	// Keyword-based insight extraction
	insightPatterns := map[string]string{
		"budget":         "Budget considerations are important to the user",
		"timeline":       "Project timeline is a key concern",
		"sustainability": "Strong focus on sustainable solutions",
		"energy":         "Energy efficiency is a priority",
		"materials":      "Material selection is important",
		"technology":     "Interest in technological integration",
		"family":         "Family needs influence design decisions",
		"privacy":        "Privacy requirements affect space planning",
		"community":      "Community integration is valued",
		"maintenance":    "Low maintenance solutions are preferred",
		"flexibility":    "Flexible and adaptable spaces are desired",
		"natural":        "Connection to nature is important",
	}

	for keyword, insight := range insightPatterns {
		if strings.Contains(content, keyword) {
			insights = append(insights, insight)
		}
	}

	return insights
}

func (h *HighlighterDefinition) identifyThemes(conversation []interface{}, focus []interface{}) []ThemeAnalysis {
	themes := []ThemeAnalysis{}

	// Predefined themes for habitat design conversations
	themeDefinitions := map[string]ThemeAnalysis{
		"sustainability": {
			Name:       "Sustainability & Environment",
			Importance: "high",
			Keywords:   []string{"sustainability", "environment", "green", "eco", "renewable", "carbon"},
			Quotes:     []string{},
			Frequency:  0,
		},
		"technology": {
			Name:       "Technology Integration",
			Importance: "medium",
			Keywords:   []string{"technology", "smart", "automation", "digital", "sensors"},
			Quotes:     []string{},
			Frequency:  0,
		},
		"budget": {
			Name:       "Budget & Cost",
			Importance: "high",
			Keywords:   []string{"budget", "cost", "price", "expensive", "affordable", "investment"},
			Quotes:     []string{},
			Frequency:  0,
		},
		"design": {
			Name:       "Design & Aesthetics",
			Importance: "medium",
			Keywords:   []string{"design", "beautiful", "style", "aesthetic", "appearance", "look"},
			Quotes:     []string{},
			Frequency:  0,
		},
		"space": {
			Name:       "Space Planning",
			Importance: "high",
			Keywords:   []string{"space", "room", "area", "layout", "size", "square"},
			Quotes:     []string{},
			Frequency:  0,
		},
	}

	// Analyze conversation for themes
	for _, msgInterface := range conversation {
		if msg, ok := msgInterface.(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok {
				contentLower := strings.ToLower(content)

				for themeKey, theme := range themeDefinitions {
					for _, keyword := range theme.Keywords {
						if strings.Contains(contentLower, keyword) {
							theme.Frequency++
							if len(content) <= 500 && msg["role"] == "user" {
								theme.Quotes = append(theme.Quotes, content)
							}
							themeDefinitions[themeKey] = theme
						}
					}
				}
			}
		}
	}

	// Convert to slice and add recommendations
	for _, theme := range themeDefinitions {
		if theme.Frequency > 0 {
			theme.SentimentScore = 0.5 // Neutral by default
			theme.Recommendations = h.generateThemeRecommendations(theme.Name)
			themes = append(themes, theme)
		}
	}

	return themes
}

func (h *HighlighterDefinition) generateThemeRecommendations(themeName string) []string {
	recommendations := map[string][]string{
		"Sustainability & Environment": {
			"Consider renewable energy systems (solar, wind)",
			"Explore sustainable building materials",
			"Implement water conservation systems",
			"Plan for carbon-neutral operations",
		},
		"Technology Integration": {
			"Evaluate smart home automation options",
			"Consider energy monitoring systems",
			"Plan for future technology upgrades",
			"Integrate sustainable technology solutions",
		},
		"Budget & Cost": {
			"Develop detailed cost breakdown",
			"Explore financing options",
			"Consider phased implementation",
			"Identify cost-saving opportunities",
		},
		"Design & Aesthetics": {
			"Work with sustainable design principles",
			"Consider local architectural styles",
			"Plan for natural lighting optimization",
			"Integrate landscape design",
		},
		"Space Planning": {
			"Optimize space utilization",
			"Plan for future needs flexibility",
			"Consider multi-functional spaces",
			"Integrate indoor-outdoor living",
		},
	}

	if recs, exists := recommendations[themeName]; exists {
		return recs
	}
	return []string{"Continue exploring this theme with the client"}
}

func (h *HighlighterDefinition) generateSummary(conversation []interface{}, themes []ThemeAnalysis, insights []string) string {
	messageCount := len(conversation)
	themeCount := len(themes)
	insightCount := len(insights)

	summary := fmt.Sprintf("📋 Analyse de Conversation - Habitat Durable 2035\n\n")
	summary += fmt.Sprintf("Messages analysés: %d | Thèmes identifiés: %d | Insights extraits: %d\n\n",
		messageCount, themeCount, insightCount)

	if len(themes) > 0 {
		summary += "🎯 Thèmes Principaux:\n"
		for _, theme := range themes {
			summary += fmt.Sprintf("• %s (%s importance, %d mentions)\n",
				theme.Name, theme.Importance, theme.Frequency)
		}
		summary += "\n"
	}

	if len(insights) > 0 {
		summary += "💡 Insights Clés:\n"
		for i, insight := range insights {
			if i < 3 { // Show top 3 insights
				summary += fmt.Sprintf("• %s\n", insight)
			}
		}
		summary += "\n"
	}

	summary += "📈 Cette analyse révèle les priorités et préférences du client pour son projet d'habitat durable, "
	summary += "fournissant une base solide pour les prochaines étapes de conception et de planification."

	return summary
}

func (h *HighlighterDefinition) extractActionableItems(conversation []interface{}, themes []ThemeAnalysis) []string {
	actionableItems := []string{}

	// Generate actionable items based on themes
	for _, theme := range themes {
		switch theme.Name {
		case "Sustainability & Environment":
			if theme.Frequency >= 2 {
				actionableItems = append(actionableItems, "Programmer une consultation avec un expert en durabilité")
				actionableItems = append(actionableItems, "Rechercher des certifications environnementales applicables")
			}
		case "Technology Integration":
			if theme.Frequency >= 2 {
				actionableItems = append(actionableItems, "Évaluer les options de maison intelligente disponibles")
				actionableItems = append(actionableItems, "Planifier l'infrastructure technologique nécessaire")
			}
		case "Budget & Cost":
			if theme.Frequency >= 2 {
				actionableItems = append(actionableItems, "Établir un budget détaillé avec alternatives")
				actionableItems = append(actionableItems, "Explorer les options de financement durable")
			}
		case "Design & Aesthetics":
			if theme.Frequency >= 2 {
				actionableItems = append(actionableItems, "Développer des concepts visuels initiaux")
				actionableItems = append(actionableItems, "Planifier une session de design collaboratif")
			}
		case "Space Planning":
			if theme.Frequency >= 2 {
				actionableItems = append(actionableItems, "Créer des plans d'aménagement préliminaires")
				actionableItems = append(actionableItems, "Analyser les besoins d'espace détaillés")
			}
		}
	}

	if len(actionableItems) == 0 {
		actionableItems = append(actionableItems, "Approfondir la discussion sur les priorités du projet")
		actionableItems = append(actionableItems, "Planifier une visite du site potentiel")
	}

	return actionableItems
}

func (h *HighlighterDefinition) calculateSentimentScore(conversation []interface{}) float64 {
	positiveWords := []string{"excellent", "parfait", "génial", "formidable", "intéressant", "j'aime", "super"}
	negativeWords := []string{"problème", "difficile", "inquiet", "cher", "compliqué", "impossible"}

	positiveCount := 0
	negativeCount := 0
	totalWords := 0

	for _, msgInterface := range conversation {
		if msg, ok := msgInterface.(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok && msg["role"] == "user" {
				contentLower := strings.ToLower(content)
				words := strings.Fields(contentLower)
				totalWords += len(words)

				for _, word := range words {
					for _, positive := range positiveWords {
						if strings.Contains(word, positive) {
							positiveCount++
						}
					}
					for _, negative := range negativeWords {
						if strings.Contains(word, negative) {
							negativeCount++
						}
					}
				}
			}
		}
	}

	if totalWords == 0 {
		return 0.0
	}

	// Calculate sentiment score between -1 and 1
	score := float64(positiveCount-negativeCount) / float64(totalWords) * 10
	if score > 1.0 {
		score = 1.0
	} else if score < -1.0 {
		score = -1.0
	}

	return score
}

func (h *HighlighterDefinition) assessComplexityLevel(conversation []interface{}, themes []ThemeAnalysis) string {
	complexityIndicators := 0
	messageCount := len(conversation)
	themeCount := len(themes)

	// Count complexity indicators
	if messageCount > 20 {
		complexityIndicators++
	}
	if themeCount > 4 {
		complexityIndicators++
	}

	// Check for complex topics in conversation
	complexTopics := []string{"automation", "integration", "sustainable", "renewable", "smart", "complex", "multiple"}
	for _, msgInterface := range conversation {
		if msg, ok := msgInterface.(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok {
				contentLower := strings.ToLower(content)
				for _, topic := range complexTopics {
					if strings.Contains(contentLower, topic) {
						complexityIndicators++
						break
					}
				}
			}
		}
	}

	// Determine complexity level
	switch {
	case complexityIndicators >= 4:
		return "highly_complex"
	case complexityIndicators >= 2:
		return "complex"
	case complexityIndicators >= 1:
		return "moderate"
	default:
		return "simple"
	}
}

func (h *HighlighterDefinition) generateRecommendations(themes []ThemeAnalysis, insights []string, complexity string) []string {
	recommendations := []string{}

	// Base recommendations based on complexity
	switch complexity {
	case "highly_complex":
		recommendations = append(recommendations, "Organiser une série de consultations spécialisées")
		recommendations = append(recommendations, "Développer un plan de projet en phases multiples")
		recommendations = append(recommendations, "Constituer une équipe multidisciplinaire")
	case "complex":
		recommendations = append(recommendations, "Planifier des consultations avec des experts pertinents")
		recommendations = append(recommendations, "Créer un calendrier de projet détaillé")
	case "moderate":
		recommendations = append(recommendations, "Approfondir quelques aspects clés identifiés")
		recommendations = append(recommendations, "Planifier des étapes de validation avec le client")
	default:
		recommendations = append(recommendations, "Procéder avec un plan de conception standard")
	}

	// Add theme-specific recommendations
	highImportanceThemes := 0
	for _, theme := range themes {
		if theme.Importance == "high" && theme.Frequency >= 2 {
			highImportanceThemes++
		}
	}

	if highImportanceThemes >= 2 {
		recommendations = append(recommendations, "Prioriser les thèmes à forte importance identifiés")
		recommendations = append(recommendations, "Développer des solutions intégrées pour les thèmes principaux")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Continuer l'exploration des besoins avec le client")
	}

	return recommendations
}

// Utility methods

func (h *HighlighterDefinition) deduplicateInsights(insights []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, insight := range insights {
		if !seen[insight] {
			seen[insight] = true
			result = append(result, insight)
		}
	}

	return result
}

func (h *HighlighterDefinition) extractThemeNames(themes []ThemeAnalysis) []string {
	names := []string{}
	for _, theme := range themes {
		names = append(names, theme.Name)
	}
	return names
}

func (h *HighlighterDefinition) calculateQualityScore(insights []string, themes []ThemeAnalysis, summary string) float64 {
	score := 0.0

	// Score based on insights count
	score += float64(len(insights)) * 0.1

	// Score based on themes count and quality
	for _, theme := range themes {
		score += 0.2
		if theme.Importance == "high" {
			score += 0.1
		}
		if theme.Frequency >= 3 {
			score += 0.1
		}
	}

	// Score based on summary length (indicates depth)
	if len(summary) > 200 {
		score += 0.3
	} else if len(summary) > 100 {
		score += 0.2
	}

	// Normalize to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// GetHighlighterConfig returns the configuration for backward compatibility
func GetHighlighterConfig() *agent.AgentConfig {
	// For backward compatibility, create a temporary instance
	temp := NewHighlighterDefinition()
	return temp.GetConfig()
}
