// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Streaming Example

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

// StreamingAgent handles streaming responses
type StreamingAgent struct {
	endpoint    string
	model       string
	history     []string
	streamMode  string
	showCursor  bool
	colorOutput bool
}

// StreamMode represents different streaming modes
type StreamMode struct {
	Name        string
	Description string
	ChunkSize   int
	Delay       time.Duration
}

var streamModes = map[string]StreamMode{
	"token": {
		Name:        "Token-by-Token",
		Description: "Stream each token individually (slowest, most granular)",
		ChunkSize:   1,
		Delay:       50 * time.Millisecond,
	},
	"word": {
		Name:        "Word-by-Word",
		Description: "Stream each word individually (balanced speed and granularity)",
		ChunkSize:   1,
		Delay:       100 * time.Millisecond,
	},
	"chunk": {
		Name:        "Chunk-by-Chunk",
		Description: "Stream in small chunks (faster, less granular)",
		ChunkSize:   5,
		Delay:       200 * time.Millisecond,
	},
	"sentence": {
		Name:        "Sentence-by-Sentence",
		Description: "Stream complete sentences (fastest, least granular)",
		ChunkSize:   1,
		Delay:       300 * time.Millisecond,
	},
}

// Color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func main() {
	fmt.Println("🌊 GoLangGraph Streaming Agent")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Welcome to the real-time streaming agent example!")
	fmt.Println()
	fmt.Println("This agent demonstrates:")
	fmt.Println("  ⚡ Real-time response streaming")
	fmt.Println("  🎨 Visual streaming effects")
	fmt.Println("  ⚙️  Multiple streaming modes")
	fmt.Println("  📊 Performance metrics")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit  - Exit the agent")
	fmt.Println("  /help          - Show help message")
	fmt.Println("  /modes         - Show streaming modes")
	fmt.Println("  /mode <name>   - Change streaming mode")
	fmt.Println("  /cursor        - Toggle typing cursor")
	fmt.Println("  /color         - Toggle color output")
	fmt.Println()

	// Initialize the streaming agent
	fmt.Println("🔍 Checking Ollama connection...")
	agent := NewStreamingAgent("http://localhost:11434", "gemma3:1b")

	if err := agent.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the model with: ollama pull gemma3:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	fmt.Printf("✅ Streaming agent initialized with mode: %s\n", agent.streamMode)
	fmt.Println("✅ Agent ready for real-time conversations")
	fmt.Println()

	// Start interactive session
	agent.startStreamingSession()
}

// NewStreamingAgent creates a new streaming agent
func NewStreamingAgent(endpoint, model string) *StreamingAgent {
	return &StreamingAgent{
		endpoint:    endpoint,
		model:       model,
		history:     make([]string, 0),
		streamMode:  "word", // Default mode
		showCursor:  true,
		colorOutput: true,
	}
}

// validateConnection checks if Ollama is running and accessible
func (s *StreamingAgent) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", s.endpoint+"/api/tags", nil)
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

// startStreamingSession runs the interactive streaming session
func (s *StreamingAgent) startStreamingSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🌊 Streaming Session Started")
	fmt.Println("Type your message to see real-time streaming")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println()

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		if userInput == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(userInput, "/") {
			if userInput == "/quit" || userInput == "/exit" {
				fmt.Println("\n👋 Streaming session ended.")
				break
			}

			if s.processCommand(userInput) {
				continue
			}

			fmt.Printf("❓ Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands.")
			continue
		}

		// Process streaming response
		s.processStreamingInput(userInput)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processCommand handles system commands
func (s *StreamingAgent) processCommand(command string) bool {
	parts := strings.Fields(command)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		s.showHelp()
		return true
	case "/modes":
		s.showStreamingModes()
		return true
	case "/mode":
		if len(parts) > 1 {
			s.setStreamingMode(parts[1])
		} else {
			fmt.Println("Usage: /mode <mode_name>")
			fmt.Println("Available modes: token, word, chunk, sentence")
		}
		return true
	case "/cursor":
		s.showCursor = !s.showCursor
		fmt.Printf("✅ Typing cursor %s\n", map[bool]string{true: "enabled", false: "disabled"}[s.showCursor])
		return true
	case "/color":
		s.colorOutput = !s.colorOutput
		fmt.Printf("✅ Color output %s\n", map[bool]string{true: "enabled", false: "disabled"}[s.colorOutput])
		return true
	default:
		return false
	}
}

// processStreamingInput handles streaming response generation
func (s *StreamingAgent) processStreamingInput(input string) {
	startTime := time.Now()

	// Add user input to history
	s.history = append(s.history, "User: "+input)

	// Create context
	context := s.buildContext(input)

	// Get response from Ollama
	response, err := s.callOllama(context)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// Add response to history
	s.history = append(s.history, "Assistant: "+response)

	// Stream the response
	fmt.Print("\n🤖 StreamingAgent: ")
	s.streamResponse(response)

	responseTime := time.Since(startTime)
	fmt.Printf("\n⏱️  Response time: %s | Mode: %s\n", formatDuration(responseTime), s.streamMode)
	fmt.Println()
}

// streamResponse streams the response with visual effects
func (s *StreamingAgent) streamResponse(response string) {
	mode := streamModes[s.streamMode]

	switch s.streamMode {
	case "token":
		s.streamByToken(response, mode)
	case "word":
		s.streamByWord(response, mode)
	case "chunk":
		s.streamByChunk(response, mode)
	case "sentence":
		s.streamBySentence(response, mode)
	default:
		s.streamByWord(response, streamModes["word"])
	}
}

// streamByToken streams character by character
func (s *StreamingAgent) streamByToken(response string, mode StreamMode) {
	for i, char := range response {
		if s.colorOutput {
			color := s.getCharColor(i)
			fmt.Printf("%s%c%s", color, char, ColorReset)
		} else {
			fmt.Printf("%c", char)
		}

		if s.showCursor {
			fmt.Print("▋")
			time.Sleep(mode.Delay)
			fmt.Print("\b \b") // Clear cursor
		} else {
			time.Sleep(mode.Delay)
		}
	}
}

// streamByWord streams word by word
func (s *StreamingAgent) streamByWord(response string, mode StreamMode) {
	words := strings.Fields(response)

	for i, word := range words {
		if s.colorOutput {
			color := s.getWordColor(i)
			fmt.Printf("%s%s%s", color, word, ColorReset)
		} else {
			fmt.Print(word)
		}

		if i < len(words)-1 {
			fmt.Print(" ")
		}

		if s.showCursor {
			fmt.Print("▋")
			time.Sleep(mode.Delay)
			fmt.Print("\b \b") // Clear cursor
		} else {
			time.Sleep(mode.Delay)
		}
	}
}

// streamByChunk streams in small chunks
func (s *StreamingAgent) streamByChunk(response string, mode StreamMode) {
	words := strings.Fields(response)

	for i := 0; i < len(words); i += mode.ChunkSize {
		end := i + mode.ChunkSize
		if end > len(words) {
			end = len(words)
		}

		chunk := strings.Join(words[i:end], " ")

		if s.colorOutput {
			color := s.getChunkColor(i / mode.ChunkSize)
			fmt.Printf("%s%s%s", color, chunk, ColorReset)
		} else {
			fmt.Print(chunk)
		}

		if end < len(words) {
			fmt.Print(" ")
		}

		if s.showCursor {
			fmt.Print("▋")
			time.Sleep(mode.Delay)
			fmt.Print("\b \b") // Clear cursor
		} else {
			time.Sleep(mode.Delay)
		}
	}
}

// streamBySentence streams sentence by sentence
func (s *StreamingAgent) streamBySentence(response string, mode StreamMode) {
	sentences := strings.Split(response, ". ")

	for i, sentence := range sentences {
		if i > 0 && i < len(sentences) {
			sentence = ". " + sentence
		}

		if s.colorOutput {
			color := s.getSentenceColor(i)
			fmt.Printf("%s%s%s", color, sentence, ColorReset)
		} else {
			fmt.Print(sentence)
		}

		if s.showCursor {
			fmt.Print("▋")
			time.Sleep(mode.Delay)
			fmt.Print("\b \b") // Clear cursor
		} else {
			time.Sleep(mode.Delay)
		}
	}
}

// Color helper functions
func (s *StreamingAgent) getCharColor(index int) string {
	colors := []string{ColorCyan, ColorGreen, ColorYellow, ColorBlue, ColorPurple}
	return colors[index%len(colors)]
}

func (s *StreamingAgent) getWordColor(index int) string {
	colors := []string{ColorWhite, ColorCyan, ColorGreen, ColorYellow}
	return colors[index%len(colors)]
}

func (s *StreamingAgent) getChunkColor(index int) string {
	colors := []string{ColorBlue, ColorPurple, ColorCyan, ColorGreen}
	return colors[index%len(colors)]
}

func (s *StreamingAgent) getSentenceColor(index int) string {
	colors := []string{ColorWhite, ColorBlue, ColorGreen, ColorCyan}
	return colors[index%len(colors)]
}

// buildContext creates context from conversation history
func (s *StreamingAgent) buildContext(currentInput string) string {
	var context strings.Builder

	// Add system prompt
	context.WriteString("You are a helpful and friendly AI assistant. Provide clear, concise, and helpful responses.\n\n")

	// Add recent conversation history (last 5 exchanges)
	start := len(s.history) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(s.history); i++ {
		context.WriteString(s.history[i] + "\n")
	}

	// Add current input
	context.WriteString("User: " + currentInput + "\n")
	context.WriteString("Assistant:")

	return context.String()
}

// callOllama makes a request to the Ollama API
func (s *StreamingAgent) callOllama(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := OllamaRequest{
		Model:  s.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
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

// setStreamingMode changes the streaming mode
func (s *StreamingAgent) setStreamingMode(mode string) {
	if _, exists := streamModes[mode]; exists {
		s.streamMode = mode
		fmt.Printf("✅ Streaming mode changed to: %s\n", streamModes[mode].Name)
		fmt.Printf("   %s\n", streamModes[mode].Description)
	} else {
		fmt.Printf("❌ Unknown streaming mode: %s\n", mode)
		fmt.Println("Available modes: token, word, chunk, sentence")
	}
}

// showHelp displays help information
func (s *StreamingAgent) showHelp() {
	fmt.Println("\n📚 Help - Streaming Agent")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("This agent demonstrates real-time streaming with visual effects.")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit, /exit   - Exit the streaming session")
	fmt.Println("  /help          - Show this help message")
	fmt.Println("  /modes         - Show available streaming modes")
	fmt.Println("  /mode <name>   - Change streaming mode")
	fmt.Println("  /cursor        - Toggle typing cursor effect")
	fmt.Println("  /color         - Toggle color output")
	fmt.Println()
	fmt.Println("Streaming modes:")
	fmt.Println("  • token    - Character-by-character (slowest, most dramatic)")
	fmt.Println("  • word     - Word-by-word (balanced speed and effect)")
	fmt.Println("  • chunk    - Small chunks (faster, less granular)")
	fmt.Println("  • sentence - Sentence-by-sentence (fastest)")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  ✅ Real-time response streaming")
	fmt.Println("  ✅ Multiple streaming modes")
	fmt.Println("  ✅ Visual effects and colors")
	fmt.Println("  ✅ Typing cursor animation")
	fmt.Println("  ✅ Performance metrics")
	fmt.Println()
}

// showStreamingModes displays all available streaming modes
func (s *StreamingAgent) showStreamingModes() {
	fmt.Println("\n🌊 Streaming Modes")
	fmt.Println("==================")
	fmt.Println()

	for key, mode := range streamModes {
		current := ""
		if key == s.streamMode {
			current = " (current)"
		}

		fmt.Printf("🔹 %s%s\n", mode.Name, current)
		fmt.Printf("   Command: /mode %s\n", key)
		fmt.Printf("   Description: %s\n", mode.Description)
		fmt.Printf("   Delay: %s\n", mode.Delay)
		fmt.Println()
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
