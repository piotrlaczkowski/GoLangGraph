// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func TestNewServer(t *testing.T) {
	config := &ServerConfig{
		Host: "localhost",
		Port: 8080,
	}
	server := NewServer(config)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	if server.config.Port != 8080 {
		t.Error("Server port should be 8080")
	}

	if server.config.Host != "localhost" {
		t.Error("Host should be localhost")
	}

	if server.router == nil {
		t.Error("Router should not be nil")
	}
}

func TestDefaultServerConfig(t *testing.T) {
	config := DefaultServerConfig()

	if config == nil {
		t.Fatal("DefaultServerConfig returned nil")
	}

	if config.Host != "0.0.0.0" {
		t.Error("Default host should be 0.0.0.0")
	}

	if config.Port != 8080 {
		t.Error("Default port should be 8080")
	}

	if !config.EnableCORS {
		t.Error("CORS should be enabled by default")
	}
}

func TestServer_HealthCheck(t *testing.T) {
	server := NewServer(nil)

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health check returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to unmarshal health check response")
	}

	if response["status"] != "healthy" {
		t.Error("Health check should return status: healthy")
	}
}

func TestServer_CORS(t *testing.T) {
	config := &ServerConfig{
		Host:       "localhost",
		Port:       8080,
		EnableCORS: true,
	}

	server := NewServer(config)

	req, err := http.NewRequest("OPTIONS", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("CORS preflight returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check CORS headers
	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS headers should be set")
	}
}

func TestServer_Middleware(t *testing.T) {
	server := NewServer(nil)

	// Test with a request that should have logging middleware
	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Request with middleware returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestServer_ErrorHandling(t *testing.T) {
	server := NewServer(nil)

	// Test 404 error
	req, err := http.NewRequest("GET", "/nonexistent", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Nonexistent endpoint should return 404, got %v", status)
	}
}

func TestServer_StartStop(t *testing.T) {
	config := &ServerConfig{
		Host: "localhost",
		Port: 0, // Use port 0 for automatic assignment
	}
	server := NewServer(config)

	// Test starting server
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server start failed: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test stopping server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Server stop failed: %v", err)
	}
}

func TestAgentManager(t *testing.T) {
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()
	manager := NewAgentManager(llmManager, toolRegistry)

	if manager == nil {
		t.Fatal("NewAgentManager returned nil")
	}

	// Test creating an agent
	config := &agent.AgentConfig{
		Name:         "test-agent",
		Type:         agent.AgentTypeChat,
		SystemPrompt: "You are a test agent",
	}

	createdAgent, err := manager.CreateAgent(config)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if createdAgent == nil {
		t.Error("Created agent should not be nil")
	}

	// Test listing agents
	agents := manager.ListAgents()
	if len(agents) == 0 {
		t.Error("Should have at least one agent")
	}

	// Test getting the agent (we need to find the ID from the created agent)
	var agentID string
	for _, id := range agents {
		agentID = id
		break
	}

	retrievedAgent, exists := manager.GetAgent(agentID)
	if !exists {
		t.Error("Agent should exist")
	}

	if retrievedAgent == nil {
		t.Error("Retrieved agent should not be nil")
	}

	// Test deleting the agent
	manager.DeleteAgent(agentID)

	// Verify agent is deleted
	_, exists = manager.GetAgent(agentID)
	if exists {
		t.Error("Agent should be deleted")
	}
}

func TestServer_SetMethods(t *testing.T) {
	server := NewServer(nil)

	// Test setting LLM manager
	llmManager := llm.NewProviderManager()
	server.SetLLMManager(llmManager)

	// Test setting tool registry
	toolRegistry := tools.NewToolRegistry()
	server.SetToolRegistry(toolRegistry)

	// Test setting agent manager
	agentManager := NewAgentManager(llmManager, toolRegistry)
	server.SetAgentManager(agentManager)

	// These methods don't return anything, so we just test they don't panic
}

// MockProvider for testing
type MockProvider struct{}

func (m *MockProvider) GetName() string { return "mock" }
func (m *MockProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"mock-model"}, nil
}

func (m *MockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{
		ID:      "mock-response",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "mock-model",
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: "Mock response",
				},
				FinishReason: "stop",
			},
		},
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		},
	}, nil
}

func (m *MockProvider) CompleteStream(ctx context.Context, req llm.CompletionRequest, callback llm.StreamCallback) error {
	return nil
}

func (m *MockProvider) IsHealthy(ctx context.Context) error           { return nil }
func (m *MockProvider) GetConfig() map[string]interface{}             { return map[string]interface{}{} }
func (m *MockProvider) SetConfig(config map[string]interface{}) error { return nil }
func (m *MockProvider) Close() error                                  { return nil }

// Benchmark tests
func BenchmarkServer_HealthCheck(b *testing.B) {
	server := NewServer(nil)

	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)
	}
}

func BenchmarkAgentManager_CreateAgent(b *testing.B) {
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()
	manager := NewAgentManager(llmManager, toolRegistry)

	config := &agent.AgentConfig{
		Name:         "benchmark-agent",
		Type:         agent.AgentTypeChat,
		SystemPrompt: "You are a benchmark agent",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateAgent(config)
	}
}
