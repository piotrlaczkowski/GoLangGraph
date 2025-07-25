// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// ServerConfig represents server configuration
type ServerConfig struct {
	Host           string        `json:"host"`
	Port           int           `json:"port"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	MaxHeaderBytes int           `json:"max_header_bytes"`
	EnableCORS     bool          `json:"enable_cors"`
	StaticDir      string        `json:"static_dir"`
	DevMode        bool          `json:"dev_mode"`
	LogLevel       string        `json:"log_level"`
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:           "0.0.0.0",
		Port:           8080,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
		EnableCORS:     true,
		StaticDir:      "./static",
		DevMode:        false,
		LogLevel:       "info",
	}
}

// Server represents the HTTP server
type Server struct {
	config   *ServerConfig
	router   *mux.Router
	server   *http.Server
	logger   *logrus.Logger
	upgrader websocket.Upgrader

	// Core components
	llmManager     *llm.ProviderManager
	toolRegistry   *tools.ToolRegistry
	agentManager   *AgentManager
	sessionManager *persistence.SessionManager

	// WebSocket connections
	wsConnections   map[string]*websocket.Conn
	wsConnectionsMu sync.RWMutex
}

// NewServer creates a new server
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	server := &Server{
		config:        config,
		router:        mux.NewRouter(),
		logger:        logrus.New(),
		wsConnections: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}

	server.setupRoutes()
	return server
}

// SetLLMManager sets the LLM provider manager
func (s *Server) SetLLMManager(manager *llm.ProviderManager) {
	s.llmManager = manager
}

// SetToolRegistry sets the tool registry
func (s *Server) SetToolRegistry(registry *tools.ToolRegistry) {
	s.toolRegistry = registry
}

// SetAgentManager sets the agent manager
func (s *Server) SetAgentManager(manager *AgentManager) {
	s.agentManager = manager
}

// SetSessionManager sets the session manager
func (s *Server) SetSessionManager(manager *persistence.SessionManager) {
	s.sessionManager = manager
}

// setupRoutes sets up HTTP routes
func (s *Server) setupRoutes() {
	// Enable CORS if configured
	if s.config.EnableCORS {
		s.router.Use(s.corsMiddleware)
	}

	// Middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.authMiddleware)

	// API routes
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET", "OPTIONS")

	// LLM providers
	api.HandleFunc("/providers", s.handleListProviders).Methods("GET")
	api.HandleFunc("/providers/{name}/models", s.handleGetProviderModels).Methods("GET")
	api.HandleFunc("/providers/{name}/health", s.handleProviderHealth).Methods("GET")

	// Agents
	api.HandleFunc("/agents", s.handleListAgents).Methods("GET")
	api.HandleFunc("/agents", s.handleCreateAgent).Methods("POST")
	api.HandleFunc("/agents/{id}", s.handleGetAgent).Methods("GET")
	api.HandleFunc("/agents/{id}", s.handleUpdateAgent).Methods("PUT")
	api.HandleFunc("/agents/{id}", s.handleDeleteAgent).Methods("DELETE")
	api.HandleFunc("/agents/{id}/execute", s.handleExecuteAgent).Methods("POST")
	api.HandleFunc("/agents/{id}/history", s.handleGetAgentHistory).Methods("GET")

	// Graphs
	api.HandleFunc("/graphs", s.handleListGraphs).Methods("GET")
	api.HandleFunc("/graphs/{id}", s.handleGetGraph).Methods("GET")
	api.HandleFunc("/graphs/{id}/topology", s.handleGetGraphTopology).Methods("GET")
	api.HandleFunc("/graphs/{id}/execute", s.handleExecuteGraph).Methods("POST")
	api.HandleFunc("/graphs/{id}/interrupt", s.handleInterruptGraph).Methods("POST")

	// Sessions and threads
	api.HandleFunc("/sessions", s.handleCreateSession).Methods("POST")
	api.HandleFunc("/sessions/{id}", s.handleGetSession).Methods("GET")
	api.HandleFunc("/threads", s.handleCreateThread).Methods("POST")
	api.HandleFunc("/threads/{id}", s.handleGetThread).Methods("GET")
	api.HandleFunc("/threads/{id}/checkpoints", s.handleListCheckpoints).Methods("GET")

	// Tools
	api.HandleFunc("/tools", s.handleListTools).Methods("GET")
	api.HandleFunc("/tools/{name}", s.handleGetTool).Methods("GET")

	// WebSocket endpoints
	api.HandleFunc("/ws/agents/{id}/stream", s.handleAgentWebSocket)
	api.HandleFunc("/ws/graphs/{id}/stream", s.handleGraphWebSocket)

	// Dev mode specific routes
	if s.config.DevMode {
		debug := s.router.PathPrefix("/debug").Subrouter()
		debug.HandleFunc("/", s.handleDebugDashboard).Methods("GET")
		debug.HandleFunc("/agents", s.handleDebugAgents).Methods("GET")
		debug.HandleFunc("/config", s.handleDebugConfig).Methods("GET")
		debug.HandleFunc("/logs", s.handleDebugLogs).Methods("GET")
		debug.HandleFunc("/metrics", s.handleDebugMetrics).Methods("GET")
		debug.HandleFunc("/reload", s.handleDebugReload).Methods("POST")

		playground := s.router.PathPrefix("/playground").Subrouter()
		playground.HandleFunc("/", s.handlePlaygroundDashboard).Methods("GET")
		playground.HandleFunc("/test", s.handlePlaygroundTest).Methods("POST")
		playground.HandleFunc("/agents/{id}/test", s.handlePlaygroundAgentTest).Methods("POST")
	}

	// Static files for UI
	if s.config.StaticDir != "" {
		s.router.PathPrefix("/").Handler(http.FileServer(http.Dir(s.config.StaticDir)))
	}
}

// Start starts the server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:        s.router,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}

	s.logger.WithFields(logrus.Fields{
		"host": s.config.Host,
		"port": s.config.Port,
	}).Info("Starting GoLangGraph server")

	return s.server.ListenAndServe()
}

// Stop stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping GoLangGraph server")
	return s.server.Shutdown(ctx)
}

// Middleware

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
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

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		s.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(start),
			"remote":   r.RemoteAddr,
		}).Info("HTTP request")
	})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple authentication - in production, implement proper JWT/OAuth
		// For now, just check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Allow requests without API key for development
		}

		next.ServeHTTP(w, r)
	})
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	// Check component health
	if s.llmManager != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		providerHealth := s.llmManager.HealthCheck(ctx)
		health["providers"] = providerHealth
	}

	s.writeJSON(w, http.StatusOK, health)
}

// Provider handlers
func (s *Server) handleListProviders(w http.ResponseWriter, r *http.Request) {
	if s.llmManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "LLM manager not available")
		return
	}

	providers := s.llmManager.ListProviders()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers": providers,
	})
}

func (s *Server) handleGetProviderModels(w http.ResponseWriter, r *http.Request) {
	if s.llmManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "LLM manager not available")
		return
	}

	vars := mux.Vars(r)
	providerName := vars["name"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	models, err := s.llmManager.GetProviderModels(ctx, providerName)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"provider": providerName,
		"models":   models,
	})
}

func (s *Server) handleProviderHealth(w http.ResponseWriter, r *http.Request) {
	if s.llmManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "LLM manager not available")
		return
	}

	vars := mux.Vars(r)
	providerName := vars["name"]

	provider, err := s.llmManager.GetProvider(providerName)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = provider.IsHealthy(ctx)
	status := "healthy"
	if err != nil {
		status = "unhealthy"
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"provider": providerName,
		"status":   status,
		"error":    err,
	})
}

// Agent handlers
func (s *Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	agents := s.agentManager.ListAgents()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
	})
}

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	var config agent.AgentConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	agentInstance, err := s.agentManager.CreateAgent(&config)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"agent": agentInstance.GetConfig(),
	})
}

func (s *Server) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	vars := mux.Vars(r)
	agentID := vars["id"]

	agentInstance, exists := s.agentManager.GetAgent(agentID)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent": agentInstance.GetConfig(),
	})
}

func (s *Server) handleExecuteAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	vars := mux.Vars(r)
	agentID := vars["id"]

	var request struct {
		Input  string `json:"input"`
		Stream bool   `json:"stream"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	agentInstance, exists := s.agentManager.GetAgent(agentID)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	execution, err := agentInstance.Execute(ctx, request.Input)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"execution": execution,
	})
}

func (s *Server) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	vars := mux.Vars(r)
	agentID := vars["id"]

	var config agent.AgentConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure the ID matches
	config.ID = agentID

	agentInstance, err := s.agentManager.CreateAgent(&config)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent": agentInstance.GetConfig(),
	})
}

func (s *Server) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	vars := mux.Vars(r)
	agentID := vars["id"]

	s.agentManager.DeleteAgent(agentID)
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Agent deleted successfully",
	})
}

func (s *Server) handleGetAgentHistory(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	vars := mux.Vars(r)
	agentID := vars["id"]

	agentInstance, exists := s.agentManager.GetAgent(agentID)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}

	// Get execution history from agent
	history := agentInstance.GetExecutionHistory()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id": agentID,
		"history":  history,
	})
}

func (s *Server) handleListGraphs(w http.ResponseWriter, r *http.Request) {
	// For now, return empty list - would need graph manager
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"graphs": []string{},
	})
}

func (s *Server) handleGetGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	// Placeholder implementation
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"graph_id": graphID,
		"nodes":    []string{},
		"edges":    []string{},
	})
}

func (s *Server) handleGetGraphTopology(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	// Placeholder implementation
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"graph_id": graphID,
		"topology": map[string]interface{}{
			"nodes": []string{},
			"edges": []string{},
		},
	})
}

func (s *Server) handleExecuteGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	var request struct {
		Input string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Placeholder implementation
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"graph_id":  graphID,
		"execution": "completed",
		"result":    "placeholder result",
	})
}

func (s *Server) handleInterruptGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	// Placeholder implementation
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"graph_id": graphID,
		"status":   "interrupted",
	})
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	if s.sessionManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Session manager not available")
		return
	}

	var request struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	session := &persistence.Session{
		ID:        uuid.New().String(),
		UserID:    request.UserID,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
	}

	err := s.sessionManager.CreateSession(ctx, session)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"session": session,
	})
}

func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
	if s.sessionManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Session manager not available")
		return
	}

	vars := mux.Vars(r)
	sessionID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	session, err := s.sessionManager.GetSession(ctx, sessionID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"session": session,
	})
}

func (s *Server) handleCreateThread(w http.ResponseWriter, r *http.Request) {
	if s.sessionManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Session manager not available")
		return
	}

	var request struct {
		SessionID string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	thread := &persistence.Thread{
		ID:        uuid.New().String(),
		Name:      fmt.Sprintf("Thread-%s", time.Now().Format("2006-01-02-15-04-05")),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.sessionManager.CreateThread(ctx, thread)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"thread": thread,
	})
}

func (s *Server) handleGetThread(w http.ResponseWriter, r *http.Request) {
	if s.sessionManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Session manager not available")
		return
	}

	vars := mux.Vars(r)
	threadID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	thread, err := s.sessionManager.GetThread(ctx, threadID)
	if err != nil {
		s.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"thread": thread,
	})
}

func (s *Server) handleListCheckpoints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	threadID := vars["id"]

	// Placeholder implementation
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"thread_id":   threadID,
		"checkpoints": []string{},
	})
}

func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	if s.toolRegistry == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Tool registry not available")
		return
	}

	tools := s.toolRegistry.ListTools()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tools": tools,
	})
}

func (s *Server) handleGetTool(w http.ResponseWriter, r *http.Request) {
	if s.toolRegistry == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Tool registry not available")
		return
	}

	vars := mux.Vars(r)
	toolName := vars["name"]

	tool, exists := s.toolRegistry.GetTool(toolName)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Tool not found")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tool": tool,
	})
}

func (s *Server) handleGraphWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	graphID := vars["id"]

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade WebSocket")
		return
	}
	defer conn.Close()

	// Store connection
	s.wsConnectionsMu.Lock()
	s.wsConnections[graphID] = conn
	s.wsConnectionsMu.Unlock()

	// Clean up on disconnect
	defer func() {
		s.wsConnectionsMu.Lock()
		delete(s.wsConnections, graphID)
		s.wsConnectionsMu.Unlock()
	}()

	// Handle WebSocket messages for graph execution
	for {
		var message struct {
			Type  string `json:"type"`
			Input string `json:"input"`
		}

		err := conn.ReadJSON(&message)
		if err != nil {
			s.logger.WithError(err).Error("WebSocket read error")
			break
		}

		// Placeholder graph execution
		if message.Type == "execute" {
			conn.WriteJSON(map[string]interface{}{
				"type":      "result",
				"graph_id":  graphID,
				"result":    "Graph execution completed",
				"timestamp": time.Now(),
			})
		}
	}
}

// WebSocket handlers
func (s *Server) handleAgentWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade WebSocket")
		return
	}
	defer conn.Close()

	// Store connection
	s.wsConnectionsMu.Lock()
	s.wsConnections[agentID] = conn
	s.wsConnectionsMu.Unlock()

	// Clean up on disconnect
	defer func() {
		s.wsConnectionsMu.Lock()
		delete(s.wsConnections, agentID)
		s.wsConnectionsMu.Unlock()
	}()

	// Handle WebSocket messages
	for {
		var message struct {
			Type  string `json:"type"`
			Input string `json:"input"`
		}

		err := conn.ReadJSON(&message)
		if err != nil {
			s.logger.WithError(err).Error("WebSocket read error")
			break
		}

		if message.Type == "execute" && s.agentManager != nil {
			agentInstance, exists := s.agentManager.GetAgent(agentID)
			if exists {
				go s.streamAgentExecution(conn, agentInstance, message.Input)
			}
		}
	}
}

func (s *Server) streamAgentExecution(conn *websocket.Conn, agent *agent.Agent, input string) {
	ctx := context.Background()

	// Send start message
	conn.WriteJSON(map[string]interface{}{
		"type":      "start",
		"timestamp": time.Now(),
	})

	// Execute agent
	execution, err := agent.Execute(ctx, input)

	if err != nil {
		conn.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": err.Error(),
		})
		return
	}

	// Send result
	conn.WriteJSON(map[string]interface{}{
		"type":      "result",
		"execution": execution,
		"timestamp": time.Now(),
	})
}

// Utility functions
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]interface{}{
		"error":     message,
		"timestamp": time.Now(),
	})
}

// Dev mode handlers
func (s *Server) handleDebugDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>GoLangGraph Debug Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .card { border: 1px solid #ddd; padding: 20px; margin: 10px 0; border-radius: 5px; }
        .nav { margin-bottom: 20px; }
        .nav a { margin-right: 15px; text-decoration: none; color: #0066cc; }
        .nav a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>GoLangGraph Debug Dashboard</h1>
    <div class="nav">
        <a href="/debug/agents">Agents</a>
        <a href="/debug/config">Configuration</a>
        <a href="/debug/logs">Logs</a>
        <a href="/debug/metrics">Metrics</a>
        <a href="/playground">Playground</a>
    </div>
    <div class="card">
        <h3>System Status</h3>
        <p>Development mode is active</p>
        <p>Server uptime: <span id="uptime">Loading...</span></p>
    </div>
    <div class="card">
        <h3>Quick Actions</h3>
        <button onclick="fetch('/debug/reload', {method: 'POST'}).then(r => r.json()).then(d => alert(JSON.stringify(d)))">Reload Configuration</button>
    </div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(dashboardHTML))
}

func (s *Server) handleDebugAgents(w http.ResponseWriter, r *http.Request) {
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	agents := s.agentManager.ListAgents()
	var agentDetails []map[string]interface{}

	for _, agentID := range agents {
		if agentInstance, exists := s.agentManager.GetAgent(agentID); exists {
			config := agentInstance.GetConfig()
			agentDetails = append(agentDetails, map[string]interface{}{
				"id":     agentID,
				"config": config,
				"graph":  agentInstance.GetGraph(),
			})
		}
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agentDetails,
		"count":  len(agents),
	})
}

func (s *Server) handleDebugConfig(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"server_config": s.config,
		"dev_mode":      s.config.DevMode,
	})
}

func (s *Server) handleDebugLogs(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you would retrieve logs from a log store
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs": []map[string]interface{}{
			{
				"timestamp": time.Now().Format(time.RFC3339),
				"level":     "info",
				"message":   "Debug logs endpoint accessed",
			},
		},
	})
}

func (s *Server) handleDebugMetrics(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you would collect actual metrics
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"metrics": map[string]interface{}{
			"requests_total":        0,
			"agents_active":         len(s.agentManager.ListAgents()),
			"websocket_connections": len(s.wsConnections),
			"memory_usage":          "N/A",
		},
	})
}

func (s *Server) handleDebugReload(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you would reload configuration
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Configuration reloaded successfully",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handlePlaygroundDashboard(w http.ResponseWriter, r *http.Request) {
	playgroundHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>GoLangGraph Playground</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .panel { border: 1px solid #ddd; padding: 20px; margin: 10px 0; border-radius: 5px; }
        textarea { width: 100%; height: 100px; }
        button { padding: 10px 20px; margin: 5px; background: #0066cc; color: white; border: none; border-radius: 3px; cursor: pointer; }
        button:hover { background: #0052a3; }
        .output { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>GoLangGraph Playground</h1>

        <div class="panel">
            <h3>Test Agent</h3>
            <textarea id="input" placeholder="Enter your message here..."></textarea>
            <br>
            <button onclick="testAgent()">Test Agent</button>
            <div id="output" class="output"></div>
        </div>

        <div class="panel">
            <h3>Available Agents</h3>
            <div id="agents">Loading...</div>
        </div>
    </div>

    <script>
        async function testAgent() {
            const input = document.getElementById('input').value;
            const output = document.getElementById('output');

            try {
                const response = await fetch('/playground/test', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ input: input })
                });
                const result = await response.json();
                output.innerHTML = '<pre>' + JSON.stringify(result, null, 2) + '</pre>';
            } catch (error) {
                output.innerHTML = '<pre>Error: ' + error.message + '</pre>';
            }
        }

        // Load agents
        fetch('/api/v1/agents')
            .then(r => r.json())
            .then(data => {
                const agentsDiv = document.getElementById('agents');
                if (data.agents && data.agents.length > 0) {
                    agentsDiv.innerHTML = data.agents.map(agent =>
                        '<div>• ' + agent + '</div>'
                    ).join('');
                } else {
                    agentsDiv.innerHTML = 'No agents available';
                }
            });
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(playgroundHTML))
}

func (s *Server) handlePlaygroundTest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Input string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Test with the first available agent
	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	agents := s.agentManager.ListAgents()
	if len(agents) == 0 {
		s.writeError(w, http.StatusNotFound, "No agents available")
		return
	}

	agentInstance, exists := s.agentManager.GetAgent(agents[0])
	if !exists {
		s.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	execution, err := agentInstance.Execute(ctx, request.Input)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id":  agents[0],
		"input":     request.Input,
		"execution": execution,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handlePlaygroundAgentTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]

	var request struct {
		Input string `json:"input"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if s.agentManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Agent manager not available")
		return
	}

	agentInstance, exists := s.agentManager.GetAgent(agentID)
	if !exists {
		s.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	execution, err := agentInstance.Execute(ctx, request.Input)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id":  agentID,
		"input":     request.Input,
		"execution": execution,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// AgentManager manages multiple agents
type AgentManager struct {
	agents       map[string]*agent.Agent
	llmManager   *llm.ProviderManager
	toolRegistry *tools.ToolRegistry
	mu           sync.RWMutex
}

// NewAgentManager creates a new agent manager
func NewAgentManager(llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) *AgentManager {
	return &AgentManager{
		agents:       make(map[string]*agent.Agent),
		llmManager:   llmManager,
		toolRegistry: toolRegistry,
	}
}

// CreateAgent creates a new agent
func (am *AgentManager) CreateAgent(config *agent.AgentConfig) (*agent.Agent, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	agentInstance := agent.NewAgent(config, am.llmManager, am.toolRegistry)
	am.agents[config.ID] = agentInstance

	return agentInstance, nil
}

// GetAgent retrieves an agent by ID
func (am *AgentManager) GetAgent(id string) (*agent.Agent, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	agentInstance, exists := am.agents[id]
	return agentInstance, exists
}

// ListAgents returns all agent IDs
func (am *AgentManager) ListAgents() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	ids := make([]string, 0, len(am.agents))
	for id := range am.agents {
		ids = append(ids, id)
	}
	return ids
}

// DeleteAgent removes an agent
func (am *AgentManager) DeleteAgent(id string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.agents, id)
}
