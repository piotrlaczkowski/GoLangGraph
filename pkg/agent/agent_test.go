package agent

import (
	"context"
	"testing"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// Mock LLM provider for testing
type mockProvider struct {
	response string
	err      error
}

func (m *mockProvider) GetName() string {
	return "mock-provider"
}

func (m *mockProvider) GetModels(ctx context.Context) ([]string, error) {
	return []string{"test-model"}, nil
}

func (m *mockProvider) Complete(ctx context.Context, req llm.CompletionRequest) (*llm.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &llm.CompletionResponse{
		ID:      "test-completion",
		Object:  "text_completion",
		Model:   "test-model",
		Created: 1234567890,
		Choices: []llm.Choice{
			{
				Index: 0,
				Message: llm.Message{
					Role:    "assistant",
					Content: m.response,
				},
				FinishReason: "stop",
			},
		},
		Usage: llm.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}, nil
}

func (m *mockProvider) CompleteStream(ctx context.Context, req llm.CompletionRequest, callback llm.StreamCallback) error {
	response, err := m.Complete(ctx, req)
	if err != nil {
		return err
	}
	return callback(*response)
}

func (m *mockProvider) IsHealthy(ctx context.Context) error {
	return nil
}

func (m *mockProvider) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"api_key":     "test-key",
		"model":       "test-model",
		"temperature": 0.7,
		"max_tokens":  1000,
	}
}

func (m *mockProvider) SetConfig(config map[string]interface{}) error {
	return nil
}

func (m *mockProvider) Close() error {
	return nil
}

// Helper function to create test agent
func createTestAgent(t testing.TB, agentType AgentType) *Agent {
	provider := &mockProvider{response: "Hello, World!"}
	llmManager := llm.NewProviderManager()
	err := llmManager.RegisterProvider("mock", provider)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	toolRegistry := tools.NewToolRegistry()

	config := &AgentConfig{
		Name:     "test-agent",
		Type:     agentType,
		Provider: "mock",
		Model:    "test-model",
	}

	return NewAgent(config, llmManager, toolRegistry)
}

func TestNewAgent(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	if agent == nil {
		t.Error("NewAgent() should return a non-nil agent")
	}

	if agent.GetConfig().Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got '%s'", agent.GetConfig().Name)
	}

	if agent.GetConfig().Type != AgentTypeChat {
		t.Errorf("Expected agent type '%s', got '%s'", AgentTypeChat, agent.GetConfig().Type)
	}
}

func TestAgent_GetConfig(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	config := agent.GetConfig()
	if config == nil {
		t.Error("GetConfig() should return a non-nil config")
	}

	if config.Name != "test-agent" {
		t.Errorf("Expected config name 'test-agent', got '%s'", config.Name)
	}
}

func TestAgent_UpdateConfig(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	newConfig := &AgentConfig{
		Name:     "updated-agent",
		Type:     AgentTypeReAct,
		Provider: "mock",
		Model:    "updated-model",
	}

	agent.UpdateConfig(newConfig)

	config := agent.GetConfig()
	if config.Name != "updated-agent" {
		t.Errorf("Expected updated name 'updated-agent', got '%s'", config.Name)
	}

	if config.Type != AgentTypeReAct {
		t.Errorf("Expected updated type '%s', got '%s'", AgentTypeReAct, config.Type)
	}
}

func TestAgent_Execute(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	ctx := context.Background()
	execution, err := agent.Execute(ctx, "Hello")

	if err != nil {
		t.Errorf("Execute() should not return an error, got: %v", err)
	}

	if execution == nil {
		t.Error("Execute() should return a non-nil execution")
	}

	if execution.Input != "Hello" {
		t.Errorf("Expected input 'Hello', got '%s'", execution.Input)
	}

	if execution.Success != true {
		t.Error("Execution should be successful")
	}
}

func TestAgent_GetConversation(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	// Initially empty
	conversation := agent.GetConversation()
	if len(conversation) != 0 {
		t.Errorf("Expected empty conversation, got %d messages", len(conversation))
	}

	// Execute to add to conversation
	ctx := context.Background()
	agent.Execute(ctx, "Hello")

	conversation = agent.GetConversation()
	if len(conversation) == 0 {
		t.Error("Expected conversation to have messages after execution")
	}
}

func TestAgent_ClearConversation(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	// Add some conversation
	ctx := context.Background()
	agent.Execute(ctx, "Hello")

	// Clear conversation
	agent.ClearConversation()

	conversation := agent.GetConversation()
	if len(conversation) != 0 {
		t.Errorf("Expected empty conversation after clear, got %d messages", len(conversation))
	}
}

func TestAgent_GetExecutionHistory(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	// Initially empty
	history := agent.GetExecutionHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d executions", len(history))
	}

	// Execute to add to history
	ctx := context.Background()
	agent.Execute(ctx, "Hello")

	history = agent.GetExecutionHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 execution in history, got %d", len(history))
	}
}

func TestAgent_IsRunning(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	// Initially not running
	if agent.IsRunning() {
		t.Error("Agent should not be running initially")
	}
}

func TestAgent_GetGraph(t *testing.T) {
	agent := createTestAgent(t, AgentTypeChat)

	graph := agent.GetGraph()
	if graph == nil {
		t.Error("GetGraph() should return a non-nil graph")
	}
}

func TestAgentTypes(t *testing.T) {
	testCases := []struct {
		name      string
		agentType AgentType
	}{
		{"Chat Agent", AgentTypeChat},
		{"ReAct Agent", AgentTypeReAct},
		{"Tool Agent", AgentTypeTool},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent := createTestAgent(t, tc.agentType)
			if agent.GetConfig().Type != tc.agentType {
				t.Errorf("Expected agent type '%s', got '%s'", tc.agentType, agent.GetConfig().Type)
			}
		})
	}
}

func TestDefaultAgentConfig(t *testing.T) {
	config := DefaultAgentConfig()

	if config == nil {
		t.Error("DefaultAgentConfig() should return a non-nil config")
	}

	if config.Type != AgentTypeChat {
		t.Errorf("Expected default type '%s', got '%s'", AgentTypeChat, config.Type)
	}

	if config.Temperature != 0.7 {
		t.Errorf("Expected default temperature 0.7, got %f", config.Temperature)
	}

	if config.MaxTokens != 1000 {
		t.Errorf("Expected default max tokens 1000, got %d", config.MaxTokens)
	}
}

func TestMultiAgentCoordinator(t *testing.T) {
	coordinator := NewMultiAgentCoordinator()

	if coordinator == nil {
		t.Error("NewMultiAgentCoordinator() should return a non-nil coordinator")
	}

	// Test adding agents
	agent1 := createTestAgent(t, AgentTypeChat)
	agent2 := createTestAgent(t, AgentTypeReAct)

	coordinator.AddAgent("agent1", agent1)
	coordinator.AddAgent("agent2", agent2)

	// Test getting agents
	retrievedAgent1, exists := coordinator.GetAgent("agent1")
	if !exists || retrievedAgent1 != agent1 {
		t.Error("Should be able to retrieve added agent")
	}

	// Test listing agents
	agentIDs := coordinator.ListAgents()
	if len(agentIDs) != 2 {
		t.Errorf("Expected 2 agents, got %d", len(agentIDs))
	}

	// Test removing agents
	coordinator.RemoveAgent("agent1")
	_, exists = coordinator.GetAgent("agent1")
	if exists {
		t.Error("Agent should be removed")
	}
}

// Benchmark tests
func BenchmarkAgent_Execute(b *testing.B) {
	agent := createTestAgent(b, AgentTypeChat)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.Execute(ctx, "Benchmark message")
	}
}

func BenchmarkAgent_GetConfig(b *testing.B) {
	agent := createTestAgent(b, AgentTypeChat)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.GetConfig()
	}
}

func BenchmarkAgent_GetConversation(b *testing.B) {
	agent := createTestAgent(b, AgentTypeChat)

	// Add some conversation history
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		agent.Execute(ctx, "Message")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent.GetConversation()
	}
}
