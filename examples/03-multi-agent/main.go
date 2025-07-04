// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Multi-Agent System Example

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
	"sync"
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

// Agent represents an individual agent in the multi-agent system
type Agent struct {
	name         string
	role         string
	expertise    string
	systemPrompt string
	endpoint     string
	model        string
}

// AgentPool manages multiple agents and their interactions
type AgentPool struct {
	agents   map[string]*Agent
	endpoint string
	model    string
	history  []string
}

// TaskResult represents the result of an agent's task
type TaskResult struct {
	AgentName string
	Task      string
	Result    string
	Duration  time.Duration
}

func main() {
	fmt.Println("👥 GoLangGraph Multi-Agent System")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Welcome to the Multi-Agent System example!")
	fmt.Println()
	fmt.Println("This system includes specialized agents:")
	fmt.Println("  🔍 Research Agent - Gathers information and facts")
	fmt.Println("  📊 Analysis Agent - Analyzes data and patterns")
	fmt.Println("  ✍️  Writing Agent - Creates content and summaries")
	fmt.Println("  🎯 Coordination Agent - Manages tasks and workflow")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit  - Exit the system")
	fmt.Println("  /help          - Show help message")
	fmt.Println("  /agents        - List all agents")
	fmt.Println("  /workflow      - Run a collaborative workflow")
	fmt.Println()

	// Initialize the agent pool
	fmt.Println("🔍 Checking Ollama connection...")
	agentPool := NewAgentPool("http://localhost:11434", "gemma3:1b")

	if err := agentPool.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the model with: ollama pull gemma3:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	fmt.Printf("✅ Multi-agent system initialized with %d agents\n", len(agentPool.agents))
	fmt.Println("✅ System ready for collaborative tasks")
	fmt.Println()

	// Start interactive session
	agentPool.startInteractiveSession()
}

// NewAgentPool creates a new agent pool with specialized agents
func NewAgentPool(endpoint, model string) *AgentPool {
	pool := &AgentPool{
		agents:   make(map[string]*Agent),
		endpoint: endpoint,
		model:    model,
		history:  make([]string, 0),
	}

	// Initialize specialized agents
	pool.initializeAgents()

	return pool
}

// initializeAgents creates specialized agents with different roles
func (ap *AgentPool) initializeAgents() {
	// Research Agent
	ap.agents["researcher"] = &Agent{
		name:         "Research Agent",
		role:         "researcher",
		expertise:    "Information gathering, fact-checking, data collection",
		systemPrompt: "You are a research specialist. Your job is to gather accurate information, verify facts, and provide comprehensive research on topics. Be thorough and cite sources when possible.",
		endpoint:     ap.endpoint,
		model:        ap.model,
	}

	// Analysis Agent
	ap.agents["analyst"] = &Agent{
		name:         "Analysis Agent",
		role:         "analyst",
		expertise:    "Data analysis, pattern recognition, insights generation",
		systemPrompt: "You are an analytical specialist. Your job is to analyze information, identify patterns, draw insights, and provide data-driven conclusions. Be logical and systematic in your approach.",
		endpoint:     ap.endpoint,
		model:        ap.model,
	}

	// Writing Agent
	ap.agents["writer"] = &Agent{
		name:         "Writing Agent",
		role:         "writer",
		expertise:    "Content creation, summarization, communication",
		systemPrompt: "You are a writing specialist. Your job is to create clear, engaging content, summarize complex information, and communicate ideas effectively. Focus on clarity and readability.",
		endpoint:     ap.endpoint,
		model:        ap.model,
	}

	// Coordination Agent
	ap.agents["coordinator"] = &Agent{
		name:         "Coordination Agent",
		role:         "coordinator",
		expertise:    "Task management, workflow coordination, decision making",
		systemPrompt: "You are a coordination specialist. Your job is to manage tasks, coordinate between different agents, make decisions about workflow, and ensure efficient collaboration.",
		endpoint:     ap.endpoint,
		model:        ap.model,
	}
}

// validateConnection checks if Ollama is running and accessible
func (ap *AgentPool) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", ap.endpoint+"/api/tags", nil)
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

// startInteractiveSession runs the interactive multi-agent session
func (ap *AgentPool) startInteractiveSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("👥 Multi-Agent Session Started")
	fmt.Println("Type your task or use commands (type /help for help)")
	fmt.Println("─────────────────────────────────────────────────────")
	fmt.Println()

	for {
		fmt.Print("Task: ")
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
				fmt.Println("\n👋 Multi-agent session ended.")
				break
			}

			if ap.processCommand(userInput) {
				continue
			}

			fmt.Printf("❓ Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands.")
			continue
		}

		// Process the task with multiple agents
		ap.processMultiAgentTask(userInput)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processCommand handles system commands
func (ap *AgentPool) processCommand(command string) bool {
	switch strings.ToLower(command) {
	case "/help":
		ap.showHelp()
		return true
	case "/agents":
		ap.showAgents()
		return true
	case "/workflow":
		ap.runSampleWorkflow()
		return true
	case "/history":
		ap.showHistory()
		return true
	default:
		return false
	}
}

// processMultiAgentTask processes a task using multiple agents
func (ap *AgentPool) processMultiAgentTask(task string) {
	fmt.Printf("\n🎯 Processing task: %s\n", task)
	fmt.Println("─────────────────────────────────────────")

	startTime := time.Now()

	// Step 1: Coordinator decides which agents to use
	coordinatorResult := ap.executeAgentTask("coordinator", fmt.Sprintf("Analyze this task and decide which agents should work on it: %s", task))
	fmt.Printf("📋 Coordinator: %s\n", coordinatorResult.Result)

	// Step 2: Execute tasks in parallel with relevant agents
	var wg sync.WaitGroup
	results := make(chan TaskResult, len(ap.agents))

	// For demonstration, use researcher and analyst in parallel
	agents := []string{"researcher", "analyst"}

	for _, agentRole := range agents {
		wg.Add(1)
		go func(role string) {
			defer wg.Done()

			var agentTask string
			switch role {
			case "researcher":
				agentTask = fmt.Sprintf("Research this topic thoroughly: %s", task)
			case "analyst":
				agentTask = fmt.Sprintf("Analyze this topic and provide insights: %s", task)
			}

			result := ap.executeAgentTask(role, agentTask)
			results <- result
		}(agentRole)
	}

	// Wait for all agents to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var agentResults []TaskResult
	for result := range results {
		agentResults = append(agentResults, result)
		fmt.Printf("✅ %s completed in %s\n", result.AgentName, formatDuration(result.Duration))
	}

	// Step 3: Writer synthesizes the results
	synthesis := ap.synthesizeResults(task, agentResults)

	totalTime := time.Since(startTime)
	fmt.Println("─────────────────────────────────────────")
	fmt.Printf("📝 Final Result:\n%s\n", synthesis)
	fmt.Printf("⏱️  Total processing time: %s\n", formatDuration(totalTime))
	fmt.Println()

	// Add to history
	ap.history = append(ap.history, fmt.Sprintf("Task: %s", task))
	ap.history = append(ap.history, fmt.Sprintf("Result: %s", synthesis))
}

// executeAgentTask executes a task with a specific agent
func (ap *AgentPool) executeAgentTask(agentRole, task string) TaskResult {
	startTime := time.Now()

	agent, exists := ap.agents[agentRole]
	if !exists {
		return TaskResult{
			AgentName: agentRole,
			Task:      task,
			Result:    fmt.Sprintf("Agent %s not found", agentRole),
			Duration:  time.Since(startTime),
		}
	}

	result := ap.callOllama(agent.systemPrompt, task)

	return TaskResult{
		AgentName: agent.name,
		Task:      task,
		Result:    result,
		Duration:  time.Since(startTime),
	}
}

// synthesizeResults combines results from multiple agents
func (ap *AgentPool) synthesizeResults(originalTask string, results []TaskResult) string {
	var combinedResults strings.Builder
	combinedResults.WriteString(fmt.Sprintf("Original task: %s\n\n", originalTask))

	for _, result := range results {
		combinedResults.WriteString(fmt.Sprintf("%s findings:\n%s\n\n", result.AgentName, result.Result))
	}

	synthesisTask := fmt.Sprintf("Synthesize these agent results into a comprehensive, well-structured response:\n%s", combinedResults.String())

	return ap.callOllama(ap.agents["writer"].systemPrompt, synthesisTask)
}

// callOllama makes a request to the Ollama API
func (ap *AgentPool) callOllama(systemPrompt, userPrompt string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fullPrompt := systemPrompt + "\n\n" + userPrompt

	reqBody := OllamaRequest{
		Model:  ap.model,
		Prompt: fullPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Sprintf("Error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ap.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("Error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Ollama returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return fmt.Sprintf("Error unmarshaling response: %v", err)
	}

	return strings.TrimSpace(ollamaResp.Response)
}

// showHelp displays help information
func (ap *AgentPool) showHelp() {
	fmt.Println("\n📚 Help - Multi-Agent System")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("This multi-agent system coordinates specialized agents to solve complex tasks.")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit, /exit   - Exit the system")
	fmt.Println("  /help          - Show this help message")
	fmt.Println("  /agents        - List all agents and their roles")
	fmt.Println("  /workflow      - Run a sample collaborative workflow")
	fmt.Println("  /history       - Show task history")
	fmt.Println()
	fmt.Println("Example tasks:")
	fmt.Println("  • 'Research the benefits of renewable energy'")
	fmt.Println("  • 'Analyze market trends in electric vehicles'")
	fmt.Println("  • 'Create a summary of AI developments in 2024'")
	fmt.Println()
	fmt.Println("How it works:")
	fmt.Println("  1. Coordinator Agent analyzes the task")
	fmt.Println("  2. Specialized agents work in parallel")
	fmt.Println("  3. Writer Agent synthesizes all results")
	fmt.Println("  4. Final comprehensive response is provided")
	fmt.Println()
}

// showAgents displays all available agents
func (ap *AgentPool) showAgents() {
	fmt.Println("\n👥 Available Agents")
	fmt.Println("==================")
	fmt.Println()

	for _, agent := range ap.agents {
		fmt.Printf("🤖 %s\n", agent.name)
		fmt.Printf("   Role: %s\n", agent.role)
		fmt.Printf("   Expertise: %s\n", agent.expertise)
		fmt.Println()
	}
}

// runSampleWorkflow demonstrates a collaborative workflow
func (ap *AgentPool) runSampleWorkflow() {
	fmt.Println("\n🔄 Running Sample Workflow")
	fmt.Println("==========================")
	fmt.Println()

	sampleTask := "Analyze the impact of artificial intelligence on future job markets"
	fmt.Printf("Sample task: %s\n", sampleTask)
	fmt.Println()

	ap.processMultiAgentTask(sampleTask)
}

// showHistory displays task history
func (ap *AgentPool) showHistory() {
	fmt.Println("\n📋 Task History")
	fmt.Println("===============")

	if len(ap.history) == 0 {
		fmt.Println("No task history yet.")
		return
	}

	for i, item := range ap.history {
		fmt.Printf("%d. %s\n", i+1, item)
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
