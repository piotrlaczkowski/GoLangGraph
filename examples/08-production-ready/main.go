// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Production-Ready Example

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration structure
type Config struct {
	Server struct {
		Host         string `mapstructure:"host"`
		Port         int    `mapstructure:"port"`
		ReadTimeout  int    `mapstructure:"read_timeout"`
		WriteTimeout int    `mapstructure:"write_timeout"`
	} `mapstructure:"server"`

	Ollama struct {
		Endpoint string `mapstructure:"endpoint"`
		Model    string `mapstructure:"model"`
		Timeout  int    `mapstructure:"timeout"`
	} `mapstructure:"ollama"`

	Logging struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"logging"`

	Metrics struct {
		Enabled bool   `mapstructure:"enabled"`
		Path    string `mapstructure:"path"`
	} `mapstructure:"metrics"`
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// ChatRequest represents an API request
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
}

// ChatResponse represents an API response
type ChatResponse struct {
	Response  string `json:"response"`
	SessionID string `json:"session_id"`
	Timestamp string `json:"timestamp"`
	Duration  string `json:"duration"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks"`
}

// ProductionAgent represents the production-ready agent
type ProductionAgent struct {
	config     *Config
	logger     *logrus.Logger
	httpClient *http.Client
	startTime  time.Time
	sessions   map[string][]string
	sessionMu  sync.RWMutex

	// Metrics
	requestsTotal     prometheus.Counter
	requestDuration   prometheus.Histogram
	activeConnections prometheus.Gauge
}

// Metrics definitions
var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "golanggraph_requests_total",
			Help: "Total number of requests processed",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "golanggraph_request_duration_seconds",
			Help: "Request duration in seconds",
		},
		[]string{"method", "endpoint"},
	)

	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "golanggraph_active_connections",
			Help: "Number of active connections",
		},
	)
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in this example
	},
}

func main() {
	fmt.Println("🚀 GoLangGraph Production-Ready Agent")
	fmt.Println("=====================================")
	fmt.Println()

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := setupLogger(config)
	logger.Info("Starting GoLangGraph Production Agent")

	// Initialize metrics
	if config.Metrics.Enabled {
		prometheus.MustRegister(requestsTotal, requestDuration, activeConnections)
		logger.Info("Metrics enabled")
	}

	// Create production agent
	agent := &ProductionAgent{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Ollama.Timeout) * time.Second,
		},
		startTime: time.Now(),
		sessions:  make(map[string][]string),
	}

	// Setup HTTP server
	server := setupServer(agent)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
		logger.Infof("Server starting on %s", addr)
		fmt.Printf("🌐 Server running on http://%s\n", addr)
		fmt.Printf("📊 Health check: http://%s/health\n", addr)
		if config.Metrics.Enabled {
			fmt.Printf("📈 Metrics: http://%s%s\n", addr, config.Metrics.Path)
		}
		fmt.Printf("🔌 WebSocket: ws://%s/ws\n", addr)
		fmt.Println()
		fmt.Println("Press Ctrl+C to stop the server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down server...")
	fmt.Println("\n🛑 Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	} else {
		logger.Info("Server shutdown complete")
		fmt.Println("✅ Server shutdown complete")
	}
}

// loadConfig loads configuration from file and environment
func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("ollama.endpoint", "http://localhost:11434")
	viper.SetDefault("ollama.model", "gemma3:1b")
	viper.SetDefault("ollama.timeout", 30)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")

	// Enable environment variables
	viper.SetEnvPrefix("GOLANGGRAPH")
	viper.AutomaticEnv()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setupLogger configures the logger
func setupLogger(config *Config) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if config.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return logger
}

// setupServer configures the HTTP server
func setupServer(agent *ProductionAgent) *http.Server {
	router := mux.NewRouter()

	// Middleware
	router.Use(agent.loggingMiddleware)
	router.Use(agent.metricsMiddleware)
	router.Use(agent.corsMiddleware)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/chat", agent.chatHandler).Methods("POST")
	api.HandleFunc("/sessions", agent.listSessionsHandler).Methods("GET")
	api.HandleFunc("/sessions/{id}", agent.getSessionHandler).Methods("GET")
	api.HandleFunc("/sessions/{id}", agent.deleteSessionHandler).Methods("DELETE")

	// WebSocket endpoint
	router.HandleFunc("/ws", agent.websocketHandler)

	// Health check
	router.HandleFunc("/health", agent.healthHandler).Methods("GET")

	// Metrics endpoint
	if agent.config.Metrics.Enabled {
		router.Handle(agent.config.Metrics.Path, promhttp.Handler())
	}

	// Static files (for demo purposes)
	router.HandleFunc("/", agent.indexHandler).Methods("GET")

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", agent.config.Server.Host, agent.config.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(agent.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(agent.config.Server.WriteTimeout) * time.Second,
	}
}

// Middleware functions

func (p *ProductionAgent) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		p.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapper.statusCode,
			"duration":   duration,
			"user_agent": r.UserAgent(),
			"remote_ip":  r.RemoteAddr,
		}).Info("Request processed")
	})
}

func (p *ProductionAgent) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.config.Metrics.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		wrapper := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		// Record metrics
		requestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", wrapper.statusCode)).Inc()
		requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}

func (p *ProductionAgent) corsMiddleware(next http.Handler) http.Handler {
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

// Response writer wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Handler functions

func (p *ProductionAgent) chatHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		p.writeError(w, "Message is required", http.StatusBadRequest)
		return
	}

	// Generate session ID if not provided
	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}

	start := time.Now()

	// Call Ollama API
	response, err := p.callOllama(req.Message, req.SessionID)
	if err != nil {
		p.logger.WithError(err).Error("Failed to call Ollama API")
		p.writeError(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	duration := time.Since(start)

	// Store in session
	p.sessionMu.Lock()
	if p.sessions[req.SessionID] == nil {
		p.sessions[req.SessionID] = make([]string, 0)
	}
	p.sessions[req.SessionID] = append(p.sessions[req.SessionID],
		fmt.Sprintf("User: %s", req.Message),
		fmt.Sprintf("Assistant: %s", response))
	p.sessionMu.Unlock()

	// Send response
	chatResp := ChatResponse{
		Response:  response,
		SessionID: req.SessionID,
		Timestamp: time.Now().Format(time.RFC3339),
		Duration:  duration.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

func (p *ProductionAgent) listSessionsHandler(w http.ResponseWriter, r *http.Request) {
	p.sessionMu.RLock()
	sessions := make(map[string]int)
	for id, messages := range p.sessions {
		sessions[id] = len(messages)
	}
	p.sessionMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (p *ProductionAgent) getSessionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	p.sessionMu.RLock()
	messages, exists := p.sessions[sessionID]
	p.sessionMu.RUnlock()

	if !exists {
		p.writeError(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": sessionID,
		"messages":   messages,
	})
}

func (p *ProductionAgent) deleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	p.sessionMu.Lock()
	delete(p.sessions, sessionID)
	p.sessionMu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (p *ProductionAgent) websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	activeConnections.Inc()
	defer activeConnections.Dec()

	p.logger.Info("WebSocket connection established")

	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			p.logger.WithError(err).Error("Failed to read WebSocket message")
			break
		}

		message, ok := msg["message"].(string)
		if !ok {
			conn.WriteJSON(map[string]string{"error": "Invalid message format"})
			continue
		}

		sessionID, _ := msg["session_id"].(string)
		if sessionID == "" {
			sessionID = fmt.Sprintf("ws_session_%d", time.Now().UnixNano())
		}

		// Process message
		response, err := p.callOllama(message, sessionID)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": err.Error()})
			continue
		}

		// Send response
		conn.WriteJSON(map[string]interface{}{
			"response":   response,
			"session_id": sessionID,
			"timestamp":  time.Now().Format(time.RFC3339),
		})
	}

	p.logger.Info("WebSocket connection closed")
}

func (p *ProductionAgent) healthHandler(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)

	// Check Ollama connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.Ollama.Endpoint+"/api/tags", nil)
	if err != nil {
		checks["ollama"] = "error"
	} else {
		resp, err := p.httpClient.Do(req)
		if err != nil || resp.StatusCode != 200 {
			checks["ollama"] = "unhealthy"
		} else {
			checks["ollama"] = "healthy"
			resp.Body.Close()
		}
	}

	// Overall status
	status := "healthy"
	for _, check := range checks {
		if check != "healthy" {
			status = "unhealthy"
			break
		}
	}

	health := HealthResponse{
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		Uptime:    time.Since(p.startTime).String(),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(health)
}

func (p *ProductionAgent) indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>GoLangGraph Production Agent</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
        .method { font-weight: bold; color: #007bff; }
        pre { background: #f8f9fa; padding: 10px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 GoLangGraph Production Agent</h1>
        <p>Welcome to the production-ready GoLangGraph agent API.</p>
        
        <h2>API Endpoints</h2>
        
        <div class="endpoint">
            <span class="method">POST</span> /api/v1/chat
            <p>Send a chat message to the agent</p>
            <pre>{"message": "Hello", "session_id": "optional"}</pre>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> /api/v1/sessions
            <p>List all active sessions</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> /api/v1/sessions/{id}
            <p>Get session messages</p>
        </div>
        
        <div class="endpoint">
            <span class="method">DELETE</span> /api/v1/sessions/{id}
            <p>Delete a session</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> /health
            <p>Health check endpoint</p>
        </div>
        
        <div class="endpoint">
            <span class="method">GET</span> /metrics
            <p>Prometheus metrics endpoint</p>
        </div>
        
        <div class="endpoint">
            <span class="method">WebSocket</span> /ws
            <p>WebSocket connection for real-time chat</p>
        </div>
        
        <h2>Example Usage</h2>
        <pre>
# Send a chat message
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hello, how are you?"}'

# Check health
curl http://localhost:8080/health

# List sessions
curl http://localhost:8080/api/v1/sessions
        </pre>
    </div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// callOllama makes a request to the Ollama API
func (p *ProductionAgent) callOllama(message, sessionID string) (string, error) {
	// Build context with session history
	var conversationContext string
	p.sessionMu.RLock()
	if messages, exists := p.sessions[sessionID]; exists {
		// Add last few messages for context
		start := len(messages) - 6
		if start < 0 {
			start = 0
		}
		for i := start; i < len(messages); i++ {
			conversationContext += messages[i] + "\n"
		}
	}
	p.sessionMu.RUnlock()

	prompt := "You are a helpful AI assistant. Provide clear and concise responses.\n\n"
	if conversationContext != "" {
		prompt += "Previous conversation:\n" + conversationContext + "\n"
	}
	prompt += "User: " + message + "\nAssistant:"

	reqBody := OllamaRequest{
		Model:  p.config.Ollama.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(p.config.Ollama.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		p.config.Ollama.Endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ollamaResp.Response, nil
}

// writeError writes an error response
func (p *ProductionAgent) writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResp := ErrorResponse{
		Error:   http.StatusText(code),
		Code:    code,
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResp)
}

func init() {
	// Set up structured logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
}
