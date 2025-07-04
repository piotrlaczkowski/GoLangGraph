package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// GeminiProvider implements the Provider interface for Google Gemini
type GeminiProvider struct {
	service  *generativelanguage.Service
	config   *ProviderConfig
	logger   *logrus.Logger
	models   []string
	lastSync time.Time
}

// GeminiRequest represents a Gemini API request
type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []GeminiSafetySetting   `json:"safetySettings,omitempty"`
	Tools            []GeminiTool            `json:"tools,omitempty"`
}

// GeminiContent represents content in a Gemini request
type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text             string                  `json:"text,omitempty"`
	FunctionCall     *GeminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
}

// GeminiFunctionCall represents a function call
type GeminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// GeminiFunctionResponse represents a function response
type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// GeminiGenerationConfig represents generation configuration
type GeminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

// GeminiSafetySetting represents safety settings
type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// GeminiTool represents a tool definition
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations"`
}

// GeminiFunctionDeclaration represents a function declaration
type GeminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GeminiResponse represents a Gemini API response
type GeminiResponse struct {
	Candidates    []GeminiCandidate    `json:"candidates"`
	UsageMetadata *GeminiUsageMetadata `json:"usageMetadata,omitempty"`
}

// GeminiCandidate represents a response candidate
type GeminiCandidate struct {
	Content       GeminiContent        `json:"content"`
	FinishReason  string               `json:"finishReason"`
	Index         int                  `json:"index"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings"`
}

// GeminiSafetyRating represents a safety rating
type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GeminiUsageMetadata represents usage metadata
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(config *ProviderConfig) (*GeminiProvider, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key is required")
	}

	ctx := context.Background()
	service, err := generativelanguage.NewService(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini service: %w", err)
	}

	provider := &GeminiProvider{
		service: service,
		config:  config,
		logger:  logrus.New(),
		models:  []string{},
	}

	return provider, nil
}

// GetName returns the provider name
func (p *GeminiProvider) GetName() string {
	return "gemini"
}

// GetModels returns available models
func (p *GeminiProvider) GetModels(ctx context.Context) ([]string, error) {
	// Cache models for 5 minutes
	if time.Since(p.lastSync) < 5*time.Minute && len(p.models) > 0 {
		return p.models, nil
	}

	// For now, return known Gemini models
	p.models = p.GetDefaultModels()
	p.lastSync = time.Now()
	return p.models, nil
}

// Complete generates a completion
func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	model := req.Model
	if model == "" {
		model = p.config.Model
		if model == "" {
			model = "gemini-pro"
		}
	}

	// Convert request to Gemini format
	geminiReq := p.convertToGeminiRequest(req)

	// Make API call
	modelPath := fmt.Sprintf("models/%s", model)
	generateReq := &generativelanguage.GenerateContentRequest{
		Contents:         []*generativelanguage.Content{},
		GenerationConfig: &generativelanguage.GenerationConfig{},
	}

	// Convert contents
	for _, content := range geminiReq.Contents {
		geminiContent := &generativelanguage.Content{
			Role:  content.Role,
			Parts: []*generativelanguage.Part{},
		}

		for _, part := range content.Parts {
			geminiPart := &generativelanguage.Part{}
			if part.Text != "" {
				geminiPart.Data = &generativelanguage.Part_Text{Text: part.Text}
			}
			geminiContent.Parts = append(geminiContent.Parts, geminiPart)
		}

		generateReq.Contents = append(generateReq.Contents, geminiContent)
	}

	// Set generation config
	if geminiReq.GenerationConfig != nil {
		if geminiReq.GenerationConfig.Temperature != nil {
			generateReq.GenerationConfig.Temperature = geminiReq.GenerationConfig.Temperature
		}
		if geminiReq.GenerationConfig.TopP != nil {
			generateReq.GenerationConfig.TopP = geminiReq.GenerationConfig.TopP
		}
		if geminiReq.GenerationConfig.TopK != nil {
			generateReq.GenerationConfig.TopK = int64(*geminiReq.GenerationConfig.TopK)
		}
		if geminiReq.GenerationConfig.MaxOutputTokens != nil {
			generateReq.GenerationConfig.MaxOutputTokens = int64(*geminiReq.GenerationConfig.MaxOutputTokens)
		}
		generateReq.GenerationConfig.StopSequences = geminiReq.GenerationConfig.StopSequences
	}

	resp, err := p.service.Models.GenerateContent(modelPath, generateReq).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Gemini completion failed: %w", err)
	}

	return p.convertFromGeminiResponse(resp), nil
}

// CompleteStream generates a streaming completion
func (p *GeminiProvider) CompleteStream(ctx context.Context, req CompletionRequest, callback StreamCallback) error {
	// Gemini doesn't support streaming in the same way, so we'll simulate it
	resp, err := p.Complete(ctx, req)
	if err != nil {
		return err
	}

	// Simulate streaming by sending chunks
	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		words := strings.Fields(content)

		for i, word := range words {
			chunk := CompletionResponse{
				ID:      resp.ID,
				Object:  "chat.completion.chunk",
				Created: resp.Created,
				Model:   resp.Model,
				Choices: []Choice{
					{
						Index: 0,
						Delta: Message{
							Role:    "assistant",
							Content: word + " ",
						},
						FinishReason: "",
					},
				},
			}

			if i == len(words)-1 {
				chunk.Choices[0].FinishReason = "stop"
			}

			if err := callback(chunk); err != nil {
				return err
			}
		}
	}

	return nil
}

// IsHealthy checks if the provider is healthy
func (p *GeminiProvider) IsHealthy(ctx context.Context) error {
	// Simple health check - try to list models
	_, err := p.GetModels(ctx)
	return err
}

// GetConfig returns provider configuration
func (p *GeminiProvider) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":        p.config.Name,
		"type":        p.config.Type,
		"model":       p.config.Model,
		"temperature": p.config.Temperature,
		"max_tokens":  p.config.MaxTokens,
	}
}

// SetConfig updates provider configuration
func (p *GeminiProvider) SetConfig(config map[string]interface{}) error {
	if model, ok := config["model"].(string); ok {
		p.config.Model = model
	}
	if temp, ok := config["temperature"].(float64); ok {
		p.config.Temperature = temp
	}
	if maxTokens, ok := config["max_tokens"].(int); ok {
		p.config.MaxTokens = maxTokens
	}
	return nil
}

// Close closes the provider and cleans up resources
func (p *GeminiProvider) Close() error {
	// Nothing to close for Gemini provider
	return nil
}

// convertToGeminiRequest converts our request format to Gemini format
func (p *GeminiProvider) convertToGeminiRequest(req CompletionRequest) GeminiRequest {
	contents := make([]GeminiContent, len(req.Messages))
	for i, msg := range req.Messages {
		parts := []GeminiPart{
			{Text: msg.Content},
		}

		// Convert tool calls
		for _, toolCall := range msg.ToolCalls {
			parts = append(parts, GeminiPart{
				FunctionCall: &GeminiFunctionCall{
					Name: toolCall.Function.Name,
					Args: map[string]interface{}{},
				},
			})
		}

		contents[i] = GeminiContent{
			Role:  p.convertRole(msg.Role),
			Parts: parts,
		}
	}

	genConfig := &GeminiGenerationConfig{}
	if req.Temperature > 0 {
		genConfig.Temperature = &req.Temperature
	}
	if req.MaxTokens > 0 {
		genConfig.MaxOutputTokens = &req.MaxTokens
	}
	if len(req.StopSequences) > 0 {
		genConfig.StopSequences = req.StopSequences
	}

	// Convert tools
	var tools []GeminiTool
	if len(req.Tools) > 0 {
		tool := GeminiTool{
			FunctionDeclarations: make([]GeminiFunctionDeclaration, len(req.Tools)),
		}
		for i, t := range req.Tools {
			tool.FunctionDeclarations[i] = GeminiFunctionDeclaration{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			}
		}
		tools = append(tools, tool)
	}

	return GeminiRequest{
		Contents:         contents,
		GenerationConfig: genConfig,
		Tools:            tools,
		SafetySettings: []GeminiSafetySetting{
			{
				Category:  "HARM_CATEGORY_HARASSMENT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_HATE_SPEECH",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: "BLOCK_MEDIUM_AND_ABOVE",
			},
		},
	}
}

// convertFromGeminiResponse converts Gemini response to our format
func (p *GeminiProvider) convertFromGeminiResponse(resp *generativelanguage.GenerateContentResponse) *CompletionResponse {
	choices := make([]Choice, len(resp.Candidates))
	for i, candidate := range resp.Candidates {
		var content string
		var toolCalls []ToolCall

		if len(candidate.Content.Parts) > 0 {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					content += part.Text
				}
				// Handle function calls if present
			}
		}

		choices[i] = Choice{
			Index: i,
			Message: Message{
				Role:      p.convertRoleBack(candidate.Content.Role),
				Content:   content,
				ToolCalls: toolCalls,
			},
			FinishReason: strings.ToLower(candidate.FinishReason),
		}
	}

	usage := Usage{}
	if resp.UsageMetadata != nil {
		usage.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		usage.CompletionTokens = int(resp.UsageMetadata.CandidatesTokenCount)
		usage.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)
	}

	return &CompletionResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   p.config.Model,
		Choices: choices,
		Usage:   usage,
	}
}

// convertRole converts our role format to Gemini role format
func (p *GeminiProvider) convertRole(role string) string {
	switch role {
	case "system":
		return "user" // Gemini doesn't have system role, use user
	case "user":
		return "user"
	case "assistant":
		return "model"
	case "tool":
		return "user" // Tool responses go as user messages
	default:
		return "user"
	}
}

// convertRoleBack converts Gemini role format back to our format
func (p *GeminiProvider) convertRoleBack(role string) string {
	switch role {
	case "user":
		return "user"
	case "model":
		return "assistant"
	default:
		return "assistant"
	}
}

// GetDefaultModels returns commonly used Gemini models
func (p *GeminiProvider) GetDefaultModels() []string {
	return []string{
		"gemini-pro",
		"gemini-pro-vision",
		"gemini-1.0-pro",
		"gemini-1.0-pro-001",
		"gemini-1.0-pro-latest",
		"gemini-1.0-pro-vision-latest",
		"gemini-1.5-pro",
		"gemini-1.5-pro-latest",
		"gemini-1.5-flash",
		"gemini-1.5-flash-latest",
	}
}

// SupportsStreaming returns true if the provider supports streaming
func (p *GeminiProvider) SupportsStreaming() bool {
	return true // Simulated streaming
}

// SupportsToolCalls returns true if the provider supports tool calls
func (p *GeminiProvider) SupportsToolCalls() bool {
	return true
}

// GetMaxTokens returns the maximum tokens for a model
func (p *GeminiProvider) GetMaxTokens(model string) int {
	switch model {
	case "gemini-pro", "gemini-1.0-pro", "gemini-1.0-pro-001", "gemini-1.0-pro-latest":
		return 32768
	case "gemini-1.5-pro", "gemini-1.5-pro-latest":
		return 1048576 // 1M tokens
	case "gemini-1.5-flash", "gemini-1.5-flash-latest":
		return 1048576 // 1M tokens
	case "gemini-pro-vision", "gemini-1.0-pro-vision-latest":
		return 16384
	default:
		return 32768
	}
}
