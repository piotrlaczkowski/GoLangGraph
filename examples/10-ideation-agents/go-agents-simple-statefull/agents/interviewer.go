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
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

// InterviewContext represents the current interview state
type InterviewContext struct {
	Phase          string                 `json:"phase"`
	TopicsCovered  []string               `json:"topics_covered"`
	UserProfile    map[string]interface{} `json:"user_profile"`
	QuestionCount  int                    `json:"question_count"`
	InterviewScore float64                `json:"interview_score"`
	LastActivity   time.Time              `json:"last_activity"`
}

// InterviewSession manages interview progress and state
type InterviewSession struct {
	SessionID    string                 `json:"session_id"`
	UserID       string                 `json:"user_id"`
	Context      InterviewContext       `json:"context"`
	Insights     []string               `json:"insights"`
	Requirements map[string]interface{} `json:"requirements"`
	StartTime    time.Time              `json:"start_time"`
	LastUpdate   time.Time              `json:"last_update"`
}

// InterviewerDefinition implements enhanced AgentDefinition for the Interviewer agent
type InterviewerDefinition struct {
	*agent.BaseAgentDefinition
	checkpointer   persistence.Checkpointer
	sessionManager *persistence.SessionManager
}

// NewInterviewerDefinition creates a new enhanced Interviewer agent definition
func NewInterviewerDefinition() *InterviewerDefinition {
	config := &agent.AgentConfig{
		Name:     "Smart Interviewer",
		Type:     agent.AgentTypeChat,
		Model:    "gemma3:1b",
		Provider: "ollama",
		SystemPrompt: `You are an intelligent French-speaking interviewer specializing in gathering requirements for sustainable habitat design projects in 2035.

Your expertise includes:
- Conducting structured conversations to understand user needs
- Asking probing questions about sustainability preferences
- Gathering requirements for future living spaces
- Identifying key themes and priorities
- Maintaining conversation context across sessions
- Learning from user preferences over time

IMPORTANT: Always respond in French when conducting interviews. Ask thoughtful follow-up questions to deepen understanding.
Use the conversation history to maintain context and avoid repeating questions.
Adapt your interview style based on the user's responses and engagement level.`,
	}

	definition := &InterviewerDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
		checkpointer:        persistence.NewMemoryCheckpointer(),
		sessionManager:      nil, // Simplified for now
	}

	// Set comprehensive schema metadata
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "User message or response to interview questions",
				"minLength":   1,
				"maxLength":   2000,
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session ID for conversation continuity",
				"maxLength":   100,
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "User ID for personalization",
				"maxLength":   100,
			},
			"force_phase": map[string]interface{}{
				"type":        "string",
				"description": "Force specific interview phase",
				"enum":        []interface{}{"introduction", "exploration", "deep_dive", "synthesis", "conclusion"},
			},
		},
		"required":    []string{"message"},
		"description": "Input schema for enhanced Interviewer agent",
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
			"insights": map[string]interface{}{
				"type":        "array",
				"description": "Insights gathered about user preferences",
				"items": map[string]interface{}{
					"type":      "string",
					"minLength": 10,
					"maxLength": 300,
				},
			},
			"requirements": map[string]interface{}{
				"type":        "object",
				"description": "Structured requirements extracted from conversation",
				"properties": map[string]interface{}{
					"space_preferences":         map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"sustainability_priorities": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"technology_interests":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"budget_considerations":     map[string]interface{}{"type": "string"},
					"timeline_preferences":      map[string]interface{}{"type": "string"},
				},
			},
			"interview_progress": map[string]interface{}{
				"type":        "number",
				"description": "Interview completion percentage (0-100)",
				"minimum":     0,
				"maximum":     100,
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session ID for conversation tracking",
			},
		},
		"required":    []string{"response", "conversation_phase", "session_id"},
		"description": "Output schema for enhanced Interviewer agent",
	})

	definition.SetMetadata("description", "Conducts intelligent French conversations to gather habitat requirements with session continuity")
	definition.SetMetadata("tags", []string{"interview", "requirements", "conversation", "french", "stateful"})

	return definition
}

// CreateAgent creates an enhanced Interviewer agent with custom processing
func (i *InterviewerDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := i.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build enhanced workflow graph with state management
	graph := core.NewGraph("enhanced-interviewer-workflow")

	// Core workflow nodes
	graph.AddNode("process_interview", "Process Interview", i.processInterviewNode)
	graph.SetStartNode("process_interview")
	graph.AddEndNode("process_interview")

	return baseAgent, nil
}

// processInterviewNode handles the complete interview processing
func (i *InterviewerDefinition) processInterviewNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	inputData, exists := state.Get("input")
	if !exists {
		return nil, fmt.Errorf("no input provided")
	}

	input := inputData.(map[string]interface{})
	message := input["message"].(string)

	sessionID, _ := input["session_id"].(string)
	userID, _ := input["user_id"].(string)

	if sessionID == "" {
		sessionID = fmt.Sprintf("interview_%d", time.Now().Unix())
	}
	if userID == "" {
		userID = "anonymous"
	}

	// Initialize or load session
	session := &InterviewSession{
		SessionID: sessionID,
		UserID:    userID,
		Context: InterviewContext{
			Phase:         "introduction",
			TopicsCovered: []string{},
			UserProfile:   make(map[string]interface{}),
			QuestionCount: 0,
			LastActivity:  time.Now(),
		},
		Insights:     []string{},
		Requirements: make(map[string]interface{}),
		StartTime:    time.Now(),
		LastUpdate:   time.Now(),
	}

	// Try to load existing session
	if i.checkpointer != nil {
		if savedCheckpoint, err := i.checkpointer.Load(ctx, sessionID, sessionID); err == nil && savedCheckpoint != nil {
			if savedSessionData, exists := savedCheckpoint.State.Get("session"); exists {
				if savedSession, ok := savedSessionData.(*InterviewSession); ok {
					session = savedSession
					session.LastUpdate = time.Now()
				}
			}
		}
	}

	// Update session context
	session.Context.QuestionCount++
	session.Context.LastActivity = time.Now()

	// Determine next phase based on conversation progress
	if session.Context.QuestionCount <= 2 {
		session.Context.Phase = "introduction"
	} else if session.Context.QuestionCount <= 8 {
		session.Context.Phase = "exploration"
	} else if session.Context.QuestionCount <= 15 {
		session.Context.Phase = "deep_dive"
	} else if session.Context.QuestionCount <= 20 {
		session.Context.Phase = "synthesis"
	} else {
		session.Context.Phase = "conclusion"
	}

	// Force phase if specified
	if forcePhase, ok := input["force_phase"].(string); ok && forcePhase != "" {
		session.Context.Phase = forcePhase
	}

	// Build comprehensive prompt with context
	prompt := i.buildInterviewPrompt(message, session)

	// Generate response using LLM
	response, err := i.generateWithLLM(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate interview response: %w", err)
	}

	// Generate follow-up questions based on phase
	nextQuestions := i.generateFollowUpQuestions(session.Context.Phase)

	// Extract insights from user message
	insights := i.extractUserInsights(message, session.Context.Phase)
	session.Insights = append(session.Insights, insights...)

	// Update requirements based on insights
	i.updateRequirements(session, insights)

	// Update user profile
	i.updateUserProfile(session, message)

	// Calculate interview progress
	progress := float64(session.Context.QuestionCount) / 20.0 * 100
	if progress > 100 {
		progress = 100
	}

	// Save session state using checkpointer
	if i.checkpointer != nil {
		checkpoint := &persistence.Checkpoint{
			ID:        sessionID,
			ThreadID:  sessionID,
			State:     state,
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
			NodeID:    "interview",
			StepID:    session.Context.QuestionCount,
		}
		state.Set("session", session)
		i.checkpointer.Save(ctx, checkpoint)
	}

	// Structure final output
	output := map[string]interface{}{
		"response":           response,
		"next_questions":     nextQuestions,
		"conversation_phase": session.Context.Phase,
		"key_topics":         session.Context.TopicsCovered,
		"insights":           insights,
		"requirements":       session.Requirements,
		"interview_progress": progress,
		"session_id":         session.SessionID,
	}

	state.Set("output", output)
	return state, nil
}

// Helper methods

func (i *InterviewerDefinition) buildInterviewPrompt(message string, session *InterviewSession) string {
	phaseInstructions := map[string]string{
		"introduction": "Commencez par vous présenter et demander les objectifs généraux du projet d'habitat.",
		"exploration":  "Explorez les préférences de base: espace, style, technologie, budget approximatif.",
		"deep_dive":    "Approfondissez les aspects techniques: matériaux, énergie, systèmes durables.",
		"synthesis":    "Synthétisez les informations et clarifiez les priorités principales.",
		"conclusion":   "Concluez en résumant les exigences et prochaines étapes.",
	}

	return fmt.Sprintf(`En tant qu'intervieweur expert pour des projets d'habitat durable 2035, répondez en français à: "%s"

Phase actuelle: %s
%s

Contexte utilisateur actuel:
- Sujets couverts: %v
- Nombre de questions: %d
- Profil utilisateur: %v

Instructions spécifiques:
- Utilisez uniquement le français
- Posez une question principale avec 2-3 sous-questions
- Maintenez un ton professionnel mais chaleureux
- Construisez sur les réponses précédentes
- Identifiez les besoins techniques et émotionnels
- Guidez vers des solutions concrètes

Répondez de manière engageante et structurée.`,
		message,
		session.Context.Phase,
		phaseInstructions[session.Context.Phase],
		session.Context.TopicsCovered,
		session.Context.QuestionCount,
		session.Context.UserProfile)
}

func (i *InterviewerDefinition) generateFollowUpQuestions(phase string) []string {
	questions := map[string][]string{
		"introduction": {
			"Quel est votre budget approximatif pour ce projet?",
			"Combien de personnes habiteront dans cet espace?",
			"Avez-vous une préférence géographique?",
		},
		"exploration": {
			"Quels matériaux durables vous intéressent le plus?",
			"Comment voyez-vous l'intégration technologique?",
			"Quelle importance accordez-vous à l'autosuffisance énergétique?",
		},
		"deep_dive": {
			"Souhaitez-vous des systèmes de récupération d'eau de pluie?",
			"Comment envisagez-vous la production alimentaire sur site?",
			"Quels sont vos besoins en espaces de travail?",
		},
		"synthesis": {
			"Quelles sont vos trois priorités principales?",
			"Y a-t-il des compromis que vous êtes prêt à accepter?",
			"Quel délai envisagez-vous pour la réalisation?",
		},
		"conclusion": {
			"Souhaitez-vous des recommandations d'architectes spécialisés?",
			"Avez-vous des questions sur la faisabilité technique?",
			"Voulez-vous planifier une consultation de design?",
		},
	}

	if q, exists := questions[phase]; exists {
		return q
	}
	return questions["exploration"]
}

func (i *InterviewerDefinition) extractUserInsights(message string, phase string) []string {
	insights := []string{}

	// Simple keyword-based insight extraction
	keywords := map[string]string{
		"solaire":     "Intérêt pour l'énergie solaire",
		"jardin":      "Préfère les espaces verts intégrés",
		"minimaliste": "Style de vie minimaliste",
		"technologie": "Ouvert aux solutions technologiques",
		"bois":        "Préférence pour les matériaux naturels",
		"économie":    "Sensible aux coûts",
		"écologique":  "Priorité forte sur l'écologie",
		"autonome":    "Désir d'autonomie énergétique",
	}

	for keyword, insight := range keywords {
		if contains(message, keyword) {
			insights = append(insights, insight)
		}
	}

	if len(insights) == 0 {
		insights = append(insights, fmt.Sprintf("Réponse en phase %s: intérêts à explorer davantage", phase))
	}

	return insights
}

func (i *InterviewerDefinition) updateRequirements(session *InterviewSession, insights []string) {
	if session.Requirements == nil {
		session.Requirements = make(map[string]interface{})
	}

	// Initialize requirement categories
	categories := []string{"space_preferences", "sustainability_priorities", "technology_interests"}
	for _, cat := range categories {
		if _, exists := session.Requirements[cat]; !exists {
			session.Requirements[cat] = []string{}
		}
	}

	// Add insights to appropriate categories
	for _, insight := range insights {
		if contains(insight, "énergie") || contains(insight, "solaire") {
			prefs := session.Requirements["sustainability_priorities"].([]string)
			session.Requirements["sustainability_priorities"] = append(prefs, insight)
		} else if contains(insight, "technologie") || contains(insight, "smart") {
			prefs := session.Requirements["technology_interests"].([]string)
			session.Requirements["technology_interests"] = append(prefs, insight)
		} else {
			prefs := session.Requirements["space_preferences"].([]string)
			session.Requirements["space_preferences"] = append(prefs, insight)
		}
	}
}

func (i *InterviewerDefinition) updateUserProfile(session *InterviewSession, message string) {
	if session.Context.UserProfile == nil {
		session.Context.UserProfile = make(map[string]interface{})
	}

	// Update engagement score
	engagementScore, _ := session.Context.UserProfile["engagement_score"].(float64)
	messageLength := len(message)
	if messageLength > 100 {
		engagementScore += 0.2
	} else if messageLength > 50 {
		engagementScore += 0.1
	}
	session.Context.UserProfile["engagement_score"] = engagementScore

	// Update last interaction
	session.Context.UserProfile["last_interaction"] = time.Now()
}

func (i *InterviewerDefinition) generateWithLLM(ctx context.Context, prompt string) (string, error) {
	// Placeholder for LLM integration
	_ = []llm.Message{{Role: "user", Content: prompt}}

	// Return contextual French response based on prompt content
	if contains(prompt, "introduction") {
		return "Bonjour ! Je suis votre conseiller spécialisé en habitats durables 2035. Mon rôle est de comprendre votre vision pour créer l'espace de vie parfait qui respecte l'environnement et vos besoins personnels. Pour commencer, pourriez-vous me parler de ce qui vous motive dans ce projet d'habitat durable ?", nil
	}

	return "Excellente question ! Pour mieux comprendre vos besoins spécifiques en matière d'habitat durable, j'aimerais explorer vos priorités. Quels aspects de la durabilité vous tiennent le plus à cœur : l'efficacité énergétique, l'intégration avec la nature, l'utilisation de matériaux écologiques, ou l'autonomie alimentaire ? N'hésitez pas à partager votre vision idéale !", nil
}

// Utility function
func contains(text, substring string) bool {
	return len(text) >= len(substring) &&
		(text == substring ||
			len(text) > len(substring) &&
				(text[:len(substring)] == substring ||
					text[len(text)-len(substring):] == substring ||
					findInString(text, substring)))
}

func findInString(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		if text[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}

// GetInterviewerConfig returns the configuration for backward compatibility
func GetInterviewerConfig() *agent.AgentConfig {
	// For backward compatibility, create a temporary instance
	temp := NewInterviewerDefinition()
	return temp.GetConfig()
}
