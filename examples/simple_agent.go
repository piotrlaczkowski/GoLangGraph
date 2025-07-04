package examples

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func SimpleAgentDemo() {
	fmt.Println("GoLangGraph Simple Agent Example")
	fmt.Println("=================================")

	// Initialize LLM Provider Manager
	llmManager := llm.NewProviderManager()

	// Add OpenAI provider if API key is available
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Println("Initializing OpenAI provider...")
		openaiProvider, err := llm.NewOpenAIProvider(&llm.ProviderConfig{
			APIKey:   apiKey,
			Endpoint: "https://api.openai.com/v1",
		})
		if err != nil {
			log.Fatalf("Failed to create OpenAI provider: %v", err)
		}
		llmManager.RegisterProvider("openai", openaiProvider)
		fmt.Println("✓ OpenAI provider registered")
	}

	// Add Ollama provider (local LLM)
	fmt.Println("Initializing Ollama provider...")
	ollamaProvider, err := llm.NewOllamaProvider(&llm.ProviderConfig{
		Endpoint: "http://localhost:11434",
	})
	if err != nil {
		fmt.Printf("⚠ Failed to create Ollama provider: %v\n", err)
		fmt.Println("Make sure Ollama is running on localhost:11434")
	} else {
		llmManager.RegisterProvider("ollama", ollamaProvider)
		fmt.Println("✓ Ollama provider registered")
	}

	// Initialize Tool Registry
	fmt.Println("Initializing tools...")
	toolRegistry := tools.NewToolRegistry()

	// Register basic tools
	toolRegistry.RegisterTool(tools.NewCalculatorTool())
	toolRegistry.RegisterTool(tools.NewWebSearchTool())
	toolRegistry.RegisterTool(tools.NewFileReadTool())
	fmt.Println("✓ Tools registered")

	// Create Agent Configuration
	agentConfig := &agent.AgentConfig{
		Name:          "helpful-assistant",
		Type:          agent.AgentTypeReAct,
		Model:         "gpt-3.5-turbo", // or "llama2" for Ollama
		Provider:      "openai",        // or "ollama"
		SystemPrompt:  "You are a helpful assistant that can use tools to answer questions. Think step by step and use the available tools when needed.",
		Temperature:   0.7,
		MaxTokens:     1000,
		MaxIterations: 5,
		Tools:         []string{"calculator", "web_search", "file_read"},
		Timeout:       30 * time.Second,
	}

	// Create Agent
	fmt.Println("Creating agent...")
	agentInstance := agent.NewAgent(agentConfig, llmManager, toolRegistry)
	fmt.Printf("✓ Agent created: %s (Type: %s)\n", agentInstance.GetConfig().Name, agentInstance.GetConfig().Type)

	// Validate the agent's graph
	graph := agentInstance.GetGraph()
	if err := graph.Validate(); err != nil {
		log.Fatalf("Graph validation failed: %v", err)
	}
	fmt.Println("✓ Agent graph validated")

	// Example 1: Simple Chat
	fmt.Println("\n--- Example 1: Simple Question ---")
	ctx := context.Background()

	execution, err := agentInstance.Execute(ctx, "What is 25 * 34?")
	if err != nil {
		log.Printf("Execution failed: %v", err)
	} else {
		fmt.Printf("Input: %s\n", execution.Input)
		fmt.Printf("Output: %s\n", execution.Output)
		fmt.Printf("Success: %v\n", execution.Success)
		fmt.Printf("Duration: %v\n", execution.Duration)
		if len(execution.ToolCalls) > 0 {
			fmt.Printf("Tools used: %d\n", len(execution.ToolCalls))
			for i, toolCall := range execution.ToolCalls {
				fmt.Printf("  %d. %s: %s\n", i+1, toolCall.Function.Name, toolCall.Function.Arguments)
			}
		}
	}

	// Example 2: Multi-step reasoning
	fmt.Println("\n--- Example 2: Multi-step Reasoning ---")

	execution2, err := agentInstance.Execute(ctx, "I need to calculate the area of a circle with radius 5, then multiply that by 3. What's the final result?")
	if err != nil {
		log.Printf("Execution failed: %v", err)
	} else {
		fmt.Printf("Input: %s\n", execution2.Input)
		fmt.Printf("Output: %s\n", execution2.Output)
		fmt.Printf("Success: %v\n", execution2.Success)
		fmt.Printf("Duration: %v\n", execution2.Duration)
		if len(execution2.ToolCalls) > 0 {
			fmt.Printf("Tools used: %d\n", len(execution2.ToolCalls))
			for i, toolCall := range execution2.ToolCalls {
				fmt.Printf("  %d. %s: %s\n", i+1, toolCall.Function.Name, toolCall.Function.Arguments)
			}
		}
	}

	// Example 3: Show execution history
	fmt.Println("\n--- Execution History ---")
	history := agentInstance.GetExecutionHistory()
	fmt.Printf("Total executions: %d\n", len(history))
	for i, exec := range history {
		fmt.Printf("%d. %s -> %s (Success: %v)\n", i+1,
			truncateString(exec.Input, 50),
			truncateString(exec.Output, 50),
			exec.Success)
	}

	// Example 4: State persistence
	fmt.Println("\n--- State Persistence Example ---")

	// Create a memory checkpointer
	checkpointer := persistence.NewMemoryCheckpointer()

	// Save a checkpoint
	checkpoint := &persistence.Checkpoint{
		ID:        "checkpoint-1",
		ThreadID:  "thread-1",
		State:     agentInstance.GetGraph().GetCurrentState(),
		Metadata:  map[string]interface{}{"example": "checkpoint"},
		CreatedAt: time.Now(),
		NodeID:    "test-node",
		StepID:    1,
	}

	err = checkpointer.Save(ctx, checkpoint)
	if err != nil {
		log.Printf("Failed to save checkpoint: %v", err)
	} else {
		fmt.Println("✓ Checkpoint saved")
	}

	// Load the checkpoint
	loadedCheckpoint, err := checkpointer.Load(ctx, "thread-1", "checkpoint-1")
	if err != nil {
		log.Printf("Failed to load checkpoint: %v", err)
	} else {
		fmt.Printf("✓ Checkpoint loaded: %s\n", loadedCheckpoint.ID)
	}

	fmt.Println("\n--- Example Complete ---")
	fmt.Println("This example demonstrated:")
	fmt.Println("• Creating and configuring an agent")
	fmt.Println("• Executing simple and complex queries")
	fmt.Println("• Using tools for calculations")
	fmt.Println("• Viewing execution history")
	fmt.Println("• Basic state persistence")
	fmt.Println("\nFor more advanced features, see the other examples in this directory.")
}

// Helper function to truncate long strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
