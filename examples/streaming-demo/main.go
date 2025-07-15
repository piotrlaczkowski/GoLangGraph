// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Example: Streaming Mode Demonstration
// This example shows how to use bulk and streaming modes with different LLM providers

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

func main() {
	fmt.Println("🚀 GoLangGraph Streaming Mode Demonstration")
	fmt.Println(strings.Repeat("=", 50))

	// Create LLM provider manager
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Register multiple providers with streaming support
	setupProviders(llmManager)

	// Demonstrate streaming modes
	demonstrateStreamingModes(llmManager)

	// Demonstrate agent streaming
	demonstrateAgentStreaming(llmManager, toolRegistry)

	// Demonstrate real-time streaming
	demonstrateRealTimeStreaming(llmManager)

	fmt.Println("\n✅ Streaming demonstration completed!")
}

func setupProviders(llmManager *llm.ProviderManager) {
	fmt.Println("\n📋 Setting up LLM providers with streaming support...")

	// Setup Ollama provider
	ollamaConfig := llm.DefaultProviderConfig()
	ollamaConfig.Type = "ollama"
	ollamaConfig.Endpoint = "http://localhost:11434"
	ollamaConfig.Model = "gemma3:1b"
	ollamaConfig.Streaming.Enabled = true
	ollamaConfig.Streaming.Mode = llm.StreamModeAuto

	if ollamaProvider, err := llm.NewOllamaProvider(ollamaConfig); err == nil {
		llmManager.RegisterProvider("ollama", ollamaProvider)
		fmt.Println("✅ Ollama provider registered with streaming support")
	} else {
		fmt.Printf("⚠️  Ollama provider setup failed: %v\n", err)
	}

	// Setup Gemini provider (mock for demo)
	geminiConfig := llm.DefaultProviderConfig()
	geminiConfig.Type = "gemini"
	geminiConfig.APIKey = "demo-key"
	geminiConfig.Model = "gemini-pro"
	geminiConfig.Streaming.Enabled = true
	geminiConfig.Streaming.Mode = llm.StreamModeForced

	if geminiProvider, err := llm.NewGeminiProvider(geminiConfig); err == nil {
		llmManager.RegisterProvider("gemini", geminiProvider)
		fmt.Println("✅ Gemini provider registered with streaming support")
	} else {
		fmt.Printf("⚠️  Gemini provider setup failed: %v\n", err)
	}

	fmt.Printf("📊 Total providers registered: %d\n", len(llmManager.ListProviders()))
}

func demonstrateStreamingModes(llmManager *llm.ProviderManager) {
	fmt.Println("\n🔄 Demonstrating Streaming Modes")
	fmt.Println(strings.Repeat("-", 30))

	ctx := context.Background()
	request := llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Explain the benefits of streaming responses in AI applications"},
		},
	}

	providers := llmManager.ListProviders()
	for _, providerName := range providers {
		fmt.Printf("\n📡 Testing provider: %s\n", providerName)

		// Test StreamModeNone - Traditional bulk response
		fmt.Println("  🔸 Mode: None (Bulk)")
		start := time.Now()
		resp, err := llmManager.CompleteWithMode(ctx, providerName, request, llm.StreamModeNone)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("    ❌ Error: %v\n", err)
		} else {
			fmt.Printf("    ✅ Response received in %v\n", duration)
			fmt.Printf("    📝 Content length: %d characters\n", len(resp.Choices[0].Message.Content))
			if len(resp.Choices[0].Message.Content) > 100 {
				fmt.Printf("    📄 Preview: %s...\n", resp.Choices[0].Message.Content[:100])
			}
		}

		// Test StreamModeForced - Streaming but collected
		fmt.Println("  🔸 Mode: Forced (Streaming Collected)")
		start = time.Now()
		resp, err = llmManager.CompleteWithMode(ctx, providerName, request, llm.StreamModeForced)
		duration = time.Since(start)

		if err != nil {
			fmt.Printf("    ❌ Error: %v\n", err)
		} else {
			fmt.Printf("    ✅ Streaming response collected in %v\n", duration)
			fmt.Printf("    📝 Content length: %d characters\n", len(resp.Choices[0].Message.Content))
		}

		// Test real streaming with callback
		fmt.Println("  🔸 Mode: Real-time Streaming")
		start = time.Now()
		var chunks []string
		var firstChunkTime time.Time
		var lastChunkTime time.Time

		callback := func(chunk llm.CompletionResponse) error {
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				if firstChunkTime.IsZero() {
					firstChunkTime = time.Now()
				}
				lastChunkTime = time.Now()
				chunks = append(chunks, chunk.Choices[0].Delta.Content)

				// Show real-time progress
				if len(chunks)%5 == 0 { // Show every 5th chunk
					fmt.Printf("    📦 Chunk %d received...\n", len(chunks))
				}
			}
			return nil
		}

		err = llmManager.CompleteStreamWithMode(ctx, providerName, request, callback, llm.StreamModeForced)
		if err != nil {
			fmt.Printf("    ❌ Streaming error: %v\n", err)
		} else {
			totalDuration := time.Since(start)
			timeToFirstChunk := firstChunkTime.Sub(start)
			streamingDuration := lastChunkTime.Sub(firstChunkTime)

			fmt.Printf("    ✅ Streaming completed: %d chunks\n", len(chunks))
			fmt.Printf("    ⏱️  Time to first chunk: %v\n", timeToFirstChunk)
			fmt.Printf("    ⏱️  Streaming duration: %v\n", streamingDuration)
			fmt.Printf("    ⏱️  Total duration: %v\n", totalDuration)

			combined := strings.Join(chunks, "")
			fmt.Printf("    📝 Total content: %d characters\n", len(combined))
		}
	}
}

func demonstrateAgentStreaming(llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) {
	fmt.Println("\n🤖 Demonstrating Agent Streaming")
	fmt.Println(strings.Repeat("-", 30))

	// Create agent with streaming disabled
	config := agent.DefaultAgentConfig()
	config.Name = "Streaming Demo Agent"
	config.Model = "gemma3:1b"
	config.Provider = "ollama"
	config.SystemPrompt = "You are a helpful AI assistant that explains concepts clearly and concisely."
	config.EnableStreaming = false

	testAgent := agent.NewAgent(config, llmManager, toolRegistry)

	ctx := context.Background()
	input := "Explain the difference between batch processing and streaming in AI"

	// Test without streaming
	fmt.Println("🔸 Agent without streaming:")
	start := time.Now()
	execution, err := testAgent.Execute(ctx, input)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  ✅ Response received in %v\n", duration)
		fmt.Printf("  📝 Response: %s\n", execution.Output[:min(200, len(execution.Output))])
	}

	// Enable streaming on agent
	fmt.Println("\n🔸 Enabling streaming on agent...")
	err = testAgent.EnableStreaming()
	if err != nil {
		fmt.Printf("  ❌ Failed to enable streaming: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Streaming enabled (mode: %s)\n", testAgent.GetStreamingMode())

	// Test with streaming
	fmt.Println("🔸 Agent with streaming:")
	start = time.Now()
	execution, err = testAgent.Execute(ctx, input)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  ✅ Streaming response received in %v\n", duration)
		fmt.Printf("  📝 Response: %s\n", execution.Output[:min(200, len(execution.Output))])
	}

	// Demonstrate different streaming modes
	fmt.Println("\n🔄 Testing different streaming modes:")

	modes := []llm.StreamMode{llm.StreamModeNone, llm.StreamModeAuto, llm.StreamModeForced}
	for _, mode := range modes {
		fmt.Printf("  🔸 Setting mode to: %s\n", mode)
		err := testAgent.SetStreamingMode(mode)
		if err != nil {
			fmt.Printf("    ❌ Failed to set mode: %v\n", err)
			continue
		}

		start := time.Now()
		_, err = testAgent.Execute(ctx, "What are the benefits of "+string(mode)+" mode?")
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("    ❌ Error: %v\n", err)
		} else {
			fmt.Printf("    ✅ %s mode completed in %v\n", mode, duration)
		}
	}
}

func demonstrateRealTimeStreaming(llmManager *llm.ProviderManager) {
	fmt.Println("\n⚡ Real-time Streaming Simulation")
	fmt.Println(strings.Repeat("-", 30))

	ctx := context.Background()
	request := llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "user", Content: "Write a short story about a robot learning to understand human emotions"},
		},
	}

	fmt.Println("📝 Simulating real-time UI updates with streaming...")

	provider := "gemini" // Using Gemini for demo since it works without external dependencies

	var content strings.Builder
	var wordCount int
	startTime := time.Now()

	// Simulate a real-time chat interface
	fmt.Println("\n┌─ Chat Interface Simulation ─────────────────┐")
	fmt.Println("│ 🤖 AI Assistant:                            │")
	fmt.Print("│ ")

	callback := func(chunk llm.CompletionResponse) error {
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			deltaContent := chunk.Choices[0].Delta.Content
			content.WriteString(deltaContent)

			// Count words
			words := strings.Fields(deltaContent)
			wordCount += len(words)

			// Simulate typing effect
			for _, char := range deltaContent {
				fmt.Print(string(char))
				time.Sleep(10 * time.Millisecond) // Simulate typing speed
			}
		}
		return nil
	}

	err := llmManager.CompleteStreamWithMode(ctx, provider, request, callback, llm.StreamModeForced)

	fmt.Println("\n│                                              │")
	fmt.Println("└──────────────────────────────────────────────┘")

	totalTime := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ Streaming error: %v\n", err)
	} else {
		fmt.Printf("\n📊 Streaming Statistics:\n")
		fmt.Printf("  ⏱️  Total time: %v\n", totalTime)
		fmt.Printf("  📝 Words generated: %d\n", wordCount)
		fmt.Printf("  🚀 Words per second: %.1f\n", float64(wordCount)/totalTime.Seconds())
		fmt.Printf("  📊 Characters: %d\n", content.Len())

		if totalTime.Milliseconds() > 0 {
			fmt.Printf("  ⚡ Characters per second: %.1f\n", float64(content.Len())/totalTime.Seconds())
		}
	}
}

// Helper function for minimum
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to repeat strings (for formatting)
func repeat(s string, count int) string {
	return strings.Repeat(s, count)
}
