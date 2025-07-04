// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Tools Integration Example

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
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
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

// Tool represents a tool that can be used by the agent
type Tool interface {
	GetName() string
	GetDescription() string
	Execute(args map[string]interface{}) (string, error)
	GetUsageExamples() []string
}

// ToolsAgent manages tools and interactions
type ToolsAgent struct {
	endpoint string
	model    string
	tools    map[string]Tool
	history  []string
}

// FileSystemTool provides file system operations
type FileSystemTool struct{}

// WebRequestTool provides web request capabilities
type WebRequestTool struct{}

// SystemMonitorTool provides system monitoring
type SystemMonitorTool struct{}

// DataProcessingTool provides data processing utilities
type DataProcessingTool struct{}

// CommandExecutorTool provides command execution
type CommandExecutorTool struct{}

func main() {
	fmt.Println("🛠️  GoLangGraph Tools Integration")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Welcome to the advanced tools integration example!")
	fmt.Println()
	fmt.Println("This agent demonstrates:")
	fmt.Println("  📁 File system operations")
	fmt.Println("  🌐 Web requests and API calls")
	fmt.Println("  📊 System monitoring")
	fmt.Println("  🔧 Data processing utilities")
	fmt.Println("  💻 Command execution")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit  - Exit the agent")
	fmt.Println("  /help          - Show help message")
	fmt.Println("  /tools         - List all available tools")
	fmt.Println("  /tool <name>   - Get detailed tool information")
	fmt.Println("  /examples      - Show tool usage examples")
	fmt.Println()

	// Initialize the tools agent
	fmt.Println("🔍 Checking Ollama connection...")
	agent := NewToolsAgent("http://localhost:11434", "gemma3:1b")

	if err := agent.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the model with: ollama pull gemma3:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	fmt.Printf("✅ Tools agent initialized with %d tools\n", len(agent.tools))
	fmt.Println("✅ Agent ready for tool-enhanced conversations")
	fmt.Println()

	// Start interactive session
	agent.startToolsSession()
}

// NewToolsAgent creates a new tools agent
func NewToolsAgent(endpoint, model string) *ToolsAgent {
	agent := &ToolsAgent{
		endpoint: endpoint,
		model:    model,
		tools:    make(map[string]Tool),
		history:  make([]string, 0),
	}

	// Register tools
	agent.registerTool(&FileSystemTool{})
	agent.registerTool(&WebRequestTool{})
	agent.registerTool(&SystemMonitorTool{})
	agent.registerTool(&DataProcessingTool{})
	agent.registerTool(&CommandExecutorTool{})

	return agent
}

// registerTool registers a tool with the agent
func (t *ToolsAgent) registerTool(tool Tool) {
	t.tools[tool.GetName()] = tool
}

// validateConnection checks if Ollama is running and accessible
func (t *ToolsAgent) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", t.endpoint+"/api/tags", nil)
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

// startToolsSession runs the interactive tools session
func (t *ToolsAgent) startToolsSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🛠️  Tools Session Started")
	fmt.Println("Ask me to use tools for various tasks")
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
				fmt.Println("\n👋 Tools session ended.")
				break
			}

			if t.processCommand(userInput) {
				continue
			}

			fmt.Printf("❓ Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands.")
			continue
		}

		// Process tools input
		t.processToolsInput(userInput)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processCommand handles system commands
func (t *ToolsAgent) processCommand(command string) bool {
	parts := strings.Fields(command)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/help":
		t.showHelp()
		return true
	case "/tools":
		t.listTools()
		return true
	case "/tool":
		if len(parts) > 1 {
			t.showToolDetails(parts[1])
		} else {
			fmt.Println("Usage: /tool <tool_name>")
		}
		return true
	case "/examples":
		t.showExamples()
		return true
	default:
		return false
	}
}

// processToolsInput handles user input and tool execution
func (t *ToolsAgent) processToolsInput(input string) {
	startTime := time.Now()

	// Add user input to history
	t.history = append(t.history, "User: "+input)

	// Analyze input and determine if tools are needed
	toolsNeeded := t.analyzeToolsNeeded(input)

	var response string
	var err error

	if len(toolsNeeded) > 0 {
		// Execute tools and get results
		toolResults := t.executeTools(toolsNeeded, input)

		// Generate response with tool results
		response, err = t.generateToolResponse(input, toolResults)
		if err != nil {
			fmt.Printf("❌ Error generating response: %v\n", err)
			return
		}
	} else {
		// Regular conversation without tools
		context := t.buildContext(input)
		response, err = t.callOllama(context)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			return
		}
	}

	// Add response to history
	t.history = append(t.history, "Assistant: "+response)

	responseTime := time.Since(startTime)
	fmt.Printf("\n🤖 Assistant: %s\n", response)
	fmt.Printf("⏱️  Response time: %s\n", formatDuration(responseTime))
	fmt.Println()
}

// analyzeToolsNeeded analyzes input to determine which tools to use
func (t *ToolsAgent) analyzeToolsNeeded(input string) []string {
	input = strings.ToLower(input)
	var toolsNeeded []string

	// File system operations
	if strings.Contains(input, "file") || strings.Contains(input, "directory") ||
		strings.Contains(input, "folder") || strings.Contains(input, "read") ||
		strings.Contains(input, "write") || strings.Contains(input, "create") ||
		strings.Contains(input, "delete") || strings.Contains(input, "list") {
		toolsNeeded = append(toolsNeeded, "filesystem")
	}

	// Web requests
	if strings.Contains(input, "http") || strings.Contains(input, "url") ||
		strings.Contains(input, "website") || strings.Contains(input, "api") ||
		strings.Contains(input, "download") || strings.Contains(input, "fetch") {
		toolsNeeded = append(toolsNeeded, "web")
	}

	// System monitoring
	if strings.Contains(input, "system") || strings.Contains(input, "cpu") ||
		strings.Contains(input, "memory") || strings.Contains(input, "disk") ||
		strings.Contains(input, "performance") || strings.Contains(input, "monitor") {
		toolsNeeded = append(toolsNeeded, "system")
	}

	// Data processing
	if strings.Contains(input, "process") || strings.Contains(input, "analyze") ||
		strings.Contains(input, "calculate") || strings.Contains(input, "sort") ||
		strings.Contains(input, "filter") || strings.Contains(input, "data") {
		toolsNeeded = append(toolsNeeded, "data")
	}

	// Command execution
	if strings.Contains(input, "run") || strings.Contains(input, "execute") ||
		strings.Contains(input, "command") || strings.Contains(input, "shell") {
		toolsNeeded = append(toolsNeeded, "command")
	}

	return toolsNeeded
}

// executeTools executes the needed tools
func (t *ToolsAgent) executeTools(toolNames []string, input string) map[string]string {
	results := make(map[string]string)

	for _, toolName := range toolNames {
		if tool, exists := t.tools[toolName]; exists {
			// Parse arguments from input (simplified)
			args := t.parseToolArguments(toolName, input)

			result, err := tool.Execute(args)
			if err != nil {
				results[toolName] = fmt.Sprintf("Error: %v", err)
			} else {
				results[toolName] = result
			}
		}
	}

	return results
}

// parseToolArguments parses tool arguments from input (simplified)
func (t *ToolsAgent) parseToolArguments(toolName, input string) map[string]interface{} {
	args := make(map[string]interface{})

	switch toolName {
	case "filesystem":
		if strings.Contains(input, "list") {
			args["action"] = "list"
			args["path"] = "."
		} else if strings.Contains(input, "read") {
			args["action"] = "read"
			// Extract filename if possible
			words := strings.Fields(input)
			for i, word := range words {
				if strings.Contains(word, ".") && i > 0 {
					args["path"] = word
					break
				}
			}
		}
	case "system":
		args["action"] = "status"
	case "web":
		// Extract URL if present
		words := strings.Fields(input)
		for _, word := range words {
			if strings.HasPrefix(word, "http") {
				args["url"] = word
				break
			}
		}
	case "data":
		args["action"] = "analyze"
		args["data"] = input
	case "command":
		// Extract command after "run" or "execute"
		words := strings.Fields(input)
		for i, word := range words {
			if (word == "run" || word == "execute") && i+1 < len(words) {
				args["command"] = strings.Join(words[i+1:], " ")
				break
			}
		}
	}

	return args
}

// generateToolResponse generates a response incorporating tool results
func (t *ToolsAgent) generateToolResponse(input string, toolResults map[string]string) (string, error) {
	var context strings.Builder

	context.WriteString("You are a helpful AI assistant with access to various tools. ")
	context.WriteString("Based on the user's request and the tool results below, provide a comprehensive response.\n\n")

	context.WriteString(fmt.Sprintf("User Request: %s\n\n", input))

	context.WriteString("Tool Results:\n")
	for toolName, result := range toolResults {
		context.WriteString(fmt.Sprintf("- %s: %s\n", toolName, result))
	}

	context.WriteString("\nPlease provide a helpful response based on these results.")

	return t.callOllama(context.String())
}

// buildContext creates context from conversation history
func (t *ToolsAgent) buildContext(currentInput string) string {
	var context strings.Builder

	context.WriteString("You are a helpful AI assistant with access to various tools. ")
	context.WriteString("Provide clear, concise, and helpful responses.\n\n")

	// Add recent conversation history
	start := len(t.history) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(t.history); i++ {
		context.WriteString(t.history[i] + "\n")
	}

	context.WriteString("User: " + currentInput + "\n")
	context.WriteString("Assistant:")

	return context.String()
}

// callOllama makes a request to the Ollama API
func (t *ToolsAgent) callOllama(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := OllamaRequest{
		Model:  t.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
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

// Tool implementations

// FileSystemTool implementation
func (f *FileSystemTool) GetName() string {
	return "filesystem"
}

func (f *FileSystemTool) GetDescription() string {
	return "Provides file system operations: list directories, read files, write files, create directories"
}

func (f *FileSystemTool) Execute(args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action parameter required")
	}

	switch action {
	case "list":
		path := "."
		if p, ok := args["path"].(string); ok {
			path = p
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return "", fmt.Errorf("failed to read directory: %w", err)
		}

		var result strings.Builder
		result.WriteString(fmt.Sprintf("Directory listing for %s:\n", path))
		for _, entry := range entries {
			if entry.IsDir() {
				result.WriteString(fmt.Sprintf("📁 %s/\n", entry.Name()))
			} else {
				info, _ := entry.Info()
				result.WriteString(fmt.Sprintf("📄 %s (%d bytes)\n", entry.Name(), info.Size()))
			}
		}

		return result.String(), nil

	case "read":
		path, ok := args["path"].(string)
		if !ok {
			return "", fmt.Errorf("path parameter required for read action")
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}

		return fmt.Sprintf("Content of %s:\n%s", path, string(content)), nil

	default:
		return "", fmt.Errorf("unsupported action: %s", action)
	}
}

func (f *FileSystemTool) GetUsageExamples() []string {
	return []string{
		"List files in current directory",
		"Read the contents of config.txt",
		"Show me the files in the documents folder",
	}
}

// WebRequestTool implementation
func (w *WebRequestTool) GetName() string {
	return "web"
}

func (w *WebRequestTool) GetDescription() string {
	return "Makes HTTP requests to web APIs and websites"
}

func (w *WebRequestTool) Execute(args map[string]interface{}) (string, error) {
	url, ok := args["url"].(string)
	if !ok {
		return "No URL provided for web request", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return fmt.Sprintf("HTTP %d response from %s:\n%s", resp.StatusCode, url, string(body)[:min(500, len(body))]), nil
}

func (w *WebRequestTool) GetUsageExamples() []string {
	return []string{
		"Fetch data from https://api.github.com/users/octocat",
		"Check the status of https://httpbin.org/status/200",
		"Download content from a URL",
	}
}

// SystemMonitorTool implementation
func (s *SystemMonitorTool) GetName() string {
	return "system"
}

func (s *SystemMonitorTool) GetDescription() string {
	return "Monitors system resources: CPU, memory, disk usage"
}

func (s *SystemMonitorTool) Execute(args map[string]interface{}) (string, error) {
	var result strings.Builder

	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		result.WriteString(fmt.Sprintf("🖥️  CPU Usage: %.1f%%\n", cpuPercent[0]))
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		result.WriteString(fmt.Sprintf("💾 Memory: %.1f%% used (%.1f GB / %.1f GB)\n",
			memInfo.UsedPercent,
			float64(memInfo.Used)/1024/1024/1024,
			float64(memInfo.Total)/1024/1024/1024))
	}

	// Disk usage
	diskInfo, err := disk.Usage("/")
	if err == nil {
		result.WriteString(fmt.Sprintf("💿 Disk: %.1f%% used (%.1f GB / %.1f GB)\n",
			diskInfo.UsedPercent,
			float64(diskInfo.Used)/1024/1024/1024,
			float64(diskInfo.Total)/1024/1024/1024))
	}

	// System info
	result.WriteString(fmt.Sprintf("📊 System: %s %s\n", runtime.GOOS, runtime.GOARCH))
	result.WriteString(fmt.Sprintf("🔧 Go routines: %d\n", runtime.NumGoroutine()))

	return result.String(), nil
}

func (s *SystemMonitorTool) GetUsageExamples() []string {
	return []string{
		"Check system performance",
		"Show me CPU and memory usage",
		"Monitor system resources",
	}
}

// DataProcessingTool implementation
func (d *DataProcessingTool) GetName() string {
	return "data"
}

func (d *DataProcessingTool) GetDescription() string {
	return "Processes and analyzes data: statistics, sorting, filtering"
}

func (d *DataProcessingTool) Execute(args map[string]interface{}) (string, error) {
	data, ok := args["data"].(string)
	if !ok {
		return "", fmt.Errorf("data parameter required")
	}

	// Extract numbers from the data
	words := strings.Fields(data)
	var numbers []float64

	for _, word := range words {
		if num, err := strconv.ParseFloat(word, 64); err == nil {
			numbers = append(numbers, num)
		}
	}

	if len(numbers) == 0 {
		return fmt.Sprintf("Data analysis for: %s\nWord count: %d\nCharacter count: %d",
			data, len(words), len(data)), nil
	}

	// Calculate statistics
	var sum float64
	min := numbers[0]
	max := numbers[0]

	for _, num := range numbers {
		sum += num
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}

	avg := sum / float64(len(numbers))

	return fmt.Sprintf("Data analysis results:\nNumbers found: %d\nSum: %.2f\nAverage: %.2f\nMin: %.2f\nMax: %.2f",
		len(numbers), sum, avg, min, max), nil
}

func (d *DataProcessingTool) GetUsageExamples() []string {
	return []string{
		"Analyze these numbers: 10 20 30 40 50",
		"Process data from a file",
		"Calculate statistics for a dataset",
	}
}

// CommandExecutorTool implementation
func (c *CommandExecutorTool) GetName() string {
	return "command"
}

func (c *CommandExecutorTool) GetDescription() string {
	return "Executes safe system commands (read-only operations)"
}

func (c *CommandExecutorTool) Execute(args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command parameter required")
	}

	// Security: only allow safe, read-only commands
	safeCommands := []string{"ls", "pwd", "date", "whoami", "uname", "ps"}
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmdName := parts[0]
	allowed := false
	for _, safe := range safeCommands {
		if cmdName == safe {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Sprintf("Command '%s' is not allowed for security reasons. Allowed commands: %s",
			cmdName, strings.Join(safeCommands, ", ")), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output)), nil
	}

	return fmt.Sprintf("Command output:\n%s", string(output)), nil
}

func (c *CommandExecutorTool) GetUsageExamples() []string {
	return []string{
		"Run ls to list files",
		"Execute pwd to show current directory",
		"Run date to show current time",
	}
}

// Helper functions

func (t *ToolsAgent) listTools() {
	fmt.Println("\n🛠️  Available Tools")
	fmt.Println("==================")

	for name, tool := range t.tools {
		fmt.Printf("🔹 %s\n", name)
		fmt.Printf("   Description: %s\n", tool.GetDescription())
		fmt.Println()
	}
}

func (t *ToolsAgent) showToolDetails(toolName string) {
	tool, exists := t.tools[toolName]
	if !exists {
		fmt.Printf("❌ Tool not found: %s\n", toolName)
		return
	}

	fmt.Printf("\n🔧 Tool Details: %s\n", tool.GetName())
	fmt.Println("============================")
	fmt.Printf("Description: %s\n", tool.GetDescription())
	fmt.Println("\nUsage Examples:")
	for _, example := range tool.GetUsageExamples() {
		fmt.Printf("  • %s\n", example)
	}
	fmt.Println()
}

func (t *ToolsAgent) showExamples() {
	fmt.Println("\n📚 Tool Usage Examples")
	fmt.Println("======================")
	fmt.Println()
	fmt.Println("File System Operations:")
	fmt.Println("  • 'List files in the current directory'")
	fmt.Println("  • 'Read the contents of README.md'")
	fmt.Println("  • 'Show me what's in the documents folder'")
	fmt.Println()
	fmt.Println("Web Requests:")
	fmt.Println("  • 'Fetch data from https://api.github.com/users/octocat'")
	fmt.Println("  • 'Check the status of https://httpbin.org/get'")
	fmt.Println()
	fmt.Println("System Monitoring:")
	fmt.Println("  • 'Check system performance'")
	fmt.Println("  • 'Show me CPU and memory usage'")
	fmt.Println("  • 'Monitor system resources'")
	fmt.Println()
	fmt.Println("Data Processing:")
	fmt.Println("  • 'Analyze these numbers: 10 20 30 40 50'")
	fmt.Println("  • 'Calculate statistics for this data'")
	fmt.Println()
	fmt.Println("Command Execution:")
	fmt.Println("  • 'Run ls to list files'")
	fmt.Println("  • 'Execute pwd to show current directory'")
	fmt.Println("  • 'Run date to show current time'")
	fmt.Println()
}

func (t *ToolsAgent) showHelp() {
	fmt.Println("\n📚 Help - Tools Integration")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("This agent integrates various tools to enhance AI capabilities.")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit, /exit   - Exit the tools session")
	fmt.Println("  /help          - Show this help message")
	fmt.Println("  /tools         - List all available tools")
	fmt.Println("  /tool <name>   - Get detailed information about a tool")
	fmt.Println("  /examples      - Show comprehensive usage examples")
	fmt.Println()
	fmt.Println("Available Tools:")
	fmt.Println("  🔹 filesystem - File and directory operations")
	fmt.Println("  🔹 web        - HTTP requests and API calls")
	fmt.Println("  🔹 system     - System monitoring and resource usage")
	fmt.Println("  🔹 data       - Data processing and analysis")
	fmt.Println("  🔹 command    - Safe system command execution")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  1. Describe what you want to do")
	fmt.Println("  2. The agent analyzes your request")
	fmt.Println("  3. Appropriate tools are automatically selected and executed")
	fmt.Println("  4. Results are incorporated into the AI response")
	fmt.Println()
	fmt.Println("Security:")
	fmt.Println("  • File operations are limited to safe read operations")
	fmt.Println("  • Command execution is restricted to safe, read-only commands")
	fmt.Println("  • Web requests have timeouts and size limits")
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
