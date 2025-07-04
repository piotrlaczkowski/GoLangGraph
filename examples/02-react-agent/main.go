// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - ReAct Agent Example

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("🧠 GoLangGraph ReAct Agent")
	fmt.Println("==========================")
	fmt.Println()
	fmt.Println("Welcome to the ReAct (Reasoning and Acting) agent example!")
	fmt.Println()
	fmt.Println("This agent can:")
	fmt.Println("  🧮 Perform calculations")
	fmt.Println("  📊 Analyze data")
	fmt.Println("  🔄 Convert units")
	fmt.Println("  💭 Reason through problems step by step")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit  - Exit the chat")
	fmt.Println("  /help          - Show help message")
	fmt.Println("  /tools         - List available tools")
	fmt.Println("  /clear         - Clear conversation history")
	fmt.Println()
	fmt.Println("Just type your message and press Enter to chat!")
	fmt.Println()

	// Initialize the ReAct agent
	fmt.Println("🔍 Checking Ollama connection...")
	agent := NewReActAgent("http://localhost:11434", "orieg/gemma3-tools:1b")

	if err := agent.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the tool-enabled model with: ollama pull orieg/gemma3-tools:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	fmt.Printf("✅ ReAct agent created with model: %s\n", agent.model)
	fmt.Printf("✅ Loaded %d tools\n", len(agent.tools))
	fmt.Println("✅ Agent ready for reasoning and acting")
	fmt.Println()

	// Start interactive chat session
	agent.startChatSession()
}

// startChatSession runs an interactive chat session
func (a *ReActAgent) startChatSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("💬 ReAct Session Started")
	fmt.Println("Type your message or use commands (type /help for help)")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println()

	for {
		// Get user input
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		// Skip empty input
		if userInput == "" {
			continue
		}

		// Check for commands
		if strings.HasPrefix(userInput, "/") {
			if userInput == "/quit" || userInput == "/exit" {
				fmt.Println("\n👋 Goodbye! ReAct session ended.")
				break
			}

			// Process other commands
			if a.processCommand(userInput) {
				continue
			}

			// If command not recognized, show help
			fmt.Printf("❓ Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands.")
			continue
		}

		// Process the user input with ReAct pattern
		a.processReActInput(userInput)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processReActInput handles ReAct reasoning and acting
func (a *ReActAgent) processReActInput(input string) {
	fmt.Printf("\n🧠 Reasoning about: %s\n", input)
	fmt.Println("─────────────────────────────────────────")

	startTime := time.Now()

	// Step 1: Analyze the input and determine if tools are needed
	reasoning := a.analyzeInput(input)
	fmt.Printf("💭 Thought: %s\n", reasoning.thought)

	// Step 2: If action is needed, execute it
	var result string
	if reasoning.needsAction {
		fmt.Printf("🎯 Action: %s\n", reasoning.action)
		actionResult := a.executeAction(reasoning.action, reasoning.actionInput)
		fmt.Printf("📊 Observation: %s\n", actionResult)
		result = actionResult
	}

	// Step 3: Generate final response
	finalResponse := a.generateResponse(input, reasoning, result)
	responseTime := time.Since(startTime)

	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("🤖 ReActAgent: %s\n", finalResponse)
	fmt.Printf("⏱️  Response time: %s\n", formatDuration(responseTime))
	fmt.Println()

	// Add to history
	a.history = append(a.history, fmt.Sprintf("User: %s", input))
	a.history = append(a.history, fmt.Sprintf("Assistant: %s", finalResponse))
}

// processCommand handles special commands
func (a *ReActAgent) processCommand(command string) bool {
	switch strings.ToLower(command) {
	case "/help":
		a.showHelp()
		return true
	case "/tools":
		a.showTools()
		return true
	case "/clear":
		a.history = make([]string, 0)
		fmt.Println("✅ Conversation history cleared.")
		return true
	case "/history":
		a.showHistory()
		return true
	case "/quit", "/exit":
		return false // Signal to exit
	default:
		return false // Not a command
	}
}

// showHelp displays help information
func (a *ReActAgent) showHelp() {
	fmt.Println("\n📚 Help - GoLangGraph ReAct Agent")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("This is a ReAct (Reasoning and Acting) agent that can:")
	fmt.Println("• Think through problems step by step")
	fmt.Println("• Use tools when needed to solve problems")
	fmt.Println("• Provide detailed reasoning for its actions")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit, /exit   - Exit the chat session")
	fmt.Println("  /help          - Show this help message")
	fmt.Println("  /tools         - Show available tools")
	fmt.Println("  /clear         - Clear conversation history")
	fmt.Println("  /history       - Show conversation history")
	fmt.Println()
	fmt.Println("Example queries:")
	fmt.Println("  • 'Calculate the square root of 144'")
	fmt.Println("  • 'Convert 100 fahrenheit to celsius'")
	fmt.Println("  • 'What's the mean of these numbers: 1, 2, 3, 4, 5'")
	fmt.Println("  • 'Calculate compound interest for $1000 at 5% for 10 years'")
	fmt.Println()
	fmt.Println("The agent will show its reasoning process:")
	fmt.Println("  💭 Thought - What the agent is thinking")
	fmt.Println("  🎯 Action - What tool it will use")
	fmt.Println("  📊 Observation - The result of the action")
	fmt.Println("  🤖 Final Response - The agent's conclusion")
	fmt.Println()
}

// showTools displays available tools
func (a *ReActAgent) showTools() {
	fmt.Println("\n🔧 Available Tools")
	fmt.Println("==================")
	fmt.Println()

	for name, tool := range a.tools {
		fmt.Printf("🛠️  %s\n", name)
		fmt.Printf("   Description: %s\n", tool.description)
		fmt.Printf("   Usage: %s\n", tool.usage)
		fmt.Println()
	}
}

// showHistory displays conversation history
func (a *ReActAgent) showHistory() {
	fmt.Println("\n💬 Conversation History")
	fmt.Println("=======================")

	if len(a.history) == 0 {
		fmt.Println("No conversation history yet.")
		return
	}

	for i, msg := range a.history {
		fmt.Printf("%d. %s\n", i+1, msg)
	}
	fmt.Println()
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func init() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
