// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
	"github.com/stretchr/testify/assert"
)

// Test Agent Definition
type TestAgentDefinition struct {
	*BaseAgentDefinition
	customSetupCalled bool
}

func NewTestAgentDefinition(name string, agentType AgentType) *TestAgentDefinition {
	config := &AgentConfig{
		Name:         name,
		Type:         agentType,
		Model:        "gpt-3.5-turbo",
		Provider:     "openai",
		SystemPrompt: "You are a test agent",
		Temperature:  0.7,
		MaxTokens:    1000,
		Tools:        []string{"test-tool"},
	}

	return &TestAgentDefinition{
		BaseAgentDefinition: NewBaseAgentDefinition(config),
		customSetupCalled:   false,
	}
}

func (tad *TestAgentDefinition) Initialize(llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) error {
	tad.customSetupCalled = true
	return tad.BaseAgentDefinition.Initialize(llmManager, toolRegistry)
}

func (tad *TestAgentDefinition) CreateAgent() (*Agent, error) {
	agent, err := tad.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Add some custom metadata
	tad.SetMetadata("custom_setup_called", tad.customSetupCalled)
	tad.SetMetadata("test_flag", true)

	return agent, nil
}

// Test Advanced Agent Definition
type TestAdvancedAgentDefinition struct {
	*AdvancedAgentDefinition
	graphBuilt bool
}

func NewTestAdvancedAgentDefinition(name string) *TestAdvancedAgentDefinition {
	config := &AgentConfig{
		Name:         name,
		Type:         AgentTypeReAct,
		Model:        "gpt-4",
		Provider:     "openai",
		SystemPrompt: "You are an advanced test agent",
		Temperature:  0.5,
		MaxTokens:    2000,
		Tools:        []string{"advanced-tool"},
	}

	def := &TestAdvancedAgentDefinition{
		AdvancedAgentDefinition: NewAdvancedAgentDefinition(config),
		graphBuilt:              false,
	}

	// Set a custom graph builder to ensure BuildGraph is called
	def.AdvancedAgentDefinition.WithGraphBuilder(func() (*core.Graph, error) {
		def.graphBuilt = true
		// Return a simple graph instead of recursing
		graph := core.NewGraph("test-graph")
		graph.AddNode("test", "Test Node", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
			return state, nil
		})
		graph.SetStartNode("test")
		graph.AddEndNode("test")
		return graph, nil
	})

	return def
}

func (taad *TestAdvancedAgentDefinition) BuildGraph() (*core.Graph, error) {
	taad.graphBuilt = true
	return taad.AdvancedAgentDefinition.BuildGraph()
}

func (taad *TestAdvancedAgentDefinition) GetCustomTools() []tools.Tool {
	return []tools.Tool{
		&TestTool{name: "custom-test-tool"},
	}
}

// Test Tool Implementation
type TestTool struct {
	name string
}

func (tt *TestTool) GetName() string {
	return tt.name
}

func (tt *TestTool) GetDescription() string {
	return "A test tool for testing"
}

func (tt *TestTool) GetDefinition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.Function{
			Name:        tt.GetName(),
			Description: tt.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]interface{}{
						"type":        "string",
						"description": "Test input",
					},
				},
				"required": []string{"input"},
			},
		},
	}
}

func (tt *TestTool) Execute(ctx context.Context, args string) (string, error) {
	var params struct {
		Input string `json:"input"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	return fmt.Sprintf("Tool executed with input: %s", params.Input), nil
}

func (tt *TestTool) Validate(args string) error {
	var params struct {
		Input string `json:"input"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if params.Input == "" {
		return fmt.Errorf("input is required")
	}

	return nil
}

func (tt *TestTool) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"name": tt.name,
	}
}

func (tt *TestTool) SetConfig(config map[string]interface{}) error {
	if name, ok := config["name"].(string); ok {
		tt.name = name
	}
	return nil
}

// Test Agent Definition Registry
func TestAgentRegistry(t *testing.T) {
	registry := NewAgentRegistry()

	// Test registering a definition
	testDef := NewTestAgentDefinition("test-agent", AgentTypeChat)
	err := registry.RegisterDefinition("test-agent", testDef)
	assert.NoError(t, err)

	// Test retrieving a definition
	retrievedDef, exists := registry.GetDefinition("test-agent")
	assert.True(t, exists)
	assert.Equal(t, testDef, retrievedDef)

	// Test registering duplicate definition
	err = registry.RegisterDefinition("test-agent", testDef)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test registering a factory
	factory := func() AgentDefinition {
		return NewTestAgentDefinition("factory-agent", AgentTypeReAct)
	}

	err = registry.RegisterFactory("factory-agent", factory)
	assert.NoError(t, err)

	// Test listing definitions and factories
	definitions := registry.ListDefinitions()
	factories := registry.ListFactories()

	assert.Contains(t, definitions, "test-agent")
	assert.Contains(t, factories, "factory-agent")

	// Test getting metadata
	metadata := registry.GetMetadata()
	assert.Contains(t, metadata, "test-agent")
	assert.Equal(t, "test-agent", metadata["test-agent"]["name"])
}

func TestAgentDefinitionBuilder(t *testing.T) {
	definition := NewAgentDefinitionBuilder().
		WithName("builder-agent").
		WithType(AgentTypeChat).
		WithModel("gpt-4").
		WithProvider("openai").
		WithSystemPrompt("You are a builder agent").
		WithTemperature(0.8).
		WithMaxTokens(1500).
		WithTools("tool1", "tool2").
		WithMetadata("version", "1.0").
		Build()

	config := definition.GetConfig()
	assert.Equal(t, "builder-agent", config.Name)
	assert.Equal(t, AgentTypeChat, config.Type)
	assert.Equal(t, "gpt-4", config.Model)
	assert.Equal(t, "openai", config.Provider)
	assert.Equal(t, 0.8, config.Temperature)
	assert.Equal(t, 1500, config.MaxTokens)
	assert.Equal(t, []string{"tool1", "tool2"}, config.Tools)

	metadata := definition.GetMetadata()
	assert.Equal(t, "1.0", metadata["version"])
}

func TestAdvancedAgentDefinitionCreation(t *testing.T) {
	// Create mock LLM manager and tool registry
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Create advanced agent definition
	advancedDef := NewTestAdvancedAgentDefinition("advanced-agent")

	// Initialize
	err := advancedDef.Initialize(llmManager, toolRegistry)
	assert.NoError(t, err)

	// Create agent
	agent, err := advancedDef.CreateAgent()
	assert.NoError(t, err)
	assert.NotNil(t, agent)

	// Check that custom graph building was called
	assert.True(t, advancedDef.graphBuilt)
}

func TestMultiAgentConfig(t *testing.T) {
	config := &MultiAgentConfig{
		Name:        "test-multi-agent",
		Version:     "1.0",
		Description: "Test multi-agent configuration",
		Agents: map[string]*AgentConfig{
			"agent1": {
				ID:           "agent1",
				Name:         "Agent 1",
				Type:         AgentTypeChat,
				Model:        "gpt-3.5-turbo",
				Provider:     "openai",
				SystemPrompt: "You are agent 1",
				Temperature:  0.7,
				MaxTokens:    1000,
				Tools:        []string{"tool1"},
			},
			"agent2": {
				ID:           "agent2",
				Name:         "Agent 2",
				Type:         AgentTypeReAct,
				Model:        "gpt-4",
				Provider:     "openai",
				SystemPrompt: "You are agent 2",
				Temperature:  0.5,
				MaxTokens:    2000,
				Tools:        []string{"tool2"},
			},
		},
		Routing: &RoutingConfig{
			Type: "path",
			Rules: []RoutingRule{
				{
					ID:       "rule1",
					Pattern:  "/agent1",
					AgentID:  "agent1",
					Method:   "POST",
					Priority: 1,
				},
				{
					ID:       "rule2",
					Pattern:  "/agent2",
					AgentID:  "agent2",
					Method:   "POST",
					Priority: 2,
				},
			},
			DefaultAgent: "agent1",
		},
	}

	// Test validation
	err := config.Validate()
	assert.NoError(t, err)

	// Test invalid config
	invalidConfig := &MultiAgentConfig{
		Name: "", // Empty name should cause validation error
	}

	err = invalidConfig.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestMultiAgentManager(t *testing.T) {
	// Create test configuration
	config := &MultiAgentConfig{
		Name:        "test-manager",
		Version:     "1.0",
		Description: "Test multi-agent manager",
		Agents: map[string]*AgentConfig{
			"chat-agent": {
				ID:           "chat-agent",
				Name:         "Chat Agent",
				Type:         AgentTypeChat,
				Model:        "gpt-3.5-turbo",
				Provider:     "openai",
				SystemPrompt: "You are a chat agent",
				Temperature:  0.7,
				MaxTokens:    1000,
				Tools:        []string{},
			},
			"react-agent": {
				ID:           "react-agent",
				Name:         "React Agent",
				Type:         AgentTypeReAct,
				Model:        "gpt-4",
				Provider:     "openai",
				SystemPrompt: "You are a react agent",
				Temperature:  0.5,
				MaxTokens:    2000,
				Tools:        []string{},
			},
		},
		Routing: &RoutingConfig{
			Type: "path",
			Rules: []RoutingRule{
				{
					ID:       "chat-rule",
					Pattern:  "/chat",
					AgentID:  "chat-agent",
					Method:   "POST",
					Priority: 1,
				},
				{
					ID:       "react-rule",
					Pattern:  "/react",
					AgentID:  "react-agent",
					Method:   "POST",
					Priority: 2,
				},
			},
			DefaultAgent: "chat-agent",
		},
		Deployment: &DeploymentConfig{
			Type:        "docker",
			Environment: "test",
			Replicas:    1,
		},
	}

	// Create managers
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Create multi-agent manager
	manager, err := NewMultiAgentManager(config, llmManager, toolRegistry)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test getting configuration
	retrievedConfig := manager.GetConfig()
	assert.Equal(t, config.Name, retrievedConfig.Name)

	// Test getting router
	router := manager.GetRouter()
	assert.NotNil(t, router)

	// Test getting metrics
	metrics := manager.GetMetrics()
	assert.NotNil(t, metrics)
	assert.Contains(t, metrics.AgentMetrics, "chat-agent")
	assert.Contains(t, metrics.AgentMetrics, "react-agent")

	// Test getting deployment state
	deploymentState := manager.GetDeploymentState()
	assert.NotNil(t, deploymentState)
	assert.Contains(t, deploymentState.AgentStates, "chat-agent")
	assert.Contains(t, deploymentState.AgentStates, "react-agent")
}

func TestMultiAgentManagerWithDefinitions(t *testing.T) {
	// Register test agent definitions
	registry := GetGlobalRegistry()

	testDef1 := NewTestAgentDefinition("definition-agent", AgentTypeChat)
	testDef2 := NewTestAgentDefinition("factory-agent", AgentTypeReAct)

	err := registry.RegisterDefinition("definition-agent", testDef1)
	assert.NoError(t, err)

	factory := func() AgentDefinition {
		return testDef2
	}
	err = registry.RegisterFactory("factory-agent", factory)
	assert.NoError(t, err)

	// Create configuration that includes both definition and config-based agents
	config := &MultiAgentConfig{
		Name:        "mixed-agent-manager",
		Version:     "1.0",
		Description: "Test mixed agent types",
		Agents: map[string]*AgentConfig{
			"definition-agent": {
				ID:           "definition-agent",
				Name:         "Definition Agent",
				Type:         AgentTypeChat,
				Model:        "gpt-3.5-turbo",
				Provider:     "openai",
				SystemPrompt: "You are a definition agent",
				Temperature:  0.7,
				MaxTokens:    1000,
				Tools:        []string{},
			},
			"factory-agent": {
				ID:           "factory-agent",
				Name:         "Factory Agent",
				Type:         AgentTypeReAct,
				Model:        "gpt-4",
				Provider:     "openai",
				SystemPrompt: "You are a factory agent",
				Temperature:  0.5,
				MaxTokens:    2000,
				Tools:        []string{},
			},
			"config-agent": {
				ID:           "config-agent",
				Name:         "Config Agent",
				Type:         AgentTypeTool,
				Model:        "gpt-3.5-turbo",
				Provider:     "openai",
				SystemPrompt: "You are a config agent",
				Temperature:  0.8,
				MaxTokens:    1500,
				Tools:        []string{},
			},
		},
		Routing: &RoutingConfig{
			Type: "path",
			Rules: []RoutingRule{
				{
					ID:       "def-rule",
					Pattern:  "/definition",
					AgentID:  "definition-agent",
					Method:   "POST",
					Priority: 1,
				},
				{
					ID:       "factory-rule",
					Pattern:  "/factory",
					AgentID:  "factory-agent",
					Method:   "POST",
					Priority: 2,
				},
				{
					ID:       "config-rule",
					Pattern:  "/config",
					AgentID:  "config-agent",
					Method:   "POST",
					Priority: 3,
				},
			},
			DefaultAgent: "config-agent",
		},
	}

	// Create managers
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Create multi-agent manager
	manager, err := NewMultiAgentManager(config, llmManager, toolRegistry)
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test that all agents were created
	deploymentState := manager.GetDeploymentState()
	assert.Contains(t, deploymentState.AgentStates, "definition-agent")
	assert.Contains(t, deploymentState.AgentStates, "factory-agent")
	assert.Contains(t, deploymentState.AgentStates, "config-agent")

	// Test that definition-based agents have their custom setup called
	assert.True(t, testDef1.customSetupCalled)
	assert.True(t, testDef2.customSetupCalled)
}

func TestMultiAgentRoutingHTTP(t *testing.T) {
	// Create test configuration
	config := &MultiAgentConfig{
		Name:        "routing-test",
		Version:     "1.0",
		Description: "Test HTTP routing",
		Agents: map[string]*AgentConfig{
			"echo-agent": {
				ID:           "echo-agent",
				Name:         "Echo Agent",
				Type:         AgentTypeChat,
				Model:        "mock-model",
				Provider:     "mock",
				SystemPrompt: "You are an echo agent",
				Temperature:  0.7,
				MaxTokens:    1000,
				Tools:        []string{},
			},
		},
		Routing: &RoutingConfig{
			Type: "path",
			Rules: []RoutingRule{
				{
					ID:       "echo-rule",
					Pattern:  "/echo",
					AgentID:  "echo-agent",
					Method:   "POST",
					Priority: 1,
				},
			},
			DefaultAgent: "echo-agent",
			Middleware:   []MiddlewareConfig{},
		},
	}

	// Create managers
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Register mock provider for testing
	mockProvider := &mockProvider{response: "Hello, World!"}
	err := llmManager.RegisterProvider("mock", mockProvider)
	assert.NoError(t, err)

	// Create multi-agent manager
	manager, err := NewMultiAgentManager(config, llmManager, toolRegistry)
	assert.NoError(t, err)

	// Create test server
	server := httptest.NewServer(manager.GetRouter())
	defer server.Close()

	// Test routing to specific agent
	requestBody := `{"input": "test message"}`
	resp, err := http.Post(server.URL+"/echo", "application/json", strings.NewReader(requestBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test health endpoint
	resp, err = http.Get(server.URL + "/health")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test metrics endpoint
	resp, err = http.Get(server.URL + "/metrics")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test agent list endpoint
	resp, err = http.Get(server.URL + "/agents")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMultiAgentConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *MultiAgentConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &MultiAgentConfig{
				Name:        "valid-config",
				Version:     "1.0",
				Description: "Valid configuration",
				Agents: map[string]*AgentConfig{
					"agent1": {
						ID:           "agent1",
						Name:         "Agent 1",
						Type:         AgentTypeChat,
						Model:        "gpt-3.5-turbo",
						Provider:     "openai",
						SystemPrompt: "You are agent 1",
						Temperature:  0.7,
						MaxTokens:    1000,
						Tools:        []string{},
					},
				},
				Routing: &RoutingConfig{
					Type: "path",
					Rules: []RoutingRule{
						{
							ID:       "rule1",
							Pattern:  "/agent1",
							AgentID:  "agent1",
							Method:   "POST",
							Priority: 1,
						},
					},
					DefaultAgent: "agent1",
				},
			},
			expectError: false,
		},
		{
			name: "empty name",
			config: &MultiAgentConfig{
				Name: "",
			},
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "no agents",
			config: &MultiAgentConfig{
				Name:        "no-agents",
				Version:     "1.0",
				Description: "No agents configuration",
				Agents:      map[string]*AgentConfig{},
			},
			expectError: true,
			errorMsg:    "at least one agent must be defined",
		},
		{
			name: "invalid agent",
			config: &MultiAgentConfig{
				Name:        "invalid-agent",
				Version:     "1.0",
				Description: "Invalid agent configuration",
				Agents: map[string]*AgentConfig{
					"agent1": {
						ID:   "agent1",
						Name: "", // Empty name
					},
				},
			},
			expectError: true,
			errorMsg:    "agent agent1: name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultiAgentManagerLifecycle(t *testing.T) {
	// Create test configuration
	config := &MultiAgentConfig{
		Name:        "lifecycle-test",
		Version:     "1.0",
		Description: "Test lifecycle management",
		Agents: map[string]*AgentConfig{
			"test-agent": {
				ID:           "test-agent",
				Name:         "Test Agent",
				Type:         AgentTypeChat,
				Model:        "gpt-3.5-turbo",
				Provider:     "openai",
				SystemPrompt: "You are a test agent",
				Temperature:  0.7,
				MaxTokens:    1000,
				Tools:        []string{},
			},
		},
		Routing: &RoutingConfig{
			Type: "path",
			Rules: []RoutingRule{
				{
					ID:       "test-rule",
					Pattern:  "/test",
					AgentID:  "test-agent",
					Method:   "POST",
					Priority: 1,
				},
			},
			DefaultAgent: "test-agent",
		},
	}

	// Create managers
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Create multi-agent manager
	manager, err := NewMultiAgentManager(config, llmManager, toolRegistry)
	assert.NoError(t, err)

	// Test starting the manager
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		startErr := manager.Start(ctx)
		assert.NoError(t, startErr)
	}()

	// Give some time for the manager to start
	time.Sleep(100 * time.Millisecond)

	// Test stopping the manager
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err = manager.Stop(stopCtx)
	assert.NoError(t, err)
}

func TestGlobalRegistry(t *testing.T) {
	// Test global registry functions
	testDef := NewTestAgentDefinition("global-test", AgentTypeChat)

	err := RegisterAgent("global-test", testDef)
	assert.NoError(t, err)

	factory := func() AgentDefinition {
		return NewTestAgentDefinition("global-factory", AgentTypeReAct)
	}

	err = RegisterAgentFactory("global-factory", factory)
	assert.NoError(t, err)

	// Test getting global registry
	registry := GetGlobalRegistry()
	assert.NotNil(t, registry)

	// Test that agents are registered
	definitions := registry.ListDefinitions()
	factories := registry.ListFactories()

	assert.Contains(t, definitions, "global-test")
	assert.Contains(t, factories, "global-factory")
}

func TestAgentInfo(t *testing.T) {
	registry := NewAgentRegistry()

	// Register test agents
	testDef := NewTestAgentDefinition("info-test", AgentTypeChat)
	err := registry.RegisterDefinition("info-test", testDef)
	assert.NoError(t, err)

	factory := func() AgentDefinition {
		return NewTestAgentDefinition("info-factory", AgentTypeReAct)
	}
	err = registry.RegisterFactory("info-factory", factory)
	assert.NoError(t, err)

	// Get agent info
	infos := registry.GetAgentInfo()
	assert.Len(t, infos, 2)

	// Check definition info
	var defInfo, factoryInfo *AgentInfo
	for i := range infos {
		if infos[i].ID == "info-test" {
			defInfo = &infos[i]
		} else if infos[i].ID == "info-factory" {
			factoryInfo = &infos[i]
		}
	}

	assert.NotNil(t, defInfo)
	assert.Equal(t, SourceDefinition, defInfo.Source)
	assert.Equal(t, "info-test", defInfo.Config.Name)

	assert.NotNil(t, factoryInfo)
	assert.Equal(t, SourceFactory, factoryInfo.Source)
	assert.Equal(t, "info-factory", factoryInfo.Config.Name)
}

func TestMultiAgentConfigSerialization(t *testing.T) {
	config := &MultiAgentConfig{
		Name:        "serialization-test",
		Version:     "1.0",
		Description: "Test serialization",
		Agents: map[string]*AgentConfig{
			"agent1": {
				ID:           "agent1",
				Name:         "Agent 1",
				Type:         AgentTypeChat,
				Model:        "gpt-3.5-turbo",
				Provider:     "openai",
				SystemPrompt: "You are agent 1",
				Temperature:  0.7,
				MaxTokens:    1000,
				Tools:        []string{"tool1"},
			},
		},
		Routing: &RoutingConfig{
			Type: "path",
			Rules: []RoutingRule{
				{
					ID:       "rule1",
					Pattern:  "/agent1",
					AgentID:  "agent1",
					Method:   "POST",
					Priority: 1,
				},
			},
			DefaultAgent: "agent1",
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(config)
	assert.NoError(t, err)

	// Test JSON deserialization
	var deserializedConfig MultiAgentConfig
	err = json.Unmarshal(jsonData, &deserializedConfig)
	assert.NoError(t, err)

	// Verify deserialized data
	assert.Equal(t, config.Name, deserializedConfig.Name)
	assert.Equal(t, config.Version, deserializedConfig.Version)
	assert.Equal(t, config.Description, deserializedConfig.Description)
	assert.Len(t, deserializedConfig.Agents, 1)
	assert.Contains(t, deserializedConfig.Agents, "agent1")
	assert.Equal(t, config.Agents["agent1"].Name, deserializedConfig.Agents["agent1"].Name)
}

// Benchmark tests
func BenchmarkAgentRegistration(b *testing.B) {
	registry := NewAgentRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testDef := NewTestAgentDefinition(fmt.Sprintf("bench-agent-%d", i), AgentTypeChat)
		registry.RegisterDefinition(fmt.Sprintf("bench-agent-%d", i), testDef)
	}
}

func BenchmarkAgentCreation(b *testing.B) {
	registry := NewAgentRegistry()
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Pre-register agents
	for i := 0; i < 100; i++ {
		testDef := NewTestAgentDefinition(fmt.Sprintf("bench-agent-%d", i), AgentTypeChat)
		registry.RegisterDefinition(fmt.Sprintf("bench-agent-%d", i), testDef)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agentID := fmt.Sprintf("bench-agent-%d", i%100)
		_, err := registry.CreateAgentFromDefinition(agentID, llmManager, toolRegistry)
		if err != nil {
			b.Fatalf("Failed to create agent: %v", err)
		}
	}
}

func BenchmarkMultiAgentRouting(b *testing.B) {
	// Create test configuration
	config := &MultiAgentConfig{
		Name:        "benchmark-routing",
		Version:     "1.0",
		Description: "Benchmark routing performance",
		Agents: map[string]*AgentConfig{
			"bench-agent": {
				ID:           "bench-agent",
				Name:         "Benchmark Agent",
				Type:         AgentTypeChat,
				Model:        "gpt-3.5-turbo",
				Provider:     "openai",
				SystemPrompt: "You are a benchmark agent",
				Temperature:  0.7,
				MaxTokens:    1000,
				Tools:        []string{},
			},
		},
		Routing: &RoutingConfig{
			Type: "path",
			Rules: []RoutingRule{
				{
					ID:       "bench-rule",
					Pattern:  "/bench",
					AgentID:  "bench-agent",
					Method:   "POST",
					Priority: 1,
				},
			},
			DefaultAgent: "bench-agent",
		},
	}

	// Create managers
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Create multi-agent manager
	manager, err := NewMultiAgentManager(config, llmManager, toolRegistry)
	if err != nil {
		b.Fatalf("Failed to create manager: %v", err)
	}

	// Create test server
	server := httptest.NewServer(manager.GetRouter())
	defer server.Close()

	requestBody := `{"input": "benchmark test"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := http.Post(server.URL+"/bench", "application/json", strings.NewReader(requestBody))
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}
}
