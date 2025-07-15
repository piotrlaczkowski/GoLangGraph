// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// AutoServer automatically generates REST endpoints for agents
type AutoServer struct {
	registry     *agent.AgentRegistry
	llmManager   *llm.ProviderManager
	toolRegistry *tools.ToolRegistry
	router       *mux.Router
	config       *AutoServerConfig
	logger       *logrus.Logger

	// Dynamic agent instances
	agentInstances map[string]*agent.Agent
	agentMetadata  map[string]map[string]interface{}

	// Metrics tracking
	startTime    time.Time
	requestCount int64
}

// AutoServerConfig configures the auto-generated server
type AutoServerConfig struct {
	Host             string                 `yaml:"host" json:"host"`
	Port             int                    `yaml:"port" json:"port"`
	BasePath         string                 `yaml:"base_path" json:"base_path"`
	EnableWebUI      bool                   `yaml:"enable_web_ui" json:"enable_web_ui"`
	EnablePlayground bool                   `yaml:"enable_playground" json:"enable_playground"`
	EnableSchemaAPI  bool                   `yaml:"enable_schema_api" json:"enable_schema_api"`
	EnableMetricsAPI bool                   `yaml:"enable_metrics_api" json:"enable_metrics_api"`
	EnableCORS       bool                   `yaml:"enable_cors" json:"enable_cors"`
	SchemaValidation bool                   `yaml:"schema_validation" json:"schema_validation"`
	OllamaEndpoint   string                 `yaml:"ollama_endpoint" json:"ollama_endpoint"`
	LLMProviders     map[string]interface{} `yaml:"llm_providers" json:"llm_providers"`
	ServerTimeout    time.Duration          `yaml:"server_timeout" json:"server_timeout"`
	MaxRequestSize   int64                  `yaml:"max_request_size" json:"max_request_size"`
	Middleware       []string               `yaml:"middleware" json:"middleware"`
}

// DefaultAutoServerConfig returns default configuration
func DefaultAutoServerConfig() *AutoServerConfig {
	return &AutoServerConfig{
		Host:             "0.0.0.0",
		Port:             8080,
		BasePath:         "/api",
		EnableWebUI:      true,
		EnablePlayground: true,
		EnableSchemaAPI:  true,
		EnableMetricsAPI: true,
		EnableCORS:       true,
		SchemaValidation: true,
		OllamaEndpoint:   "http://localhost:11434",
		ServerTimeout:    30 * time.Second,
		MaxRequestSize:   10 * 1024 * 1024, // 10MB
		Middleware:       []string{"cors", "logging", "recovery"},
	}
}

// NewAutoServer creates a new auto-server instance
func NewAutoServer(config *AutoServerConfig) *AutoServer {
	if config == nil {
		config = DefaultAutoServerConfig()
	}

	router := mux.NewRouter()
	logger := logrus.New()

	// Initialize managers
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Setup LLM providers from config
	setupLLMProviders(llmManager, config)

	return &AutoServer{
		registry:       agent.GetGlobalRegistry(),
		llmManager:     llmManager,
		toolRegistry:   toolRegistry,
		router:         router,
		config:         config,
		logger:         logger,
		agentInstances: make(map[string]*agent.Agent),
		agentMetadata:  make(map[string]map[string]interface{}),
		startTime:      time.Now(),
		requestCount:   0,
	}
}

// LoadAgentsFromDirectory loads agent definitions from a directory
func (as *AutoServer) LoadAgentsFromDirectory(directory string) error {
	as.logger.WithField("directory", directory).Info("Loading agents from directory")

	// This would scan for Go files and load agent definitions
	// For now, we'll use the existing registry system
	definitions := as.registry.ListDefinitions()

	as.logger.WithField("count", len(definitions)).Info("Loaded agent definitions")
	return nil
}

// LoadAgentsFromConfig loads agents from a multi-agent config file
func (as *AutoServer) LoadAgentsFromConfig(configPath string) error {
	config, err := agent.LoadMultiAgentConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load multi-agent config: %w", err)
	}

	// Register agent configs as definitions
	for agentID, agentConfig := range config.Agents {
		definition := agent.NewBaseAgentDefinition(agentConfig)
		if err := as.registry.RegisterDefinition(agentID, definition); err != nil {
			as.logger.WithError(err).WithField("agent_id", agentID).Warn("Failed to register agent definition")
		}
	}

	as.logger.WithField("agents", len(config.Agents)).Info("Loaded agents from config")
	return nil
}

// RegisterAgent registers a single agent programmatically
func (as *AutoServer) RegisterAgent(id string, definition agent.AgentDefinition) error {
	return as.registry.RegisterDefinition(id, definition)
}

// GenerateEndpoints automatically generates REST endpoints for all registered agents
func (as *AutoServer) GenerateEndpoints() error {
	as.logger.Info("Generating dynamic endpoints for agents")

	// Apply middleware
	as.applyMiddleware()

	// Generate core system endpoints
	as.generateSystemEndpoints()

	// Generate agent-specific endpoints
	if err := as.generateAgentEndpoints(); err != nil {
		return fmt.Errorf("failed to generate agent endpoints: %w", err)
	}

	// Generate web interfaces if enabled
	if as.config.EnableWebUI {
		as.generateWebInterfaces()
	}

	// Generate schema endpoints if enabled
	if as.config.EnableSchemaAPI {
		as.generateSchemaEndpoints()
	}

	// Generate metrics endpoints if enabled
	if as.config.EnableMetricsAPI {
		as.generateMetricsEndpoints()
	}

	as.logger.Info("Successfully generated all endpoints")
	return nil
}

// generateSystemEndpoints creates core system endpoints
func (as *AutoServer) generateSystemEndpoints() {
	// Health check
	as.router.HandleFunc("/health", as.handleHealth).Methods("GET", "OPTIONS")

	// Agent capabilities
	as.router.HandleFunc("/capabilities", as.handleCapabilities).Methods("GET", "OPTIONS")

	// List agents
	as.router.HandleFunc("/agents", as.handleListAgents).Methods("GET", "OPTIONS")

	// Agent info
	as.router.HandleFunc("/agents/{agentId}", as.handleAgentInfo).Methods("GET", "OPTIONS")

	as.logger.Info("Generated system endpoints")
}

// generateAgentEndpoints creates dynamic endpoints for each agent
func (as *AutoServer) generateAgentEndpoints() error {
	definitions := as.registry.ListDefinitions()

	for _, agentID := range definitions {
		definition, exists := as.registry.GetDefinition(agentID)
		if !exists {
			continue
		}

		// Create agent instance
		agentInstance, err := as.registry.CreateAgentFromDefinition(agentID, as.llmManager, as.toolRegistry)
		if err != nil {
			as.logger.WithError(err).WithField("agent_id", agentID).Error("Failed to create agent instance")
			continue
		}

		as.agentInstances[agentID] = agentInstance
		as.agentMetadata[agentID] = definition.GetMetadata()

		// Generate endpoints for this agent
		basePath := fmt.Sprintf("%s/%s", as.config.BasePath, agentID)

		// Main agent execution endpoint
		as.router.HandleFunc(basePath, as.createAgentHandler(agentID)).Methods("POST", "OPTIONS")

		// Agent stream endpoint (if supported)
		as.router.HandleFunc(basePath+"/stream", as.createAgentStreamHandler(agentID)).Methods("POST", "OPTIONS")

		// Agent conversation endpoint
		as.router.HandleFunc(basePath+"/conversation", as.createConversationHandler(agentID)).Methods("GET", "POST", "DELETE")

		// Agent status endpoint
		as.router.HandleFunc(basePath+"/status", as.createStatusHandler(agentID)).Methods("GET")

		as.logger.WithField("agent_id", agentID).WithField("base_path", basePath).Info("Generated endpoints for agent")
	}

	return nil
}

// generateWebInterfaces creates web UI endpoints
func (as *AutoServer) generateWebInterfaces() {
	// Root handler redirects to chat
	as.router.HandleFunc("/", as.handleChatInterface).Methods("GET")

	// Main chat interface
	as.router.HandleFunc("/chat", as.handleChatInterface).Methods("GET")

	// Playground interface
	if as.config.EnablePlayground {
		as.router.HandleFunc("/playground", as.handlePlayground).Methods("GET")
	}

	// Debug interface
	as.router.HandleFunc("/debug", as.handleDebug).Methods("GET")

	as.logger.Info("Generated web interfaces")
}

// generateSchemaEndpoints creates schema API endpoints
func (as *AutoServer) generateSchemaEndpoints() {
	// All schemas
	as.router.HandleFunc("/schemas", as.handleSchemas).Methods("GET")

	// Specific agent schema
	as.router.HandleFunc("/schemas/{agentId}", as.handleAgentSchema).Methods("GET")

	// Schema validation endpoint
	as.router.HandleFunc("/validate/{agentId}", as.handleValidateSchema).Methods("POST")

	as.logger.Info("Generated schema endpoints")
}

// generateMetricsEndpoints creates metrics API endpoints
func (as *AutoServer) generateMetricsEndpoints() {
	// System metrics
	as.router.HandleFunc("/metrics", as.handleMetrics).Methods("GET")

	// Agent-specific metrics
	as.router.HandleFunc("/metrics/{agentId}", as.handleAgentMetrics).Methods("GET")

	as.logger.Info("Generated metrics endpoints")
}

// applyMiddleware applies configured middleware
func (as *AutoServer) applyMiddleware() {
	// Always apply metrics middleware
	as.router.Use(as.metricsMiddleware())

	for _, middleware := range as.config.Middleware {
		switch middleware {
		case "cors":
			if as.config.EnableCORS {
				as.router.Use(corsMiddleware)
			}
		case "logging":
			as.router.Use(loggingMiddleware(as.logger))
		case "recovery":
			as.router.Use(recoveryMiddleware(as.logger))
		}
	}
}

// Start starts the auto-server
func (as *AutoServer) Start(ctx context.Context) error {
	if err := as.GenerateEndpoints(); err != nil {
		return fmt.Errorf("failed to generate endpoints: %w", err)
	}

	address := fmt.Sprintf("%s:%d", as.config.Host, as.config.Port)

	server := &http.Server{
		Addr:         address,
		Handler:      as.router,
		ReadTimeout:  as.config.ServerTimeout,
		WriteTimeout: as.config.ServerTimeout,
	}

	as.logger.WithField("address", address).Info("Starting auto-generated multi-agent server")

	// Print available endpoints
	as.printEndpoints()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			as.logger.WithError(err).Error("Server failed to start")
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	as.logger.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

// printEndpoints prints all available endpoints
func (as *AutoServer) printEndpoints() {
	as.logger.Info("🌐 Auto-Generated Endpoints:")
	as.logger.Info("📋 System Endpoints:")
	as.logger.Info("   GET  /health - Health check")
	as.logger.Info("   GET  /capabilities - System capabilities")
	as.logger.Info("   GET  /agents - List all agents")
	as.logger.Info("   GET  /agents/{agentId} - Agent information")

	if as.config.EnableWebUI {
		as.logger.Info("🎨 Web Interfaces:")
		as.logger.Info("   GET  /chat - Interactive chat interface")
		as.logger.Info("   GET  /debug - Debug interface")
		if as.config.EnablePlayground {
			as.logger.Info("   GET  /playground - API playground")
		}
	}

	if as.config.EnableSchemaAPI {
		as.logger.Info("📄 Schema API:")
		as.logger.Info("   GET  /schemas - All agent schemas")
		as.logger.Info("   GET  /schemas/{agentId} - Specific agent schema")
		as.logger.Info("   POST /validate/{agentId} - Validate agent input/output")
	}

	if as.config.EnableMetricsAPI {
		as.logger.Info("📊 Metrics API:")
		as.logger.Info("   GET  /metrics - System metrics")
		as.logger.Info("   GET  /metrics/{agentId} - Agent metrics")
	}

	as.logger.Info("🤖 Agent Endpoints:")
	for agentID := range as.agentInstances {
		basePath := fmt.Sprintf("%s/%s", as.config.BasePath, agentID)
		as.logger.WithField("agent", agentID).Info(fmt.Sprintf("   POST %s - Execute agent", basePath))
		as.logger.WithField("agent", agentID).Info(fmt.Sprintf("   POST %s/stream - Stream agent response", basePath))
		as.logger.WithField("agent", agentID).Info(fmt.Sprintf("   GET  %s/conversation - Get conversation history", basePath))
		as.logger.WithField("agent", agentID).Info(fmt.Sprintf("   POST %s/conversation - Add to conversation", basePath))
		as.logger.WithField("agent", agentID).Info(fmt.Sprintf("   DELETE %s/conversation - Clear conversation", basePath))
		as.logger.WithField("agent", agentID).Info(fmt.Sprintf("   GET  %s/status - Agent status", basePath))
	}
}

// setupLLMProviders initializes LLM providers based on configuration
func setupLLMProviders(manager *llm.ProviderManager, config *AutoServerConfig) {
	// Setup Ollama if endpoint is configured
	if config.OllamaEndpoint != "" {
		ollamaProvider, err := llm.NewOllamaProvider(&llm.ProviderConfig{
			Endpoint: config.OllamaEndpoint,
		})
		if err != nil {
			// Just skip this provider if it fails
			return
		}
		manager.RegisterProvider("ollama", ollamaProvider)
	}

	// Setup other providers from config
	for providerName, providerConfig := range config.LLMProviders {
		// This would setup providers based on their configuration
		// Implementation depends on the specific provider types
		_ = providerName
		_ = providerConfig
	}
}

// metricsMiddleware tracks request counts
func (as *AutoServer) metricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			as.requestCount++
			next.ServeHTTP(w, r)
		})
	}
}

// Middleware functions
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.WithFields(logrus.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"duration": time.Since(start),
				"ip":       r.RemoteAddr,
			}).Info("Request processed")
		})
	}
}

func recoveryMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithField("error", err).Error("Panic recovered")
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
