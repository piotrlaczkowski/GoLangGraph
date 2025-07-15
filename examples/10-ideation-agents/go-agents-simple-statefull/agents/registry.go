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
	"fmt"
	"go-agents-simple-statefull/database"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
)

// StatefulAgentRegistry manages all enhanced stateful agents
type StatefulAgentRegistry struct {
	*agent.AgentRegistry
	registeredAgents map[string]AgentInfo
}

// AgentInfo contains metadata about registered agents
type AgentInfo struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	Description      string    `json:"description"`
	Tags             []string  `json:"tags"`
	StatefulFeatures []string  `json:"stateful_features"`
	RegisteredAt     time.Time `json:"registered_at"`
	Capabilities     []string  `json:"capabilities"`
	Version          string    `json:"version"`
}

// NewStatefulAgentRegistry creates a new enhanced agent registry
func NewStatefulAgentRegistry() *StatefulAgentRegistry {
	return &StatefulAgentRegistry{
		AgentRegistry:    agent.NewAgentRegistry(),
		registeredAgents: make(map[string]AgentInfo),
	}
}

// RegisterAllAgents registers all enhanced stateful ideation agents
func (sar *StatefulAgentRegistry) RegisterAllAgents(databaseManager *database.DatabaseManager) error {
	// Register Designer agent with database manager
	designerDef := NewDesignerDefinition(databaseManager)
	if err := sar.AgentRegistry.RegisterDefinition("designer", designerDef); err != nil {
		return fmt.Errorf("failed to register designer agent: %w", err)
	}
	sar.registeredAgents["designer"] = AgentInfo{
		ID:          "designer",
		Name:        "Enhanced Visual Designer",
		Type:        "stateful_chat",
		Description: "Creates sustainable habitat designs with session continuity and user preference learning",
		Tags:        []string{"design", "architecture", "sustainability", "creativity", "stateful", "persistent"},
		StatefulFeatures: []string{
			"Session Management",
			"Design History Tracking",
			"User Preference Learning",
			"Iterative Design Process",
			"Context-Aware Recommendations",
		},
		RegisteredAt: time.Now(),
		Capabilities: []string{
			"Sustainable Habitat Design",
			"3D Visualization Concepts",
			"Material Selection Guidance",
			"Energy Efficiency Planning",
			"Cost Estimation",
			"Timeline Planning",
		},
		Version: "2.0.0",
	}

	// Register Interviewer agent
	interviewerDef := NewInterviewerDefinition()
	if err := sar.AgentRegistry.RegisterDefinition("interviewer", interviewerDef); err != nil {
		return fmt.Errorf("failed to register interviewer agent: %w", err)
	}
	sar.registeredAgents["interviewer"] = AgentInfo{
		ID:          "interviewer",
		Name:        "Smart French Interviewer",
		Type:        "stateful_chat",
		Description: "Conducts intelligent French conversations to gather habitat requirements with session continuity",
		Tags:        []string{"interview", "requirements", "conversation", "french", "stateful", "persistent"},
		StatefulFeatures: []string{
			"Session Persistence",
			"Conversation Context Tracking",
			"Interview Progress Monitoring",
			"User Profile Building",
			"Requirement Extraction",
		},
		RegisteredAt: time.Now(),
		Capabilities: []string{
			"French Language Interviews",
			"Requirement Gathering",
			"User Profile Analysis",
			"Context-Aware Questioning",
			"Progress Tracking",
			"Insight Extraction",
		},
		Version: "2.0.0",
	}

	// Register Highlighter agent
	highlighterDef := NewHighlighterDefinition()
	if err := sar.AgentRegistry.RegisterDefinition("highlighter", highlighterDef); err != nil {
		return fmt.Errorf("failed to register highlighter agent: %w", err)
	}
	sar.registeredAgents["highlighter"] = AgentInfo{
		ID:          "highlighter",
		Name:        "Insight Highlighter",
		Type:        "stateful_chat",
		Description: "Extracts comprehensive insights and themes from conversations with sentiment analysis",
		Tags:        []string{"analysis", "insights", "themes", "sentiment", "recommendations", "stateful"},
		StatefulFeatures: []string{
			"Analysis History Tracking",
			"Cumulative Insight Building",
			"Theme Evolution Monitoring",
			"Quality Score Calculation",
			"Progressive Analysis Improvement",
		},
		RegisteredAt: time.Now(),
		Capabilities: []string{
			"Conversation Analysis",
			"Theme Identification",
			"Sentiment Analysis",
			"Insight Extraction",
			"Actionable Recommendations",
			"Complexity Assessment",
		},
		Version: "2.0.0",
	}

	// Register Storymaker agent
	storymakerDef := NewStorymakerDefinition()
	if err := sar.AgentRegistry.RegisterDefinition("storymaker", storymakerDef); err != nil {
		return fmt.Errorf("failed to register storymaker agent: %w", err)
	}
	sar.registeredAgents["storymaker"] = AgentInfo{
		ID:          "storymaker",
		Name:        "Story Creator",
		Type:        "stateful_chat",
		Description: "Creates engaging narratives about future sustainable habitat scenarios with educational value",
		Tags:        []string{"storytelling", "narrative", "futures", "sustainability", "education", "stateful"},
		StatefulFeatures: []string{
			"Story History Tracking",
			"Character Bank Management",
			"Theme Preference Learning",
			"Genre Adaptation",
			"Quality Evolution Tracking",
		},
		RegisteredAt: time.Now(),
		Capabilities: []string{
			"Narrative Creation",
			"Character Development",
			"Sustainability Education",
			"Multi-Genre Storytelling",
			"Educational Value Assessment",
			"Reading Time Calculation",
		},
		Version: "2.0.0",
	}

	return nil
}

// GetAllAgentConfigs returns all agent configurations for backward compatibility
func (sar *StatefulAgentRegistry) GetAllAgentConfigs() map[string]*agent.AgentConfig {
	return map[string]*agent.AgentConfig{
		"designer":    GetDesignerConfig(),
		"interviewer": GetInterviewerConfig(),
		"highlighter": GetHighlighterConfig(),
		"storymaker":  GetStorymakerConfig(),
	}
}

// GetAgentInfo returns detailed information about a specific agent
func (sar *StatefulAgentRegistry) GetAgentInfo(agentID string) (AgentInfo, bool) {
	info, exists := sar.registeredAgents[agentID]
	return info, exists
}

// GetAllAgentInfo returns information about all registered agents
func (sar *StatefulAgentRegistry) GetAllAgentInfo() map[string]AgentInfo {
	result := make(map[string]AgentInfo)
	for id, info := range sar.registeredAgents {
		result[id] = info
	}
	return result
}

// GetAgentsByTag returns agents that have a specific tag
func (sar *StatefulAgentRegistry) GetAgentsByTag(tag string) []AgentInfo {
	var result []AgentInfo
	for _, info := range sar.registeredAgents {
		for _, agentTag := range info.Tags {
			if agentTag == tag {
				result = append(result, info)
				break
			}
		}
	}
	return result
}

// GetStatefulAgents returns only agents with stateful capabilities
func (sar *StatefulAgentRegistry) GetStatefulAgents() []AgentInfo {
	var result []AgentInfo
	for _, info := range sar.registeredAgents {
		if len(info.StatefulFeatures) > 0 {
			result = append(result, info)
		}
	}
	return result
}

// GetAgentCapabilities returns all unique capabilities across all agents
func (sar *StatefulAgentRegistry) GetAgentCapabilities() []string {
	capabilitySet := make(map[string]bool)
	for _, info := range sar.registeredAgents {
		for _, capability := range info.Capabilities {
			capabilitySet[capability] = true
		}
	}

	var capabilities []string
	for capability := range capabilitySet {
		capabilities = append(capabilities, capability)
	}
	return capabilities
}

// GetStatefulFeatures returns all unique stateful features across all agents
func (sar *StatefulAgentRegistry) GetStatefulFeatures() []string {
	featureSet := make(map[string]bool)
	for _, info := range sar.registeredAgents {
		for _, feature := range info.StatefulFeatures {
			featureSet[feature] = true
		}
	}

	var features []string
	for feature := range featureSet {
		features = append(features, feature)
	}
	return features
}

// GenerateAgentSummary generates a comprehensive summary of all registered agents
func (sar *StatefulAgentRegistry) GenerateAgentSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"total_agents":       len(sar.registeredAgents),
		"stateful_agents":    len(sar.GetStatefulAgents()),
		"agent_types":        sar.getAgentTypes(),
		"total_capabilities": len(sar.GetAgentCapabilities()),
		"stateful_features":  len(sar.GetStatefulFeatures()),
		"agents":             sar.registeredAgents,
		"capabilities":       sar.GetAgentCapabilities(),
		"features":           sar.GetStatefulFeatures(),
		"generated_at":       time.Now(),
	}
	return summary
}

// getAgentTypes returns unique agent types
func (sar *StatefulAgentRegistry) getAgentTypes() []string {
	typeSet := make(map[string]bool)
	for _, info := range sar.registeredAgents {
		typeSet[info.Type] = true
	}

	var types []string
	for agentType := range typeSet {
		types = append(types, agentType)
	}
	return types
}

// ValidateRegistration checks if all required agents are properly registered
func (sar *StatefulAgentRegistry) ValidateRegistration() error {
	requiredAgents := []string{"designer", "interviewer", "highlighter", "storymaker"}

	for _, required := range requiredAgents {
		if _, exists := sar.registeredAgents[required]; !exists {
			return fmt.Errorf("required agent %s is not registered", required)
		}

		// Check if agent is properly registered in the base registry
		if _, exists := sar.AgentRegistry.GetDefinition(required); !exists {
			return fmt.Errorf("agent %s is not properly registered in base registry", required)
		}
	}

	return nil
}

// Global registry instance
var GlobalStatefulRegistry *StatefulAgentRegistry

// InitializeGlobalRegistry initializes the global stateful agent registry
func InitializeGlobalRegistry() error {
	GlobalStatefulRegistry = NewStatefulAgentRegistry()
	// Note: Agent registration is deferred until database manager is available
	return nil
}

// GetGlobalRegistry returns the global stateful agent registry
func GetGlobalRegistry() *StatefulAgentRegistry {
	if GlobalStatefulRegistry == nil {
		if err := InitializeGlobalRegistry(); err != nil {
			panic(fmt.Sprintf("Failed to initialize global registry: %v", err))
		}
	}
	return GlobalStatefulRegistry
}

// Legacy functions for backward compatibility

// RegisterAllAgents registers all ideation agents with the provided registry (legacy)
func RegisterAllAgents(registry *agent.AgentRegistry, databaseManager *database.DatabaseManager) error {
	statefulRegistry := &StatefulAgentRegistry{
		AgentRegistry:    registry,
		registeredAgents: make(map[string]AgentInfo),
	}
	return statefulRegistry.RegisterAllAgents(databaseManager)
}

// GetAllAgentConfigs returns all agent configurations for backward compatibility (legacy)
func GetAllAgentConfigs() map[string]*agent.AgentConfig {
	return map[string]*agent.AgentConfig{
		"designer":    GetDesignerConfig(),
		"interviewer": GetInterviewerConfig(),
		"highlighter": GetHighlighterConfig(),
		"storymaker":  GetStorymakerConfig(),
	}
}
