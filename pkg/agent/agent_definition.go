// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"fmt"
	"plugin"
	"reflect"
	"sync"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// AgentDefinition represents a programmatic agent definition
type AgentDefinition interface {
	// GetConfig returns the base configuration for the agent
	GetConfig() *AgentConfig

	// Initialize sets up the agent with the provided managers
	Initialize(llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) error

	// CreateAgent creates and returns the configured agent instance
	CreateAgent() (*Agent, error)

	// GetMetadata returns additional metadata about the agent
	GetMetadata() map[string]interface{}

	// Validate validates the agent definition
	Validate() error
}

// CustomAgentDefinition allows for completely custom agent creation
type CustomAgentDefinition interface {
	AgentDefinition

	// BuildGraph allows custom graph construction
	BuildGraph() (*core.Graph, error)

	// GetCustomTools returns custom tools specific to this agent
	GetCustomTools() []tools.Tool

	// GetCustomMiddleware returns custom middleware for this agent
	GetCustomMiddleware() []func(next func(*core.BaseState) (*core.BaseState, error)) func(*core.BaseState) (*core.BaseState, error)
}

// AgentFactory is a function type for creating agent definitions
type AgentFactory func() AgentDefinition

// AgentRegistry manages programmatically defined agents
type AgentRegistry struct {
	definitions map[string]AgentDefinition
	factories   map[string]AgentFactory
	plugins     map[string]*plugin.Plugin
	mu          sync.RWMutex
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		definitions: make(map[string]AgentDefinition),
		factories:   make(map[string]AgentFactory),
		plugins:     make(map[string]*plugin.Plugin),
	}
}

// RegisterDefinition registers an agent definition with a unique ID
func (ar *AgentRegistry) RegisterDefinition(id string, definition AgentDefinition) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if _, exists := ar.definitions[id]; exists {
		return fmt.Errorf("agent definition with ID %s already exists", id)
	}

	if err := definition.Validate(); err != nil {
		return fmt.Errorf("invalid agent definition: %w", err)
	}

	ar.definitions[id] = definition
	return nil
}

// RegisterFactory registers an agent factory function
func (ar *AgentRegistry) RegisterFactory(id string, factory AgentFactory) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if _, exists := ar.factories[id]; exists {
		return fmt.Errorf("agent factory with ID %s already exists", id)
	}

	ar.factories[id] = factory
	return nil
}

// LoadFromPlugin loads agent definitions from a Go plugin
func (ar *AgentRegistry) LoadFromPlugin(pluginPath string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", pluginPath, err)
	}

	// Look for GetAgentDefinitions function
	sym, err := p.Lookup("GetAgentDefinitions")
	if err != nil {
		return fmt.Errorf("plugin %s does not export GetAgentDefinitions function: %w", pluginPath, err)
	}

	getDefinitions, ok := sym.(func() map[string]AgentDefinition)
	if !ok {
		return fmt.Errorf("plugin %s GetAgentDefinitions function has incorrect signature", pluginPath)
	}

	definitions := getDefinitions()
	for id, definition := range definitions {
		if err := definition.Validate(); err != nil {
			return fmt.Errorf("invalid agent definition %s from plugin %s: %w", id, pluginPath, err)
		}
		ar.definitions[id] = definition
	}

	ar.plugins[pluginPath] = p
	return nil
}

// GetDefinition retrieves an agent definition by ID
func (ar *AgentRegistry) GetDefinition(id string) (AgentDefinition, bool) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	definition, exists := ar.definitions[id]
	return definition, exists
}

// CreateAgentFromDefinition creates an agent from a registered definition
func (ar *AgentRegistry) CreateAgentFromDefinition(id string, llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) (*Agent, error) {
	definition, exists := ar.GetDefinition(id)
	if !exists {
		return nil, fmt.Errorf("agent definition %s not found", id)
	}

	if err := definition.Initialize(llmManager, toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to initialize agent definition %s: %w", id, err)
	}

	return definition.CreateAgent()
}

// CreateAgentFromFactory creates an agent using a registered factory
func (ar *AgentRegistry) CreateAgentFromFactory(id string, llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) (*Agent, error) {
	ar.mu.RLock()
	factory, exists := ar.factories[id]
	ar.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent factory %s not found", id)
	}

	definition := factory()
	if err := definition.Initialize(llmManager, toolRegistry); err != nil {
		return nil, fmt.Errorf("failed to initialize agent from factory %s: %w", id, err)
	}

	return definition.CreateAgent()
}

// ListDefinitions returns all registered agent definition IDs
func (ar *AgentRegistry) ListDefinitions() []string {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	ids := make([]string, 0, len(ar.definitions))
	for id := range ar.definitions {
		ids = append(ids, id)
	}
	return ids
}

// ListFactories returns all registered agent factory IDs
func (ar *AgentRegistry) ListFactories() []string {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	ids := make([]string, 0, len(ar.factories))
	for id := range ar.factories {
		ids = append(ids, id)
	}
	return ids
}

// GetMetadata returns metadata for all registered agents
func (ar *AgentRegistry) GetMetadata() map[string]map[string]interface{} {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	metadata := make(map[string]map[string]interface{})

	for id, definition := range ar.definitions {
		metadata[id] = definition.GetMetadata()
	}

	return metadata
}

// BaseAgentDefinition provides a base implementation of AgentDefinition
type BaseAgentDefinition struct {
	config       *AgentConfig
	llmManager   *llm.ProviderManager
	toolRegistry *tools.ToolRegistry
	metadata     map[string]interface{}
}

// NewBaseAgentDefinition creates a new base agent definition
func NewBaseAgentDefinition(config *AgentConfig) *BaseAgentDefinition {
	return &BaseAgentDefinition{
		config:   config,
		metadata: make(map[string]interface{}),
	}
}

// GetConfig returns the agent configuration
func (bad *BaseAgentDefinition) GetConfig() *AgentConfig {
	return bad.config
}

// Initialize sets up the agent with managers
func (bad *BaseAgentDefinition) Initialize(llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) error {
	bad.llmManager = llmManager
	bad.toolRegistry = toolRegistry
	return nil
}

// CreateAgent creates a standard agent instance
func (bad *BaseAgentDefinition) CreateAgent() (*Agent, error) {
	if bad.llmManager == nil || bad.toolRegistry == nil {
		return nil, fmt.Errorf("agent definition not properly initialized")
	}

	return NewAgent(bad.config, bad.llmManager, bad.toolRegistry), nil
}

// GetMetadata returns agent metadata
func (bad *BaseAgentDefinition) GetMetadata() map[string]interface{} {
	metadata := make(map[string]interface{})
	for k, v := range bad.metadata {
		metadata[k] = v
	}

	// Add configuration metadata
	metadata["name"] = bad.config.Name
	metadata["type"] = bad.config.Type
	metadata["model"] = bad.config.Model
	metadata["provider"] = bad.config.Provider
	metadata["tools"] = bad.config.Tools

	return metadata
}

// SetMetadata sets a metadata value
func (bad *BaseAgentDefinition) SetMetadata(key string, value interface{}) {
	bad.metadata[key] = value
}

// Validate validates the agent definition
func (bad *BaseAgentDefinition) Validate() error {
	if bad.config == nil {
		return fmt.Errorf("agent configuration is required")
	}

	if bad.config.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if bad.config.Type == "" {
		return fmt.Errorf("agent type is required")
	}

	if bad.config.Model == "" {
		return fmt.Errorf("agent model is required")
	}

	if bad.config.Provider == "" {
		return fmt.Errorf("agent provider is required")
	}

	return nil
}

// AgentDefinitionBuilder provides a fluent interface for building agent definitions
type AgentDefinitionBuilder struct {
	config   *AgentConfig
	metadata map[string]interface{}
}

// NewAgentDefinitionBuilder creates a new agent definition builder
func NewAgentDefinitionBuilder() *AgentDefinitionBuilder {
	return &AgentDefinitionBuilder{
		config:   DefaultAgentConfig(),
		metadata: make(map[string]interface{}),
	}
}

// WithName sets the agent name
func (adb *AgentDefinitionBuilder) WithName(name string) *AgentDefinitionBuilder {
	adb.config.Name = name
	return adb
}

// WithType sets the agent type
func (adb *AgentDefinitionBuilder) WithType(agentType AgentType) *AgentDefinitionBuilder {
	adb.config.Type = agentType
	return adb
}

// WithModel sets the LLM model
func (adb *AgentDefinitionBuilder) WithModel(model string) *AgentDefinitionBuilder {
	adb.config.Model = model
	return adb
}

// WithProvider sets the LLM provider
func (adb *AgentDefinitionBuilder) WithProvider(provider string) *AgentDefinitionBuilder {
	adb.config.Provider = provider
	return adb
}

// WithSystemPrompt sets the system prompt
func (adb *AgentDefinitionBuilder) WithSystemPrompt(prompt string) *AgentDefinitionBuilder {
	adb.config.SystemPrompt = prompt
	return adb
}

// WithTemperature sets the temperature
func (adb *AgentDefinitionBuilder) WithTemperature(temperature float64) *AgentDefinitionBuilder {
	adb.config.Temperature = temperature
	return adb
}

// WithMaxTokens sets the max tokens
func (adb *AgentDefinitionBuilder) WithMaxTokens(maxTokens int) *AgentDefinitionBuilder {
	adb.config.MaxTokens = maxTokens
	return adb
}

// WithTools sets the tools
func (adb *AgentDefinitionBuilder) WithTools(tools ...string) *AgentDefinitionBuilder {
	adb.config.Tools = tools
	return adb
}

// WithMetadata sets metadata
func (adb *AgentDefinitionBuilder) WithMetadata(key string, value interface{}) *AgentDefinitionBuilder {
	adb.metadata[key] = value
	return adb
}

// Build creates the agent definition
func (adb *AgentDefinitionBuilder) Build() *BaseAgentDefinition {
	definition := NewBaseAgentDefinition(adb.config)
	for k, v := range adb.metadata {
		definition.SetMetadata(k, v)
	}
	return definition
}

// AdvancedAgentDefinition provides additional customization capabilities
type AdvancedAgentDefinition struct {
	*BaseAgentDefinition
	customGraph        *core.Graph
	customTools        []tools.Tool
	customMiddleware   []func(next func(*core.BaseState) (*core.BaseState, error)) func(*core.BaseState) (*core.BaseState, error)
	graphBuilder       func() (*core.Graph, error)
	toolsProvider      func() []tools.Tool
	middlewareProvider func() []func(next func(*core.BaseState) (*core.BaseState, error)) func(*core.BaseState) (*core.BaseState, error)
}

// NewAdvancedAgentDefinition creates a new advanced agent definition
func NewAdvancedAgentDefinition(config *AgentConfig) *AdvancedAgentDefinition {
	return &AdvancedAgentDefinition{
		BaseAgentDefinition: NewBaseAgentDefinition(config),
		customTools:         make([]tools.Tool, 0),
		customMiddleware:    make([]func(next func(*core.BaseState) (*core.BaseState, error)) func(*core.BaseState) (*core.BaseState, error), 0),
	}
}

// WithCustomGraph sets a custom graph
func (aad *AdvancedAgentDefinition) WithCustomGraph(graph *core.Graph) *AdvancedAgentDefinition {
	aad.customGraph = graph
	return aad
}

// WithGraphBuilder sets a custom graph builder function
func (aad *AdvancedAgentDefinition) WithGraphBuilder(builder func() (*core.Graph, error)) *AdvancedAgentDefinition {
	aad.graphBuilder = builder
	return aad
}

// WithCustomTools adds custom tools
func (aad *AdvancedAgentDefinition) WithCustomTools(tools ...tools.Tool) *AdvancedAgentDefinition {
	aad.customTools = append(aad.customTools, tools...)
	return aad
}

// WithToolsProvider sets a custom tools provider function
func (aad *AdvancedAgentDefinition) WithToolsProvider(provider func() []tools.Tool) *AdvancedAgentDefinition {
	aad.toolsProvider = provider
	return aad
}

// WithCustomMiddleware adds custom middleware
func (aad *AdvancedAgentDefinition) WithCustomMiddleware(middleware ...func(next func(*core.BaseState) (*core.BaseState, error)) func(*core.BaseState) (*core.BaseState, error)) *AdvancedAgentDefinition {
	aad.customMiddleware = append(aad.customMiddleware, middleware...)
	return aad
}

// WithMiddlewareProvider sets a custom middleware provider function
func (aad *AdvancedAgentDefinition) WithMiddlewareProvider(provider func() []func(next func(*core.BaseState) (*core.BaseState, error)) func(*core.BaseState) (*core.BaseState, error)) *AdvancedAgentDefinition {
	aad.middlewareProvider = provider
	return aad
}

// BuildGraph builds the custom graph
func (aad *AdvancedAgentDefinition) BuildGraph() (*core.Graph, error) {
	if aad.graphBuilder != nil {
		return aad.graphBuilder()
	}

	if aad.customGraph != nil {
		return aad.customGraph, nil
	}

	// Fall back to default graph building
	agent := NewAgent(aad.config, aad.llmManager, aad.toolRegistry)
	return agent.GetGraph(), nil
}

// GetCustomTools returns custom tools
func (aad *AdvancedAgentDefinition) GetCustomTools() []tools.Tool {
	if aad.toolsProvider != nil {
		return aad.toolsProvider()
	}
	return aad.customTools
}

// GetCustomMiddleware returns custom middleware
func (aad *AdvancedAgentDefinition) GetCustomMiddleware() []func(next func(*core.BaseState) (*core.BaseState, error)) func(*core.BaseState) (*core.BaseState, error) {
	if aad.middlewareProvider != nil {
		return aad.middlewareProvider()
	}
	return aad.customMiddleware
}

// CreateAgent creates an advanced agent with custom components
func (aad *AdvancedAgentDefinition) CreateAgent() (*Agent, error) {
	// Create base agent
	agent, err := aad.BaseAgentDefinition.CreateAgent()
	if err != nil {
		return nil, err
	}

	// Register custom tools
	for _, tool := range aad.GetCustomTools() {
		aad.toolRegistry.RegisterTool(tool)
	}

	// Apply custom graph if available
	if aad.customGraph != nil || aad.graphBuilder != nil {
		graph, err := aad.BuildGraph()
		if err != nil {
			return nil, fmt.Errorf("failed to build custom graph: %w", err)
		}

		// Note: In a real implementation, you'd need a way to set the graph on the agent
		// For now, we'll use reflection or provide a setter method
		aad.setAgentGraph(agent, graph)
	}

	return agent, nil
}

// setAgentGraph sets the graph on an agent using reflection
func (aad *AdvancedAgentDefinition) setAgentGraph(agent *Agent, graph *core.Graph) {
	// Use reflection to set the graph field
	agentValue := reflect.ValueOf(agent).Elem()
	graphField := agentValue.FieldByName("graph")
	if graphField.IsValid() && graphField.CanSet() {
		graphField.Set(reflect.ValueOf(graph))
	}
}

// FileBasedAgentLoader loads agent definitions from Go files
type FileBasedAgentLoader struct {
	registry *AgentRegistry
}

// NewFileBasedAgentLoader creates a new file-based agent loader
func NewFileBasedAgentLoader(registry *AgentRegistry) *FileBasedAgentLoader {
	return &FileBasedAgentLoader{
		registry: registry,
	}
}

// LoadFromDirectory loads all agent definitions from a directory
func (fbal *FileBasedAgentLoader) LoadFromDirectory(directory string) error {
	// In a real implementation, this would:
	// 1. Scan the directory for .go files
	// 2. Compile them as plugins or use go/ast to analyze them
	// 3. Load the agent definitions

	// For now, we'll provide a placeholder implementation
	return fmt.Errorf("file-based loading not yet implemented - use plugin loading instead")
}

// AgentSource represents where an agent definition comes from
type AgentSource string

const (
	SourceConfig     AgentSource = "config"
	SourceDefinition AgentSource = "definition"
	SourceFactory    AgentSource = "factory"
	SourcePlugin     AgentSource = "plugin"
)

// AgentInfo provides information about a registered agent
type AgentInfo struct {
	ID       string                 `json:"id"`
	Source   AgentSource            `json:"source"`
	Config   *AgentConfig           `json:"config,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// GetAgentInfo returns information about all registered agents
func (ar *AgentRegistry) GetAgentInfo() []AgentInfo {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	var infos []AgentInfo

	// Add definitions
	for id, definition := range ar.definitions {
		infos = append(infos, AgentInfo{
			ID:       id,
			Source:   SourceDefinition,
			Config:   definition.GetConfig(),
			Metadata: definition.GetMetadata(),
		})
	}

	// Add factories
	for id := range ar.factories {
		// Create a temporary instance to get metadata
		factory := ar.factories[id]
		tempDef := factory()

		infos = append(infos, AgentInfo{
			ID:       id,
			Source:   SourceFactory,
			Config:   tempDef.GetConfig(),
			Metadata: tempDef.GetMetadata(),
		})
	}

	return infos
}

// Global agent registry instance
var globalRegistry = NewAgentRegistry()

// RegisterAgent registers an agent definition globally
func RegisterAgent(id string, definition AgentDefinition) error {
	return globalRegistry.RegisterDefinition(id, definition)
}

// RegisterAgentFactory registers an agent factory globally
func RegisterAgentFactory(id string, factory AgentFactory) error {
	return globalRegistry.RegisterFactory(id, factory)
}

// GetGlobalRegistry returns the global agent registry
func GetGlobalRegistry() *AgentRegistry {
	return globalRegistry
}
