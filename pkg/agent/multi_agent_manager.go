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
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// MultiAgentManager manages multiple agents with routing and deployment capabilities
type MultiAgentManager struct {
	config          *MultiAgentConfig
	agents          map[string]*Agent
	llmManager      *llm.ProviderManager
	toolRegistry    *tools.ToolRegistry
	router          *mux.Router
	middleware      []MiddlewareFunc
	deploymentState *DeploymentState
	logger          *logrus.Logger
	mu              sync.RWMutex

	// Health monitoring
	healthCheckers map[string]*HealthChecker
	healthMu       sync.RWMutex

	// Metrics and monitoring
	metrics *MultiAgentMetrics
}

// MiddlewareFunc defines middleware function signature
type MiddlewareFunc func(next http.Handler) http.Handler

// DeploymentState tracks the deployment state of agents
type DeploymentState struct {
	Status      string                 `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	AgentStates map[string]*AgentState `json:"agent_states"`
	ErrorCount  int                    `json:"error_count"`
	LastError   string                 `json:"last_error"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AgentState tracks the state of individual agents
type AgentState struct {
	ID           string                 `json:"id"`
	Status       string                 `json:"status"` // "starting", "running", "stopping", "stopped", "error"
	StartedAt    time.Time              `json:"started_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	RequestCount int64                  `json:"request_count"`
	ErrorCount   int64                  `json:"error_count"`
	LastRequest  time.Time              `json:"last_request"`
	LastError    string                 `json:"last_error"`
	HealthStatus string                 `json:"health_status"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// HealthChecker performs health checks for agents
type HealthChecker struct {
	AgentID          string
	Config           *HealthCheckConfig
	LastCheck        time.Time
	Status           string
	ConsecutiveFails int
	Logger           *logrus.Logger
}

// MultiAgentMetrics tracks metrics for multi-agent system
type MultiAgentMetrics struct {
	TotalRequests  int64                    `json:"total_requests"`
	TotalErrors    int64                    `json:"total_errors"`
	AgentMetrics   map[string]*AgentMetrics `json:"agent_metrics"`
	RoutingMetrics *RoutingMetrics          `json:"routing_metrics"`
	LastUpdated    time.Time                `json:"last_updated"`
	mu             sync.RWMutex
}

// AgentMetrics tracks metrics for individual agents
type AgentMetrics struct {
	RequestCount   int64         `json:"request_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	LastRequest    time.Time     `json:"last_request"`
	TotalLatency   time.Duration `json:"total_latency"`
}

// RoutingMetrics tracks routing statistics
type RoutingMetrics struct {
	RoutingDecisions map[string]int64 `json:"routing_decisions"`
	DefaultRoutes    int64            `json:"default_routes"`
	FailedRoutes     int64            `json:"failed_routes"`
}

// NewMultiAgentManager creates a new multi-agent manager
func NewMultiAgentManager(config *MultiAgentConfig, llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) (*MultiAgentManager, error) {
	if config == nil {
		return nil, fmt.Errorf("multi-agent config is required")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid multi-agent config: %w", err)
	}

	manager := &MultiAgentManager{
		config:         config,
		agents:         make(map[string]*Agent),
		llmManager:     llmManager,
		toolRegistry:   toolRegistry,
		router:         mux.NewRouter(),
		middleware:     []MiddlewareFunc{},
		healthCheckers: make(map[string]*HealthChecker),
		logger:         logrus.New(),
		deploymentState: &DeploymentState{
			Status:      "initialized",
			StartedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			AgentStates: make(map[string]*AgentState),
			Metadata:    make(map[string]interface{}),
		},
		metrics: &MultiAgentMetrics{
			AgentMetrics: make(map[string]*AgentMetrics),
			RoutingMetrics: &RoutingMetrics{
				RoutingDecisions: make(map[string]int64),
			},
			LastUpdated: time.Now(),
		},
	}

	// Initialize agents
	if err := manager.initializeAgents(); err != nil {
		return nil, fmt.Errorf("failed to initialize agents: %w", err)
	}

	// Setup routing
	if err := manager.setupRouting(); err != nil {
		return nil, fmt.Errorf("failed to setup routing: %w", err)
	}

	// Setup health checking
	if err := manager.setupHealthChecking(); err != nil {
		return nil, fmt.Errorf("failed to setup health checking: %w", err)
	}

	return manager, nil
}

// initializeAgents creates and initializes all agents
func (mam *MultiAgentManager) initializeAgents() error {
	mam.mu.Lock()
	defer mam.mu.Unlock()

	registry := GetGlobalRegistry()

	for agentID, agentConfig := range mam.config.Agents {
		// Ensure agent has an ID
		if agentConfig.ID == "" {
			agentConfig.ID = agentID
		}

		// Create agent instance
		var agent *Agent
		var err error

		// Check if agent is defined programmatically first
		if _, exists := registry.GetDefinition(agentID); exists {
			agent, err = registry.CreateAgentFromDefinition(agentID, mam.llmManager, mam.toolRegistry)
			if err != nil {
				return fmt.Errorf("failed to create agent %s from definition: %w", agentID, err)
			}
			mam.logger.WithField("agent_id", agentID).Info("Agent created from definition")
		} else {
			// Check for factory-based creation
			factoryIDs := registry.ListFactories()

			isFactory := false
			for _, id := range factoryIDs {
				if id == agentID {
					isFactory = true
					break
				}
			}

			if isFactory {
				agent, err = registry.CreateAgentFromFactory(agentID, mam.llmManager, mam.toolRegistry)
				if err != nil {
					return fmt.Errorf("failed to create agent %s from factory: %w", agentID, err)
				}
				mam.logger.WithField("agent_id", agentID).Info("Agent created from factory")
			} else {
				// Fall back to config-based agent creation
				agent = NewAgent(agentConfig, mam.llmManager, mam.toolRegistry)
				mam.logger.WithField("agent_id", agentID).Info("Agent created from config")
			}
		}

		mam.agents[agentID] = agent

		// Initialize agent state
		mam.deploymentState.AgentStates[agentID] = &AgentState{
			ID:           agentID,
			Status:       "initialized",
			StartedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			RequestCount: 0,
			ErrorCount:   0,
			HealthStatus: "unknown",
			Metadata:     make(map[string]interface{}),
		}

		// Initialize agent metrics
		mam.metrics.AgentMetrics[agentID] = &AgentMetrics{
			RequestCount:   0,
			ErrorCount:     0,
			AverageLatency: 0,
			TotalLatency:   0,
		}

		mam.logger.WithField("agent_id", agentID).Info("Agent initialized")
	}

	return nil
}

// setupRouting configures HTTP routing for multi-agent requests
func (mam *MultiAgentManager) setupRouting() error {
	// Setup global middleware
	for _, middlewareConfig := range mam.config.Routing.Middleware {
		if middlewareConfig.Enabled {
			middleware := mam.createMiddleware(middlewareConfig)
			if middleware != nil {
				mam.middleware = append(mam.middleware, middleware)
			}
		}
	}

	// Apply middleware to router
	for _, middleware := range mam.middleware {
		mam.router.Use(mux.MiddlewareFunc(middleware))
	}

	// Add metrics middleware
	mam.router.Use(mux.MiddlewareFunc(mam.metricsMiddleware))

	// Sort routing rules by priority
	rules := make([]RoutingRule, len(mam.config.Routing.Rules))
	copy(rules, mam.config.Routing.Rules)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	// Setup routing rules
	for _, rule := range rules {
		mam.setupRoutingRule(rule)
	}

	// Setup default route if configured
	if mam.config.Routing.DefaultAgent != "" {
		mam.router.PathPrefix("/").HandlerFunc(mam.createAgentHandler(mam.config.Routing.DefaultAgent, true))
	}

	// Setup management endpoints
	mam.setupManagementEndpoints()

	return nil
}

// setupRoutingRule sets up a single routing rule
func (mam *MultiAgentManager) setupRoutingRule(rule RoutingRule) {
	handler := mam.createAgentHandler(rule.AgentID, false)

	var route *mux.Route
	switch mam.config.Routing.Type {
	case "path":
		route = mam.router.Path(rule.Pattern)
	case "host":
		route = mam.router.Host(rule.Pattern)
	case "header":
		// Extract header key and value from pattern
		parts := strings.Split(rule.Pattern, ":")
		if len(parts) == 2 {
			route = mam.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
				return r.Header.Get(parts[0]) == parts[1]
			})
		}
	case "query":
		// Extract query key and value from pattern
		parts := strings.Split(rule.Pattern, "=")
		if len(parts) == 2 {
			route = mam.router.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
				return r.URL.Query().Get(parts[0]) == parts[1]
			})
		}
	default:
		route = mam.router.Path(rule.Pattern)
	}

	if route != nil {
		if rule.Method != "" {
			route = route.Methods(rule.Method)
		}
		route.Handler(handler)

		mam.logger.WithFields(logrus.Fields{
			"rule_id":  rule.ID,
			"pattern":  rule.Pattern,
			"agent_id": rule.AgentID,
			"method":   rule.Method,
			"priority": rule.Priority,
		}).Info("Routing rule configured")
	}
}

// createAgentHandler creates an HTTP handler for a specific agent
func (mam *MultiAgentManager) createAgentHandler(agentID string, isDefault bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get agent
		agent, exists := mam.getAgent(agentID)
		if !exists {
			mam.recordMetrics(agentID, time.Since(start), true)
			http.Error(w, fmt.Sprintf("Agent %s not found", agentID), http.StatusNotFound)
			return
		}

		// Update routing metrics
		mam.updateRoutingMetrics(agentID, isDefault)

		// Parse request
		var input string
		switch r.Method {
		case "GET":
			input = r.URL.Query().Get("input")
			if input == "" {
				input = r.URL.Query().Get("q")
			}
		case "POST":
			var requestData struct {
				Input string `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
				mam.recordMetrics(agentID, time.Since(start), true)
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			input = requestData.Input
		default:
			mam.recordMetrics(agentID, time.Since(start), true)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if input == "" {
			mam.recordMetrics(agentID, time.Since(start), true)
			http.Error(w, "Input is required", http.StatusBadRequest)
			return
		}

		// Execute agent
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		execution, err := agent.Execute(ctx, input)
		if err != nil {
			mam.recordMetrics(agentID, time.Since(start), true)
			mam.updateAgentError(agentID, err)
			http.Error(w, fmt.Sprintf("Agent execution failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Record successful metrics
		mam.recordMetrics(agentID, time.Since(start), false)
		mam.updateAgentSuccess(agentID)

		// Return response
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"agent_id":  agentID,
			"execution": execution,
			"timestamp": time.Now(),
		}
		json.NewEncoder(w).Encode(response)
	}
}

// setupManagementEndpoints sets up management and monitoring endpoints
func (mam *MultiAgentManager) setupManagementEndpoints() {
	// Health check endpoint
	mam.router.HandleFunc("/health", mam.handleHealth).Methods("GET")
	mam.router.HandleFunc("/health/{agent_id}", mam.handleAgentHealth).Methods("GET")

	// Metrics endpoint
	mam.router.HandleFunc("/metrics", mam.handleMetrics).Methods("GET")

	// Agent management endpoints
	mam.router.HandleFunc("/agents", mam.handleListAgents).Methods("GET")
	mam.router.HandleFunc("/agents/{agent_id}", mam.handleGetAgent).Methods("GET")
	mam.router.HandleFunc("/agents/{agent_id}/status", mam.handleAgentStatus).Methods("GET")

	// Configuration endpoints
	mam.router.HandleFunc("/config", mam.handleGetConfig).Methods("GET")
	mam.router.HandleFunc("/routing", mam.handleGetRouting).Methods("GET")

	// Deployment endpoints
	mam.router.HandleFunc("/deployment/status", mam.handleDeploymentStatus).Methods("GET")
	mam.router.HandleFunc("/deployment/restart", mam.handleRestart).Methods("POST")
}

// setupHealthChecking initializes health checking for agents
func (mam *MultiAgentManager) setupHealthChecking() error {
	if mam.config.Deployment == nil || mam.config.Deployment.HealthCheck == nil || !mam.config.Deployment.HealthCheck.Enabled {
		return nil
	}

	mam.healthMu.Lock()
	defer mam.healthMu.Unlock()

	for agentID := range mam.config.Agents {
		healthConfig := mam.config.Deployment.HealthCheck

		// Check for agent-specific health check config
		if agentSpecific, exists := healthConfig.AgentSpecific[agentID]; exists {
			healthConfig = agentSpecific
		}

		checker := &HealthChecker{
			AgentID: agentID,
			Config:  healthConfig,
			Status:  "unknown",
			Logger:  logrus.New(),
		}

		mam.healthCheckers[agentID] = checker

		// Start health checking goroutine
		go mam.runHealthChecker(checker)
	}

	return nil
}

// runHealthChecker runs health checks for an agent
func (mam *MultiAgentManager) runHealthChecker(checker *HealthChecker) {
	ticker := time.NewTicker(time.Duration(checker.Config.PeriodSeconds) * time.Second)
	defer ticker.Stop()

	// Initial delay
	time.Sleep(time.Duration(checker.Config.InitialDelaySeconds) * time.Second)

	for range ticker.C {
		mam.performHealthCheck(checker)
	}
}

// performHealthCheck performs a single health check
func (mam *MultiAgentManager) performHealthCheck(checker *HealthChecker) {
	checker.LastCheck = time.Now()

	// Simple health check - in a real implementation, this would make HTTP requests
	agent, exists := mam.getAgent(checker.AgentID)
	if !exists {
		checker.Status = "not_found"
		checker.ConsecutiveFails++
		mam.updateAgentHealthStatus(checker.AgentID, "unhealthy")
		return
	}

	// Check if agent is responsive (simplified check)
	if agent.IsRunning() {
		checker.Status = "healthy"
		checker.ConsecutiveFails = 0
		mam.updateAgentHealthStatus(checker.AgentID, "healthy")
	} else {
		checker.Status = "unhealthy"
		checker.ConsecutiveFails++
		mam.updateAgentHealthStatus(checker.AgentID, "unhealthy")
	}

	// Log health status changes
	if checker.ConsecutiveFails == checker.Config.FailureThreshold {
		checker.Logger.Warn("Agent health check failing consistently")
	} else if checker.ConsecutiveFails == 0 && checker.Status == "healthy" {
		checker.Logger.Info("Agent health check recovered")
	}
}

// Middleware creation
func (mam *MultiAgentManager) createMiddleware(config MiddlewareConfig) MiddlewareFunc {
	switch config.Type {
	case "cors":
		return mam.corsMiddleware
	case "auth":
		return mam.authMiddleware
	case "logging":
		return mam.loggingMiddleware
	case "rate_limit":
		return mam.rateLimitMiddleware
	default:
		mam.logger.WithField("type", config.Type).Warn("Unknown middleware type")
		return nil
	}
}

// Middleware implementations
func (mam *MultiAgentManager) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mam.config.Shared.Security.CORS.Enabled {
			cors := mam.config.Shared.Security.CORS

			origin := r.Header.Get("Origin")
			if origin != "" && mam.isAllowedOrigin(origin, cors.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cors.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cors.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(cors.ExposedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cors.MaxAge))

			if cors.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (mam *MultiAgentManager) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple API key authentication
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.URL.Query().Get("api_key")
		}

		// In a real implementation, validate the API key
		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (mam *MultiAgentManager) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		next.ServeHTTP(w, r)

		mam.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": time.Since(startTime),
			"remote":   r.RemoteAddr,
		}).Info("HTTP request")
	})
}

func (mam *MultiAgentManager) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple rate limiting - in a real implementation, use a proper rate limiter
		next.ServeHTTP(w, r)
	})
}

func (mam *MultiAgentManager) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = time.Now() // Could be used for request duration tracking

		next.ServeHTTP(w, r)

		// Update global metrics
		mam.metrics.mu.Lock()
		mam.metrics.TotalRequests++
		mam.metrics.LastUpdated = time.Now()
		mam.metrics.mu.Unlock()
	})
}

// Helper methods
func (mam *MultiAgentManager) getAgent(agentID string) (*Agent, bool) {
	mam.mu.RLock()
	defer mam.mu.RUnlock()

	agent, exists := mam.agents[agentID]
	return agent, exists
}

func (mam *MultiAgentManager) isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (mam *MultiAgentManager) recordMetrics(agentID string, duration time.Duration, isError bool) {
	mam.metrics.mu.Lock()
	defer mam.metrics.mu.Unlock()

	if agentMetrics, exists := mam.metrics.AgentMetrics[agentID]; exists {
		agentMetrics.RequestCount++
		agentMetrics.LastRequest = time.Now()
		agentMetrics.TotalLatency += duration
		agentMetrics.AverageLatency = agentMetrics.TotalLatency / time.Duration(agentMetrics.RequestCount)

		if isError {
			agentMetrics.ErrorCount++
			mam.metrics.TotalErrors++
		}
	}
}

func (mam *MultiAgentManager) updateRoutingMetrics(agentID string, isDefault bool) {
	mam.metrics.mu.Lock()
	defer mam.metrics.mu.Unlock()

	if isDefault {
		mam.metrics.RoutingMetrics.DefaultRoutes++
	} else {
		mam.metrics.RoutingMetrics.RoutingDecisions[agentID]++
	}
}

func (mam *MultiAgentManager) updateAgentError(agentID string, err error) {
	mam.mu.Lock()
	defer mam.mu.Unlock()

	if state, exists := mam.deploymentState.AgentStates[agentID]; exists {
		state.ErrorCount++
		state.LastError = err.Error()
		state.UpdatedAt = time.Now()
	}
}

func (mam *MultiAgentManager) updateAgentSuccess(agentID string) {
	mam.mu.Lock()
	defer mam.mu.Unlock()

	if state, exists := mam.deploymentState.AgentStates[agentID]; exists {
		state.RequestCount++
		state.LastRequest = time.Now()
		state.UpdatedAt = time.Now()
		state.Status = "running"
	}
}

func (mam *MultiAgentManager) updateAgentHealthStatus(agentID, status string) {
	mam.mu.Lock()
	defer mam.mu.Unlock()

	if state, exists := mam.deploymentState.AgentStates[agentID]; exists {
		state.HealthStatus = status
		state.UpdatedAt = time.Now()
	}
}

// HTTP Handlers
func (mam *MultiAgentManager) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"agents":    make(map[string]string),
	}

	mam.mu.RLock()
	for agentID, state := range mam.deploymentState.AgentStates {
		health["agents"].(map[string]string)[agentID] = state.HealthStatus
	}
	mam.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (mam *MultiAgentManager) handleAgentHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	mam.mu.RLock()
	state, exists := mam.deploymentState.AgentStates[agentID]
	mam.mu.RUnlock()

	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	health := map[string]interface{}{
		"agent_id":      agentID,
		"status":        state.HealthStatus,
		"timestamp":     time.Now(),
		"request_count": state.RequestCount,
		"error_count":   state.ErrorCount,
		"last_request":  state.LastRequest,
		"started_at":    state.StartedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (mam *MultiAgentManager) handleMetrics(w http.ResponseWriter, r *http.Request) {
	mam.metrics.mu.RLock()
	defer mam.metrics.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mam.metrics)
}

func (mam *MultiAgentManager) handleListAgents(w http.ResponseWriter, r *http.Request) {
	mam.mu.RLock()
	agents := make(map[string]interface{})
	for agentID, state := range mam.deploymentState.AgentStates {
		agents[agentID] = map[string]interface{}{
			"status":        state.Status,
			"health_status": state.HealthStatus,
			"request_count": state.RequestCount,
			"error_count":   state.ErrorCount,
			"started_at":    state.StartedAt,
		}
	}
	mam.mu.RUnlock()

	response := map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (mam *MultiAgentManager) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	agent, exists := mam.getAgent(agentID)
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	mam.mu.RLock()
	state := mam.deploymentState.AgentStates[agentID]
	mam.mu.RUnlock()

	response := map[string]interface{}{
		"agent_id": agentID,
		"config":   agent.GetConfig(),
		"state":    state,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (mam *MultiAgentManager) handleAgentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	mam.mu.RLock()
	state, exists := mam.deploymentState.AgentStates[agentID]
	mam.mu.RUnlock()

	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (mam *MultiAgentManager) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mam.config)
}

func (mam *MultiAgentManager) handleGetRouting(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mam.config.Routing)
}

func (mam *MultiAgentManager) handleDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	mam.mu.RLock()
	state := *mam.deploymentState
	mam.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

func (mam *MultiAgentManager) handleRestart(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would restart agents
	mam.logger.Info("Restart requested")

	response := map[string]interface{}{
		"status":    "restart_initiated",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Public methods

// Start starts the multi-agent manager
func (mam *MultiAgentManager) Start(ctx context.Context) error {
	mam.logger.Info("Starting multi-agent manager")

	mam.mu.Lock()
	mam.deploymentState.Status = "starting"
	mam.deploymentState.UpdatedAt = time.Now()

	// Start all agents
	for agentID := range mam.agents {
		if state, exists := mam.deploymentState.AgentStates[agentID]; exists {
			state.Status = "starting"
			state.UpdatedAt = time.Now()
		}
	}
	mam.mu.Unlock()

	// Mark as running
	mam.mu.Lock()
	mam.deploymentState.Status = "running"
	mam.deploymentState.UpdatedAt = time.Now()

	for agentID := range mam.agents {
		if state, exists := mam.deploymentState.AgentStates[agentID]; exists {
			state.Status = "running"
			state.UpdatedAt = time.Now()
		}
	}
	mam.mu.Unlock()

	mam.logger.Info("Multi-agent manager started successfully")
	return nil
}

// Stop stops the multi-agent manager
func (mam *MultiAgentManager) Stop(ctx context.Context) error {
	mam.logger.Info("Stopping multi-agent manager")

	mam.mu.Lock()
	mam.deploymentState.Status = "stopping"
	mam.deploymentState.UpdatedAt = time.Now()

	// Stop all agents
	for agentID := range mam.agents {
		if state, exists := mam.deploymentState.AgentStates[agentID]; exists {
			state.Status = "stopping"
			state.UpdatedAt = time.Now()
		}
	}
	mam.mu.Unlock()

	// Mark as stopped
	mam.mu.Lock()
	mam.deploymentState.Status = "stopped"
	mam.deploymentState.UpdatedAt = time.Now()

	for agentID := range mam.agents {
		if state, exists := mam.deploymentState.AgentStates[agentID]; exists {
			state.Status = "stopped"
			state.UpdatedAt = time.Now()
		}
	}
	mam.mu.Unlock()

	mam.logger.Info("Multi-agent manager stopped")
	return nil
}

// GetRouter returns the HTTP router
func (mam *MultiAgentManager) GetRouter() *mux.Router {
	return mam.router
}

// GetConfig returns the multi-agent configuration
func (mam *MultiAgentManager) GetConfig() *MultiAgentConfig {
	return mam.config
}

// GetMetrics returns current metrics
func (mam *MultiAgentManager) GetMetrics() *MultiAgentMetrics {
	mam.metrics.mu.RLock()
	defer mam.metrics.mu.RUnlock()

	// Create a copy of metrics without copying the mutex
	metricsCopy := MultiAgentMetrics{
		TotalRequests: mam.metrics.TotalRequests,
		TotalErrors:   mam.metrics.TotalErrors,
		AgentMetrics:  make(map[string]*AgentMetrics),
		RoutingMetrics: &RoutingMetrics{
			RoutingDecisions: make(map[string]int64),
			DefaultRoutes:    mam.metrics.RoutingMetrics.DefaultRoutes,
			FailedRoutes:     mam.metrics.RoutingMetrics.FailedRoutes,
		},
		LastUpdated: mam.metrics.LastUpdated,
	}

	// Copy agent metrics
	for k, v := range mam.metrics.AgentMetrics {
		metricsCopy.AgentMetrics[k] = &AgentMetrics{
			RequestCount:   v.RequestCount,
			ErrorCount:     v.ErrorCount,
			AverageLatency: v.AverageLatency,
			LastRequest:    v.LastRequest,
			TotalLatency:   v.TotalLatency,
		}
	}

	// Copy routing decisions
	for k, v := range mam.metrics.RoutingMetrics.RoutingDecisions {
		metricsCopy.RoutingMetrics.RoutingDecisions[k] = v
	}

	return &metricsCopy
}

// GetDeploymentState returns current deployment state
func (mam *MultiAgentManager) GetDeploymentState() *DeploymentState {
	mam.mu.RLock()
	defer mam.mu.RUnlock()

	state := *mam.deploymentState
	return &state
}

// LoadConfigFromFile loads multi-agent configuration from a file
func LoadMultiAgentConfigFromFile(filename string) (*MultiAgentConfig, error) {
	data, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	content, err := os.ReadFile(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config MultiAgentConfig
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}
