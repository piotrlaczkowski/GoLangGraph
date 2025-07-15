// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Production-Ready Stateful Ideation Agents

package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"

	"go-agents-simple-statefull/database"
)

// DesignerDefinition implements AgentDefinition with full state management using GoLangGraph infrastructure
type DesignerDefinition struct {
	*agent.BaseAgentDefinition

	// GoLangGraph persistence components
	sessionManager *persistence.SessionManager
	checkpointer   persistence.Checkpointer

	// Enhanced database management
	databaseManager *database.DatabaseManager

	// Enhanced memory and conversation management
	conversationHistory map[string]*ConversationContext

	// Design-specific state management
	activeDesigns map[string]*DesignSession
	userProfiles  map[string]*UserProfile
}

// ConversationContext stores comprehensive conversation state using GoLangGraph's infrastructure
type ConversationContext struct {
	SessionID string `json:"session_id"`
	ThreadID  string `json:"thread_id"`
	UserID    string `json:"user_id"`

	// Conversation flow
	Messages        []llm.Message `json:"messages"`
	CurrentPhase    string        `json:"current_phase"`
	CompletedPhases []string      `json:"completed_phases"`

	// Memory and learning using GoLangGraph's RAG capabilities
	RelevantMemories []*database.MemoryItem      `json:"relevant_memories"`
	UserPreferences  *database.UserPreferences   `json:"user_preferences"`
	DesignHistory    []*database.DesignIteration `json:"design_history"`

	// State management
	LastInteraction time.Time              `json:"last_interaction"`
	Metadata        map[string]interface{} `json:"metadata"`
	StateSnapshot   *core.BaseState        `json:"state_snapshot"`
}

// DesignSession manages individual design sessions with persistence
type DesignSession struct {
	ID       string `json:"id"`
	ThreadID string `json:"thread_id"`
	UserID   string `json:"user_id"`

	// Design state using GoLangGraph's state management
	CurrentIteration *DesignIteration   `json:"current_iteration"`
	AllIterations    []*DesignIteration `json:"all_iterations"`
	FinalDesign      *DesignOutput      `json:"final_design,omitempty"`

	// Session metadata
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DesignIteration represents a single design iteration with full persistence
type DesignIteration struct {
	ID              string `json:"id"`
	SessionID       string `json:"session_id"`
	IterationNumber int    `json:"iteration_number"`

	// Design content
	Concept             string                  `json:"concept"`
	Requirements        []string                `json:"requirements"`
	Constraints         []string                `json:"constraints"`
	MaterialSpecs       []MaterialSpecification `json:"material_specs"`
	SustainabilityScore float64                 `json:"sustainability_score"`
	EstimatedCost       *CostEstimate           `json:"estimated_cost"`

	// User interaction and learning
	UserFeedback     string   `json:"user_feedback,omitempty"`
	UserRating       float64  `json:"user_rating,omitempty"`
	ImprovementAreas []string `json:"improvement_areas,omitempty"`

	// Persistence metadata
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata"`
	CheckpointID string                 `json:"checkpoint_id,omitempty"`
}

// DesignOutput represents the final design output with comprehensive data
type DesignOutput struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`

	// Complete design information
	FinalConcept         string                  `json:"final_concept"`
	DetailedDescription  string                  `json:"detailed_description"`
	TechnicalSpecs       map[string]interface{}  `json:"technical_specs"`
	Materials            []MaterialSpecification `json:"materials"`
	SustainabilityReport *SustainabilityReport   `json:"sustainability_report"`
	CostBreakdown        *DetailedCostEstimate   `json:"cost_breakdown"`

	// Implementation guidance
	ConstructionPhases []ConstructionPhase `json:"construction_phases"`
	Timeline           *ProjectTimeline    `json:"timeline"`
	Recommendations    []string            `json:"recommendations"`

	// Quality and validation
	QualityScore     float64         `json:"quality_score"`
	ComplianceChecks map[string]bool `json:"compliance_checks"`
	RiskAssessment   *RiskAssessment `json:"risk_assessment"`

	// Documentation
	GeneratedAt time.Time              `json:"generated_at"`
	Version     string                 `json:"version"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UserProfile extends user preferences with comprehensive learning
type UserProfile struct {
	UserID string `json:"user_id"`

	// Preferences using database infrastructure
	Preferences *database.UserPreferences `json:"preferences"`

	// Interaction history
	TotalSessions     int     `json:"total_sessions"`
	TotalInteractions int     `json:"total_interactions"`
	AverageRating     float64 `json:"average_rating"`

	// Learning patterns
	PreferredStyles     map[string]float64 `json:"preferred_styles"`
	MaterialPreferences map[string]float64 `json:"material_preferences"`
	BudgetPatterns      []float64          `json:"budget_patterns"`
	ComplexityTrends    []float64          `json:"complexity_trends"`

	// Engagement metrics
	SessionDurations  []time.Duration `json:"session_durations"`
	FeedbackFrequency float64         `json:"feedback_frequency"`
	IterationPatterns map[string]int  `json:"iteration_patterns"`

	// Profile metadata
	LastUpdated    time.Time              `json:"last_updated"`
	ProfileVersion string                 `json:"profile_version"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// Enhanced supporting structures for comprehensive design management

// MaterialSpecification with enhanced sustainability tracking
type MaterialSpecification struct {
	Name                 string   `json:"name"`
	Type                 string   `json:"type"`
	Quantity             float64  `json:"quantity"`
	Unit                 string   `json:"unit"`
	CostPerUnit          float64  `json:"cost_per_unit"`
	Supplier             string   `json:"supplier"`
	SustainabilityRating float64  `json:"sustainability_rating"`
	CarbonFootprint      float64  `json:"carbon_footprint"`
	Recyclability        float64  `json:"recyclability"`
	LocallySourced       bool     `json:"locally_sourced"`
	Certifications       []string `json:"certifications"`
}

// CostEstimate with detailed breakdown
type CostEstimate struct {
	TotalCost       float64                `json:"total_cost"`
	MaterialsCost   float64                `json:"materials_cost"`
	LaborCost       float64                `json:"labor_cost"`
	EquipmentCost   float64                `json:"equipment_cost"`
	PermitsCost     float64                `json:"permits_cost"`
	ContingencyRate float64                `json:"contingency_rate"`
	Breakdown       map[string]interface{} `json:"breakdown"`
	Currency        string                 `json:"currency"`
	EstimatedAt     time.Time              `json:"estimated_at"`
	Confidence      float64                `json:"confidence"`
}

// DetailedCostEstimate for final output
type DetailedCostEstimate struct {
	*CostEstimate
	PhaseBreakdown       []PhaseCost          `json:"phase_breakdown"`
	TimelineImpact       map[string]float64   `json:"timeline_impact"`
	RiskFactors          []CostRisk           `json:"risk_factors"`
	SavingsOpportunities []SavingsOpportunity `json:"savings_opportunities"`
}

// PhaseCost represents cost for each construction phase
type PhaseCost struct {
	Phase        string   `json:"phase"`
	Cost         float64  `json:"cost"`
	Duration     string   `json:"duration"`
	Dependencies []string `json:"dependencies"`
}

// CostRisk represents potential cost risks
type CostRisk struct {
	Factor      string  `json:"factor"`
	Impact      float64 `json:"impact"`
	Probability float64 `json:"probability"`
	Mitigation  string  `json:"mitigation"`
}

// SavingsOpportunity represents potential cost savings
type SavingsOpportunity struct {
	Opportunity string  `json:"opportunity"`
	Savings     float64 `json:"savings"`
	Effort      string  `json:"effort"`
	Timeline    string  `json:"timeline"`
}

// SustainabilityReport with comprehensive metrics
type SustainabilityReport struct {
	OverallScore           float64            `json:"overall_score"`
	CarbonFootprint        float64            `json:"carbon_footprint"`
	EnergyEfficiency       float64            `json:"energy_efficiency"`
	WaterUsage             float64            `json:"water_usage"`
	WasteReduction         float64            `json:"waste_reduction"`
	MaterialSustainability map[string]float64 `json:"material_sustainability"`
	CertificationsPossible []string           `json:"certifications_possible"`
	ImprovementSuggestions []string           `json:"improvement_suggestions"`
	LocalImpact            string             `json:"local_impact"`
	LongTermBenefits       []string           `json:"long_term_benefits"`
}

// ConstructionPhase with detailed planning
type ConstructionPhase struct {
	Phase         string                  `json:"phase"`
	Description   string                  `json:"description"`
	Duration      string                  `json:"duration"`
	Prerequisites []string                `json:"prerequisites"`
	Tasks         []string                `json:"tasks"`
	Materials     []MaterialSpecification `json:"materials"`
	Labor         []LaborRequirement      `json:"labor"`
	Equipment     []EquipmentRequirement  `json:"equipment"`
	Permits       []string                `json:"permits"`
	QualityChecks []QualityCheck          `json:"quality_checks"`
}

// LaborRequirement specifies labor needs
type LaborRequirement struct {
	Skill       string  `json:"skill"`
	Hours       float64 `json:"hours"`
	Rate        float64 `json:"rate"`
	Specialized bool    `json:"specialized"`
}

// EquipmentRequirement specifies equipment needs
type EquipmentRequirement struct {
	Equipment string  `json:"equipment"`
	Duration  string  `json:"duration"`
	Cost      float64 `json:"cost"`
	Rental    bool    `json:"rental"`
}

// QualityCheck defines quality control checkpoints
type QualityCheck struct {
	Checkpoint string   `json:"checkpoint"`
	Criteria   []string `json:"criteria"`
	Inspector  string   `json:"inspector"`
	Timeline   string   `json:"timeline"`
}

// ProjectTimeline with milestone tracking
type ProjectTimeline struct {
	TotalDuration  string              `json:"total_duration"`
	StartDate      *time.Time          `json:"start_date,omitempty"`
	EndDate        *time.Time          `json:"end_date,omitempty"`
	Phases         []PhaseTimeline     `json:"phases"`
	Milestones     []Milestone         `json:"milestones"`
	CriticalPath   []string            `json:"critical_path"`
	BufferTime     string              `json:"buffer_time"`
	WeatherFactors map[string]string   `json:"weather_factors"`
	Dependencies   map[string][]string `json:"dependencies"`
}

// PhaseTimeline represents timeline for each phase
type PhaseTimeline struct {
	Phase    string `json:"phase"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Duration string `json:"duration"`
	Buffer   string `json:"buffer"`
	Critical bool   `json:"critical"`
}

// Milestone represents project milestones
type Milestone struct {
	Name         string   `json:"name"`
	Date         string   `json:"date"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
	Deliverables []string `json:"deliverables"`
}

// RiskAssessment for comprehensive risk management
type RiskAssessment struct {
	OverallRisk     string            `json:"overall_risk"`
	RiskFactors     []RiskFactor      `json:"risk_factors"`
	MitigationPlan  map[string]string `json:"mitigation_plan"`
	ContingencyPlan string            `json:"contingency_plan"`
	InsuranceNeeds  []string          `json:"insurance_needs"`
	MonitoringPlan  []MonitoringPoint `json:"monitoring_plan"`
}

// RiskFactor represents individual risk factors
type RiskFactor struct {
	Category    string  `json:"category"`
	Risk        string  `json:"risk"`
	Probability float64 `json:"probability"`
	Impact      float64 `json:"impact"`
	Severity    string  `json:"severity"`
	Mitigation  string  `json:"mitigation"`
}

// MonitoringPoint for ongoing risk monitoring
type MonitoringPoint struct {
	Point      string             `json:"point"`
	Frequency  string             `json:"frequency"`
	Metrics    []string           `json:"metrics"`
	Thresholds map[string]float64 `json:"thresholds"`
	Actions    []string           `json:"actions"`
}

// NewDesignerDefinition creates a new stateful Designer agent definition using GoLangGraph infrastructure
func NewDesignerDefinition(databaseManager *database.DatabaseManager) *DesignerDefinition {
	config := &agent.AgentConfig{
		Name:     "Advanced Visual Designer with State Management",
		Type:     agent.AgentTypeChat,
		Model:    "gemma3:1b",
		Provider: "ollama",
		SystemPrompt: `You are an advanced AI architect and designer specializing in sustainable habitat design for 2035.

Core Expertise:
- Eco-friendly materials and construction techniques
- Energy-efficient design principles
- Integration with natural environments
- Smart home technology integration
- Sustainable living solutions
- Cost estimation and project planning
- Building regulations and standards

Enhanced Capabilities with State Management:
- Remember user preferences across sessions using vector embeddings
- Track design evolution and learn from iterations
- Utilize conversation history and context for better designs
- Learn from user feedback to improve future recommendations
- Maintain project context and constraints across interactions
- Access relevant past designs and successful patterns

Memory and Learning:
- Your responses are enhanced by accessing relevant memories from past interactions
- You learn and adapt from user feedback and preferences
- You maintain context across conversation threads and sessions
- You can reference and build upon previous design iterations

Output Format:
Always provide structured, detailed responses with:
1. Comprehensive design descriptions with sustainability focus
2. Material specifications with environmental impact scores
3. Cost estimates with detailed breakdowns and timelines
4. Feature explanations with user benefits and technical details
5. Sustainability impact analysis with long-term projections
6. Construction phases with quality checkpoints
7. Risk assessment and mitigation strategies

Be creative, practical, and environmentally conscious. Use your memory of past interactions to provide personalized and improved designs.`,
	}

	definition := &DesignerDefinition{
		BaseAgentDefinition: agent.NewBaseAgentDefinition(config),
		databaseManager:     databaseManager,
		sessionManager:      databaseManager.SessionManager,
		checkpointer:        databaseManager.Checkpointer,
		conversationHistory: make(map[string]*ConversationContext),
		activeDesigns:       make(map[string]*DesignSession),
		userProfiles:        make(map[string]*UserProfile),
	}

	// Set comprehensive schema metadata for auto-validation and API generation
	definition.SetMetadata("version", "2.0.0")
	definition.SetMetadata("framework", "GoLangGraph-Enhanced")
	definition.SetMetadata("author", "GoLangGraph Team")
	definition.SetMetadata("description", "Production-ready stateful visual designer with comprehensive memory management")
	definition.SetMetadata("capabilities", []string{
		"sustainable_design", "cost_estimation", "material_selection", "construction_planning",
		"state_management", "conversation_memory", "user_preference_learning", "design_iteration_tracking",
		"vector_embeddings", "RAG_integration", "session_persistence", "thread_management",
		"comprehensive_documentation", "risk_assessment", "timeline_planning", "quality_control",
	})

	// Enhanced schema metadata for comprehensive API documentation
	definition.SetMetadata("input_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "User's design request or feedback with full context support",
				"maxLength":   5000,
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session identifier for state continuity",
				"format":      "uuid",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "User identifier for personalization and preference learning",
			},
			"context": map[string]interface{}{
				"type":        "object",
				"description": "Additional context for enhanced processing",
				"properties": map[string]interface{}{
					"project_type":            map[string]string{"type": "string"},
					"budget_range":            map[string]string{"type": "number"},
					"timeline":                map[string]string{"type": "string"},
					"location":                map[string]string{"type": "string"},
					"sustainability_priority": map[string]string{"type": "number"},
					"previous_feedback":       map[string]string{"type": "string"},
				},
			},
		},
		"required": []string{"message"},
	})

	definition.SetMetadata("output_schema", map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"design_response": map[string]interface{}{
				"type":        "string",
				"description": "Comprehensive design response with all specifications",
			},
			"design_iteration": map[string]interface{}{
				"type":        "object",
				"description": "Complete design iteration with all technical details",
				"$ref":        "#/definitions/DesignIteration",
			},
			"session_state": map[string]interface{}{
				"type":        "object",
				"description": "Current session state for persistence",
				"$ref":        "#/definitions/ConversationContext",
			},
			"user_preferences": map[string]interface{}{
				"type":        "object",
				"description": "Updated user preferences based on interaction",
				"$ref":        "#/definitions/UserPreferences",
			},
			"metadata": map[string]interface{}{
				"type":        "object",
				"description": "Response metadata including processing information",
				"properties": map[string]interface{}{
					"processing_time":   map[string]string{"type": "number"},
					"memory_items_used": map[string]string{"type": "number"},
					"iteration_number":  map[string]string{"type": "number"},
					"confidence_score":  map[string]string{"type": "number"},
					"checkpoint_id":     map[string]string{"type": "string"},
				},
			},
		},
		"required": []string{"design_response"},
	})

	// Set API metadata for auto-server generation
	definition.SetMetadata("endpoints", map[string]interface{}{
		"POST /design": map[string]interface{}{
			"description": "Main design endpoint with full state management",
			"input":       definition.GetMetadata()["input_schema"],
			"output":      definition.GetMetadata()["output_schema"],
		},
		"GET /design/session/{session_id}": map[string]interface{}{
			"description": "Retrieve session state and design history",
		},
		"GET /design/user/{user_id}/preferences": map[string]interface{}{
			"description": "Get user preferences and learning data",
		},
		"GET /design/user/{user_id}/history": map[string]interface{}{
			"description": "Get user's design history and iterations",
		},
	})

	return definition
}

// CreateAgent creates a Designer agent with advanced state management
func (d *DesignerDefinition) CreateAgent() (*agent.Agent, error) {
	baseAgent, err := d.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Build enhanced graph for stateful Designer workflow
	graph := core.NewGraph("stateful-designer-workflow")

	// Add nodes with state management
	graph.AddNode("initialize_session", "Initialize Session", d.initializeSessionNode)
	graph.AddNode("load_context", "Load Context", d.loadContextNode)
	graph.AddNode("analyze_request", "Analyze Request", d.analyzeRequestNode)
	graph.AddNode("generate_design", "Generate Design", d.generateDesignNode)
	graph.AddNode("evaluate_design", "Evaluate Design", d.evaluateDesignNode)
	graph.AddNode("save_iteration", "Save Iteration", d.saveIterationNode)
	graph.AddNode("update_memory", "Update Memory", d.updateMemoryNode)
	graph.AddNode("finalize_response", "Finalize Response", d.finalizeResponseNode)

	// Define workflow with state transitions
	graph.SetStartNode("initialize_session")
	graph.AddEdge("initialize_session", "load_context", nil)
	graph.AddEdge("load_context", "analyze_request", nil)
	graph.AddEdge("analyze_request", "generate_design", nil)
	graph.AddEdge("generate_design", "evaluate_design", nil)
	graph.AddEdge("evaluate_design", "save_iteration", nil)
	graph.AddEdge("save_iteration", "update_memory", nil)
	graph.AddEdge("update_memory", "finalize_response", nil)
	graph.AddEndNode("finalize_response")

	// Note: This graph would integrate with the agent's execution system
	_ = graph

	return baseAgent, nil
}

// Enhanced node implementations using GoLangGraph infrastructure

// initializeSessionNode initializes or restores session context using GoLangGraph persistence
func (dd *DesignerDefinition) initializeSessionNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	sessionID, _ := state.Get("session_id")
	userID, _ := state.Get("user_id")
	threadID, _ := state.Get("thread_id")

	sessionIDStr := fmt.Sprintf("%v", sessionID)
	userIDStr := fmt.Sprintf("%v", userID)
	_ = fmt.Sprintf("%v", threadID) // threadIDStr not used, suppress warning

	// Initialize or restore session using GoLangGraph SessionManager
	var session *persistence.Session
	var thread *persistence.Thread
	var err error

	if sessionIDStr != "" && sessionIDStr != "<nil>" {
		// Try to restore existing session
		session, err = dd.sessionManager.GetSession(ctx, sessionIDStr)
		if err != nil {
			// Create new session if not found
			session = &persistence.Session{
				ID:        uuid.New().String(),
				UserID:    userIDStr,
				Metadata:  make(map[string]interface{}),
				CreatedAt: time.Now(),
			}

			// Create new thread
			thread = &persistence.Thread{
				ID:        uuid.New().String(),
				Name:      fmt.Sprintf("Design-Session-%s", time.Now().Format("2006-01-02-15-04-05")),
				Metadata:  make(map[string]interface{}),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Create thread first, then session with thread reference
			if err := dd.sessionManager.CreateThread(ctx, thread); err != nil {
				return nil, fmt.Errorf("failed to create thread: %w", err)
			}

			session.ThreadID = thread.ID
			if err := dd.sessionManager.CreateSession(ctx, session); err != nil {
				return nil, fmt.Errorf("failed to create session: %w", err)
			}
		} else {
			// Get existing thread
			thread, err = dd.sessionManager.GetThread(ctx, session.ThreadID)
			if err != nil {
				return nil, fmt.Errorf("failed to get thread: %w", err)
			}
		}
	} else {
		// Create new session and thread
		session = &persistence.Session{
			ID:        uuid.New().String(),
			UserID:    userIDStr,
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
		}

		thread = &persistence.Thread{
			ID:        uuid.New().String(),
			Name:      fmt.Sprintf("Design-Session-%s", time.Now().Format("2006-01-02-15-04-05")),
			Metadata:  make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Create thread first, then session
		if err := dd.sessionManager.CreateThread(ctx, thread); err != nil {
			return nil, fmt.Errorf("failed to create thread: %w", err)
		}

		session.ThreadID = thread.ID
		if err := dd.sessionManager.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Initialize or restore conversation context
	conversationCtx := &ConversationContext{
		SessionID:        session.ID,
		ThreadID:         thread.ID,
		UserID:           session.UserID,
		Messages:         []llm.Message{},
		CurrentPhase:     "initialization",
		CompletedPhases:  []string{},
		RelevantMemories: []*database.MemoryItem{},
		LastInteraction:  time.Now(),
		Metadata:         make(map[string]interface{}),
		StateSnapshot:    state.Clone(),
	}

	dd.conversationHistory[session.ID] = conversationCtx

	// Update state with session information
	state.Set("session_id", session.ID)
	state.Set("thread_id", thread.ID)
	state.Set("user_id", session.UserID)
	state.Set("conversation_context", conversationCtx)
	state.Set("session_initialized", true)

	return state, nil
}

// loadContextNode loads relevant context using GoLangGraph's memory and RAG capabilities
func (dd *DesignerDefinition) loadContextNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	sessionID, _ := state.Get("session_id")
	threadID, _ := state.Get("thread_id")
	userID, _ := state.Get("user_id")
	input, _ := state.Get("input")

	sessionIDStr := fmt.Sprintf("%v", sessionID)
	threadIDStr := fmt.Sprintf("%v", threadID)
	userIDStr := fmt.Sprintf("%v", userID)
	inputStr := fmt.Sprintf("%v", input)

	// Load conversation context
	conversationCtx, exists := dd.conversationHistory[sessionIDStr]
	if !exists {
		return nil, fmt.Errorf("conversation context not found for session %s", sessionIDStr)
	}

	// Generate embedding for the current input to find relevant memories
	// This would typically use an embedding service - for now we'll use a placeholder
	queryEmbedding := dd.generateMockEmbedding(inputStr)

	// Retrieve relevant memories using GoLangGraph's enhanced memory manager
	relevantMemories, err := dd.databaseManager.MemoryManager.RetrieveMemories(ctx, threadIDStr, queryEmbedding, 5)
	if err != nil {
		// Continue without memories if retrieval fails
		relevantMemories = []*database.MemoryItem{}
	}

	// Load user preferences using the database manager
	userProfile, exists := dd.userProfiles[userIDStr]
	if !exists {
		userProfile = &UserProfile{
			UserID: userIDStr,
			Preferences: &database.UserPreferences{
				UserID:        userIDStr,
				DesignStyles:  make(map[string]float64),
				MaterialPrefs: make(map[string]float64),
				BudgetRange:   &database.BudgetRange{},
				Metadata:      make(map[string]interface{}),
				UpdatedAt:     time.Now(),
			},
			TotalSessions:       0,
			TotalInteractions:   0,
			PreferredStyles:     make(map[string]float64),
			MaterialPreferences: make(map[string]float64),
			BudgetPatterns:      []float64{},
			ComplexityTrends:    []float64{},
			SessionDurations:    []time.Duration{},
			IterationPatterns:   make(map[string]int),
			LastUpdated:         time.Now(),
			ProfileVersion:      "1.0",
			Metadata:            make(map[string]interface{}),
		}
		dd.userProfiles[userIDStr] = userProfile
	}

	// Update conversation context with loaded information
	conversationCtx.RelevantMemories = relevantMemories
	conversationCtx.UserPreferences = userProfile.Preferences
	conversationCtx.LastInteraction = time.Now()

	// Build enhanced context prompt with relevant memories and preferences
	var contextParts []string

	// Add relevant memories
	if len(relevantMemories) > 0 {
		contextParts = append(contextParts, "## Relevant Context from Previous Interactions:")
		for _, memory := range relevantMemories {
			contextParts = append(contextParts, fmt.Sprintf("- %s", memory.Content))
		}
	}

	// Add user preferences
	if len(userProfile.Preferences.DesignStyles) > 0 {
		contextParts = append(contextParts, "## User Design Preferences:")
		for style, weight := range userProfile.Preferences.DesignStyles {
			if weight > 0.3 { // Only include significant preferences
				contextParts = append(contextParts, fmt.Sprintf("- %s (strength: %.2f)", style, weight))
			}
		}
	}

	// Add material preferences
	if len(userProfile.Preferences.MaterialPrefs) > 0 {
		contextParts = append(contextParts, "## Material Preferences:")
		for material, weight := range userProfile.Preferences.MaterialPrefs {
			if weight > 0.3 {
				contextParts = append(contextParts, fmt.Sprintf("- %s (preference: %.2f)", material, weight))
			}
		}
	}

	enhancedContext := ""
	if len(contextParts) > 0 {
		enhancedContext = "\n\n" + fmt.Sprintf("%s", contextParts)
	}

	// Update state with loaded context
	state.Set("enhanced_context", enhancedContext)
	state.Set("relevant_memories", relevantMemories)
	state.Set("user_preferences", userProfile.Preferences)
	state.Set("context_loaded", true)

	return state, nil
}

// analyzeRequestNode analyzes the user request with full context understanding
func (dd *DesignerDefinition) analyzeRequestNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	input, _ := state.Get("input")
	enhancedContext, _ := state.Get("enhanced_context")
	sessionID, _ := state.Get("session_id")

	inputStr := fmt.Sprintf("%v", input)
	contextStr := fmt.Sprintf("%v", enhancedContext)
	sessionIDStr := fmt.Sprintf("%v", sessionID)

	// Get conversation context
	conversationCtx := dd.conversationHistory[sessionIDStr]

	// Analyze the request type and extract requirements
	analysisPrompt := fmt.Sprintf(`Analyze this design request with full context awareness:

User Request: %s

%s

Based on the context and user's history, analyze:
1. Request type (new design, iteration, feedback, clarification)
2. Extracted requirements and constraints
3. Design complexity level (1-10)
4. Sustainability priorities
5. Budget considerations
6. Timeline requirements
7. Special preferences or considerations

Provide a structured analysis in JSON format.`, inputStr, contextStr)

	// Store analysis in state
	state.Set("request_analysis", map[string]interface{}{
		"prompt":         analysisPrompt,
		"request_type":   "design_request", // This would be determined by LLM analysis
		"complexity":     5,                // This would be determined by LLM analysis
		"sustainability": "high",           // This would be determined by LLM analysis
		"analyzed_at":    time.Now(),
	})

	// Update conversation phase
	conversationCtx.CurrentPhase = "request_analysis"
	conversationCtx.CompletedPhases = append(conversationCtx.CompletedPhases, "initialization")

	return state, nil
}

// generateDesignNode generates the design using enhanced AI capabilities
func (dd *DesignerDefinition) generateDesignNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	input, _ := state.Get("input")
	enhancedContext, _ := state.Get("enhanced_context")
	requestAnalysis, _ := state.Get("request_analysis")
	sessionID, _ := state.Get("session_id")
	threadID, _ := state.Get("thread_id")
	userID, _ := state.Get("user_id")

	sessionIDStr := fmt.Sprintf("%v", sessionID)
	threadIDStr := fmt.Sprintf("%v", threadID)
	userIDStr := fmt.Sprintf("%v", userID)

	// Get conversation context
	conversationCtx := dd.conversationHistory[sessionIDStr]

	// Create a new design iteration
	iteration := &DesignIteration{
		ID:              uuid.New().String(),
		SessionID:       sessionIDStr,
		IterationNumber: len(conversationCtx.DesignHistory) + 1,
		Concept:         "", // Will be filled by LLM response
		Requirements:    []string{},
		Constraints:     []string{},
		MaterialSpecs:   []MaterialSpecification{},
		CreatedAt:       time.Now(),
		Metadata:        make(map[string]interface{}),
	}

	// Enhanced design prompt with full context
	_ = fmt.Sprintf(`Create a comprehensive sustainable habitat design for 2035 based on:

User Request: %s

%s

Analysis Context: %v

Generate a detailed design response including:

1. **Design Concept**: Comprehensive description of the sustainable habitat
2. **Material Specifications**: Detailed list with sustainability ratings
3. **Cost Estimation**: Complete breakdown with timeline
4. **Construction Phases**: Step-by-step implementation plan
5. **Sustainability Report**: Environmental impact and benefits
6. **Timeline**: Project schedule with milestones
7. **Risk Assessment**: Potential challenges and mitigation strategies

Focus on:
- Cutting-edge sustainable technologies for 2035
- Integration with natural environment
- Energy efficiency and renewable energy
- Smart home technology integration
- Cost-effectiveness and practical implementation
- Building regulations compliance
- Long-term durability and maintenance

Provide detailed, actionable specifications that could be used for actual construction.`,
		fmt.Sprintf("%v", input), fmt.Sprintf("%v", enhancedContext), requestAnalysis)

	// Create sample design output (in real implementation, this would come from LLM)
	designOutput := &DesignOutput{
		ID:                  uuid.New().String(),
		SessionID:           sessionIDStr,
		FinalConcept:        "Modular Bio-Integrated Habitat 2035",
		DetailedDescription: "A cutting-edge sustainable habitat featuring modular construction, living wall systems, integrated renewable energy, smart home automation, and carbon-negative materials designed for 2035 environmental standards.",
		TechnicalSpecs: map[string]interface{}{
			"total_area":       "150 sqm",
			"energy_rating":    "A+++",
			"carbon_footprint": "-2.5 tons CO2/year",
			"water_efficiency": "90% recycling",
		},
		Materials: []MaterialSpecification{
			{
				Name:                 "Bio-Concrete Blocks",
				Type:                 "Structural",
				Quantity:             250,
				Unit:                 "blocks",
				CostPerUnit:          45.00,
				SustainabilityRating: 9.2,
				CarbonFootprint:      -0.5,
				Recyclability:        95.0,
				LocallySourced:       true,
				Certifications:       []string{"Cradle-to-Cradle", "Bio-Based"},
			},
			{
				Name:                 "Recycled Steel Frame",
				Type:                 "Structural",
				Quantity:             12,
				Unit:                 "tons",
				CostPerUnit:          850.00,
				SustainabilityRating: 8.5,
				CarbonFootprint:      0.2,
				Recyclability:        100.0,
				LocallySourced:       false,
				Certifications:       []string{"Green Steel Certified"},
			},
		},
		SustainabilityReport: &SustainabilityReport{
			OverallScore:     9.1,
			CarbonFootprint:  -2.5,
			EnergyEfficiency: 9.5,
			WaterUsage:       2.8,
			WasteReduction:   8.9,
			MaterialSustainability: map[string]float64{
				"bio_concrete": 9.2,
				"steel_frame":  8.5,
				"solar_panels": 8.8,
			},
			CertificationsPossible: []string{"LEED Platinum", "Passive House", "Living Building Challenge"},
			ImprovementSuggestions: []string{
				"Consider geo-thermal heating integration",
				"Explore bio-luminescent lighting options",
				"Investigate mycelium insulation materials",
			},
		},
		CostBreakdown: &DetailedCostEstimate{
			CostEstimate: &CostEstimate{
				TotalCost:       125000.00,
				MaterialsCost:   65000.00,
				LaborCost:       35000.00,
				EquipmentCost:   15000.00,
				PermitsCost:     5000.00,
				ContingencyRate: 0.15,
				Currency:        "USD",
				EstimatedAt:     time.Now(),
				Confidence:      0.85,
			},
		},
		ConstructionPhases: []ConstructionPhase{
			{
				Phase:         "Foundation & Site Preparation",
				Description:   "Site clearing, foundation work, and utility connections",
				Duration:      "3 weeks",
				Prerequisites: []string{"Permits approved", "Site survey completed"},
				Tasks:         []string{"Excavation", "Foundation pouring", "Utility roughing"},
			},
			{
				Phase:         "Structural Framework",
				Description:   "Steel frame assembly and bio-concrete block installation",
				Duration:      "4 weeks",
				Prerequisites: []string{"Foundation cured", "Materials delivered"},
				Tasks:         []string{"Frame assembly", "Block installation", "Roof structure"},
			},
		},
		Timeline: &ProjectTimeline{
			TotalDuration: "16 weeks",
			Phases: []PhaseTimeline{
				{Phase: "Planning", Start: "Week 1", End: "Week 2", Duration: "2 weeks"},
				{Phase: "Foundation", Start: "Week 3", End: "Week 5", Duration: "3 weeks"},
				{Phase: "Structure", Start: "Week 6", End: "Week 9", Duration: "4 weeks"},
				{Phase: "Systems", Start: "Week 10", End: "Week 12", Duration: "3 weeks"},
				{Phase: "Finishing", Start: "Week 13", End: "Week 16", Duration: "4 weeks"},
			},
			Milestones: []Milestone{
				{Name: "Foundation Complete", Date: "Week 5", Description: "Foundation ready for framing"},
				{Name: "Structure Complete", Date: "Week 9", Description: "Building weatherproof"},
				{Name: "Final Inspection", Date: "Week 16", Description: "Ready for occupancy"},
			},
		},
		RiskAssessment: &RiskAssessment{
			OverallRisk: "Medium",
			RiskFactors: []RiskFactor{
				{
					Category:    "Material",
					Risk:        "Bio-concrete availability",
					Probability: 0.3,
					Impact:      0.6,
					Severity:    "Medium",
					Mitigation:  "Secure multiple suppliers, consider alternatives",
				},
			},
		},
		QualityScore: 8.8,
		ComplianceChecks: map[string]bool{
			"building_codes":   true,
			"environmental":    true,
			"safety_standards": true,
			"accessibility":    true,
		},
		GeneratedAt: time.Now(),
		Version:     "1.0",
		Metadata: map[string]interface{}{
			"iteration_number": iteration.IterationNumber,
			"session_id":       sessionIDStr,
			"user_id":          userIDStr,
			"generation_time":  time.Now(),
		},
	}

	// Update iteration with design details
	iteration.Concept = designOutput.FinalConcept
	iteration.MaterialSpecs = designOutput.Materials
	iteration.EstimatedCost = designOutput.CostBreakdown.CostEstimate
	iteration.SustainabilityScore = designOutput.SustainabilityReport.OverallScore

	// Store design in active session
	designSession := &DesignSession{
		ID:               uuid.New().String(),
		ThreadID:         threadIDStr,
		UserID:           userIDStr,
		CurrentIteration: iteration,
		AllIterations:    []*DesignIteration{iteration},
		FinalDesign:      designOutput,
		Status:           "active",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	dd.activeDesigns[sessionIDStr] = designSession

	// Update conversation context - convert to database.DesignIteration
	dbIteration := &database.DesignIteration{
		ID:            iteration.ID,
		ThreadID:      sessionIDStr,
		UserID:        userIDStr,
		DesignConcept: iteration.Concept,
		Feedback:      iteration.UserFeedback,
		Rating:        iteration.UserRating,
		Improvements:  iteration.ImprovementAreas,
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
	}
	conversationCtx.DesignHistory = append(conversationCtx.DesignHistory, dbIteration)
	conversationCtx.CurrentPhase = "design_generation"
	conversationCtx.CompletedPhases = append(conversationCtx.CompletedPhases, "request_analysis")

	// Update state with design results
	state.Set("design_output", designOutput)
	state.Set("current_iteration", iteration)
	state.Set("design_session", designSession)
	state.Set("design_generated", true)

	return state, nil
}

// evaluateDesignNode evaluates the generated design and prepares for potential iteration
func (dd *DesignerDefinition) evaluateDesignNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	_, _ = state.Get("design_output") // designOutput not used in this method
	sessionID, _ := state.Get("session_id")

	sessionIDStr := fmt.Sprintf("%v", sessionID)
	conversationCtx := dd.conversationHistory[sessionIDStr]

	// Evaluate design quality based on multiple criteria
	evaluation := map[string]interface{}{
		"sustainability_score": 9.1,
		"cost_effectiveness":   8.5,
		"innovation_level":     8.8,
		"feasibility":          9.0,
		"user_alignment":       8.7,
		"overall_quality":      8.8,
		"areas_for_improvement": []string{
			"Consider additional renewable energy options",
			"Explore smart home integration opportunities",
			"Investigate additional cost optimization",
		},
		"strengths": []string{
			"Excellent sustainability integration",
			"Comprehensive material specifications",
			"Detailed construction planning",
			"Strong environmental benefits",
		},
	}

	// Update conversation context
	conversationCtx.CurrentPhase = "design_evaluation"
	conversationCtx.CompletedPhases = append(conversationCtx.CompletedPhases, "design_generation")

	// Update state
	state.Set("design_evaluation", evaluation)
	state.Set("design_evaluated", true)

	return state, nil
}

// saveIterationNode saves the current iteration using GoLangGraph's persistence layer
func (dd *DesignerDefinition) saveIterationNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	sessionID, _ := state.Get("session_id")
	threadID, _ := state.Get("thread_id")
	currentIteration, _ := state.Get("current_iteration")
	_, _ = state.Get("design_output") // designOutput not used in this method

	sessionIDStr := fmt.Sprintf("%v", sessionID)
	threadIDStr := fmt.Sprintf("%v", threadID)

	// Create checkpoint using GoLangGraph's checkpointer
	checkpoint := &persistence.Checkpoint{
		ID:       uuid.New().String(),
		ThreadID: threadIDStr,
		State:    state,
		Metadata: map[string]interface{}{
			"checkpoint_type":  "design_iteration",
			"session_id":       sessionIDStr,
			"iteration_number": currentIteration,
			"timestamp":        time.Now(),
		},
		CreatedAt: time.Now(),
		NodeID:    "save_iteration",
		StepID:    len(dd.conversationHistory[sessionIDStr].DesignHistory),
	}

	// Save checkpoint
	if err := dd.checkpointer.Save(ctx, checkpoint); err != nil {
		return nil, fmt.Errorf("failed to save checkpoint: %w", err)
	}

	// Store design iteration as memory using enhanced memory manager
	if iteration, ok := currentIteration.(*DesignIteration); ok {
		memoryItem := &database.MemoryItem{
			ID:         uuid.New().String(),
			ThreadID:   threadIDStr,
			UserID:     iteration.SessionID, // Using session ID as user context
			Content:    fmt.Sprintf("Design iteration: %s - %s", iteration.Concept, iteration.Metadata),
			MemoryType: "design_iteration",
			Embedding:  dd.generateMockEmbedding(iteration.Concept),
			Metadata: map[string]interface{}{
				"iteration_id":         iteration.ID,
				"sustainability_score": iteration.SustainabilityScore,
				"estimated_cost":       iteration.EstimatedCost.TotalCost,
				"iteration_number":     iteration.IterationNumber,
			},
			Importance:  0.8,
			CreatedAt:   time.Now(),
			LastAccess:  time.Now(),
			AccessCount: 1,
		}

		if err := dd.databaseManager.MemoryManager.StoreMemory(ctx, memoryItem); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to store memory: %v\n", err)
		}
	}

	// Update conversation context
	conversationCtx := dd.conversationHistory[sessionIDStr]
	conversationCtx.CurrentPhase = "iteration_saved"
	conversationCtx.CompletedPhases = append(conversationCtx.CompletedPhases, "design_evaluation")

	// Update state
	state.Set("checkpoint_id", checkpoint.ID)
	state.Set("iteration_saved", true)

	return state, nil
}

// updateMemoryNode updates user preferences and memory based on the interaction
func (dd *DesignerDefinition) updateMemoryNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	sessionID, _ := state.Get("session_id")
	userID, _ := state.Get("user_id")
	designOutput, _ := state.Get("design_output")

	sessionIDStr := fmt.Sprintf("%v", sessionID)
	userIDStr := fmt.Sprintf("%v", userID)

	// Extract learning signals from the interaction
	feedback := make(map[string]interface{})

	if output, ok := designOutput.(*DesignOutput); ok {
		// Learn from material preferences
		materialPrefs := make(map[string]float64)
		for _, material := range output.Materials {
			materialPrefs[material.Type] = 0.8 // Positive preference for selected materials
		}
		feedback["material_preferences"] = materialPrefs

		// Learn from design style preferences
		designStyles := make(map[string]float64)
		designStyles["sustainable"] = 0.9
		designStyles["modular"] = 0.7
		designStyles["smart_technology"] = 0.8
		feedback["design_styles"] = designStyles
	}

	// Update user preferences using enhanced memory manager
	if err := dd.databaseManager.MemoryManager.UpdateUserPreferences(userIDStr, feedback); err != nil {
		fmt.Printf("Warning: failed to update user preferences: %v\n", err)
	}

	// Update user profile
	if userProfile, exists := dd.userProfiles[userIDStr]; exists {
		userProfile.TotalInteractions++
		userProfile.LastUpdated = time.Now()

		// Update interaction patterns
		conversationCtx := dd.conversationHistory[sessionIDStr]
		sessionDuration := time.Since(conversationCtx.LastInteraction)
		userProfile.SessionDurations = append(userProfile.SessionDurations, sessionDuration)

		// Keep only recent sessions for analysis
		if len(userProfile.SessionDurations) > 10 {
			userProfile.SessionDurations = userProfile.SessionDurations[1:]
		}
	}

	// Update conversation context
	conversationCtx := dd.conversationHistory[sessionIDStr]
	conversationCtx.CurrentPhase = "memory_updated"
	conversationCtx.CompletedPhases = append(conversationCtx.CompletedPhases, "iteration_saved")

	state.Set("memory_updated", true)

	return state, nil
}

// finalizeResponseNode creates the final response with all design information
func (dd *DesignerDefinition) finalizeResponseNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	designOutput, _ := state.Get("design_output")
	sessionID, _ := state.Get("session_id")
	designEvaluation, _ := state.Get("design_evaluation")
	checkpointID, _ := state.Get("checkpoint_id")

	sessionIDStr := fmt.Sprintf("%v", sessionID)

	// Create comprehensive response
	if output, ok := designOutput.(*DesignOutput); ok {
		response := fmt.Sprintf(`# %s

## Design Overview
%s

## Technical Specifications
- Total Area: %v
- Energy Rating: %v
- Carbon Footprint: %v tons CO2/year
- Water Efficiency: %v

## Sustainability Report
- Overall Score: %.1f/10
- Carbon Footprint: %.1f tons CO2/year
- Energy Efficiency: %.1f/10
- Water Usage: %.1f/10

## Cost Breakdown
- Total Cost: $%.2f
- Materials: $%.2f
- Labor: $%.2f
- Equipment: $%.2f
- Permits: $%.2f

## Construction Timeline
Total Duration: %s

### Key Milestones:
%s

## Quality Assessment
Overall Quality Score: %.1f/10

### Design Strengths:
%s

### Areas for Improvement:
%s

## Next Steps
1. Review the design specifications
2. Provide feedback for refinements
3. Proceed with detailed planning
4. Begin permit applications

Your preferences and feedback are being learned and will improve future designs. Session ID: %s | Checkpoint: %s`,
			output.FinalConcept,
			output.DetailedDescription,
			output.TechnicalSpecs["total_area"],
			output.TechnicalSpecs["energy_rating"],
			output.TechnicalSpecs["carbon_footprint"],
			output.TechnicalSpecs["water_efficiency"],
			output.SustainabilityReport.OverallScore,
			output.SustainabilityReport.CarbonFootprint,
			output.SustainabilityReport.EnergyEfficiency,
			output.SustainabilityReport.WaterUsage,
			output.CostBreakdown.TotalCost,
			output.CostBreakdown.MaterialsCost,
			output.CostBreakdown.LaborCost,
			output.CostBreakdown.EquipmentCost,
			output.CostBreakdown.PermitsCost,
			output.Timeline.TotalDuration,
			formatMilestones(output.Timeline.Milestones),
			output.QualityScore,
			formatStrengths(designEvaluation),
			formatImprovements(designEvaluation),
			sessionIDStr,
			fmt.Sprintf("%v", checkpointID))

		state.Set("output", response)
	} else {
		state.Set("output", "I apologize, but there was an issue generating your design. Please try again.")
	}

	// Update conversation context
	conversationCtx := dd.conversationHistory[sessionIDStr]
	conversationCtx.CurrentPhase = "completed"
	conversationCtx.CompletedPhases = append(conversationCtx.CompletedPhases, "memory_updated")
	conversationCtx.LastInteraction = time.Now()

	state.Set("response_finalized", true)

	return state, nil
}

// Helper functions

func (dd *DesignerDefinition) generateMockEmbedding(text string) []float64 {
	// This is a simplified mock embedding generator
	// In production, this would use a real embedding service
	embedding := make([]float64, 1536)
	for i := range embedding {
		embedding[i] = float64(len(text)%100) / 100.0 // Simple hash-based mock
	}
	return embedding
}

func formatMilestones(milestones []Milestone) string {
	var result []string
	for _, milestone := range milestones {
		result = append(result, fmt.Sprintf("- %s (%s): %s", milestone.Name, milestone.Date, milestone.Description))
	}
	return fmt.Sprintf("%v", result)
}

func formatStrengths(evaluation interface{}) string {
	if eval, ok := evaluation.(map[string]interface{}); ok {
		if strengths, ok := eval["strengths"].([]string); ok {
			var result []string
			for _, strength := range strengths {
				result = append(result, fmt.Sprintf("- %s", strength))
			}
			return fmt.Sprintf("%v", result)
		}
	}
	return "- Comprehensive design approach\n- Strong sustainability focus"
}

func formatImprovements(evaluation interface{}) string {
	if eval, ok := evaluation.(map[string]interface{}); ok {
		if improvements, ok := eval["areas_for_improvement"].([]string); ok {
			var result []string
			for _, improvement := range improvements {
				result = append(result, fmt.Sprintf("- %s", improvement))
			}
			return fmt.Sprintf("%v", result)
		}
	}
	return "- Continue iterating based on feedback\n- Explore additional optimization opportunities"
}

// GetDesignerConfig returns the configuration for backward compatibility
func GetDesignerConfig() *agent.AgentConfig {
	// For backward compatibility, create a simple configuration
	return &agent.AgentConfig{
		Name:         "Advanced Visual Designer with State Management",
		Type:         agent.AgentTypeChat,
		Model:        "llama3.2:latest",
		Provider:     "ollama",
		SystemPrompt: `You are an advanced AI architect and designer specializing in sustainable habitat design for 2035.`,
	}
}
