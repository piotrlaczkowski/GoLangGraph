// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client   *openai.Client
	config   *ProviderConfig
	logger   *logrus.Logger
	models   []string
	lastSync time.Time
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(config *ProviderConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	if config.Endpoint != "" {
		clientConfig.BaseURL = config.Endpoint
	}

	client := openai.NewClientWithConfig(clientConfig)

	provider := &OpenAIProvider{
		client: client,
		config: config,
		logger: logrus.New(),
		models: []string{},
	}

	return provider, nil
}

// GetName returns the provider name
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetModels returns available models
func (p *OpenAIProvider) GetModels(ctx context.Context) ([]string, error) {
	// Cache models for 5 minutes
	if time.Since(p.lastSync) < 5*time.Minute && len(p.models) > 0 {
		return p.models, nil
	}

	models, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	p.models = make([]string, len(models.Models))
	for i, model := range models.Models {
		p.models[i] = model.ID
	}

	p.lastSync = time.Now()
	return p.models, nil
}

// Complete generates a completion
func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	openaiReq := p.convertToOpenAIRequest(req)

	resp, err := p.client.CreateChatCompletion(ctx, openaiReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI completion failed: %w", err)
	}

	return p.convertFromOpenAIResponse(resp), nil
}

// CompleteStream generates a streaming completion
func (p *OpenAIProvider) CompleteStream(ctx context.Context, req CompletionRequest, callback StreamCallback) error {
	openaiReq := p.convertToOpenAIRequest(req)

	stream, err := p.client.CreateChatCompletionStream(ctx, openaiReq)
	if err != nil {
		return fmt.Errorf("OpenAI streaming failed: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("stream error: %w", err)
		}

		// Convert to our format and call callback
		converted := p.convertFromOpenAIStreamResponse(response)
		if err := callback(converted); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}
	}

	return nil
}

// IsHealthy checks if the provider is healthy
func (p *OpenAIProvider) IsHealthy(ctx context.Context) error {
	// Try to list models as a health check
	_, err := p.client.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("OpenAI health check failed: %w", err)
	}
	return nil
}

// GetConfig returns provider configuration
func (p *OpenAIProvider) GetConfig() map[string]interface{} {
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
func (p *OpenAIProvider) SetConfig(config map[string]interface{}) error {
	if name, ok := config["name"].(string); ok {
		p.config.Name = name
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
func (p *OpenAIProvider) Close() error {
	// OpenAI client doesn't need explicit closing
	return nil
}

// convertToOpenAIRequest converts our request format to OpenAI format
func (p *OpenAIProvider) convertToOpenAIRequest(req CompletionRequest) openai.ChatCompletionRequest {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Convert tool calls
		if len(msg.ToolCalls) > 0 {
			toolCalls := make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				toolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			messages[i].ToolCalls = toolCalls
		}

		// Set tool call ID for tool messages
		if msg.ToolCallID != "" {
			messages[i].ToolCallID = msg.ToolCallID
		}
	}

	openaiReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: float32(req.Temperature),
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Stop:        req.StopSequences,
	}

	// Use default model if not specified
	if openaiReq.Model == "" {
		openaiReq.Model = p.config.Model
		if openaiReq.Model == "" {
			openaiReq.Model = "gpt-3.5-turbo"
		}
	}

	// Use default temperature if not specified
	if openaiReq.Temperature == 0 {
		openaiReq.Temperature = float32(p.config.Temperature)
	}

	// Use default max tokens if not specified
	if openaiReq.MaxTokens == 0 {
		openaiReq.MaxTokens = p.config.MaxTokens
	}

	// Convert tools
	if len(req.Tools) > 0 {
		tools := make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			tools[i] = openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
		openaiReq.Tools = tools
	}

	// Handle tool choice
	if req.ToolChoice != nil {
		switch tc := req.ToolChoice.(type) {
		case string:
			if tc == "auto" || tc == "none" {
				openaiReq.ToolChoice = tc
			}
		case map[string]interface{}:
			if toolType, ok := tc["type"].(string); ok && toolType == "function" {
				if function, ok := tc["function"].(map[string]interface{}); ok {
					if name, ok := function["name"].(string); ok {
						openaiReq.ToolChoice = openai.ToolChoice{
							Type: "function",
							Function: openai.ToolFunction{
								Name: name,
							},
						}
					}
				}
			}
		}
	}

	return openaiReq
}

// convertFromOpenAIResponse converts OpenAI response to our format
func (p *OpenAIProvider) convertFromOpenAIResponse(resp openai.ChatCompletionResponse) *CompletionResponse {
	choices := make([]Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		message := Message{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
			Name:    choice.Message.Name,
		}

		// Convert tool calls
		if len(choice.Message.ToolCalls) > 0 {
			toolCalls := make([]ToolCall, len(choice.Message.ToolCalls))
			for j, tc := range choice.Message.ToolCalls {
				toolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: string(tc.Type),
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			message.ToolCalls = toolCalls
		}

		choices[i] = Choice{
			Index:        choice.Index,
			Message:      message,
			FinishReason: string(choice.FinishReason),
		}
	}

	return &CompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		SystemFingerprint: resp.SystemFingerprint,
	}
}

// convertFromOpenAIStreamResponse converts OpenAI stream response to our format
func (p *OpenAIProvider) convertFromOpenAIStreamResponse(resp openai.ChatCompletionStreamResponse) CompletionResponse {
	choices := make([]Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		delta := Message{
			Role:    choice.Delta.Role,
			Content: choice.Delta.Content,
		}

		// Convert tool calls in delta
		if len(choice.Delta.ToolCalls) > 0 {
			toolCalls := make([]ToolCall, len(choice.Delta.ToolCalls))
			for j, tc := range choice.Delta.ToolCalls {
				toolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: string(tc.Type),
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
			delta.ToolCalls = toolCalls
		}

		choices[i] = Choice{
			Index:        choice.Index,
			Delta:        delta,
			FinishReason: string(choice.FinishReason),
		}
	}

	return CompletionResponse{
		ID:                resp.ID,
		Object:            resp.Object,
		Created:           resp.Created,
		Model:             resp.Model,
		Choices:           choices,
		SystemFingerprint: resp.SystemFingerprint,
	}
}

// GetDefaultModels returns commonly used OpenAI models
func (p *OpenAIProvider) GetDefaultModels() []string {
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4-turbo-preview",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
		"gpt-4-vision-preview",
		"gpt-4-1106-preview",
		"gpt-3.5-turbo-1106",
	}
}

// SupportsStreaming returns true if the provider supports streaming
func (p *OpenAIProvider) SupportsStreaming() bool {
	return true
}

// GetStreamingConfig returns the current streaming configuration
func (p *OpenAIProvider) GetStreamingConfig() *StreamingConfig {
	if p.config.Streaming == nil {
		p.config.Streaming = DefaultStreamingConfig()
	}
	return p.config.Streaming
}

// SetStreamingConfig updates the streaming configuration
func (p *OpenAIProvider) SetStreamingConfig(config *StreamingConfig) error {
	if config == nil {
		return fmt.Errorf("streaming config cannot be nil")
	}
	p.config.Streaming = config
	return nil
}

// CompleteWithMode generates a completion with explicit streaming mode
func (p *OpenAIProvider) CompleteWithMode(ctx context.Context, req CompletionRequest, mode StreamMode) (*CompletionResponse, error) {
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
func (p *OpenAIProvider) CompleteStreamWithMode(ctx context.Context, req CompletionRequest, callback StreamCallback, mode StreamMode) error {
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
func (p *OpenAIProvider) completeNonStreaming(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	return p.Complete(ctx, req)
}

// completeStreamingCollected forces streaming but collects all chunks into single response
func (p *OpenAIProvider) completeStreamingCollected(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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
func (p *OpenAIProvider) SupportsToolCalls() bool {
	return true
}

// GetMaxTokens returns the maximum tokens for a model
func (p *OpenAIProvider) GetMaxTokens(model string) int {
	switch {
	case strings.Contains(model, "gpt-4-turbo"):
		return 128000
	case strings.Contains(model, "gpt-4"):
		return 8192
	case strings.Contains(model, "gpt-3.5-turbo-16k"):
		return 16384
	case strings.Contains(model, "gpt-3.5-turbo"):
		return 4096
	default:
		return 4096
	}
}

// GetTokenLimit returns the token limit for a model
func (p *OpenAIProvider) GetTokenLimit(model string) int {
	return p.GetMaxTokens(model)
}

// EstimateTokens estimates the number of tokens in a text
func (p *OpenAIProvider) EstimateTokens(text string) int {
	// Rough approximation: 1 token ≈ 4 characters for English text
	return len(text) / 4
}

// EstimateMessagesTokens estimates the number of tokens in messages
func (p *OpenAIProvider) EstimateMessagesTokens(messages []Message) int {
	total := 0
	for _, msg := range messages {
		// Add some overhead for message formatting
		total += p.EstimateTokens(msg.Content) + 10

		// Add tokens for tool calls
		for _, tc := range msg.ToolCalls {
			total += p.EstimateTokens(tc.Function.Name) + p.EstimateTokens(tc.Function.Arguments) + 20
		}
	}
	return total
}

// ValidateModel checks if a model is valid for this provider
func (p *OpenAIProvider) ValidateModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Check if it's a known OpenAI model pattern
	validPrefixes := []string{"gpt-", "text-", "code-", "davinci", "curie", "babbage", "ada"}
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(model, prefix) {
			return nil
		}
	}

	return fmt.Errorf("model %s does not appear to be a valid OpenAI model", model)
}

// GetProviderInfo returns information about the provider
func (p *OpenAIProvider) GetProviderInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":               "OpenAI",
		"type":               "openai",
		"supports_streaming": true,
		"supports_tools":     true,
		"supports_vision":    true,
		"max_context_length": 128000,
		"default_model":      "gpt-3.5-turbo",
		"api_version":        "v1",
	}
}
