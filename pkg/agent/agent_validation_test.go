// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *AgentConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &AgentConfig{
				Name:        "test-agent",
				Type:        AgentTypeChat,
				Model:       "gpt-3.5-turbo",
				Provider:    "openai",
				MaxTokens:   1000,
				Temperature: 0.7,
			},
			expectError: false,
		},
		{
			name: "empty name",
			config: &AgentConfig{
				Type:     AgentTypeChat,
				Model:    "gpt-3.5-turbo",
				Provider: "openai",
			},
			expectError: true,
			errorMsg:    "agent name is required",
		},
		{
			name: "empty type",
			config: &AgentConfig{
				Name:     "test-agent",
				Model:    "gpt-3.5-turbo",
				Provider: "openai",
			},
			expectError: true,
			errorMsg:    "agent type is required",
		},
		{
			name: "empty model",
			config: &AgentConfig{
				Name:     "test-agent",
				Type:     AgentTypeChat,
				Provider: "openai",
			},
			expectError: true,
			errorMsg:    "agent model is required",
		},
		{
			name: "empty provider",
			config: &AgentConfig{
				Name:  "test-agent",
				Type:  AgentTypeChat,
				Model: "gpt-3.5-turbo",
			},
			expectError: true,
			errorMsg:    "agent provider is required",
		},
		{
			name: "zero MaxTokens",
			config: &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: 0,
			},
			expectError: true,
			errorMsg:    "MaxTokens must be greater than 0",
		},
		{
			name: "negative MaxTokens",
			config: &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: -100,
			},
			expectError: true,
			errorMsg:    "MaxTokens must be greater than 0",
		},
		{
			name: "too large MaxTokens",
			config: &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: 200000,
			},
			expectError: true,
			errorMsg:    "MaxTokens too large",
		},
		{
			name: "negative temperature",
			config: &AgentConfig{
				Name:        "test-agent",
				Type:        AgentTypeChat,
				Model:       "gpt-3.5-turbo",
				Provider:    "openai",
				MaxTokens:   1000,
				Temperature: -0.1,
			},
			expectError: true,
			errorMsg:    "temperature must be between 0 and 2.0",
		},
		{
			name: "too high temperature",
			config: &AgentConfig{
				Name:        "test-agent",
				Type:        AgentTypeChat,
				Model:       "gpt-3.5-turbo",
				Provider:    "openai",
				MaxTokens:   1000,
				Temperature: 2.5,
			},
			expectError: true,
			errorMsg:    "temperature must be between 0 and 2.0",
		},
		{
			name: "too many MaxIterations",
			config: &AgentConfig{
				Name:          "test-agent",
				Type:          AgentTypeChat,
				Model:         "gpt-3.5-turbo",
				Provider:      "openai",
				MaxTokens:     1000,
				Temperature:   0.7,
				MaxIterations: 150,
			},
			expectError: true,
			errorMsg:    "MaxIterations too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAgentConfigSanitization(t *testing.T) {
	tests := []struct {
		name            string
		config          *AgentConfig
		expectedTokens  int
		expectedTimeout time.Duration
	}{
		{
			name: "sanitize low MaxTokens",
			config: &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: 50, // Below minimum
			},
			expectedTokens: 500, // Should be increased
		},
		{
			name: "preserve valid MaxTokens",
			config: &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: 1000,
			},
			expectedTokens: 1000, // Should remain unchanged
		},
		{
			name: "set default timeout",
			config: &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: 1000,
				Timeout:   0, // Not set
			},
			expectedTimeout: 30 * time.Second,
		},
		{
			name: "preserve valid timeout",
			config: &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: 1000,
				Timeout:   60 * time.Second,
			},
			expectedTimeout: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateAndSanitize()
			require.NoError(t, err)

			if tt.expectedTokens > 0 {
				assert.Equal(t, tt.expectedTokens, tt.config.MaxTokens)
			}

			if tt.expectedTimeout > 0 {
				assert.Equal(t, tt.expectedTimeout, tt.config.Timeout)
			}
		})
	}
}

func TestAgentConfigDefaults(t *testing.T) {
	config := DefaultAgentConfig()

	assert.NotEmpty(t, config.ID)
	assert.Equal(t, AgentTypeChat, config.Type)
	assert.Equal(t, 0.7, config.Temperature)
	assert.Equal(t, 1000, config.MaxTokens)
	assert.Equal(t, 10, config.MaxIterations)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.False(t, config.EnableStreaming)
	assert.NotNil(t, config.Tools)
	assert.NotNil(t, config.Metadata)
}

func TestMaxTokensPrevention(t *testing.T) {
	// Test that dangerously low MaxTokens values are caught
	dangerousConfigs := []int{0, -1, 1, 5, 10, 25}

	for _, tokens := range dangerousConfigs {
		t.Run(fmt.Sprintf("MaxTokens_%d", tokens), func(t *testing.T) {
			config := &AgentConfig{
				Name:      "test-agent",
				Type:      AgentTypeChat,
				Model:     "gpt-3.5-turbo",
				Provider:  "openai",
				MaxTokens: tokens,
			}

			err := config.Validate()
			assert.Error(t, err, "MaxTokens %d should be rejected", tokens)

			// Different error messages for different cases
			if tokens <= 0 {
				assert.Contains(t, err.Error(), "MaxTokens must be greater than 0")
			} else {
				assert.Contains(t, err.Error(), "minimum required is 100 to prevent response truncation")
			}
		})
	}
}

func TestTokenTruncationPrevention(t *testing.T) {
	// Test that configurations that could cause truncation are handled
	config := &AgentConfig{
		Name:      "test-agent",
		Type:      AgentTypeChat,
		Model:     "gpt-3.5-turbo",
		Provider:  "openai",
		MaxTokens: 50, // Low but valid
	}

	err := config.ValidateAndSanitize()
	require.NoError(t, err)

	// Should be automatically increased to prevent truncation
	assert.Equal(t, 500, config.MaxTokens)
}
