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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/server"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"

	"go-agents-simple-statefull/agents"
	"go-agents-simple-statefull/database"
)

// ApplicationConfig holds the complete application configuration
type ApplicationConfig struct {
	Server   *server.AutoServerConfig         `json:"server"`
	Database *database.EnhancedDatabaseConfig `json:"database"`
	LLM      *LLMConfig                       `json:"llm"`
	Agents   *AgentsConfig                    `json:"agents"`
	Features *FeaturesConfig                  `json:"features"`
}

// LLMConfig configures LLM providers
type LLMConfig struct {
	DefaultProvider string                     `json:"default_provider"`
	Providers       map[string]*ProviderConfig `json:"providers"`
}

// ProviderConfig configures individual LLM providers
type ProviderConfig struct {
	Type     string                 `json:"type"`
	Endpoint string                 `json:"endpoint"`
	APIKey   string                 `json:"api_key"`
	Models   []string               `json:"models"`
	Config   map[string]interface{} `json:"config"`
}

// AgentsConfig configures agent behavior
type AgentsConfig struct {
	EnableAutoDiscovery bool                          `json:"enable_auto_discovery"`
	DefaultModel        string                        `json:"default_model"`
	DefaultProvider     string                        `json:"default_provider"`
	AgentConfigs        map[string]*agent.AgentConfig `json:"agent_configs"`
}

// FeaturesConfig enables/disables application features
type FeaturesConfig struct {
	EnableWebUI        bool `json:"enable_web_ui"`
	EnablePlayground   bool `json:"enable_playground"`
	EnableMetrics      bool `json:"enable_metrics"`
	EnableMonitoring   bool `json:"enable_monitoring"`
	EnablePersistence  bool `json:"enable_persistence"`
	EnableRAG          bool `json:"enable_rag"`
	EnableVectorSearch bool `json:"enable_vector_search"`
	EnableSessionMgmt  bool `json:"enable_session_management"`
}

// StatefulAgentSystem represents the complete stateful agent system
type StatefulAgentSystem struct {
	config            *ApplicationConfig
	databaseManager   *database.DatabaseManager
	llmManager        *llm.ProviderManager
	toolRegistry      *tools.ToolRegistry
	agentRegistry     *agent.AgentRegistry
	autoServer        *server.AutoServer
	multiAgentManager *agent.MultiAgentManager
	logger            *logrus.Logger
}

func main() {
	// Initialize logging
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.Info("Starting GoLangGraph Stateful Ideation Agents System")

	// Load configuration
	config := loadConfiguration()

	// Create the stateful agent system
	system, err := NewStatefulAgentSystem(config, logger)
	if err != nil {
		log.Fatalf("Failed to create stateful agent system: %v", err)
	}
	defer system.Shutdown()

	// Initialize all components
	if err := system.Initialize(); err != nil {
		log.Fatalf("Failed to initialize system: %v", err)
	}

	// Register agents
	if err := system.RegisterAgents(); err != nil {
		log.Fatalf("Failed to register agents: %v", err)
	}

	// Start the system
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Start the auto-server with all registered agents
	logger.Info("Starting auto-generated multi-agent server...")
	if err := system.Start(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	logger.Info("GoLangGraph Stateful Ideation Agents System shut down successfully")
}

// NewStatefulAgentSystem creates a new stateful agent system
func NewStatefulAgentSystem(config *ApplicationConfig, logger *logrus.Logger) (*StatefulAgentSystem, error) {
	system := &StatefulAgentSystem{
		config: config,
		logger: logger,
	}

	return system, nil
}

// Initialize initializes all system components
func (s *StatefulAgentSystem) Initialize() error {
	s.logger.Info("Initializing stateful agent system components...")

	// Initialize database manager with enhanced capabilities
	if err := s.initializeDatabaseManager(); err != nil {
		return fmt.Errorf("failed to initialize database manager: %w", err)
	}

	// Initialize LLM manager
	if err := s.initializeLLMManager(); err != nil {
		return fmt.Errorf("failed to initialize LLM manager: %w", err)
	}

	// Initialize tool registry
	if err := s.initializeToolRegistry(); err != nil {
		return fmt.Errorf("failed to initialize tool registry: %w", err)
	}

	// Initialize agent registry
	if err := s.initializeAgentRegistry(); err != nil {
		return fmt.Errorf("failed to initialize agent registry: %w", err)
	}

	// Initialize auto-server with enhanced capabilities
	if err := s.initializeAutoServer(); err != nil {
		return fmt.Errorf("failed to initialize auto-server: %w", err)
	}

	s.logger.Info("All system components initialized successfully")
	return nil
}

// initializeDatabaseManager sets up the enhanced database manager
func (s *StatefulAgentSystem) initializeDatabaseManager() error {
	s.logger.Info("Initializing enhanced database manager with GoLangGraph persistence...")

	var err error
	s.databaseManager, err = database.NewDatabaseManager(s.config.Database)
	if err != nil {
		return fmt.Errorf("failed to create database manager: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"primary_db":     s.config.Database.Primary.Type,
		"cache_db":       s.config.Database.Cache.Type,
		"vector_db":      s.config.Database.Vector.Type,
		"rag_enabled":    s.config.Database.RAG.Enabled,
		"memory_enabled": s.config.Database.Memory.EnableEmbeddings,
	}).Info("Database manager initialized with full persistence stack")

	return nil
}

// initializeLLMManager sets up the LLM provider manager
func (s *StatefulAgentSystem) initializeLLMManager() error {
	s.logger.Info("Initializing LLM provider manager...")

	s.llmManager = llm.NewProviderManager()

	// Configure LLM providers based on configuration
	for providerName, providerConfig := range s.config.LLM.Providers {
		switch providerConfig.Type {
		case "ollama":
			provider, err := llm.NewOllamaProvider(&llm.ProviderConfig{
				Endpoint: providerConfig.Endpoint,
			})
			if err != nil {
				s.logger.WithError(err).WithField("provider", providerName).Warn("Failed to initialize LLM provider")
				continue
			}
			s.llmManager.RegisterProvider(providerName, provider)
			s.logger.WithField("provider", providerName).Info("LLM provider registered")

		case "openai":
			provider, err := llm.NewOpenAIProvider(&llm.ProviderConfig{
				APIKey: providerConfig.APIKey,
			})
			if err != nil {
				s.logger.WithError(err).WithField("provider", providerName).Warn("Failed to initialize LLM provider")
				continue
			}
			s.llmManager.RegisterProvider(providerName, provider)
			s.logger.WithField("provider", providerName).Info("LLM provider registered")

		case "gemini":
			provider, err := llm.NewGeminiProvider(&llm.ProviderConfig{
				APIKey: providerConfig.APIKey,
			})
			if err != nil {
				s.logger.WithError(err).WithField("provider", providerName).Warn("Failed to initialize LLM provider")
				continue
			}
			s.llmManager.RegisterProvider(providerName, provider)
			s.logger.WithField("provider", providerName).Info("LLM provider registered")

		default:
			s.logger.WithField("provider", providerName).WithField("type", providerConfig.Type).Warn("Unknown LLM provider type")
		}
	}

	s.logger.WithField("providers", len(s.config.LLM.Providers)).Info("LLM manager initialized")
	return nil
}

// initializeToolRegistry sets up the tool registry
func (s *StatefulAgentSystem) initializeToolRegistry() error {
	s.logger.Info("Initializing tool registry...")

	s.toolRegistry = tools.NewToolRegistry()

	// Tools are automatically registered in NewToolRegistry()

	s.logger.WithField("tools", len(s.toolRegistry.ListTools())).Info("Tool registry initialized")
	return nil
}

// initializeAgentRegistry sets up the agent registry
func (s *StatefulAgentSystem) initializeAgentRegistry() error {
	s.logger.Info("Initializing agent registry...")

	s.agentRegistry = agent.GetGlobalRegistry()

	s.logger.Info("Agent registry initialized")
	return nil
}

// initializeAutoServer sets up the auto-server with enhanced capabilities
func (s *StatefulAgentSystem) initializeAutoServer() error {
	s.logger.Info("Initializing auto-server with stateful capabilities...")

	// Enhanced auto-server configuration
	autoServerConfig := &server.AutoServerConfig{
		Host:             s.config.Server.Host,
		Port:             s.config.Server.Port,
		BasePath:         s.config.Server.BasePath,
		EnableWebUI:      s.config.Features.EnableWebUI,
		EnablePlayground: s.config.Features.EnablePlayground,
		EnableSchemaAPI:  true,
		EnableMetricsAPI: s.config.Features.EnableMetrics,
		EnableCORS:       true,
		SchemaValidation: true,
		OllamaEndpoint:   getOllamaEndpoint(s.config.LLM.Providers),
		ServerTimeout:    30 * time.Second,
		MaxRequestSize:   10 * 1024 * 1024, // 10MB
		Middleware:       []string{"cors", "logging", "recovery"},
		LLMProviders:     convertLLMProviders(s.config.LLM.Providers),
	}

	s.autoServer = server.NewAutoServer(autoServerConfig)

	s.logger.WithFields(logrus.Fields{
		"host":       autoServerConfig.Host,
		"port":       autoServerConfig.Port,
		"web_ui":     autoServerConfig.EnableWebUI,
		"playground": autoServerConfig.EnablePlayground,
		"metrics":    autoServerConfig.EnableMetricsAPI,
	}).Info("Auto-server initialized with enhanced capabilities")

	return nil
}

// RegisterAgents registers all stateful agents
func (s *StatefulAgentSystem) RegisterAgents() error {
	s.logger.Info("Registering stateful agents...")

	// Get the global stateful registry
	statefulRegistry := agents.GetGlobalRegistry()

	// First, register all agents with the database manager
	if err := statefulRegistry.RegisterAllAgents(s.databaseManager); err != nil {
		return fmt.Errorf("failed to register agents with database manager: %w", err)
	}

	// Register all agents from the stateful registry
	agentInfo := statefulRegistry.GetAllAgentInfo()

	registeredCount := 0
	for agentID, info := range agentInfo {
		// Get the agent definition from the stateful registry
		if agentDef, exists := statefulRegistry.GetDefinition(agentID); exists {
			// Register with the main registry
			if err := s.agentRegistry.RegisterDefinition(agentID, agentDef); err != nil {
				s.logger.WithError(err).WithField("agent", agentID).Warn("Failed to register agent")
				continue
			}

			// Register with the auto-server for automatic endpoint generation
			if err := s.autoServer.RegisterAgent(agentID, agentDef); err != nil {
				s.logger.WithError(err).WithField("agent", agentID).Warn("Failed to register agent with auto-server")
				continue
			}

			s.logger.WithFields(logrus.Fields{
				"agent_id":   agentID,
				"agent_name": info.Name,
				"version":    info.Version,
			}).Info("Agent registered successfully")

			registeredCount++
		}
	}

	s.logger.WithFields(logrus.Fields{
		"agents_registered": registeredCount,
		"total_agents":      len(agentInfo),
		"definitions":       len(s.agentRegistry.ListDefinitions()),
	}).Info("Stateful agents registered successfully")

	return nil
}

// Start starts the complete system
func (s *StatefulAgentSystem) Start(ctx context.Context) error {
	s.logger.Info("Starting stateful agent system...")

	// Generate dynamic endpoints for all registered agents
	if err := s.autoServer.GenerateEndpoints(); err != nil {
		return fmt.Errorf("failed to generate endpoints: %w", err)
	}

	// Print system information
	s.printSystemInfo()

	// Start the auto-server
	return s.autoServer.Start(ctx)
}

// Shutdown gracefully shuts down the system
func (s *StatefulAgentSystem) Shutdown() {
	s.logger.Info("Shutting down stateful agent system...")

	// Close database connections
	if s.databaseManager != nil {
		if err := s.databaseManager.Close(); err != nil {
			s.logger.WithError(err).Error("Error closing database manager")
		}
	}

	s.logger.Info("System shutdown complete")
}

// printSystemInfo prints comprehensive system information
func (s *StatefulAgentSystem) printSystemInfo() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🤖 GoLangGraph Stateful Ideation Agents System")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("📡 Server: http://%s:%d\n", s.config.Server.Host, s.config.Server.Port)
	fmt.Printf("🎯 API Base: %s\n", s.config.Server.BasePath)

	fmt.Println("\n🔗 Available Endpoints:")
	fmt.Printf("  • Health Check:     GET  /health\n")
	fmt.Printf("  • Agent List:       GET  /agents\n")
	fmt.Printf("  • Designer Agent:   POST %s/designer\n", s.config.Server.BasePath)
	fmt.Printf("  • Agent Stream:     POST %s/designer/stream\n", s.config.Server.BasePath)
	fmt.Printf("  • Conversation:     GET  %s/designer/conversation\n", s.config.Server.BasePath)
	fmt.Printf("  • Agent Status:     GET  %s/designer/status\n", s.config.Server.BasePath)

	if s.config.Features.EnableWebUI {
		fmt.Printf("  • Web Interface:    GET  /\n")
		fmt.Printf("  • Chat Interface:   GET  /chat\n")
	}

	if s.config.Features.EnablePlayground {
		fmt.Printf("  • API Playground:   GET  /playground\n")
	}

	if s.config.Features.EnableMetrics {
		fmt.Printf("  • Metrics:          GET  /metrics\n")
	}

	fmt.Println("\n🧠 Enhanced Capabilities:")
	fmt.Printf("  • Session Management:   %v\n", s.config.Features.EnableSessionMgmt)
	fmt.Printf("  • Persistent Memory:    %v\n", s.config.Features.EnablePersistence)
	fmt.Printf("  • RAG Integration:      %v\n", s.config.Features.EnableRAG)
	fmt.Printf("  • Vector Search:        %v\n", s.config.Features.EnableVectorSearch)
	fmt.Printf("  • User Learning:        %v\n", s.config.Database.Memory.EnableEmbeddings)

	fmt.Println("\n🗄️ Database Stack:")
	fmt.Printf("  • Primary DB:           %s\n", s.config.Database.Primary.Type)
	fmt.Printf("  • Cache DB:             %s\n", s.config.Database.Cache.Type)
	fmt.Printf("  • Vector DB:            %s\n", s.config.Database.Vector.Type)
	fmt.Printf("  • RAG Enabled:          %v\n", s.config.Database.RAG.Enabled)

	fmt.Println("\n🤖 LLM Providers:")
	for name, provider := range s.config.LLM.Providers {
		fmt.Printf("  • %s (%s): %s\n", name, provider.Type, provider.Endpoint)
	}

	fmt.Println("\n📊 System Features:")
	fmt.Printf("  • Auto-Generated API:   ✅\n")
	fmt.Printf("  • Schema Validation:    ✅\n")
	fmt.Printf("  • OpenAPI Documentation: ✅\n")
	fmt.Printf("  • Real-time Streaming:  ✅\n")
	fmt.Printf("  • State Persistence:    ✅\n")
	fmt.Printf("  • Memory Management:    ✅\n")
	fmt.Printf("  • User Preference Learning: ✅\n")

	fmt.Println("\n💡 Usage Example:")
	fmt.Printf(`curl -X POST http://%s:%d%s/designer \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Design a sustainable tiny house for $50k",
    "user_id": "user123",
    "context": {
      "project_type": "residential",
      "budget_range": 50000,
      "sustainability_priority": 9
    }
  }'`, s.config.Server.Host, s.config.Server.Port, s.config.Server.BasePath)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🚀 System ready! All agents are online and stateful!")
	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// loadConfiguration loads the application configuration
func loadConfiguration() *ApplicationConfig {
	return &ApplicationConfig{
		Server: &server.AutoServerConfig{
			Host:             getEnv("SERVER_HOST", "0.0.0.0"),
			Port:             getEnvAsInt("SERVER_PORT", 8080),
			BasePath:         getEnv("API_BASE_PATH", "/api"),
			EnableWebUI:      getEnvAsBool("ENABLE_WEB_UI", true),
			EnablePlayground: getEnvAsBool("ENABLE_PLAYGROUND", true),
			EnableSchemaAPI:  true,
			EnableMetricsAPI: getEnvAsBool("ENABLE_METRICS", true),
			EnableCORS:       true,
			SchemaValidation: true,
		},
		Database: database.NewEnhancedDatabaseConfig(),
		LLM: &LLMConfig{
			DefaultProvider: getEnv("DEFAULT_LLM_PROVIDER", "ollama"),
			Providers: map[string]*ProviderConfig{
				"ollama": {
					Type:     "ollama",
					Endpoint: getEnv("OLLAMA_ENDPOINT", "http://localhost:11434"),
					Models:   []string{"gemma3:1b", "gemma3:2b", "mistral:7b"},
				},
				"openai": {
					Type:   "openai",
					APIKey: getEnv("OPENAI_API_KEY", ""),
					Models: []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"},
				},
				"gemini": {
					Type:   "gemini",
					APIKey: getEnv("GEMINI_API_KEY", ""),
					Models: []string{"gemini-pro", "gemini-pro-vision"},
				},
			},
		},
		Agents: &AgentsConfig{
			EnableAutoDiscovery: true,
			DefaultModel:        getEnv("DEFAULT_MODEL", "gemma3:1b"),
			DefaultProvider:     getEnv("DEFAULT_PROVIDER", "ollama"),
		},
		Features: &FeaturesConfig{
			EnableWebUI:        getEnvAsBool("ENABLE_WEB_UI", true),
			EnablePlayground:   getEnvAsBool("ENABLE_PLAYGROUND", true),
			EnableMetrics:      getEnvAsBool("ENABLE_METRICS", true),
			EnableMonitoring:   getEnvAsBool("ENABLE_MONITORING", true),
			EnablePersistence:  getEnvAsBool("ENABLE_PERSISTENCE", true),
			EnableRAG:          getEnvAsBool("ENABLE_RAG", true),
			EnableVectorSearch: getEnvAsBool("ENABLE_VECTOR_SEARCH", true),
			EnableSessionMgmt:  getEnvAsBool("ENABLE_SESSION_MGMT", true),
		},
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getOllamaEndpoint(providers map[string]*ProviderConfig) string {
	if ollama, exists := providers["ollama"]; exists {
		return ollama.Endpoint
	}
	return "http://localhost:11434"
}

func convertLLMProviders(providers map[string]*ProviderConfig) map[string]interface{} {
	result := make(map[string]interface{})
	for name, config := range providers {
		result[name] = map[string]interface{}{
			"type":     config.Type,
			"endpoint": config.Endpoint,
			"models":   config.Models,
		}
	}
	return result
}
