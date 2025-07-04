// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

//go:build integration
// +build integration

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/builder"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// OllamaE2ETestSuite provides end-to-end testing with Ollama and Gemma 3:1B
type OllamaE2ETestSuite struct {
	suite.Suite
	ollamaRunning bool
	modelPulled   bool
	ctx           context.Context
	cancel        context.CancelFunc
	llmManager    *llm.ProviderManager
	toolRegistry  *tools.ToolRegistry
}

// SetupSuite runs before all tests in the suite
func (suite *OllamaE2ETestSuite) SetupSuite() {
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 15*time.Minute)

	// Check if tests should run
	if os.Getenv("SKIP_OLLAMA_TESTS") == "true" {
		suite.T().Skip("Skipping Ollama integration tests (SKIP_OLLAMA_TESTS=true)")
	}

	// Setup Ollama and Gemma 3:1B
	suite.setupOllama()

	// Setup LLM manager and tools
	suite.setupLLMManager()
	suite.setupTools()
}

// TearDownSuite runs after all tests in the suite
func (suite *OllamaE2ETestSuite) TearDownSuite() {
	if suite.llmManager != nil {
		suite.llmManager.Close()
	}
	if suite.cancel != nil {
		suite.cancel()
	}
}

// setupOllama ensures Ollama is running and Gemma 3:1B is available
func (suite *OllamaE2ETestSuite) setupOllama() {
	// Check if Ollama is installed
	if !suite.isOllamaInstalled() {
		suite.T().Skip("Ollama not installed. Install with: curl -fsSL https://ollama.ai/install.sh | sh")
	}

	// Start Ollama if not running
	if !suite.isOllamaRunning() {
		suite.T().Log("Starting Ollama server...")
		if err := suite.startOllama(); err != nil {
			suite.T().Skipf("Failed to start Ollama: %v", err)
		}
		// Wait for Ollama to be ready
		suite.waitForOllama()
	}
	suite.ollamaRunning = true

	// Pull Gemma 3:1B model if not available
	if !suite.isModelAvailable("gemma3:1b") {
		suite.T().Log("Pulling Gemma 3:1B model (this may take a few minutes)...")
		if err := suite.pullModel("gemma3:1b"); err != nil {
			suite.T().Skipf("Failed to pull Gemma 3:1B model: %v", err)
		}
	}
	suite.modelPulled = true

	suite.T().Log("✅ Ollama setup complete with Gemma 3:1B model")
}

// setupLLMManager creates and configures the LLM manager
func (suite *OllamaE2ETestSuite) setupLLMManager() {
	suite.llmManager = llm.NewProviderManager()

	// Create Ollama provider
	ollamaConfig := &llm.ProviderConfig{
		Type:        "ollama",
		Endpoint:    "http://localhost:11434",
		Model:       "gemma3:1b",
		Temperature: 0.1,
		MaxTokens:   200,
		Timeout:     60 * time.Second,
	}

	ollamaProvider, err := llm.NewOllamaProvider(ollamaConfig)
	if err != nil {
		suite.T().Skipf("Failed to create Ollama provider: %v", err)
	}

	err = suite.llmManager.RegisterProvider("ollama", ollamaProvider)
	if err != nil {
		suite.T().Skipf("Failed to register Ollama provider: %v", err)
	}

	// Set as default provider
	err = suite.llmManager.SetDefaultProvider("ollama")
	if err != nil {
		suite.T().Skipf("Failed to set default provider: %v", err)
	}
}

// setupTools creates and configures the tool registry
func (suite *OllamaE2ETestSuite) setupTools() {
	suite.toolRegistry = tools.NewToolRegistry()

	// Register common tools
	suite.toolRegistry.RegisterTool(tools.NewCalculatorTool())
	suite.toolRegistry.RegisterTool(tools.NewTimeTool())
	suite.toolRegistry.RegisterTool(tools.NewWebSearchTool())
	suite.toolRegistry.RegisterTool(tools.NewFileReadTool())
	suite.toolRegistry.RegisterTool(tools.NewFileWriteTool())
}

// Test basic agent functionality with Ollama
func (suite *OllamaE2ETestSuite) TestBasicChatAgent() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Create agent configuration
	config := &agent.AgentConfig{
		Name:        "test-chat",
		Type:        agent.AgentTypeChat,
		Provider:    "ollama",
		Model:       "gemma3:1b",
		Temperature: 0.1,
		MaxTokens:   100,
	}

	// Create chat agent
	chatAgent := agent.NewAgent(config, suite.llmManager, suite.toolRegistry)
	require.NotNil(suite.T(), chatAgent)

	// Test basic conversation
	execution, err := chatAgent.Execute(suite.ctx, "Hello! Please respond with exactly: 'Hello from Gemma'")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), execution)
	require.True(suite.T(), execution.Success)

	suite.T().Logf("Chat Agent Response: %s", execution.Output)
	assert.Contains(suite.T(), strings.ToLower(execution.Output), "hello")
}

// Test ReAct agent with tools
func (suite *OllamaE2ETestSuite) TestReActAgent() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Create agent configuration
	config := &agent.AgentConfig{
		Name:          "test-react",
		Type:          agent.AgentTypeReAct,
		Provider:      "ollama",
		Model:         "gemma3:1b",
		Temperature:   0.1,
		MaxTokens:     200,
		MaxIterations: 3,
		Tools:         []string{"calculator"},
	}

	// Create ReAct agent
	reactAgent := agent.NewAgent(config, suite.llmManager, suite.toolRegistry)
	require.NotNil(suite.T(), reactAgent)

	// Test with calculator tool
	execution, err := reactAgent.Execute(suite.ctx, "What is 15 + 27? Please calculate this step by step.")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), execution)

	suite.T().Logf("ReAct Agent Response: %s", execution.Output)
	// Should contain some mathematical result
	assert.True(suite.T(), strings.Contains(execution.Output, "42") || strings.Contains(execution.Output, "15") || strings.Contains(execution.Output, "27"))
}

// Test multi-agent coordination
func (suite *OllamaE2ETestSuite) TestMultiAgentCoordination() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Create agent configurations
	researcherConfig := &agent.AgentConfig{
		Name:         "researcher",
		Type:         agent.AgentTypeChat,
		Provider:     "ollama",
		Model:        "gemma3:1b",
		Temperature:  0.2,
		MaxTokens:    150,
		SystemPrompt: "You are a researcher. Provide factual information about the topic.",
	}

	writerConfig := &agent.AgentConfig{
		Name:         "writer",
		Type:         agent.AgentTypeChat,
		Provider:     "ollama",
		Model:        "gemma3:1b",
		Temperature:  0.3,
		MaxTokens:    150,
		SystemPrompt: "You are a writer. Create a summary based on the provided information.",
	}

	// Create agents
	researcher := agent.NewAgent(researcherConfig, suite.llmManager, suite.toolRegistry)
	writer := agent.NewAgent(writerConfig, suite.llmManager, suite.toolRegistry)

	// Create multi-agent coordinator
	coordinator := agent.NewMultiAgentCoordinator()
	coordinator.AddAgent("researcher", researcher)
	coordinator.AddAgent("writer", writer)

	// Execute sequential workflow
	results, err := coordinator.ExecuteSequential(suite.ctx,
		[]string{"researcher", "writer"},
		"Research the topic 'Go programming language' and then write a brief summary")

	require.NoError(suite.T(), err)
	require.Len(suite.T(), results, 2)

	suite.T().Logf("Researcher Response: %s", results[0].Output)
	suite.T().Logf("Writer Response: %s", results[1].Output)

	// Both should have some content
	assert.NotEmpty(suite.T(), results[0].Output)
	assert.NotEmpty(suite.T(), results[1].Output)
}

// Test quick builder patterns
func (suite *OllamaE2ETestSuite) TestQuickBuilderPatterns() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Test one-line builders with custom configuration
	quick := builder.Quick().WithConfig(&builder.QuickConfig{
		DefaultModel:   "gemma3:1b",
		OllamaURL:      "http://localhost:11434",
		Temperature:    0.1,
		MaxTokens:      100,
		EnableAllTools: true,
	})

	// Test chat agent
	chatAgent := quick.Chat("quick-chat")
	require.NotNil(suite.T(), chatAgent)

	execution, err := chatAgent.Execute(suite.ctx, "Say 'Quick chat works!'")
	require.NoError(suite.T(), err)
	require.True(suite.T(), execution.Success)
	suite.T().Logf("Quick Chat Response: %s", execution.Output)

	// Test specialized agents
	researcher := quick.Researcher("quick-researcher")
	require.NotNil(suite.T(), researcher)

	execution, err = researcher.Execute(suite.ctx, "Research: What is artificial intelligence?")
	require.NoError(suite.T(), err)
	require.True(suite.T(), execution.Success)
	suite.T().Logf("Quick Researcher Response: %s", execution.Output)
	assert.Contains(suite.T(), strings.ToLower(execution.Output), "artificial")
}

// Test graph execution with custom nodes
func (suite *OllamaE2ETestSuite) TestGraphExecution() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Create a custom graph with multiple nodes
	graph := core.NewGraph("test-graph")

	// Add input processing node
	inputNode := graph.AddNode("input", "Input Processing", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		input := state.Get("input")
		processed := fmt.Sprintf("Processed: %v", input)
		state.Set("processed_input", processed)
		return state, nil
	})
	require.NotNil(suite.T(), inputNode)

	// Add LLM processing node
	llmNode := graph.AddNode("llm", "LLM Processing", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		processedInput := state.Get("processed_input")

		// Create completion request
		request := llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: "user", Content: fmt.Sprintf("%v", processedInput)},
			},
			Model:       "gemma3:1b",
			Temperature: 0.1,
			MaxTokens:   100,
		}

		// Get response from LLM
		response, err := suite.llmManager.Complete(ctx, "ollama", request)
		if err != nil {
			return state, err
		}

		if len(response.Choices) == 0 {
			return state, fmt.Errorf("no response from LLM")
		}

		state.Set("llm_response", response.Choices[0].Message.Content)
		return state, nil
	})
	require.NotNil(suite.T(), llmNode)

	// Add output processing node
	outputNode := graph.AddNode("output", "Output Processing", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		llmResponse := state.Get("llm_response")
		finalOutput := fmt.Sprintf("Final: %v", llmResponse)
		state.Set("final_output", finalOutput)
		return state, nil
	})
	require.NotNil(suite.T(), outputNode)

	// Add edges
	graph.AddEdge("input", "llm", nil)
	graph.AddEdge("llm", "output", nil)

	// Set start and end nodes
	graph.SetStartNode("input")
	graph.AddEndNode("output")

	// Execute graph
	state := core.NewBaseState()
	state.Set("input", "Hello from graph execution test")

	result, err := graph.Execute(suite.ctx, state)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), result)

	finalOutput := result.Get("final_output")
	require.NotNil(suite.T(), finalOutput)
	suite.T().Logf("Graph Execution Result: %s", finalOutput)
	assert.Contains(suite.T(), fmt.Sprintf("%v", finalOutput), "Final:")
}

// Test streaming responses
func (suite *OllamaE2ETestSuite) TestStreamingResponse() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Create streaming request
	request := llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Count from 1 to 3"},
		},
		Model:       "gemma3:1b",
		Temperature: 0.1,
		MaxTokens:   50,
		Stream:      true,
	}

	// Collect streaming responses
	var fullResponse strings.Builder
	responseCount := 0

	// Create callback for streaming
	callback := func(chunk llm.CompletionResponse) error {
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Message.Content
			if chunk.Choices[0].Delta.Content != "" {
				content = chunk.Choices[0].Delta.Content
			}
			fullResponse.WriteString(content)
			responseCount++
		}
		return nil
	}

	// Start streaming
	err := suite.llmManager.CompleteStream(suite.ctx, "ollama", request, callback)
	require.NoError(suite.T(), err)

	suite.T().Logf("Streaming Response (%d chunks): %s", responseCount, fullResponse.String())
	assert.Greater(suite.T(), responseCount, 0, "Should receive streaming chunks")
	assert.NotEmpty(suite.T(), fullResponse.String(), "Should receive content")
}

// Test conversation history and context
func (suite *OllamaE2ETestSuite) TestConversationHistory() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Create chat agent
	config := &agent.AgentConfig{
		Name:        "conversation-test",
		Type:        agent.AgentTypeChat,
		Provider:    "ollama",
		Model:       "gemma3:1b",
		Temperature: 0.1,
		MaxTokens:   100,
	}

	chatAgent := agent.NewAgent(config, suite.llmManager, suite.toolRegistry)
	require.NotNil(suite.T(), chatAgent)

	// First message
	execution1, err := chatAgent.Execute(suite.ctx, "My name is Alice. Remember this.")
	require.NoError(suite.T(), err)
	require.True(suite.T(), execution1.Success)
	suite.T().Logf("First Response: %s", execution1.Output)

	// Second message referencing the first
	execution2, err := chatAgent.Execute(suite.ctx, "What is my name?")
	require.NoError(suite.T(), err)
	require.True(suite.T(), execution2.Success)
	suite.T().Logf("Second Response: %s", execution2.Output)

	// Check conversation history
	history := chatAgent.GetConversation()
	assert.GreaterOrEqual(suite.T(), len(history), 2) // At least user messages

	// Verify responses exist
	assert.NotEmpty(suite.T(), execution1.Output)
	assert.NotEmpty(suite.T(), execution2.Output)
}

// Test error handling and recovery
func (suite *OllamaE2ETestSuite) TestErrorHandling() {
	if !suite.ollamaRunning || !suite.modelPulled {
		suite.T().Skip("Ollama or model not available")
	}

	// Test with invalid model
	config := &agent.AgentConfig{
		Name:        "error-test",
		Type:        agent.AgentTypeChat,
		Provider:    "ollama",
		Model:       "nonexistent-model",
		Temperature: 0.1,
		MaxTokens:   100,
	}

	errorAgent := agent.NewAgent(config, suite.llmManager, suite.toolRegistry)
	require.NotNil(suite.T(), errorAgent)

	// This should fail gracefully
	execution, err := errorAgent.Execute(suite.ctx, "This should fail")

	// Either the execution should fail or the agent should handle it gracefully
	if err != nil {
		suite.T().Logf("Expected error occurred: %v", err)
	} else {
		suite.T().Logf("Agent handled error gracefully: %s", execution.Output)
		assert.False(suite.T(), execution.Success)
	}
}

// Helper methods

func (suite *OllamaE2ETestSuite) isOllamaInstalled() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

func (suite *OllamaE2ETestSuite) isOllamaRunning() bool {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func (suite *OllamaE2ETestSuite) startOllama() error {
	cmd := exec.Command("ollama", "serve")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ollama: %w", err)
	}
	return nil
}

func (suite *OllamaE2ETestSuite) waitForOllama() {
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		if suite.isOllamaRunning() {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (suite *OllamaE2ETestSuite) isModelAvailable(model string) bool {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var tags struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return false
	}

	for _, m := range tags.Models {
		if strings.HasPrefix(m.Name, model) {
			return true
		}
	}
	return false
}

func (suite *OllamaE2ETestSuite) pullModel(model string) error {
	cmd := exec.CommandContext(suite.ctx, "ollama", "pull", model)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull model %s: %w\nOutput: %s", model, err, output)
	}
	return nil
}

// Run the test suite
func TestOllamaE2ETestSuite(t *testing.T) {
	suite.Run(t, new(OllamaE2ETestSuite))
}
