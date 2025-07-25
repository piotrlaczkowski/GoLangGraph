// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	client   *http.Client
	config   *ProviderConfig
	logger   *logrus.Logger
	models   []string
	lastSync time.Time
}

// OllamaRequest represents an Ollama API request
type OllamaRequest struct {
	Model     string          `json:"model"`
	Messages  []OllamaMessage `json:"messages"`
	Stream    bool            `json:"stream,omitempty"`
	Options   OllamaOptions   `json:"options,omitempty"`
	Format    string          `json:"format,omitempty"`
	KeepAlive string          `json:"keep_alive,omitempty"`
}

// OllamaMessage represents an Ollama message
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaOptions represents Ollama generation options
type OllamaOptions struct {
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	TopK        int      `json:"top_k,omitempty"`
	NumPredict  int      `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// OllamaResponse represents an Ollama API response
type OllamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt time.Time     `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
	Error     string        `json:"error,omitempty"`
}

// OllamaModelInfo represents information about an Ollama model
type OllamaModelInfo struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
}

// OllamaModelsResponse represents the response from the models endpoint
type OllamaModelsResponse struct {
	Models []OllamaModelInfo `json:"models"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config *ProviderConfig) (*OllamaProvider, error) {
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	provider := &OllamaProvider{
		client: client,
		config: config,
		logger: logrus.New(),
		models: []string{},
	}

	// Set the endpoint in config
	provider.config.Endpoint = endpoint

	return provider, nil
}

// GetName returns the provider name
func (p *OllamaProvider) GetName() string {
	return "ollama"
}

// GetModels returns available models
func (p *OllamaProvider) GetModels(ctx context.Context) ([]string, error) {
	// Cache models for 5 minutes
	if time.Since(p.lastSync) < 5*time.Minute && len(p.models) > 0 {
		return p.models, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.config.Endpoint+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get models: status %d", resp.StatusCode)
	}

	var modelsResp OllamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	p.models = make([]string, len(modelsResp.Models))
	for i, model := range modelsResp.Models {
		p.models[i] = model.Name
	}

	p.lastSync = time.Now()
	return p.models, nil
}

// Complete generates a completion
func (p *OllamaProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	ollamaReq := p.convertToOllamaRequest(req)

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log request being sent to Ollama
	p.logger.WithFields(logrus.Fields{
		"endpoint": p.config.Endpoint + "/api/chat",
		"model":    ollamaReq.Model,
	}).Debug("Sending request to Ollama")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Read all streaming chunks until done=true
	var completeResponse strings.Builder
	var finalModel string
	var finalRole string

	decoder := json.NewDecoder(resp.Body)
	for {
		var ollamaResp OllamaResponse
		if err := decoder.Decode(&ollamaResp); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		if ollamaResp.Error != "" {
			return nil, fmt.Errorf("Ollama API error: %s", ollamaResp.Error)
		}

		// Accumulate the response content
		completeResponse.WriteString(ollamaResp.Message.Content)
		finalModel = ollamaResp.Model
		finalRole = ollamaResp.Message.Role

		// Break when done
		if ollamaResp.Done {
			break
		}
	}

	finalContent := completeResponse.String()

	// Log the complete response from Ollama
	p.logger.WithFields(logrus.Fields{
		"model":          finalModel,
		"content_length": len(finalContent),
	}).Debug("Received complete response from Ollama")

	// Create final response object
	finalResponse := OllamaResponse{
		Model:     finalModel,
		CreatedAt: time.Now(),
		Message: OllamaMessage{
			Role:    finalRole,
			Content: finalContent,
		},
		Done: true,
	}

	return p.convertFromOllamaResponse(finalResponse), nil
}

// CompleteStream generates a streaming completion
func (p *OllamaProvider) CompleteStream(ctx context.Context, req CompletionRequest, callback StreamCallback) error {
	ollamaReq := p.convertToOllamaRequest(req)
	ollamaReq.Stream = true

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var ollamaResp OllamaResponse
		if err := decoder.Decode(&ollamaResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode stream response: %w", err)
		}

		if ollamaResp.Error != "" {
			return fmt.Errorf("Ollama API error: %s", ollamaResp.Error)
		}

		// Convert to our format and call callback
		converted := p.convertFromOllamaStreamResponse(ollamaResp)
		if err := callback(converted); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}

		if ollamaResp.Done {
			break
		}
	}

	return nil
}

// IsHealthy checks if the provider is healthy
func (p *OllamaProvider) IsHealthy(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.Endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// GetConfig returns provider configuration
func (p *OllamaProvider) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":        p.config.Name,
		"type":        p.config.Type,
		"endpoint":    p.config.Endpoint,
		"model":       p.config.Model,
		"temperature": p.config.Temperature,
		"max_tokens":  p.config.MaxTokens,
		"timeout":     p.config.Timeout,
		"retry_count": p.config.RetryCount,
		"retry_delay": p.config.RetryDelay,
	}
}

// SetConfig updates provider configuration
func (p *OllamaProvider) SetConfig(config map[string]interface{}) error {
	if name, ok := config["name"].(string); ok {
		p.config.Name = name
	}
	if endpoint, ok := config["endpoint"].(string); ok {
		p.config.Endpoint = endpoint
	}
	if model, ok := config["model"].(string); ok {
		p.config.Model = model
	}
	if temp, ok := config["temperature"].(float64); ok {
		p.config.Temperature = temp
	}
	if maxTokens, ok := config["max_tokens"].(int); ok {
		p.config.MaxTokens = maxTokens
	}
	if timeout, ok := config["timeout"].(time.Duration); ok {
		p.config.Timeout = timeout
		p.client.Timeout = timeout
	}
	if retryCount, ok := config["retry_count"].(int); ok {
		p.config.RetryCount = retryCount
	}
	if retryDelay, ok := config["retry_delay"].(time.Duration); ok {
		p.config.RetryDelay = retryDelay
	}

	return nil
}

// Close closes the provider and cleans up resources
func (p *OllamaProvider) Close() error {
	// HTTP client doesn't need explicit closing
	return nil
}

// convertToOllamaRequest converts our request format to Ollama format
func (p *OllamaProvider) convertToOllamaRequest(req CompletionRequest) OllamaRequest {
	var systemPrompt string
	var filteredMessages []OllamaMessage

	// Extract system prompt and filter messages
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			// Combine multiple system messages if present
			if systemPrompt != "" {
				systemPrompt += "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
		} else if msg.Content != "" {
			// Only add non-empty non-system messages
			filteredMessages = append(filteredMessages, OllamaMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// If we have a system prompt but no user messages, create a user message to trigger response
	if systemPrompt != "" && len(filteredMessages) == 0 {
		filteredMessages = append(filteredMessages, OllamaMessage{
			Role:    "user",
			Content: "Please respond according to your instructions.",
		})
	}

	// If we have a system prompt and user messages, prepend system prompt to first user message
	if systemPrompt != "" && len(filteredMessages) > 0 && filteredMessages[0].Role == "user" {
		filteredMessages[0].Content = systemPrompt + "\n\n" + filteredMessages[0].Content
	}

	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "llama2" // Default model
		}
	}

	temperature := req.Temperature
	if temperature == 0 {
		temperature = p.config.Temperature
		if temperature == 0 {
			temperature = 0.7
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.config.MaxTokens
		if maxTokens == 0 {
			maxTokens = 1000
		}
	}

	// CRITICAL: Ensure minimum MaxTokens to prevent truncation
	if maxTokens < 50 {
		p.logger.WithFields(logrus.Fields{
			"requested_tokens": maxTokens,
			"applied_tokens":   500,
		}).Warn("MaxTokens too low, applying minimum safe value")
		maxTokens = 500
	}

	// VALIDATION: Check for suspicious configurations
	if maxTokens > 0 && maxTokens < 100 {
		p.logger.WithFields(logrus.Fields{
			"model":       model,
			"maxTokens":   maxTokens,
			"temperature": temperature,
		}).Error("CRITICAL: MaxTokens dangerously low - this will cause truncation!")
	}

	ollamaReq := OllamaRequest{
		Model:    model,
		Messages: filteredMessages,
		Stream:   req.Stream,
		Options: OllamaOptions{
			Temperature: temperature,
			NumPredict:  maxTokens,
			Stop:        req.StopSequences,
		},
		KeepAlive: "5m",
	}

	// Log request details
	p.logger.WithFields(logrus.Fields{
		"model":         model,
		"temperature":   temperature,
		"num_predict":   maxTokens,
		"message_count": len(filteredMessages),
	}).Debug("Ollama request configuration")

	return ollamaReq
}

// convertFromOllamaResponse converts Ollama response to our format
func (p *OllamaProvider) convertFromOllamaResponse(resp OllamaResponse) *CompletionResponse {
	message := Message{
		Role:    resp.Message.Role,
		Content: resp.Message.Content,
	}

	choice := Choice{
		Index:   0,
		Message: message,
	}

	if resp.Done {
		choice.FinishReason = "stop"
	}

	return &CompletionResponse{
		ID:      fmt.Sprintf("ollama-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: resp.CreatedAt.Unix(),
		Model:   resp.Model,
		Choices: []Choice{choice},
		Usage: Usage{
			PromptTokens:     0, // Ollama doesn't provide token counts
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}
}

// convertFromOllamaStreamResponse converts Ollama stream response to our format
func (p *OllamaProvider) convertFromOllamaStreamResponse(resp OllamaResponse) CompletionResponse {
	delta := Message{
		Role:    resp.Message.Role,
		Content: resp.Message.Content,
	}

	choice := Choice{
		Index: 0,
		Delta: delta,
	}

	if resp.Done {
		choice.FinishReason = "stop"
	}

	return CompletionResponse{
		ID:      fmt.Sprintf("ollama-%d", time.Now().Unix()),
		Object:  "chat.completion.chunk",
		Created: resp.CreatedAt.Unix(),
		Model:   resp.Model,
		Choices: []Choice{choice},
	}
}

// GetDefaultModels returns commonly used Ollama models
func (p *OllamaProvider) GetDefaultModels() []string {
	return []string{
		"llama2",
		"llama2:13b",
		"llama2:70b",
		"codellama",
		"codellama:13b",
		"codellama:34b",
		"mistral",
		"mixtral",
		"neural-chat",
		"starling-lm",
		"dolphin-mixtral",
		"llama2-uncensored",
		"orca-mini",
		"vicuna",
		"wizard-vicuna-uncensored",
	}
}

// SupportsStreaming returns true if the provider supports streaming
func (p *OllamaProvider) SupportsStreaming() bool {
	return true
}

// GetStreamingConfig returns the current streaming configuration
func (p *OllamaProvider) GetStreamingConfig() *StreamingConfig {
	if p.config.Streaming == nil {
		p.config.Streaming = DefaultStreamingConfig()
	}
	return p.config.Streaming
}

// SetStreamingConfig updates the streaming configuration
func (p *OllamaProvider) SetStreamingConfig(config *StreamingConfig) error {
	if config == nil {
		return fmt.Errorf("streaming config cannot be nil")
	}
	p.config.Streaming = config
	return nil
}

// CompleteWithMode generates a completion with explicit streaming mode
func (p *OllamaProvider) CompleteWithMode(ctx context.Context, req CompletionRequest, mode StreamMode) (*CompletionResponse, error) {
	switch mode {
	case StreamModeNone:
		// Force non-streaming mode
		return p.completeNonStreaming(ctx, req)
	case StreamModeForced:
		// Force streaming mode but collect all chunks
		return p.completeStreamingCollected(ctx, req)
	case StreamModeAuto:
		// Auto-detect based on request.Stream flag
		if req.Stream {
			return p.completeStreamingCollected(ctx, req)
		}
		return p.completeNonStreaming(ctx, req)
	default:
		return p.Complete(ctx, req)
	}
}

// CompleteStreamWithMode generates a streaming completion with explicit mode
func (p *OllamaProvider) CompleteStreamWithMode(ctx context.Context, req CompletionRequest, callback StreamCallback, mode StreamMode) error {
	switch mode {
	case StreamModeNone:
		// Convert to non-streaming
		resp, err := p.completeNonStreaming(ctx, req)
		if err != nil {
			return err
		}
		// Send as single chunk
		return callback(*resp)
	case StreamModeForced, StreamModeAuto:
		// Use normal streaming
		return p.CompleteStream(ctx, req, callback)
	default:
		return p.CompleteStream(ctx, req, callback)
	}
}

// completeNonStreaming forces non-streaming completion (current Complete method)
func (p *OllamaProvider) completeNonStreaming(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	return p.Complete(ctx, req)
}

// completeStreamingCollected forces streaming but collects all chunks into single response
func (p *OllamaProvider) completeStreamingCollected(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	var completeContent strings.Builder
	var finalResponse *CompletionResponse

	err := p.CompleteStream(ctx, req, func(chunk CompletionResponse) error {
		if len(chunk.Choices) > 0 {
			completeContent.WriteString(chunk.Choices[0].Delta.Content)
			finalResponse = &chunk
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if finalResponse != nil {
		// Convert delta to complete message
		finalResponse.Choices[0].Message = Message{
			Role:    finalResponse.Choices[0].Delta.Role,
			Content: completeContent.String(),
		}
		finalResponse.Choices[0].Delta = Message{} // Clear delta
		finalResponse.Object = "chat.completion"   // Change from chunk to completion
	}

	return finalResponse, nil
}

// SupportsToolCalls returns true if the provider supports tool calls
func (p *OllamaProvider) SupportsToolCalls() bool {
	return false // Ollama doesn't support tool calls natively
}

// GetMaxTokens returns the maximum tokens for a model
func (p *OllamaProvider) GetMaxTokens(model string) int {
	switch {
	case strings.Contains(model, "70b"):
		return 4096
	case strings.Contains(model, "34b"):
		return 4096
	case strings.Contains(model, "13b"):
		return 4096
	case strings.Contains(model, "7b"):
		return 4096
	default:
		return 4096
	}
}

// GetTokenLimit returns the token limit for a model
func (p *OllamaProvider) GetTokenLimit(model string) int {
	return p.GetMaxTokens(model)
}

// EstimateTokens estimates the number of tokens in a text
func (p *OllamaProvider) EstimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters for English text
	return len(text) / 4
}

// EstimateMessagesTokens estimates the number of tokens in messages
func (p *OllamaProvider) EstimateMessagesTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += p.EstimateTokens(msg.Content) + 5 // Add overhead for message formatting
	}
	return total
}

// ValidateModel checks if a model is valid for this provider
func (p *OllamaProvider) ValidateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// For Ollama, we accept any model name as it depends on what's installed locally
	return nil
}

// GetProviderInfo returns information about the provider
func (p *OllamaProvider) GetProviderInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":               "Ollama",
		"type":               "ollama",
		"supports_streaming": true,
		"supports_tools":     false,
		"supports_vision":    false,
		"max_context_length": 4096,
		"default_model":      "llama2",
		"local":              true,
	}
}

// PullModel pulls a model from the Ollama registry
func (p *OllamaProvider) PullModel(ctx context.Context, model string) error {
	reqBody := map[string]string{
		"name": model,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal pull request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint+"/api/pull", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to pull model: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Read the streaming response to completion
	decoder := json.NewDecoder(resp.Body)
	for {
		var pullResp map[string]interface{}
		if err := decoder.Decode(&pullResp); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode pull response: %w", err)
		}

		// Check for errors in the response
		if errMsg, ok := pullResp["error"].(string); ok && errMsg != "" {
			return fmt.Errorf("pull model error: %s", errMsg)
		}

		// Check if pull is complete
		if status, ok := pullResp["status"].(string); ok && status == "success" {
			break
		}
	}

	return nil
}

// DeleteModel deletes a model from Ollama
func (p *OllamaProvider) DeleteModel(ctx context.Context, model string) error {
	reqBody := map[string]string{
		"name": model,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", p.config.Endpoint+"/api/delete", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete model: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
