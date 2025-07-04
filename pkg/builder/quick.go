package builder

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/server"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// QuickBuilder provides the ultimate minimal code experience for creating agents
type QuickBuilder struct {
	llmManager   *llm.ProviderManager
	toolRegistry *tools.ToolRegistry
	checkpointer persistence.Checkpointer
	config       *QuickConfig
}

// QuickConfig holds configuration for quick agent creation
type QuickConfig struct {
	// LLM Provider settings
	OpenAIKey    string
	OllamaURL    string
	GeminiKey    string
	DefaultModel string

	// Agent settings
	SystemPrompt  string
	Temperature   float64
	MaxTokens     int
	MaxIterations int

	// Persistence settings
	UseMemory   bool
	DatabaseURL string

	// Tools settings
	EnableAllTools bool
	CustomTools    []string
}

// DefaultQuickConfig returns a sensible default configuration
func DefaultQuickConfig() *QuickConfig {
	return &QuickConfig{
		OpenAIKey:      os.Getenv("OPENAI_API_KEY"),
		OllamaURL:      "http://localhost:11434",
		GeminiKey:      os.Getenv("GEMINI_API_KEY"),
		DefaultModel:   "gpt-3.5-turbo",
		SystemPrompt:   "You are a helpful AI assistant.",
		Temperature:    0.7,
		MaxTokens:      1000,
		MaxIterations:  10,
		UseMemory:      true,
		EnableAllTools: true,
	}
}

// NewQuickBuilder creates a new quick builder with auto-configuration
func NewQuickBuilder() *QuickBuilder {
	config := DefaultQuickConfig()

	// Auto-initialize LLM providers
	llmManager := llm.NewProviderManager()

	// Add OpenAI if key is available
	if config.OpenAIKey != "" {
		openaiProvider, err := llm.NewOpenAIProvider(&llm.ProviderConfig{
			APIKey:   config.OpenAIKey,
			Endpoint: "https://api.openai.com/v1",
		})
		if err == nil {
			llmManager.RegisterProvider("openai", openaiProvider)
		}
	}

	// Add Ollama
	ollamaProvider, err := llm.NewOllamaProvider(&llm.ProviderConfig{
		Endpoint: config.OllamaURL,
	})
	if err == nil {
		llmManager.RegisterProvider("ollama", ollamaProvider)
	}

	// Add Gemini if key is available
	if config.GeminiKey != "" {
		geminiProvider, err := llm.NewGeminiProvider(&llm.ProviderConfig{
			APIKey: config.GeminiKey,
		})
		if err == nil {
			llmManager.RegisterProvider("gemini", geminiProvider)
		}
	}

	// Auto-initialize tools
	toolRegistry := tools.NewToolRegistry()
	if config.EnableAllTools {
		toolRegistry.RegisterTool(tools.NewCalculatorTool())
		toolRegistry.RegisterTool(tools.NewWebSearchTool())
		toolRegistry.RegisterTool(tools.NewFileReadTool())
		toolRegistry.RegisterTool(tools.NewFileWriteTool())
		toolRegistry.RegisterTool(tools.NewShellTool())
		toolRegistry.RegisterTool(tools.NewHTTPTool())
		toolRegistry.RegisterTool(tools.NewTimeTool())
	}

	// Auto-initialize checkpointer
	var checkpointer persistence.Checkpointer
	if config.UseMemory {
		checkpointer = persistence.NewMemoryCheckpointer()
	}

	return &QuickBuilder{
		llmManager:   llmManager,
		toolRegistry: toolRegistry,
		checkpointer: checkpointer,
		config:       config,
	}
}

// WithConfig allows customizing the configuration
func (qb *QuickBuilder) WithConfig(config *QuickConfig) *QuickBuilder {
	qb.config = config
	return qb
}

// WithLLM adds or configures an LLM provider
func (qb *QuickBuilder) WithLLM(provider string, config interface{}) *QuickBuilder {
	switch provider {
	case "openai":
		if cfg, ok := config.(*llm.ProviderConfig); ok {
			openaiProvider, err := llm.NewOpenAIProvider(cfg)
			if err == nil {
				qb.llmManager.RegisterProvider("openai", openaiProvider)
			}
		}
	case "ollama":
		if cfg, ok := config.(*llm.ProviderConfig); ok {
			ollamaProvider, err := llm.NewOllamaProvider(cfg)
			if err == nil {
				qb.llmManager.RegisterProvider("ollama", ollamaProvider)
			}
		}
	case "gemini":
		if cfg, ok := config.(*llm.ProviderConfig); ok {
			geminiProvider, err := llm.NewGeminiProvider(cfg)
			if err == nil {
				qb.llmManager.RegisterProvider("gemini", geminiProvider)
			}
		}
	}
	return qb
}

// WithTools adds custom tools
func (qb *QuickBuilder) WithTools(tools ...tools.Tool) *QuickBuilder {
	for _, tool := range tools {
		qb.toolRegistry.RegisterTool(tool)
	}
	return qb
}

// WithPersistence configures persistence
func (qb *QuickBuilder) WithPersistence(checkpointer persistence.Checkpointer) *QuickBuilder {
	qb.checkpointer = checkpointer
	return qb
}

// ========== ULTRA-MINIMAL AGENT CREATION ==========

// Chat creates a simple chat agent in 1 line
func (qb *QuickBuilder) Chat(name ...string) *agent.Agent {
	agentName := "ChatAgent"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:         agentName,
		Type:         agent.AgentTypeChat,
		SystemPrompt: qb.config.SystemPrompt,
		Temperature:  qb.config.Temperature,
		MaxTokens:    qb.config.MaxTokens,
		Provider:     qb.getBestProvider(),
		Model:        qb.config.DefaultModel,
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// ReAct creates a ReAct agent with reasoning capabilities
func (qb *QuickBuilder) ReAct(name ...string) *agent.Agent {
	agentName := "ReActAgent"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:          agentName,
		Type:          agent.AgentTypeReAct,
		SystemPrompt:  "You are a helpful assistant that can reason step by step and use tools when needed.",
		Temperature:   qb.config.Temperature,
		MaxTokens:     qb.config.MaxTokens,
		MaxIterations: qb.config.MaxIterations,
		Provider:      qb.getBestProvider(),
		Model:         qb.config.DefaultModel,
		Tools:         qb.toolRegistry.ListTools(),
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// Tool creates a tool-focused agent
func (qb *QuickBuilder) Tool(name ...string) *agent.Agent {
	agentName := "ToolAgent"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:         agentName,
		Type:         agent.AgentTypeTool,
		SystemPrompt: "You are a helpful assistant that specializes in using tools to accomplish tasks.",
		Temperature:  qb.config.Temperature,
		MaxTokens:    qb.config.MaxTokens,
		Provider:     qb.getBestProvider(),
		Model:        qb.config.DefaultModel,
		Tools:        qb.toolRegistry.ListTools(),
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// RAG creates a RAG (Retrieval-Augmented Generation) agent
func (qb *QuickBuilder) RAG(name ...string) *agent.Agent {
	agentName := "RAGAgent"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:         agentName,
		Type:         agent.AgentTypeChat,
		SystemPrompt: "You are a helpful assistant that can search and retrieve information from documents to answer questions accurately.",
		Temperature:  qb.config.Temperature,
		MaxTokens:    qb.config.MaxTokens,
		Provider:     qb.getBestProvider(),
		Model:        qb.config.DefaultModel,
		Tools:        []string{"web_search", "file_read"},
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// Multi creates a multi-agent coordinator
func (qb *QuickBuilder) Multi() *agent.MultiAgentCoordinator {
	return agent.NewMultiAgentCoordinator()
}

// ========== SPECIALIZED AGENTS ==========

// Researcher creates a research-focused agent
func (qb *QuickBuilder) Researcher(name ...string) *agent.Agent {
	agentName := "Researcher"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:         agentName,
		Type:         agent.AgentTypeReAct,
		SystemPrompt: "You are a research specialist. You excel at finding, analyzing, and synthesizing information from multiple sources.",
		Temperature:  0.3, // Lower temperature for more focused research
		MaxTokens:    2000,
		Provider:     qb.getBestProvider(),
		Model:        qb.config.DefaultModel,
		Tools:        []string{"web_search", "file_read", "http_request"},
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// Writer creates a writing-focused agent
func (qb *QuickBuilder) Writer(name ...string) *agent.Agent {
	agentName := "Writer"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:         agentName,
		Type:         agent.AgentTypeChat,
		SystemPrompt: "You are a skilled technical writer. You excel at creating clear, well-structured, and engaging content.",
		Temperature:  0.8, // Higher temperature for more creative writing
		MaxTokens:    2000,
		Provider:     qb.getBestProvider(),
		Model:        qb.config.DefaultModel,
		Tools:        []string{"file_write"},
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// Analyst creates a data analysis agent
func (qb *QuickBuilder) Analyst(name ...string) *agent.Agent {
	agentName := "Analyst"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:         agentName,
		Type:         agent.AgentTypeReAct,
		SystemPrompt: "You are a data analyst. You excel at analyzing data, identifying patterns, and providing insights.",
		Temperature:  0.2, // Low temperature for precise analysis
		MaxTokens:    1500,
		Provider:     qb.getBestProvider(),
		Model:        qb.config.DefaultModel,
		Tools:        []string{"calculator", "file_read", "shell"},
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// Coder creates a coding assistant agent
func (qb *QuickBuilder) Coder(name ...string) *agent.Agent {
	agentName := "Coder"
	if len(name) > 0 {
		agentName = name[0]
	}

	config := &agent.AgentConfig{
		Name:         agentName,
		Type:         agent.AgentTypeReAct,
		SystemPrompt: "You are a coding assistant. You excel at writing, debugging, and explaining code in multiple programming languages.",
		Temperature:  0.3,
		MaxTokens:    2000,
		Provider:     qb.getBestProvider(),
		Model:        qb.config.DefaultModel,
		Tools:        []string{"file_read", "file_write", "shell"},
	}

	return agent.NewAgent(config, qb.llmManager, qb.toolRegistry)
}

// ========== WORKFLOW BUILDERS ==========

// Pipeline creates a sequential agent pipeline
func (qb *QuickBuilder) Pipeline(agents ...*agent.Agent) *AgentPipeline {
	return &AgentPipeline{
		agents:      agents,
		coordinator: agent.NewMultiAgentCoordinator(),
	}
}

// Swarm creates a parallel agent swarm
func (qb *QuickBuilder) Swarm(agents ...*agent.Agent) *AgentSwarm {
	return &AgentSwarm{
		agents:      agents,
		coordinator: agent.NewMultiAgentCoordinator(),
	}
}

// ========== SERVER BUILDER ==========

// Server creates a ready-to-use server
func (qb *QuickBuilder) Server(port ...int) *server.Server {
	serverPort := 8080
	if len(port) > 0 {
		serverPort = port[0]
	}

	config := &server.ServerConfig{
		Host:           "0.0.0.0",
		Port:           serverPort,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		EnableCORS:     true,
	}

	srv := server.NewServer(config)
	srv.SetLLMManager(qb.llmManager)
	srv.SetToolRegistry(qb.toolRegistry)

	// Create and set agent manager
	agentManager := server.NewAgentManager(qb.llmManager, qb.toolRegistry)
	srv.SetAgentManager(agentManager)

	// Create and set session manager
	sessionManager := persistence.NewSessionManager(nil, nil)
	srv.SetSessionManager(sessionManager)

	return srv
}

// ========== HELPER METHODS ==========

// getBestProvider returns the best available provider
func (qb *QuickBuilder) getBestProvider() string {
	providers := qb.llmManager.ListProviders()

	// Priority order: OpenAI, Gemini, Ollama
	for _, preferred := range []string{"openai", "gemini", "ollama"} {
		for _, available := range providers {
			if available == preferred {
				return preferred
			}
		}
	}

	// Return first available provider
	if len(providers) > 0 {
		return providers[0]
	}

	return "mock" // Fallback
}

// ========== WORKFLOW TYPES ==========

// AgentPipeline represents a sequential workflow
type AgentPipeline struct {
	agents      []*agent.Agent
	coordinator *agent.MultiAgentCoordinator
}

// Execute runs the pipeline sequentially
func (ap *AgentPipeline) Execute(ctx context.Context, input string) ([]agent.AgentExecution, error) {
	agentIDs := make([]string, len(ap.agents))

	for i, agent := range ap.agents {
		id := fmt.Sprintf("agent_%d", i)
		agentIDs[i] = id
		ap.coordinator.AddAgent(id, agent)
	}

	return ap.coordinator.ExecuteSequential(ctx, agentIDs, input)
}

// AgentSwarm represents a parallel workflow
type AgentSwarm struct {
	agents      []*agent.Agent
	coordinator *agent.MultiAgentCoordinator
}

// Execute runs the swarm in parallel
func (as *AgentSwarm) Execute(ctx context.Context, input string) ([]agent.AgentExecution, error) {
	agentIDs := make([]string, len(as.agents))

	for i, agent := range as.agents {
		id := fmt.Sprintf("agent_%d", i)
		agentIDs[i] = id
		as.coordinator.AddAgent(id, agent)
	}

	results, err := as.coordinator.ExecuteParallel(ctx, agentIDs, input)
	if err != nil {
		return nil, err
	}

	// Convert map to slice
	executions := make([]agent.AgentExecution, 0, len(results))
	for _, execution := range results {
		executions = append(executions, execution)
	}

	return executions, nil
}

// ========== GLOBAL QUICK FUNCTIONS ==========

// Quick returns a global quick builder instance
func Quick() *QuickBuilder {
	return NewQuickBuilder()
}

// OneLineChat creates a chat agent in one line
func OneLineChat(name ...string) *agent.Agent {
	return Quick().Chat(name...)
}

// OneLineReAct creates a ReAct agent in one line
func OneLineReAct(name ...string) *agent.Agent {
	return Quick().ReAct(name...)
}

// OneLineTool creates a tool agent in one line
func OneLineTool(name ...string) *agent.Agent {
	return Quick().Tool(name...)
}

// OneLineRAG creates a RAG agent in one line
func OneLineRAG(name ...string) *agent.Agent {
	return Quick().RAG(name...)
}

// OneLineServer creates a server in one line
func OneLineServer(port ...int) *server.Server {
	return Quick().Server(port...)
}

// OneLinePipeline creates a pipeline in one line
func OneLinePipeline(agents ...*agent.Agent) *AgentPipeline {
	return Quick().Pipeline(agents...)
}

// OneLineSwarm creates a swarm in one line
func OneLineSwarm(agents ...*agent.Agent) *AgentSwarm {
	return Quick().Swarm(agents...)
}
