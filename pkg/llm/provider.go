// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Message represents a message in a conversation
type Message struct {
	Role       string                 `json:"role"` // "system", "user", "assistant", "tool"
	Content    string                 `json:"content"`
	Name       string                 `json:"name,omitempty"`
	ToolCalls  []ToolCall             `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCall           `json:"function"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDefinition represents a tool that can be called by the LLM
type ToolDefinition struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// CompletionRequest represents a request for completion
type CompletionRequest struct {
	Messages      []Message        `json:"messages"`
	Model         string           `json:"model,omitempty"`
	Temperature   float64          `json:"temperature,omitempty"`
	MaxTokens     int              `json:"max_tokens,omitempty"`
	Tools         []ToolDefinition `json:"tools,omitempty"`
	ToolChoice    interface{}      `json:"tool_choice,omitempty"`
	Stream        bool             `json:"stream,omitempty"`
	SystemPrompt  string           `json:"system_prompt,omitempty"`
	StopSequences []string         `json:"stop_sequences,omitempty"`
}

// CompletionResponse represents a response from completion
type CompletionResponse struct {
	ID                string                 `json:"id"`
	Object            string                 `json:"object"`
	Created           int64                  `json:"created"`
	Model             string                 `json:"model"`
	Choices           []Choice               `json:"choices"`
	Usage             Usage                  `json:"usage"`
	SystemFingerprint string                 `json:"system_fingerprint,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// Choice represents a choice in the completion response
type Choice struct {
	Index        int       `json:"index"`
	Message      Message   `json:"message"`
	Delta        Message   `json:"delta,omitempty"`
	FinishReason string    `json:"finish_reason"`
	Logprobs     *Logprobs `json:"logprobs,omitempty"`
}

// Logprobs represents log probabilities
type Logprobs struct {
	Tokens        []string             `json:"tokens"`
	TokenLogprobs []float64            `json:"token_logprobs"`
	TopLogprobs   []map[string]float64 `json:"top_logprobs"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamCallback is called for each streaming chunk
type StreamCallback func(chunk CompletionResponse) error

// Provider represents an LLM provider interface
type Provider interface {
	// GetName returns the provider name
	GetName() string

	// GetModels returns available models
	GetModels(ctx context.Context) ([]string, error)

	// Complete generates a completion
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// CompleteStream generates a streaming completion
	CompleteStream(ctx context.Context, req CompletionRequest, callback StreamCallback) error

	// IsHealthy checks if the provider is healthy
	IsHealthy(ctx context.Context) error

	// GetConfig returns provider configuration
	GetConfig() map[string]interface{}

	// SetConfig updates provider configuration
	SetConfig(config map[string]interface{}) error

	// Close closes the provider and cleans up resources
	Close() error
}

// ProviderConfig represents provider configuration
type ProviderConfig struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Endpoint    string                 `json:"endpoint,omitempty"`
	APIKey      string                 `json:"api_key,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RetryCount  int                    `json:"retry_count,omitempty"`
	RetryDelay  time.Duration          `json:"retry_delay,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultProviderConfig returns default provider configuration
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Temperature: 0.7,
		MaxTokens:   1000,
		Timeout:     30 * time.Second,
		RetryCount:  3,
		RetryDelay:  1 * time.Second,
		Headers:     make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}
}

// ProviderManager manages multiple LLM providers
type ProviderManager struct {
	providers       map[string]Provider
	defaultProvider string
	mu              sync.RWMutex
	logger          *logrus.Logger
}

// NewProviderManager creates a new provider manager
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: make(map[string]Provider),
		logger:    logrus.New(),
	}
}

// RegisterProvider registers a new provider
func (pm *ProviderManager) RegisterProvider(name string, provider Provider) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	pm.providers[name] = provider

	// Set as default if it's the first provider
	if pm.defaultProvider == "" {
		pm.defaultProvider = name
	}

	pm.logger.WithField("provider", name).Info("Provider registered")
	return nil
}

// UnregisterProvider removes a provider
func (pm *ProviderManager) UnregisterProvider(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	provider, exists := pm.providers[name]
	if !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	// Close the provider
	if err := provider.Close(); err != nil {
		pm.logger.WithField("provider", name).WithError(err).Warn("Error closing provider")
	}

	delete(pm.providers, name)

	// Update default provider if necessary
	if pm.defaultProvider == name {
		pm.defaultProvider = ""
		// Set new default if other providers exist
		for providerName := range pm.providers {
			pm.defaultProvider = providerName
			break
		}
	}

	pm.logger.WithField("provider", name).Info("Provider unregistered")
	return nil
}

// GetProvider returns a provider by name
func (pm *ProviderManager) GetProvider(name string) (Provider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	provider, exists := pm.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// GetDefaultProvider returns the default provider
func (pm *ProviderManager) GetDefaultProvider() (Provider, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.defaultProvider == "" {
		return nil, fmt.Errorf("no default provider set")
	}

	return pm.providers[pm.defaultProvider], nil
}

// SetDefaultProvider sets the default provider
func (pm *ProviderManager) SetDefaultProvider(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}

	pm.defaultProvider = name
	pm.logger.WithField("provider", name).Info("Default provider set")
	return nil
}

// ListProviders returns all registered provider names
func (pm *ProviderManager) ListProviders() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	names := make([]string, 0, len(pm.providers))
	for name := range pm.providers {
		names = append(names, name)
	}
	return names
}

// Complete generates a completion using the specified provider (or default)
func (pm *ProviderManager) Complete(ctx context.Context, providerName string, req CompletionRequest) (*CompletionResponse, error) {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = pm.GetDefaultProvider()
	} else {
		provider, err = pm.GetProvider(providerName)
	}

	if err != nil {
		return nil, err
	}

	return provider.Complete(ctx, req)
}

// CompleteStream generates a streaming completion using the specified provider (or default)
func (pm *ProviderManager) CompleteStream(ctx context.Context, providerName string, req CompletionRequest, callback StreamCallback) error {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = pm.GetDefaultProvider()
	} else {
		provider, err = pm.GetProvider(providerName)
	}

	if err != nil {
		return err
	}

	return provider.CompleteStream(ctx, req, callback)
}

// HealthCheck checks the health of all providers
func (pm *ProviderManager) HealthCheck(ctx context.Context) map[string]error {
	pm.mu.RLock()
	providers := make(map[string]Provider)
	for name, provider := range pm.providers {
		providers[name] = provider
	}
	pm.mu.RUnlock()

	results := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, provider := range providers {
		wg.Add(1)
		go func(n string, p Provider) {
			defer wg.Done()
			err := p.IsHealthy(ctx)
			mu.Lock()
			results[n] = err
			mu.Unlock()
		}(name, provider)
	}

	wg.Wait()
	return results
}

// GetProviderModels returns available models for a provider
func (pm *ProviderManager) GetProviderModels(ctx context.Context, providerName string) ([]string, error) {
	var provider Provider
	var err error

	if providerName == "" {
		provider, err = pm.GetDefaultProvider()
	} else {
		provider, err = pm.GetProvider(providerName)
	}

	if err != nil {
		return nil, err
	}

	return provider.GetModels(ctx)
}

// GetAllModels returns all available models from all providers
func (pm *ProviderManager) GetAllModels(ctx context.Context) (map[string][]string, error) {
	pm.mu.RLock()
	providers := make(map[string]Provider)
	for name, provider := range pm.providers {
		providers[name] = provider
	}
	pm.mu.RUnlock()

	results := make(map[string][]string)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, provider := range providers {
		wg.Add(1)
		go func(n string, p Provider) {
			defer wg.Done()
			models, err := p.GetModels(ctx)
			mu.Lock()
			if err != nil {
				pm.logger.WithField("provider", n).WithError(err).Warn("Failed to get models")
				results[n] = []string{}
			} else {
				results[n] = models
			}
			mu.Unlock()
		}(name, provider)
	}

	wg.Wait()
	return results, nil
}

// Close closes all providers
func (pm *ProviderManager) Close() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var errors []error
	for name, provider := range pm.providers {
		if err := provider.Close(); err != nil {
			errors = append(errors, fmt.Errorf("error closing provider %s: %w", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing providers: %v", errors)
	}

	return nil
}

// ConversationHistory manages conversation history
type ConversationHistory struct {
	messages []Message
	mu       sync.RWMutex
}

// NewConversationHistory creates a new conversation history
func NewConversationHistory() *ConversationHistory {
	return &ConversationHistory{
		messages: make([]Message, 0),
	}
}

// AddMessage adds a message to the conversation history
func (ch *ConversationHistory) AddMessage(message Message) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.messages = append(ch.messages, message)
}

// GetMessages returns all messages in the conversation
func (ch *ConversationHistory) GetMessages() []Message {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	messages := make([]Message, len(ch.messages))
	copy(messages, ch.messages)
	return messages
}

// GetLastN returns the last N messages
func (ch *ConversationHistory) GetLastN(n int) []Message {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if n <= 0 || n > len(ch.messages) {
		n = len(ch.messages)
	}

	start := len(ch.messages) - n
	messages := make([]Message, n)
	copy(messages, ch.messages[start:])
	return messages
}

// Clear clears the conversation history
func (ch *ConversationHistory) Clear() {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.messages = ch.messages[:0]
}

// Size returns the number of messages in the conversation
func (ch *ConversationHistory) Size() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return len(ch.messages)
}

// TokenCounter interface for counting tokens
type TokenCounter interface {
	CountTokens(text string) (int, error)
	CountMessagesTokens(messages []Message) (int, error)
}

// SimpleTokenCounter is a simple token counter implementation
type SimpleTokenCounter struct{}

// NewSimpleTokenCounter creates a new simple token counter
func NewSimpleTokenCounter() *SimpleTokenCounter {
	return &SimpleTokenCounter{}
}

// CountTokens counts tokens in text (rough approximation)
func (stc *SimpleTokenCounter) CountTokens(text string) (int, error) {
	// Simple approximation: ~4 characters per token
	return len(text) / 4, nil
}

// CountMessagesTokens counts tokens in messages
func (stc *SimpleTokenCounter) CountMessagesTokens(messages []Message) (int, error) {
	total := 0
	for _, msg := range messages {
		tokens, err := stc.CountTokens(msg.Content)
		if err != nil {
			return 0, err
		}
		total += tokens
	}
	return total, nil
}
