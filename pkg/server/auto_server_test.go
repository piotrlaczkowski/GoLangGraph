// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
)

// Test configuration creation
func TestDefaultAutoServerConfig(t *testing.T) {
	config := DefaultAutoServerConfig()

	if config == nil {
		t.Fatal("DefaultAutoServerConfig returned nil")
	}
	if config.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Host)
	}
	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}
	if config.BasePath != "/api" {
		t.Errorf("Expected base path /api, got %s", config.BasePath)
	}
	if !config.EnableWebUI {
		t.Error("Expected EnableWebUI to be true")
	}
	if !config.EnablePlayground {
		t.Error("Expected EnablePlayground to be true")
	}
	if !config.EnableSchemaAPI {
		t.Error("Expected EnableSchemaAPI to be true")
	}
	if !config.EnableMetricsAPI {
		t.Error("Expected EnableMetricsAPI to be true")
	}
	if !config.EnableCORS {
		t.Error("Expected EnableCORS to be true")
	}
	if !config.SchemaValidation {
		t.Error("Expected SchemaValidation to be true")
	}
	if config.OllamaEndpoint != "http://localhost:11434" {
		t.Errorf("Expected ollama endpoint http://localhost:11434, got %s", config.OllamaEndpoint)
	}
	if config.ServerTimeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.ServerTimeout)
	}
	if config.MaxRequestSize != int64(10*1024*1024) {
		t.Errorf("Expected max request size 10MB, got %d", config.MaxRequestSize)
	}
}

// Test server creation
func TestNewAutoServer(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		server := NewAutoServer(nil)

		if server == nil {
			t.Fatal("NewAutoServer returned nil")
		}
		if server.config == nil {
			t.Error("Server config should not be nil")
		}
		if server.router == nil {
			t.Error("Server router should not be nil")
		}
		if server.logger == nil {
			t.Error("Server logger should not be nil")
		}
		if server.registry == nil {
			t.Error("Server registry should not be nil")
		}
		if server.llmManager == nil {
			t.Error("Server llmManager should not be nil")
		}
		if server.toolRegistry == nil {
			t.Error("Server toolRegistry should not be nil")
		}
		if server.agentInstances == nil {
			t.Error("Server agentInstances should not be nil")
		}
		if server.agentMetadata == nil {
			t.Error("Server agentMetadata should not be nil")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &AutoServerConfig{
			Host:             "127.0.0.1",
			Port:             9090,
			BasePath:         "/custom",
			EnableWebUI:      false,
			EnablePlayground: false,
			SchemaValidation: false,
		}

		server := NewAutoServer(config)

		if server == nil {
			t.Fatal("NewAutoServer returned nil")
		}
		if server.config.Host != "127.0.0.1" {
			t.Errorf("Expected host 127.0.0.1, got %s", server.config.Host)
		}
		if server.config.Port != 9090 {
			t.Errorf("Expected port 9090, got %d", server.config.Port)
		}
		if server.config.BasePath != "/custom" {
			t.Errorf("Expected base path /custom, got %s", server.config.BasePath)
		}
		if server.config.EnableWebUI {
			t.Error("Expected EnableWebUI to be false")
		}
		if server.config.EnablePlayground {
			t.Error("Expected EnablePlayground to be false")
		}
		if server.config.SchemaValidation {
			t.Error("Expected SchemaValidation to be false")
		}
	})
}

// Test agent registration
func TestAutoServerAgentRegistration(t *testing.T) {
	server := NewAutoServer(nil)

	// Create a test agent config
	agentConfig := &agent.AgentConfig{
		Name:         "TestAgent",
		Type:         agent.AgentTypeChat,
		Model:        "llama3.2",
		Provider:     "ollama",
		Temperature:  0.7,
		MaxTokens:    2048,
		SystemPrompt: "You are a test agent.",
		Tools:        []string{},
	}

	// Create agent definition from config
	agentDefinition := agent.NewBaseAgentDefinition(agentConfig)

	// Register the agent
	err := server.RegisterAgent("test", agentDefinition)
	if err != nil {
		t.Errorf("Expected no error registering agent, got %v", err)
	}

	// Check agent was registered in the global registry
	registry := agent.GetGlobalRegistry()
	_, exists := registry.GetDefinition("test")
	if !exists {
		t.Error("Agent should be registered in global registry")
	}

	// Test duplicate registration (should error)
	err = server.RegisterAgent("test", agentDefinition)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

// Test endpoint generation
func TestAutoServerEndpointGeneration(t *testing.T) {
	server := NewAutoServer(nil)

	// Register a test agent
	agentConfig := &agent.AgentConfig{
		Name:     "TestAgent",
		Type:     agent.AgentTypeChat,
		Model:    "llama3.2",
		Provider: "ollama",
	}
	agentDefinition := agent.NewBaseAgentDefinition(agentConfig)
	server.RegisterAgent("test", agentDefinition)

	// Generate endpoints
	err := server.GenerateEndpoints()
	if err != nil {
		t.Errorf("Expected no error generating endpoints, got %v", err)
	}

	// Test various endpoint paths exist by making requests
	routes := []string{
		"/health",
		"/capabilities",
		"/agents",
		"/agents/test",
	}

	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		// Most endpoints should exist (not 404)
		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s should exist, got 404", route)
		}
	}
}

// Test health endpoint
func TestAutoServerHealthEndpoint(t *testing.T) {
	server := NewAutoServer(nil)
	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
	if _, exists := response["timestamp"]; !exists {
		t.Error("Response should contain timestamp")
	}
	if _, exists := response["agents"]; !exists {
		t.Error("Response should contain agents")
	}
	if _, exists := response["agent_count"]; !exists {
		t.Error("Response should contain agent_count")
	}
	if _, exists := response["schema_validation"]; !exists {
		t.Error("Response should contain schema_validation")
	}
	if _, exists := response["ollama_endpoint"]; !exists {
		t.Error("Response should contain ollama_endpoint")
	}
}

// Test capabilities endpoint
func TestAutoServerCapabilitiesEndpoint(t *testing.T) {
	server := NewAutoServer(nil)
	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	req := httptest.NewRequest("GET", "/capabilities", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	requiredFields := []string{"agents", "llm_providers", "tools", "features", "system_info", "timestamp"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Response should contain %s", field)
		}
	}

	features, ok := response["features"].(map[string]interface{})
	if !ok {
		t.Error("Features should be a map")
	} else {
		expectedFeatures := []string{"web_ui", "playground", "schema_api", "streaming", "conversations"}
		for _, feature := range expectedFeatures {
			if _, exists := features[feature]; !exists {
				t.Errorf("Features should contain %s", feature)
			}
		}
	}
}

// Test agents list endpoint
func TestAutoServerAgentsListEndpoint(t *testing.T) {
	server := NewAutoServer(nil)

	// Register test agents
	agentConfig := &agent.AgentConfig{
		Name:     "TestAgent",
		Type:     agent.AgentTypeChat,
		Model:    "llama3.2",
		Provider: "ollama",
	}

	// Clear any existing registrations for clean test
	_ = agent.GetGlobalRegistry()

	agentDefinition1 := agent.NewBaseAgentDefinition(agentConfig)
	agentDefinition2 := agent.NewBaseAgentDefinition(agentConfig)

	server.RegisterAgent("test1", agentDefinition1)
	server.RegisterAgent("test2", agentDefinition2)

	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	req := httptest.NewRequest("GET", "/agents", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if _, exists := response["agents"]; !exists {
		t.Error("Response should contain agents")
	}
	if _, exists := response["total_count"]; !exists {
		t.Error("Response should contain total_count")
	}
	if _, exists := response["timestamp"]; !exists {
		t.Error("Response should contain timestamp")
	}

	agents, ok := response["agents"].([]interface{})
	if !ok {
		t.Error("Agents should be an array")
	} else if len(agents) < 2 {
		t.Errorf("Expected at least 2 agents, got %d", len(agents))
	}
}

// Test agent info endpoint
func TestAutoServerAgentInfoEndpoint(t *testing.T) {
	server := NewAutoServer(nil)

	// Register a test agent
	agentConfig := &agent.AgentConfig{
		Name:     "TestAgent",
		Type:     agent.AgentTypeChat,
		Model:    "llama3.2",
		Provider: "ollama",
	}
	agentDefinition := agent.NewBaseAgentDefinition(agentConfig)
	server.RegisterAgent("info_test", agentDefinition)

	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	t.Run("existing agent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/agents/info_test", nil)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		expectedFields := []string{"id", "name", "description", "endpoint", "schema", "metadata"}
		for _, field := range expectedFields {
			if _, exists := response[field]; !exists {
				t.Errorf("Response should contain %s", field)
			}
		}

		if response["id"] != "info_test" {
			t.Errorf("Expected id 'info_test', got %v", response["id"])
		}
		if response["name"] != "TestAgent" {
			t.Errorf("Expected name 'TestAgent', got %v", response["name"])
		}
	})

	t.Run("non-existing agent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/agents/nonexistent", nil)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

// Test agent execution endpoint (just check that endpoint exists - full execution testing requires Ollama)
func TestAutoServerAgentExecutionEndpoint(t *testing.T) {
	server := NewAutoServer(nil)

	// Register a test agent
	agentConfig := &agent.AgentConfig{
		Name:     "TestAgent",
		Type:     agent.AgentTypeChat,
		Model:    "llama3.2",
		Provider: "ollama",
		Tools:    []string{},
	}
	agentDefinition := agent.NewBaseAgentDefinition(agentConfig)
	server.RegisterAgent("exec_test", agentDefinition)

	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	t.Run("endpoint exists", func(t *testing.T) {
		payload := map[string]interface{}{
			"message": "Hello, test agent!",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/exec_test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		// Endpoint should exist (not 404), though it may fail due to no real LLM
		if w.Code == http.StatusNotFound {
			t.Error("Agent execution endpoint should exist")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/exec_test", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}
	})
}

// Test CORS middleware
func TestAutoServerCORS(t *testing.T) {
	config := DefaultAutoServerConfig()
	config.EnableCORS = true
	server := NewAutoServer(config)
	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	req := httptest.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Check CORS headers are set
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	allowMethods := w.Header().Get("Access-Control-Allow-Methods")

	if allowOrigin == "" {
		t.Error("Access-Control-Allow-Origin header should be set")
	}
	if allowMethods == "" {
		t.Error("Access-Control-Allow-Methods header should be set")
	}
}

// Test web UI endpoint
func TestAutoServerWebUIEndpoint(t *testing.T) {
	config := DefaultAutoServerConfig()
	config.EnableWebUI = true
	server := NewAutoServer(config)
	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}

	// Check that the response contains some expected HTML elements
	body := w.Body.String()
	expectedElements := []string{"<html", "GoLangGraph", "Agent Chat"}
	for _, element := range expectedElements {
		if !strings.Contains(body, element) {
			t.Errorf("Response should contain '%s'", element)
		}
	}
}

// Test playground endpoint
func TestAutoServerPlaygroundEndpoint(t *testing.T) {
	config := DefaultAutoServerConfig()
	config.EnablePlayground = true
	server := NewAutoServer(config)
	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	req := httptest.NewRequest("GET", "/playground", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}

	// Check that the response contains playground-specific content
	body := w.Body.String()
	expectedElements := []string{"<html", "API Playground"}
	for _, element := range expectedElements {
		if !strings.Contains(body, element) {
			t.Errorf("Response should contain '%s'", element)
		}
	}
}

// Test metrics endpoint
func TestAutoServerMetricsEndpoint(t *testing.T) {
	config := DefaultAutoServerConfig()
	config.EnableMetricsAPI = true
	server := NewAutoServer(config)
	err := server.GenerateEndpoints()
	if err != nil {
		t.Fatalf("Failed to generate endpoints: %v", err)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	expectedFields := []string{"agents", "requests", "uptime", "timestamp"}
	for _, field := range expectedFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Response should contain %s", field)
		}
	}
}

// Test server lifecycle methods
func TestAutoServerLifecycle(t *testing.T) {
	server := NewAutoServer(nil)

	// Test that server can generate endpoints
	err := server.GenerateEndpoints()
	if err != nil {
		t.Errorf("Expected no error generating endpoints, got %v", err)
	}

	// Verify router is properly configured
	if server.router == nil {
		t.Error("Router should not be nil after generating endpoints")
	}
}

// Benchmark agent registration
func BenchmarkAutoServerAgentRegistration(b *testing.B) {
	server := NewAutoServer(nil)

	agentConfig := &agent.AgentConfig{
		Name:     "BenchmarkAgent",
		Type:     agent.AgentTypeChat,
		Model:    "llama3.2",
		Provider: "ollama",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agentID := fmt.Sprintf("agent_%d", i)
		agentDefinition := agent.NewBaseAgentDefinition(agentConfig)
		server.RegisterAgent(agentID, agentDefinition)
	}
}

// Benchmark endpoint generation
func BenchmarkAutoServerEndpointGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		server := NewAutoServer(nil)
		server.GenerateEndpoints()
	}
}
