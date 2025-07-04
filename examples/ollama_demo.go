// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/builder"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
	fmt.Println("🚀 GoLangGraph Ollama Demo with Gemma 3:1B")
	fmt.Println("===========================================")

	ctx := context.Background()

	// Test 1: Basic Chat Agent
	fmt.Println("\n1. Testing Basic Chat Agent...")
	if err := testBasicChatAgent(ctx); err != nil {
		log.Printf("❌ Basic chat test failed: %v", err)
	} else {
		fmt.Println("✅ Basic chat test passed!")
	}

	// Test 2: ReAct Agent with Tools
	fmt.Println("\n2. Testing ReAct Agent with Tools...")
	if err := testReActAgent(ctx); err != nil {
		log.Printf("❌ ReAct agent test failed: %v", err)
	} else {
		fmt.Println("✅ ReAct agent test passed!")
	}

	// Test 3: Multi-Agent Coordination
	fmt.Println("\n3. Testing Multi-Agent Coordination...")
	if err := testMultiAgentCoordination(ctx); err != nil {
		log.Printf("❌ Multi-agent test failed: %v", err)
	} else {
		fmt.Println("✅ Multi-agent test passed!")
	}

	// Test 4: Quick Builder Pattern
	fmt.Println("\n4. Testing Quick Builder Pattern...")
	if err := testQuickBuilder(ctx); err != nil {
		log.Printf("❌ Quick builder test failed: %v", err)
	} else {
		fmt.Println("✅ Quick builder test passed!")
	}

	// Test 5: Graph Execution
	fmt.Println("\n5. Testing Graph Execution...")
	if err := testGraphExecution(ctx); err != nil {
		log.Printf("❌ Graph execution test failed: %v", err)
	} else {
		fmt.Println("✅ Graph execution test passed!")
	}

	// Test 6: Streaming Response
	fmt.Println("\n6. Testing Streaming Response...")
	if err := testStreamingResponse(ctx); err != nil {
		log.Printf("❌ Streaming test failed: %v", err)
	} else {
		fmt.Println("✅ Streaming test passed!")
	}

	fmt.Println("\n🎉 All tests completed! GoLangGraph is working with Ollama and Gemma 3:1B")
}

// setupLLMManager creates and configures the LLM manager with Ollama
func setupLLMManager() (*llm.ProviderManager, error) {
	manager := llm.NewProviderManager()

	// Create Ollama provider configuration
	config := &llm.ProviderConfig{
		Type:        "ollama",
		Endpoint:    "http://localhost:11434",
		Model:       "gemma3:1b",
		Temperature: 0.1,
		MaxTokens:   200,
		Timeout:     60 * time.Second,
	}

	// Create and register Ollama provider
	provider, err := llm.NewOllamaProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama provider: %w", err)
	}

	err = manager.RegisterProvider("ollama", provider)
	if err != nil {
		return nil, fmt.Errorf("failed to register Ollama provider: %w", err)
	}

	// Set as default provider
	err = manager.SetDefaultProvider("ollama")
	if err != nil {
		return nil, fmt.Errorf("failed to set default provider: %w", err)
	}

	return manager, nil
}

// setupTools creates and configures the tool registry
func setupTools() *tools.ToolRegistry {
	registry := tools.NewToolRegistry()

	// Register common tools
	registry.RegisterTool(tools.NewCalculatorTool())
	registry.RegisterTool(tools.NewTimeTool())
	registry.RegisterTool(tools.NewWebSearchTool())
	registry.RegisterTool(tools.NewFileReadTool())
	registry.RegisterTool(tools.NewFileWriteTool())

	return registry
}

// testBasicChatAgent tests basic chat functionality
func testBasicChatAgent(ctx context.Context) error {
	fmt.Println("  Creating LLM manager and tools...")
	llmManager, err := setupLLMManager()
	if err != nil {
		return err
	}
	defer llmManager.Close()

	toolRegistry := setupTools()

	fmt.Println("  Creating chat agent...")
	config := &agent.AgentConfig{
		Name:         "demo-chat",
		Type:         agent.AgentTypeChat,
		Provider:     "ollama",
		Model:        "gemma3:1b",
		Temperature:  0.1,
		MaxTokens:    100,
		SystemPrompt: "You are a helpful AI assistant. Be concise and friendly.",
	}

	chatAgent := agent.NewAgent(config, llmManager, toolRegistry)

	fmt.Println("  Executing chat...")
	execution, err := chatAgent.Execute(ctx, "Hello! Please say 'Hello from Gemma 3:1B!'")
	if err != nil {
		return fmt.Errorf("chat execution failed: %w", err)
	}

	if !execution.Success {
		return fmt.Errorf("chat execution was not successful")
	}

	fmt.Printf("  📝 Response: %s\n", execution.Output)
	return nil
}

// testReActAgent tests ReAct agent with tools
func testReActAgent(ctx context.Context) error {
	fmt.Println("  Creating LLM manager and tools...")
	llmManager, err := setupLLMManager()
	if err != nil {
		return err
	}
	defer llmManager.Close()

	toolRegistry := setupTools()

	fmt.Println("  Creating ReAct agent...")
	config := &agent.AgentConfig{
		Name:          "demo-react",
		Type:          agent.AgentTypeReAct,
		Provider:      "ollama",
		Model:         "gemma3:1b",
		Temperature:   0.1,
		MaxTokens:     200,
		MaxIterations: 3,
		Tools:         []string{"calculator"},
		SystemPrompt:  "You are a helpful assistant that can reason and use tools. Think step by step.",
	}

	reactAgent := agent.NewAgent(config, llmManager, toolRegistry)

	fmt.Println("  Executing ReAct agent...")
	execution, err := reactAgent.Execute(ctx, "What is 25 + 17? Please calculate this.")
	if err != nil {
		return fmt.Errorf("ReAct execution failed: %w", err)
	}

	fmt.Printf("  📝 Response: %s\n", execution.Output)
	return nil
}

// testMultiAgentCoordination tests multi-agent coordination
func testMultiAgentCoordination(ctx context.Context) error {
	fmt.Println("  Creating LLM manager and tools...")
	llmManager, err := setupLLMManager()
	if err != nil {
		return err
	}
	defer llmManager.Close()

	toolRegistry := setupTools()

	fmt.Println("  Creating multiple agents...")
	// Researcher agent
	researcherConfig := &agent.AgentConfig{
		Name:         "researcher",
		Type:         agent.AgentTypeChat,
		Provider:     "ollama",
		Model:        "gemma3:1b",
		Temperature:  0.2,
		MaxTokens:    150,
		SystemPrompt: "You are a researcher. Provide factual information about the given topic.",
	}

	// Writer agent
	writerConfig := &agent.AgentConfig{
		Name:         "writer",
		Type:         agent.AgentTypeChat,
		Provider:     "ollama",
		Model:        "gemma3:1b",
		Temperature:  0.3,
		MaxTokens:    150,
		SystemPrompt: "You are a technical writer. Create a clear summary based on the provided information.",
	}

	researcher := agent.NewAgent(researcherConfig, llmManager, toolRegistry)
	writer := agent.NewAgent(writerConfig, llmManager, toolRegistry)

	fmt.Println("  Creating coordinator...")
	coordinator := agent.NewMultiAgentCoordinator()
	coordinator.AddAgent("researcher", researcher)
	coordinator.AddAgent("writer", writer)

	fmt.Println("  Executing sequential workflow...")
	results, err := coordinator.ExecuteSequential(ctx,
		[]string{"researcher", "writer"},
		"Research and summarize: What is machine learning?")

	if err != nil {
		return fmt.Errorf("multi-agent execution failed: %w", err)
	}

	if len(results) != 2 {
		return fmt.Errorf("expected 2 results, got %d", len(results))
	}

	fmt.Printf("  📝 Researcher: %s\n", results[0].Output)
	fmt.Printf("  📝 Writer: %s\n", results[1].Output)
	return nil
}

// testQuickBuilder tests the quick builder pattern
func testQuickBuilder(ctx context.Context) error {
	fmt.Println("  Creating quick builder...")
	quick := builder.Quick().WithConfig(&builder.QuickConfig{
		DefaultModel:   "gemma3:1b",
		OllamaURL:      "http://localhost:11434",
		Temperature:    0.1,
		MaxTokens:      100,
		EnableAllTools: true,
	})

	fmt.Println("  Creating chat agent with quick builder...")
	chatAgent := quick.Chat("quick-demo")

	fmt.Println("  Executing quick chat...")
	execution, err := chatAgent.Execute(ctx, "Say 'Quick builder works!'")
	if err != nil {
		return fmt.Errorf("quick builder execution failed: %w", err)
	}

	if !execution.Success {
		return fmt.Errorf("quick builder execution was not successful")
	}

	fmt.Printf("  📝 Response: %s\n", execution.Output)

	fmt.Println("  Testing specialized agent...")
	researcher := quick.Researcher("quick-researcher")
	execution, err = researcher.Execute(ctx, "What is artificial intelligence?")
	if err != nil {
		return fmt.Errorf("quick researcher execution failed: %w", err)
	}

	fmt.Printf("  📝 Researcher: %s\n", execution.Output)
	return nil
}

// testGraphExecution tests custom graph execution
func testGraphExecution(ctx context.Context) error {
	fmt.Println("  Creating LLM manager...")
	llmManager, err := setupLLMManager()
	if err != nil {
		return err
	}
	defer llmManager.Close()

	fmt.Println("  Creating custom graph...")
	graph := core.NewGraph("demo-graph")

	// Input processing node
	graph.AddNode("input", "Input Processing", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		input, _ := state.Get("input")
		processed := fmt.Sprintf("Processed: %v", input)
		state.Set("processed_input", processed)
		return state, nil
	})

	// LLM processing node
	graph.AddNode("llm", "LLM Processing", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		processedInput, _ := state.Get("processed_input")

		request := llm.CompletionRequest{
			Messages: []llm.Message{
				{Role: "user", Content: fmt.Sprintf("%v", processedInput)},
			},
			Model:       "gemma3:1b",
			Temperature: 0.1,
			MaxTokens:   100,
		}

		response, err := llmManager.Complete(ctx, "ollama", request)
		if err != nil {
			return state, err
		}

		if len(response.Choices) == 0 {
			return state, fmt.Errorf("no response from LLM")
		}

		state.Set("llm_response", response.Choices[0].Message.Content)
		return state, nil
	})

	// Output processing node
	graph.AddNode("output", "Output Processing", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		llmResponse, _ := state.Get("llm_response")
		finalOutput := fmt.Sprintf("Final: %v", llmResponse)
		state.Set("final_output", finalOutput)
		return state, nil
	})

	// Add edges
	graph.AddEdge("input", "llm", nil)
	graph.AddEdge("llm", "output", nil)

	// Set start and end nodes
	graph.SetStartNode("input")
	graph.AddEndNode("output")

	fmt.Println("  Executing graph...")
	state := core.NewBaseState()
	state.Set("input", "Hello from graph execution!")

	result, err := graph.Execute(ctx, state)
	if err != nil {
		return fmt.Errorf("graph execution failed: %w", err)
	}

	finalOutput, exists := result.Get("final_output")
	if !exists {
		return fmt.Errorf("no final output from graph")
	}

	fmt.Printf("  📝 Graph Result: %s\n", finalOutput)
	return nil
}

// testStreamingResponse tests streaming functionality
func testStreamingResponse(ctx context.Context) error {
	fmt.Println("  Creating LLM manager...")
	llmManager, err := setupLLMManager()
	if err != nil {
		return err
	}
	defer llmManager.Close()

	fmt.Println("  Creating streaming request...")
	request := llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Count from 1 to 5"},
		},
		Model:       "gemma3:1b",
		Temperature: 0.1,
		MaxTokens:   50,
		Stream:      true,
	}

	fmt.Println("  Starting stream...")
	var fullResponse string
	responseCount := 0

	callback := func(chunk llm.CompletionResponse) error {
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Message.Content
			if chunk.Choices[0].Delta.Content != "" {
				content = chunk.Choices[0].Delta.Content
			}
			fullResponse += content
			responseCount++
		}
		return nil
	}

	err = llmManager.CompleteStream(ctx, "ollama", request, callback)
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	if responseCount == 0 {
		return fmt.Errorf("no streaming chunks received")
	}

	fmt.Printf("  📝 Streaming Response (%d chunks): %s\n", responseCount, fullResponse)
	return nil
}
