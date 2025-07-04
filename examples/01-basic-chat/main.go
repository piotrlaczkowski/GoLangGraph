// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Basic Chat Agent Example

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// ChatAgent represents a simple chat agent
type ChatAgent struct {
	endpoint string
	model    string
	history  []string
}

// NewChatAgent creates a new chat agent
func NewChatAgent(endpoint, model string) *ChatAgent {
	return &ChatAgent{
		endpoint: endpoint,
		model:    model,
		history:  make([]string, 0),
	}
}

func main() {
	fmt.Println("🤖 GoLangGraph Basic Chat Agent")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Welcome to the basic chat agent example!")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit  - Exit the chat")
	fmt.Println("  /help          - Show help message")
	fmt.Println("  /clear         - Clear conversation history")
	fmt.Println()
	fmt.Println("Just type your message and press Enter to chat!")
	fmt.Println()

	// Initialize the chat agent
	fmt.Println("🔍 Checking Ollama connection...")
	agent := NewChatAgent("http://localhost:11434", "gemma3:1b")

	if err := agent.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the model with: ollama pull gemma3:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	fmt.Printf("✅ Chat agent created with model: %s\n", agent.model)
	fmt.Println("✅ Agent ready for conversation")
	fmt.Println()

	// Start interactive chat session
	agent.startChatSession()
}

// validateConnection checks if Ollama is running and accessible
func (a *ChatAgent) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", a.endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status: %d", resp.StatusCode)
	}

	return nil
}

// startChatSession runs an interactive chat session
func (a *ChatAgent) startChatSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("💬 Chat Session Started")
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
				fmt.Println("\n👋 Goodbye! Chat session ended.")
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

		// Process the user input
		a.processUserInput(userInput)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processUserInput handles a single user input and displays the response
func (a *ChatAgent) processUserInput(input string) {
	// Record start time for performance measurement
	startTime := time.Now()

	// Add user input to history
	a.history = append(a.history, "User: "+input)

	// Create context for the conversation
	context := a.buildContext(input)

	// Call Ollama API
	response, err := a.callOllama(context)
	responseTime := time.Since(startTime)

	if err != nil {
		a.handleError(err)
		return
	}

	// Add response to history
	a.history = append(a.history, "Assistant: "+response)

	// Display the response with formatting
	fmt.Printf("\n🤖 BasicChat: %s\n", response)

	// Display performance metrics
	fmt.Printf("⏱️  Response time: %s\n", formatDuration(responseTime))
	fmt.Println()
}

// buildContext creates context from conversation history
func (a *ChatAgent) buildContext(currentInput string) string {
	var context strings.Builder

	// Add system prompt
	context.WriteString("You are a helpful and friendly AI assistant. Provide clear, concise, and helpful responses.\n\n")

	// Add recent conversation history (last 5 exchanges)
	start := len(a.history) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(a.history); i++ {
		context.WriteString(a.history[i] + "\n")
	}

	// Add current input
	context.WriteString("User: " + currentInput + "\n")
	context.WriteString("Assistant:")

	return context.String()
}

// callOllama makes a request to the Ollama API
func (a *ChatAgent) callOllama(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := OllamaRequest{
		Model:  a.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// processCommand handles special commands
func (a *ChatAgent) processCommand(command string) bool {
	switch strings.ToLower(command) {
	case "/help":
		a.showHelp()
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
func (a *ChatAgent) showHelp() {
	fmt.Println("\n📚 Help - GoLangGraph Basic Chat Agent")
	fmt.Println("======================================")
	fmt.Println()
	fmt.Println("This is a basic chat agent powered by Ollama.")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit, /exit   - Exit the chat session")
	fmt.Println("  /help          - Show this help message")
	fmt.Println("  /clear         - Clear conversation history")
	fmt.Println("  /history       - Show conversation history")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  ✅ Real-time conversation with Gemma3 1B model")
	fmt.Println("  ✅ Conversation history maintained")
	fmt.Println("  ✅ Performance metrics displayed")
	fmt.Println("  ✅ Simple and lightweight implementation")
	fmt.Println()
	fmt.Println("Tips:")
	fmt.Println("  • Ask questions about any topic")
	fmt.Println("  • Request explanations or examples")
	fmt.Println("  • The agent remembers previous messages in the conversation")
	fmt.Println("  • Use /clear to start a fresh conversation")
	fmt.Println()
}

// showHistory displays conversation history
func (a *ChatAgent) showHistory() {
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

// handleError provides user-friendly error messages
func (a *ChatAgent) handleError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("❌ Error: %v\n", err)

	// Provide helpful suggestions based on error type
	errorStr := err.Error()
	switch {
	case strings.Contains(errorStr, "connection refused"):
		fmt.Println("💡 Suggestion: Make sure Ollama is running:")
		fmt.Println("   ollama serve")
	case strings.Contains(errorStr, "model not found"):
		fmt.Println("💡 Suggestion: Pull the required model:")
		fmt.Println("   ollama pull gemma3:1b")
	case strings.Contains(errorStr, "timeout"):
		fmt.Println("💡 Suggestion: The request timed out. Try:")
		fmt.Println("   - Checking your system resources")
		fmt.Println("   - Using a smaller model if available")
	default:
		fmt.Println("💡 Check the troubleshooting section in the README.md")
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
