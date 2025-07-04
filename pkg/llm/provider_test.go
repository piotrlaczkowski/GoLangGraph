package llm

import (
	"testing"
)

func TestNewOpenAIProvider(t *testing.T) {
	// Test creating OpenAI provider
	config := &ProviderConfig{
		Type:   "openai",
		APIKey: "test-key",
		Model:  "gpt-3.5-turbo",
	}

	provider, err := NewOpenAIProvider(config)
	if err != nil {
		t.Errorf("NewOpenAIProvider() failed: %v", err)
	}

	if provider == nil {
		t.Error("NewOpenAIProvider() should return a provider")
	}

	if provider.GetName() != "openai" {
		t.Error("Provider name should be 'openai'")
	}

	// Test creating OpenAI provider without API key
	invalidConfig := &ProviderConfig{
		Type:  "openai",
		Model: "gpt-3.5-turbo",
	}

	_, err = NewOpenAIProvider(invalidConfig)
	if err == nil {
		t.Error("NewOpenAIProvider() should return error without API key")
	}
}

func TestNewOllamaProvider(t *testing.T) {
	// Test creating Ollama provider
	config := &ProviderConfig{
		Type:     "ollama",
		Endpoint: "http://localhost:11434",
		Model:    "llama2",
	}

	provider, err := NewOllamaProvider(config)
	if err != nil {
		t.Errorf("NewOllamaProvider() failed: %v", err)
	}

	if provider == nil {
		t.Error("NewOllamaProvider() should return a provider")
	}

	if provider.GetName() != "ollama" {
		t.Error("Provider name should be 'ollama'")
	}

	// Test creating Ollama provider without endpoint (should use default)
	defaultConfig := &ProviderConfig{
		Type:  "ollama",
		Model: "llama2",
	}

	defaultProvider, err := NewOllamaProvider(defaultConfig)
	if err != nil {
		t.Errorf("NewOllamaProvider() should work with default endpoint: %v", err)
	}

	if defaultProvider == nil {
		t.Error("NewOllamaProvider() should return a provider with default endpoint")
	}
}

func TestProviderConfig_Fields(t *testing.T) {
	// Test valid OpenAI config
	validConfig := &ProviderConfig{
		Type:        "openai",
		APIKey:      "test-key",
		Model:       "gpt-3.5-turbo",
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	if validConfig.Type != "openai" {
		t.Error("Config type should be openai")
	}

	if validConfig.APIKey != "test-key" {
		t.Error("Config API key should be test-key")
	}

	if validConfig.Model != "gpt-3.5-turbo" {
		t.Error("Config model should be gpt-3.5-turbo")
	}

	if validConfig.Temperature != 0.7 {
		t.Error("Config temperature should be 0.7")
	}

	if validConfig.MaxTokens != 1000 {
		t.Error("Config max tokens should be 1000")
	}
}

func TestDefaultProviderConfig(t *testing.T) {
	config := DefaultProviderConfig()

	if config == nil {
		t.Fatal("DefaultProviderConfig() should not return nil")
	}

	if config.Temperature != 0.7 {
		t.Error("Default temperature should be 0.7")
	}

	if config.MaxTokens != 1000 {
		t.Error("Default max tokens should be 1000")
	}

	if config.Headers == nil {
		t.Error("Default headers should be initialized")
	}

	if config.Metadata == nil {
		t.Error("Default metadata should be initialized")
	}
}

func TestMessage_Fields(t *testing.T) {
	// Test valid message
	validMessage := &Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	if validMessage.Role != "user" {
		t.Error("Message role should be user")
	}

	if validMessage.Content != "Hello, world!" {
		t.Error("Message content should be 'Hello, world!'")
	}

	// Test message with tool calls
	messageWithTools := &Message{
		Role:    "assistant",
		Content: "",
		ToolCalls: []ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location": "New York"}`,
				},
			},
		},
	}

	if len(messageWithTools.ToolCalls) != 1 {
		t.Error("Message should have one tool call")
	}

	if messageWithTools.ToolCalls[0].ID != "call_1" {
		t.Error("Tool call ID should be call_1")
	}

	if messageWithTools.ToolCalls[0].Function.Name != "get_weather" {
		t.Error("Function name should be get_weather")
	}
}

func TestCompletionRequest_Fields(t *testing.T) {
	// Test valid completion request
	validRequest := &CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
		Model:       "gpt-3.5-turbo",
	}

	if len(validRequest.Messages) != 1 {
		t.Error("Request should have one message")
	}

	if validRequest.Messages[0].Role != "user" {
		t.Error("Message role should be user")
	}

	if validRequest.Messages[0].Content != "Hello" {
		t.Error("Message content should be Hello")
	}

	if validRequest.Temperature != 0.7 {
		t.Error("Request temperature should be 0.7")
	}

	if validRequest.MaxTokens != 1000 {
		t.Error("Request max tokens should be 1000")
	}

	if validRequest.Model != "gpt-3.5-turbo" {
		t.Error("Request model should be gpt-3.5-turbo")
	}

	// Test request with tools
	requestWithTools := &CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "What's the weather?"},
		},
		Tools: []ToolDefinition{
			{
				Type: "function",
				Function: Function{
					Name:        "get_weather",
					Description: "Get the current weather",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		},
	}

	if len(requestWithTools.Tools) != 1 {
		t.Error("Request should have one tool")
	}

	if requestWithTools.Tools[0].Function.Name != "get_weather" {
		t.Error("Tool function name should be get_weather")
	}
}

func TestCompletionResponse_Fields(t *testing.T) {
	// Test response with choices
	response := &CompletionResponse{
		ID:      "test-id",
		Object:  "chat.completion",
		Model:   "gpt-3.5-turbo",
		Created: 1234567890,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "Hello, how can I help you?",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	if response.ID != "test-id" {
		t.Error("Response ID should be test-id")
	}

	if response.Object != "chat.completion" {
		t.Error("Response object should be chat.completion")
	}

	if response.Model != "gpt-3.5-turbo" {
		t.Error("Response model should be gpt-3.5-turbo")
	}

	if len(response.Choices) != 1 {
		t.Error("Response should have one choice")
	}

	if response.Choices[0].Message.Content != "Hello, how can I help you?" {
		t.Error("Choice message content should match")
	}

	if response.Usage.TotalTokens != 30 {
		t.Error("Usage total tokens should be 30")
	}
}

func TestUsage_Fields(t *testing.T) {
	// Test empty usage
	emptyUsage := &Usage{}
	if emptyUsage.PromptTokens != 0 {
		t.Error("Empty usage prompt tokens should be 0")
	}

	if emptyUsage.CompletionTokens != 0 {
		t.Error("Empty usage completion tokens should be 0")
	}

	if emptyUsage.TotalTokens != 0 {
		t.Error("Empty usage total tokens should be 0")
	}

	// Test non-empty usage
	usage := &Usage{
		PromptTokens:     10,
		CompletionTokens: 20,
		TotalTokens:      30,
	}

	if usage.PromptTokens != 10 {
		t.Error("Usage prompt tokens should be 10")
	}

	if usage.CompletionTokens != 20 {
		t.Error("Usage completion tokens should be 20")
	}

	if usage.TotalTokens != 30 {
		t.Error("Usage total tokens should be 30")
	}
}

func TestChoice_Fields(t *testing.T) {
	// Test choice with message
	choice := Choice{
		Index: 0,
		Message: Message{
			Role:    "assistant",
			Content: "Hello",
		},
		FinishReason: "stop",
	}

	if choice.Index != 0 {
		t.Error("Choice index should be 0")
	}

	if choice.Message.Role != "assistant" {
		t.Error("Choice message role should be assistant")
	}

	if choice.Message.Content != "Hello" {
		t.Error("Choice message content should be Hello")
	}

	if choice.FinishReason != "stop" {
		t.Error("Choice finish reason should be stop")
	}

	// Test choice with delta (for streaming)
	streamChoice := Choice{
		Index: 0,
		Delta: Message{
			Role:    "assistant",
			Content: "Hi",
		},
		FinishReason: "",
	}

	if streamChoice.Delta.Content != "Hi" {
		t.Error("Choice delta content should be Hi")
	}
}

func TestProviderManager(t *testing.T) {
	// Test creating new provider manager
	manager := NewProviderManager()
	if manager == nil {
		t.Fatal("NewProviderManager() should not return nil")
	}

	// Test listing providers (should be empty initially)
	providers := manager.ListProviders()
	if len(providers) != 0 {
		t.Error("New provider manager should have no providers")
	}

	// Test getting default provider (should fail)
	_, err := manager.GetDefaultProvider()
	if err == nil {
		t.Error("Should return error when no default provider is set")
	}
}

func TestConversationHistory(t *testing.T) {
	// Test creating new conversation history
	history := NewConversationHistory()
	if history == nil {
		t.Fatal("NewConversationHistory() should not return nil")
	}

	// Test initial state
	if history.Size() != 0 {
		t.Error("New conversation history should be empty")
	}

	messages := history.GetMessages()
	if len(messages) != 0 {
		t.Error("New conversation history should have no messages")
	}

	// Test adding messages
	msg1 := Message{Role: "user", Content: "Hello"}
	history.AddMessage(msg1)

	if history.Size() != 1 {
		t.Error("History should have one message after adding")
	}

	msg2 := Message{Role: "assistant", Content: "Hi there!"}
	history.AddMessage(msg2)

	if history.Size() != 2 {
		t.Error("History should have two messages after adding second")
	}

	// Test getting messages
	allMessages := history.GetMessages()
	if len(allMessages) != 2 {
		t.Error("Should get all messages")
	}

	if allMessages[0].Content != "Hello" {
		t.Error("First message should be 'Hello'")
	}

	if allMessages[1].Content != "Hi there!" {
		t.Error("Second message should be 'Hi there!'")
	}

	// Test getting last N messages
	lastOne := history.GetLastN(1)
	if len(lastOne) != 1 {
		t.Error("Should get last one message")
	}

	if lastOne[0].Content != "Hi there!" {
		t.Error("Last message should be 'Hi there!'")
	}

	// Test clearing history
	history.Clear()
	if history.Size() != 0 {
		t.Error("History should be empty after clearing")
	}
}

func TestSimpleTokenCounter(t *testing.T) {
	// Test creating token counter
	counter := NewSimpleTokenCounter()
	if counter == nil {
		t.Fatal("NewSimpleTokenCounter() should not return nil")
	}

	// Test counting tokens in text
	text := "Hello, world!"
	tokens, err := counter.CountTokens(text)
	if err != nil {
		t.Errorf("CountTokens() failed: %v", err)
	}

	// Simple approximation: ~4 characters per token
	expectedTokens := len(text) / 4
	if tokens != expectedTokens {
		t.Errorf("Expected %d tokens, got %d", expectedTokens, tokens)
	}

	// Test counting tokens in messages
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	totalTokens, err := counter.CountMessagesTokens(messages)
	if err != nil {
		t.Errorf("CountMessagesTokens() failed: %v", err)
	}

	expectedTotal := (len("Hello") + len("Hi there!")) / 4
	if totalTokens != expectedTotal {
		t.Errorf("Expected %d total tokens, got %d", expectedTotal, totalTokens)
	}
}

// Benchmark tests
func BenchmarkProviderConfig_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &ProviderConfig{
			Type:        "openai",
			APIKey:      "test-key",
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
			MaxTokens:   1000,
		}
	}
}

func BenchmarkMessage_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Message{
			Role:    "user",
			Content: "Hello, world!",
		}
	}
}

func BenchmarkCompletionRequest_Creation(b *testing.B) {
	messages := []Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &CompletionRequest{
			Messages:    messages,
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
			MaxTokens:   1000,
		}
	}
}

func BenchmarkTokenCounter_CountTokens(b *testing.B) {
	counter := NewSimpleTokenCounter()
	text := "Hello, world! This is a test message for benchmarking token counting."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = counter.CountTokens(text)
	}
}

func BenchmarkConversationHistory_AddMessage(b *testing.B) {
	history := NewConversationHistory()
	message := Message{Role: "user", Content: "Hello"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		history.AddMessage(message)
	}
}
