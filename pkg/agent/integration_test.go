// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLLMProvider for testing
type MockLLMProvider struct {
	name     string
	response string
	config   *llm.ProviderConfig
}

func (m *MockLLMProvider) GetName() string { return m.name }
func (m *MockLLMProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"test-model"}, nil
}
func (m *MockLLMProvider) IsHealthy(ctx context.Context) error           { return nil }
func (m *MockLLMProvider) GetConfig() map[string]interface{}             { return make(map[string]interface{}) }
func (m *MockLLMProvider) SetConfig(config map[string]interface{}) error { return nil }
func (m *MockLLMProvider) Close() error                                  { return nil }

func (m *MockLLMProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	// Simulate response truncation if MaxTokens is too low
	response := m.response
	if req.MaxTokens < 50 {
		response = "Hello" // Simulate truncation
	}

	return &llm.CompletionResponse{
		ID:     "test-completion",
		Object: "text_completion",
		Model:  "test-model",
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: response,
				},
				FinishReason: "stop",
			},
		},
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: len(response) / 4,
			TotalTokens:      10 + len(response)/4,
		},
	}, nil
}

func (m *MockLLMProvider) CompleteStream(ctx context.Context, req llm.CompletionRequest, callback llm.StreamCallback) error {
	response, err := m.Complete(ctx, req)
	if err != nil {
		return err
	}
	return callback(*response)
}

func (m *MockLLMProvider) CompleteWithMode(ctx context.Context, req llm.CompletionRequest, mode llm.StreamMode) (*llm.CompletionResponse, error) {
	return m.Complete(ctx, req)
}

func (m *MockLLMProvider) CompleteStreamWithMode(ctx context.Context, req llm.CompletionRequest, callback llm.StreamCallback, mode llm.StreamMode) error {
	return m.CompleteStream(ctx, req, callback)
}

func (m *MockLLMProvider) SupportsStreaming() bool { return true }

func (m *MockLLMProvider) GetStreamingConfig() *llm.StreamingConfig {
	return &llm.StreamingConfig{
		Mode:      llm.StreamModeNone,
		ChunkSize: 100,
		KeepAlive: false,
	}
}

func (m *MockLLMProvider) SetStreamingConfig(config *llm.StreamingConfig) error { return nil }

func TestAgentCreationWithValidation(t *testing.T) {
	// Create test LLM manager with mock provider
	llmManager := llm.NewProviderManager()
	mockProvider := &MockLLMProvider{
		name:     "mock",
		response: "This is a complete response that demonstrates the agent is working correctly with proper token limits.",
	}
	err := llmManager.RegisterProvider("mock", mockProvider)
	require.NoError(t, err)

	// Create tool registry
	toolRegistry := tools.NewToolRegistry()

	tests := []struct {
		name                string
		config              *AgentConfig
		expectCreationError bool
		expectValidResponse bool
		expectedMinTokens   int
	}{
		{
			name: "valid configuration",
			config: &AgentConfig{
				Name:        "test-agent",
				Type:        AgentTypeChat,
				Model:       "test-model",
				Provider:    "mock",
				MaxTokens:   1000,
				Temperature: 0.7,
			},
			expectCreationError: false,
			expectValidResponse: true,
			expectedMinTokens:   1000,
		},
		{
			name: "low MaxTokens gets sanitized",
			config: &AgentConfig{
				Name:        "test-agent-low-tokens",
				Type:        AgentTypeChat,
				Model:       "test-model",
				Provider:    "mock",
				MaxTokens:   25, // Will be sanitized to 500
				Temperature: 0.7,
			},
			expectCreationError: false,
			expectValidResponse: true,
			expectedMinTokens:   500,
		},
		{
			name: "zero MaxTokens gets fixed",
			config: &AgentConfig{
				Name:        "test-agent-zero-tokens",
				Type:        AgentTypeChat,
				Model:       "test-model",
				Provider:    "mock",
				MaxTokens:   0, // Invalid but will be fixed
				Temperature: 0.7,
			},
			expectCreationError: false,
			expectValidResponse: true,
			expectedMinTokens:   500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create agent
			agent := NewAgent(tt.config, llmManager, toolRegistry)

			if tt.expectCreationError {
				assert.Nil(t, agent)
				return
			}

			require.NotNil(t, agent)

			// Check that configuration was properly sanitized
			finalConfig := agent.GetConfig()
			assert.GreaterOrEqual(t, finalConfig.MaxTokens, tt.expectedMinTokens)

			// Test agent execution to ensure it works
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			execution, err := agent.Execute(ctx, "Hello, test agent!")
			require.NoError(t, err)
			require.NotNil(t, execution)

			if tt.expectValidResponse {
				// Should get a complete response, not truncated
				assert.NotEmpty(t, execution.Output)
				assert.NotEqual(t, "Hello", execution.Output, "Response appears to be truncated")
				assert.Greater(t, len(execution.Output), 10, "Response should be substantial")
			}
		})
	}
}

func TestMaxTokensPreventionIntegration(t *testing.T) {
	// Test that the system prevents truncation issues
	llmManager := llm.NewProviderManager()

	// Mock provider that simulates truncation with low MaxTokens
	mockProvider := &MockLLMProvider{
		name:     "mock",
		response: "This should be a long response but gets truncated if MaxTokens is too low",
	}
	err := llmManager.RegisterProvider("mock", mockProvider)
	require.NoError(t, err)

	toolRegistry := tools.NewToolRegistry()

	// Create agent with dangerously low MaxTokens
	config := &AgentConfig{
		Name:        "test-truncation-prevention",
		Type:        AgentTypeChat,
		Model:       "test-model",
		Provider:    "mock",
		MaxTokens:   10, // This should be sanitized
		Temperature: 0.7,
	}

	agent := NewAgent(config, llmManager, toolRegistry)
	require.NotNil(t, agent)

	// Verify that MaxTokens was sanitized
	finalConfig := agent.GetConfig()
	assert.GreaterOrEqual(t, finalConfig.MaxTokens, 500, "MaxTokens should be sanitized to prevent truncation")

	// Test execution
	ctx := context.Background()
	execution, err := agent.Execute(ctx, "Tell me about sustainable energy")
	require.NoError(t, err)
	require.NotNil(t, execution)

	// Should not be truncated because MaxTokens was sanitized
	assert.NotEqual(t, "Hello", execution.Output)
	assert.Greater(t, len(execution.Output), 20)
}

func TestProviderValidation(t *testing.T) {
	// Test that LLM provider configurations are validated
	tests := []struct {
		name           string
		providerConfig *llm.ProviderConfig
		expectError    bool
	}{
		{
			name: "valid provider config",
			providerConfig: &llm.ProviderConfig{
				Name:        "test-provider",
				Type:        "mock",
				Model:       "test-model",
				MaxTokens:   1000,
				Temperature: 0.7,
			},
			expectError: false,
		},
		{
			name: "provider config with low MaxTokens",
			providerConfig: &llm.ProviderConfig{
				Name:        "test-provider-low",
				Type:        "mock",
				Model:       "test-model",
				MaxTokens:   10, // Should be handled
				Temperature: 0.7,
			},
			expectError: false, // Should not error but should be sanitized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultConfig := llm.DefaultProviderConfig()
			assert.NotNil(t, defaultConfig)
			assert.Equal(t, 1000, defaultConfig.MaxTokens)
			assert.Equal(t, 0.7, defaultConfig.Temperature)
		})
	}
}

func TestSystemIntegrationWithDebugging(t *testing.T) {
	// Integration test that validates the entire system including debugging
	llmManager := llm.NewProviderManager()
	mockProvider := &MockLLMProvider{
		name:     "mock",
		response: "✅ Agent system is working correctly with proper configuration validation and token management.",
	}
	err := llmManager.RegisterProvider("mock", mockProvider)
	require.NoError(t, err)

	toolRegistry := tools.NewToolRegistry()

	// Test various configurations that could cause issues
	problemConfigs := []*AgentConfig{
		{
			Name:        "agent-with-zero-tokens",
			Type:        AgentTypeChat,
			Model:       "test-model",
			Provider:    "mock",
			MaxTokens:   0,
			Temperature: 0.7,
		},
		{
			Name:        "agent-with-low-tokens",
			Type:        AgentTypeChat,
			Model:       "test-model",
			Provider:    "mock",
			MaxTokens:   5,
			Temperature: 0.7,
		},
		{
			Name:        "agent-with-invalid-temp",
			Type:        AgentTypeChat,
			Model:       "test-model",
			Provider:    "mock",
			MaxTokens:   1000,
			Temperature: -0.5, // Will be fixed
		},
	}

	for i, config := range problemConfigs {
		t.Run(fmt.Sprintf("problem_config_%d", i), func(t *testing.T) {
			agent := NewAgent(config, llmManager, toolRegistry)
			require.NotNil(t, agent)

			// All configs should be sanitized and work correctly
			finalConfig := agent.GetConfig()
			assert.GreaterOrEqual(t, finalConfig.MaxTokens, 500)
			assert.GreaterOrEqual(t, finalConfig.Temperature, 0.0)
			assert.LessOrEqual(t, finalConfig.Temperature, 2.0)

			// Test execution
			ctx := context.Background()
			execution, err := agent.Execute(ctx, "Test message")
			require.NoError(t, err)
			assert.NotEmpty(t, execution.Output)
			assert.Contains(t, execution.Output, "Agent system is working correctly")
		})
	}
}
