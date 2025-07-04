// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - Complex Workflow Graph with ReAct Agent

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
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
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

// WorkflowState represents the state that flows through the graph
type WorkflowState struct {
	ID          string                 `json:"id"`
	Input       string                 `json:"input"`
	CurrentNode string                 `json:"current_node"`
	Context     map[string]interface{} `json:"context"`
	History     []NodeExecution        `json:"history"`
	Result      string                 `json:"result"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]string      `json:"metadata"`
}

// NodeExecution represents a single node execution
type NodeExecution struct {
	NodeID    string                 `json:"node_id"`
	NodeType  string                 `json:"node_type"`
	Input     interface{}            `json:"input"`
	Output    interface{}            `json:"output"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Node represents a workflow node
type Node interface {
	GetID() string
	GetType() string
	Execute(state *WorkflowState) (*WorkflowState, error)
	GetNextNodes(state *WorkflowState) []string
	GetDescription() string
}

// Edge represents a connection between nodes
type Edge struct {
	From      string                                    `json:"from"`
	To        string                                    `json:"to"`
	Condition func(state *WorkflowState) bool           `json:"-"`
	Transform func(state *WorkflowState) *WorkflowState `json:"-"`
	Label     string                                    `json:"label"`
}

// WorkflowGraph represents the complete workflow graph
type WorkflowGraph struct {
	Nodes          map[string]Node   `json:"-"`
	Edges          []Edge            `json:"edges"`
	StartNode      string            `json:"start_node"`
	EndNodes       []string          `json:"end_nodes"`
	Metadata       map[string]string `json:"metadata"`
	ollamaEndpoint string
	ollamaModel    string
}

// ReActAgent implements the ReAct (Reasoning and Acting) pattern
type ReActAgent struct {
	ID       string
	Tools    map[string]Tool
	MaxSteps int
	endpoint string
	model    string
}

// Tool interface for ReAct agent tools
type Tool interface {
	GetName() string
	GetDescription() string
	Execute(args map[string]interface{}) (string, error)
	GetSchema() map[string]interface{}
}

// Specific Node implementations

// InputNode - Entry point of the workflow
type InputNode struct {
	ID string
}

// AnalysisNode - Analyzes the input and determines workflow path
type AnalysisNode struct {
	ID       string
	endpoint string
	model    string
}

// ReActNode - Implements ReAct agent reasoning
type ReActNode struct {
	ID    string
	Agent *ReActAgent
}

// DecisionNode - Makes routing decisions based on analysis
type DecisionNode struct {
	ID string
}

// TaskExecutionNode - Executes specific tasks
type TaskExecutionNode struct {
	ID       string
	TaskType string
	endpoint string
	model    string
}

// AggregationNode - Combines results from multiple paths
type AggregationNode struct {
	ID       string
	endpoint string
	model    string
}

// OutputNode - Final output formatting
type OutputNode struct {
	ID string
}

// Tool implementations for ReAct agent

// CalculatorTool for mathematical operations
type CalculatorTool struct{}

// WebSearchTool for web searches (simulated)
type WebSearchTool struct{}

// DataAnalysisTool for data analysis
type DataAnalysisTool struct{}

// PlannerTool for task planning
type PlannerTool struct{}

func main() {
	fmt.Println("🔄 GoLangGraph Complex Workflow Graph")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("Welcome to the advanced workflow graph example!")
	fmt.Println()
	fmt.Println("This system demonstrates:")
	fmt.Println("  🔄 Complex multi-node workflows")
	fmt.Println("  🎯 Conditional edges and routing")
	fmt.Println("  🤖 ReAct agent integration")
	fmt.Println("  📊 State management and tracking")
	fmt.Println("  🔧 Dynamic workflow execution")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /quit or /exit  - Exit the system")
	fmt.Println("  /help          - Show help message")
	fmt.Println("  /graph         - Show workflow graph structure")
	fmt.Println("  /state         - Show current workflow state")
	fmt.Println("  /history       - Show execution history")
	fmt.Println("  /reset         - Reset workflow state")
	fmt.Println()

	// Initialize workflow graph
	fmt.Println("🔍 Checking Ollama connection...")
	graph, err := NewWorkflowGraph("http://localhost:11434", "gemma3:1b")
	if err != nil {
		fmt.Printf("❌ Failed to initialize workflow graph: %v\n", err)
		return
	}

	if err := graph.validateConnection(); err != nil {
		fmt.Printf("❌ Ollama connection failed: %v\n", err)
		fmt.Println("Please ensure Ollama is running and accessible at http://localhost:11434")
		fmt.Println("Start Ollama with: ollama serve")
		fmt.Println("Pull the model with: ollama pull gemma3:1b")
		return
	}
	fmt.Println("✅ Ollama connection successful")

	fmt.Println("✅ Workflow graph initialized")
	fmt.Printf("   📊 Nodes: %d | Edges: %d\n", len(graph.Nodes), len(graph.Edges))
	fmt.Println("✅ ReAct agent ready with tools")
	fmt.Println()

	// Start interactive session
	graph.startWorkflowSession()
}

// NewWorkflowGraph creates and configures a new workflow graph
func NewWorkflowGraph(endpoint, model string) (*WorkflowGraph, error) {
	graph := &WorkflowGraph{
		Nodes:          make(map[string]Node),
		Edges:          make([]Edge, 0),
		EndNodes:       []string{"output"},
		Metadata:       make(map[string]string),
		ollamaEndpoint: endpoint,
		ollamaModel:    model,
	}

	// Initialize ReAct agent with tools
	agent := &ReActAgent{
		ID:       "react-agent-1",
		Tools:    make(map[string]Tool),
		MaxSteps: 10,
		endpoint: endpoint,
		model:    model,
	}

	// Register tools
	agent.Tools["calculator"] = &CalculatorTool{}
	agent.Tools["web_search"] = &WebSearchTool{}
	agent.Tools["data_analysis"] = &DataAnalysisTool{}
	agent.Tools["planner"] = &PlannerTool{}

	// Create nodes
	graph.Nodes["input"] = &InputNode{ID: "input"}
	graph.Nodes["analysis"] = &AnalysisNode{ID: "analysis", endpoint: endpoint, model: model}
	graph.Nodes["react"] = &ReActNode{ID: "react", Agent: agent}
	graph.Nodes["decision"] = &DecisionNode{ID: "decision"}
	graph.Nodes["task_math"] = &TaskExecutionNode{ID: "task_math", TaskType: "mathematical", endpoint: endpoint, model: model}
	graph.Nodes["task_research"] = &TaskExecutionNode{ID: "task_research", TaskType: "research", endpoint: endpoint, model: model}
	graph.Nodes["task_analysis"] = &TaskExecutionNode{ID: "task_analysis", TaskType: "analysis", endpoint: endpoint, model: model}
	graph.Nodes["aggregation"] = &AggregationNode{ID: "aggregation", endpoint: endpoint, model: model}
	graph.Nodes["output"] = &OutputNode{ID: "output"}

	// Define edges with conditions
	graph.Edges = []Edge{
		{From: "input", To: "analysis", Label: "initial_analysis"},
		{From: "analysis", To: "react", Label: "needs_reasoning"},
		{From: "react", To: "decision", Label: "routing_decision"},
		{
			From:  "decision",
			To:    "task_math",
			Label: "mathematical_task",
			Condition: func(state *WorkflowState) bool {
				taskType, exists := state.Context["task_type"].(string)
				return exists && strings.Contains(strings.ToLower(taskType), "math")
			},
		},
		{
			From:  "decision",
			To:    "task_research",
			Label: "research_task",
			Condition: func(state *WorkflowState) bool {
				taskType, exists := state.Context["task_type"].(string)
				return exists && strings.Contains(strings.ToLower(taskType), "research")
			},
		},
		{
			From:  "decision",
			To:    "task_analysis",
			Label: "analysis_task",
			Condition: func(state *WorkflowState) bool {
				taskType, exists := state.Context["task_type"].(string)
				return exists && strings.Contains(strings.ToLower(taskType), "analysis")
			},
		},
		{From: "task_math", To: "aggregation", Label: "math_result"},
		{From: "task_research", To: "aggregation", Label: "research_result"},
		{From: "task_analysis", To: "aggregation", Label: "analysis_result"},
		{From: "aggregation", To: "output", Label: "final_output"},
	}

	graph.StartNode = "input"

	return graph, nil
}

// validateConnection checks Ollama connectivity
func (g *WorkflowGraph) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", g.ollamaEndpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// ExecuteWorkflow runs the complete workflow
func (g *WorkflowGraph) ExecuteWorkflow(input string) (*WorkflowState, error) {
	// Initialize state
	state := &WorkflowState{
		ID:          uuid.New().String(),
		Input:       input,
		CurrentNode: g.StartNode,
		Context:     make(map[string]interface{}),
		History:     make([]NodeExecution, 0),
		Metadata:    make(map[string]string),
	}

	state.Context["start_time"] = time.Now()
	state.Metadata["workflow_id"] = state.ID

	fmt.Printf("🚀 Starting workflow execution: %s\n", state.ID)
	fmt.Printf("📝 Input: %s\n", input)
	fmt.Println()

	// Execute workflow
	for state.CurrentNode != "" && !g.isEndNode(state.CurrentNode) {
		node, exists := g.Nodes[state.CurrentNode]
		if !exists {
			return state, fmt.Errorf("node not found: %s", state.CurrentNode)
		}

		fmt.Printf("⚡ Executing node: %s (%s)\n", state.CurrentNode, node.GetType())

		// Execute node
		start := time.Now()
		newState, err := node.Execute(state)
		duration := time.Since(start)

		// Record execution
		execution := NodeExecution{
			NodeID:    state.CurrentNode,
			NodeType:  node.GetType(),
			Input:     state.Input,
			Duration:  duration,
			Timestamp: time.Now(),
			Success:   err == nil,
			Metadata:  make(map[string]interface{}),
		}

		if err != nil {
			execution.Error = err.Error()
			state.Error = err.Error()
			fmt.Printf("❌ Node execution failed: %v\n", err)
		} else {
			state = newState
			execution.Output = state.Result
			fmt.Printf("✅ Node completed in %s\n", duration)
		}

		state.History = append(state.History, execution)

		if err != nil {
			break
		}

		// Determine next node
		nextNode := g.getNextNode(state)
		state.CurrentNode = nextNode

		if nextNode != "" {
			fmt.Printf("➡️  Next node: %s\n", nextNode)
		}
		fmt.Println()
	}

	// Final processing
	if g.isEndNode(state.CurrentNode) {
		node := g.Nodes[state.CurrentNode]
		finalState, err := node.Execute(state)
		if err == nil {
			state = finalState
		}
	}

	state.Context["end_time"] = time.Now()
	totalDuration := state.Context["end_time"].(time.Time).Sub(state.Context["start_time"].(time.Time))
	state.Context["total_duration"] = totalDuration

	fmt.Printf("🏁 Workflow completed in %s\n", totalDuration)
	return state, nil
}

// getNextNode determines the next node based on edges and conditions
func (g *WorkflowGraph) getNextNode(state *WorkflowState) string {
	for _, edge := range g.Edges {
		if edge.From == state.CurrentNode {
			if edge.Condition == nil || edge.Condition(state) {
				if edge.Transform != nil {
					state = edge.Transform(state)
				}
				return edge.To
			}
		}
	}
	return ""
}

// isEndNode checks if a node is an end node
func (g *WorkflowGraph) isEndNode(nodeID string) bool {
	for _, endNode := range g.EndNodes {
		if nodeID == endNode {
			return true
		}
	}
	return false
}

// startWorkflowSession runs the interactive session
func (g *WorkflowGraph) startWorkflowSession() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🔄 Workflow Graph Session Started")
	fmt.Println("Enter your task to see the workflow in action")
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
				fmt.Println("\n👋 Workflow session ended.")
				break
			}

			if g.processCommand(userInput) {
				continue
			}

			fmt.Printf("❓ Unknown command: %s\n", userInput)
			fmt.Println("Type /help to see available commands.")
			continue
		}

		// Execute workflow
		state, err := g.ExecuteWorkflow(userInput)
		if err != nil {
			fmt.Printf("❌ Workflow execution failed: %v\n", err)
		} else {
			fmt.Printf("📋 Final Result: %s\n", state.Result)
			fmt.Printf("📊 Nodes executed: %d\n", len(state.History))
		}
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ Error reading input: %v", err)
	}
}

// processCommand handles system commands
func (g *WorkflowGraph) processCommand(command string) bool {
	switch strings.ToLower(command) {
	case "/help":
		g.showHelp()
		return true
	case "/graph":
		g.showGraphStructure()
		return true
	case "/state":
		fmt.Println("State tracking is per-execution. Run a task to see state.")
		return true
	case "/history":
		fmt.Println("History tracking is per-execution. Run a task to see history.")
		return true
	case "/reset":
		fmt.Println("✅ Ready for new workflow execution")
		return true
	default:
		return false
	}
}

// Node implementations

// InputNode implementation
func (n *InputNode) GetID() string                              { return n.ID }
func (n *InputNode) GetType() string                            { return "input" }
func (n *InputNode) GetDescription() string                     { return "Workflow entry point" }
func (n *InputNode) GetNextNodes(state *WorkflowState) []string { return []string{"analysis"} }

func (n *InputNode) Execute(state *WorkflowState) (*WorkflowState, error) {
	state.Context["original_input"] = state.Input
	state.Context["input_length"] = len(state.Input)
	state.Context["input_timestamp"] = time.Now()
	return state, nil
}

// AnalysisNode implementation
func (n *AnalysisNode) GetID() string                              { return n.ID }
func (n *AnalysisNode) GetType() string                            { return "analysis" }
func (n *AnalysisNode) GetDescription() string                     { return "Input analysis and classification" }
func (n *AnalysisNode) GetNextNodes(state *WorkflowState) []string { return []string{"react"} }

func (n *AnalysisNode) Execute(state *WorkflowState) (*WorkflowState, error) {
	prompt := fmt.Sprintf(`Analyze the following task and classify it. Determine:
1. Task type (mathematical, research, analysis, general)
2. Complexity level (low, medium, high)
3. Required tools or capabilities
4. Estimated steps needed

Task: %s

Provide a brief analysis in this format:
Task Type: [type]
Complexity: [level]
Tools Needed: [list]
Steps: [number]
Summary: [brief description]`, state.Input)

	response, err := n.callOllama(prompt)
	if err != nil {
		return state, err
	}

	// Parse analysis response
	taskType := extractField(response, "Task Type")
	complexity := extractField(response, "Complexity")

	state.Context["analysis_result"] = response
	state.Context["task_type"] = taskType
	state.Context["complexity"] = complexity
	state.Context["analysis_timestamp"] = time.Now()

	return state, nil
}

// ReActNode implementation
func (n *ReActNode) GetID() string                              { return n.ID }
func (n *ReActNode) GetType() string                            { return "react" }
func (n *ReActNode) GetDescription() string                     { return "ReAct reasoning and planning" }
func (n *ReActNode) GetNextNodes(state *WorkflowState) []string { return []string{"decision"} }

func (n *ReActNode) Execute(state *WorkflowState) (*WorkflowState, error) {
	return n.Agent.Execute(state)
}

// DecisionNode implementation
func (n *DecisionNode) GetID() string          { return n.ID }
func (n *DecisionNode) GetType() string        { return "decision" }
func (n *DecisionNode) GetDescription() string { return "Workflow routing decision" }
func (n *DecisionNode) GetNextNodes(state *WorkflowState) []string {
	return []string{"task_math", "task_research", "task_analysis"}
}

func (n *DecisionNode) Execute(state *WorkflowState) (*WorkflowState, error) {
	// Decision logic based on analysis
	taskType, exists := state.Context["task_type"].(string)
	if !exists {
		taskType = "general"
	}

	state.Context["routing_decision"] = taskType
	state.Context["decision_timestamp"] = time.Now()

	return state, nil
}

// TaskExecutionNode implementation
func (n *TaskExecutionNode) GetID() string   { return n.ID }
func (n *TaskExecutionNode) GetType() string { return "task_execution" }
func (n *TaskExecutionNode) GetDescription() string {
	return fmt.Sprintf("Execute %s task", n.TaskType)
}
func (n *TaskExecutionNode) GetNextNodes(state *WorkflowState) []string {
	return []string{"aggregation"}
}

func (n *TaskExecutionNode) Execute(state *WorkflowState) (*WorkflowState, error) {
	var prompt string

	switch n.TaskType {
	case "mathematical":
		prompt = fmt.Sprintf("Solve this mathematical problem step by step: %s", state.Input)
	case "research":
		prompt = fmt.Sprintf("Research and provide information about: %s", state.Input)
	case "analysis":
		prompt = fmt.Sprintf("Analyze and provide insights on: %s", state.Input)
	default:
		prompt = fmt.Sprintf("Process this request: %s", state.Input)
	}

	response, err := n.callOllama(prompt)
	if err != nil {
		return state, err
	}

	state.Context[fmt.Sprintf("%s_result", n.TaskType)] = response
	state.Context[fmt.Sprintf("%s_timestamp", n.TaskType)] = time.Now()

	return state, nil
}

// AggregationNode implementation
func (n *AggregationNode) GetID() string                              { return n.ID }
func (n *AggregationNode) GetType() string                            { return "aggregation" }
func (n *AggregationNode) GetDescription() string                     { return "Aggregate and synthesize results" }
func (n *AggregationNode) GetNextNodes(state *WorkflowState) []string { return []string{"output"} }

func (n *AggregationNode) Execute(state *WorkflowState) (*WorkflowState, error) {
	// Collect all results
	var results []string
	for key, value := range state.Context {
		if strings.HasSuffix(key, "_result") {
			if strValue, ok := value.(string); ok {
				results = append(results, strValue)
			}
		}
	}

	if len(results) == 0 {
		return state, fmt.Errorf("no results to aggregate")
	}

	prompt := fmt.Sprintf(`Synthesize and summarize the following results into a coherent response:

Original Task: %s

Results:
%s

Provide a comprehensive, well-structured final answer.`,
		state.Input, strings.Join(results, "\n\n"))

	response, err := n.callOllama(prompt)
	if err != nil {
		return state, err
	}

	state.Context["aggregated_result"] = response
	state.Context["aggregation_timestamp"] = time.Now()

	return state, nil
}

// OutputNode implementation
func (n *OutputNode) GetID() string                              { return n.ID }
func (n *OutputNode) GetType() string                            { return "output" }
func (n *OutputNode) GetDescription() string                     { return "Format final output" }
func (n *OutputNode) GetNextNodes(state *WorkflowState) []string { return []string{} }

func (n *OutputNode) Execute(state *WorkflowState) (*WorkflowState, error) {
	// Get the aggregated result
	if result, exists := state.Context["aggregated_result"].(string); exists {
		state.Result = result
	} else {
		state.Result = "No final result available"
	}

	state.Context["output_timestamp"] = time.Now()
	return state, nil
}

// ReAct Agent implementation
func (agent *ReActAgent) Execute(state *WorkflowState) (*WorkflowState, error) {
	prompt := fmt.Sprintf(`You are a ReAct (Reasoning and Acting) agent. Given the task and analysis, create a step-by-step plan.

Task: %s
Analysis: %s

Available tools: %s

Think through this step by step:
1. Understand the task
2. Plan the approach
3. Identify needed tools
4. Create action sequence

Provide your reasoning and action plan.`,
		state.Input,
		state.Context["analysis_result"],
		agent.getToolsList())

	response, err := agent.callOllama(prompt)
	if err != nil {
		return state, err
	}

	state.Context["react_reasoning"] = response
	state.Context["react_timestamp"] = time.Now()

	return state, nil
}

func (agent *ReActAgent) getToolsList() string {
	var tools []string
	for name, tool := range agent.Tools {
		tools = append(tools, fmt.Sprintf("%s: %s", name, tool.GetDescription()))
	}
	return strings.Join(tools, ", ")
}

func (agent *ReActAgent) callOllama(prompt string) (string, error) {
	return callOllamaAPI(agent.endpoint, agent.model, prompt)
}

// Tool implementations

// CalculatorTool
func (t *CalculatorTool) GetName() string        { return "calculator" }
func (t *CalculatorTool) GetDescription() string { return "Perform mathematical calculations" }
func (t *CalculatorTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "Mathematical expression to evaluate",
			},
		},
	}
}

func (t *CalculatorTool) Execute(args map[string]interface{}) (string, error) {
	expression, ok := args["expression"].(string)
	if !ok {
		return "", fmt.Errorf("expression parameter required")
	}

	// Simple calculation (in a real implementation, use a proper math parser)
	return fmt.Sprintf("Calculation result for '%s': [calculated value]", expression), nil
}

// WebSearchTool
func (t *WebSearchTool) GetName() string        { return "web_search" }
func (t *WebSearchTool) GetDescription() string { return "Search the web for information" }
func (t *WebSearchTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
		},
	}
}

func (t *WebSearchTool) Execute(args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter required")
	}

	return fmt.Sprintf("Web search results for '%s': [search results]", query), nil
}

// DataAnalysisTool
func (t *DataAnalysisTool) GetName() string        { return "data_analysis" }
func (t *DataAnalysisTool) GetDescription() string { return "Analyze data and provide insights" }
func (t *DataAnalysisTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"data": map[string]interface{}{
				"type":        "string",
				"description": "Data to analyze",
			},
		},
	}
}

func (t *DataAnalysisTool) Execute(args map[string]interface{}) (string, error) {
	data, ok := args["data"].(string)
	if !ok {
		return "", fmt.Errorf("data parameter required")
	}

	return fmt.Sprintf("Data analysis for '%s': [analysis results]", data), nil
}

// PlannerTool
func (t *PlannerTool) GetName() string        { return "planner" }
func (t *PlannerTool) GetDescription() string { return "Create task plans and strategies" }
func (t *PlannerTool) GetSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task": map[string]interface{}{
				"type":        "string",
				"description": "Task to plan",
			},
		},
	}
}

func (t *PlannerTool) Execute(args map[string]interface{}) (string, error) {
	task, ok := args["task"].(string)
	if !ok {
		return "", fmt.Errorf("task parameter required")
	}

	return fmt.Sprintf("Plan for '%s': [detailed plan]", task), nil
}

// Helper functions

func (n *AnalysisNode) callOllama(prompt string) (string, error) {
	return callOllamaAPI(n.endpoint, n.model, prompt)
}

func (n *TaskExecutionNode) callOllama(prompt string) (string, error) {
	return callOllamaAPI(n.endpoint, n.model, prompt)
}

func (n *AggregationNode) callOllama(prompt string) (string, error) {
	return callOllamaAPI(n.endpoint, n.model, prompt)
}

func callOllamaAPI(endpoint, model, prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint+"/api/generate", bytes.NewBuffer(jsonData))
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

func extractField(text, field string) string {
	pattern := fmt.Sprintf(`%s:\s*(.+)`, regexp.QuoteMeta(field))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return "unknown"
}

func (g *WorkflowGraph) showHelp() {
	fmt.Println("\n📚 Help - Complex Workflow Graph")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("This system demonstrates advanced workflow orchestration with:")
	fmt.Println()
	fmt.Println("🔄 Graph Architecture:")
	fmt.Println("  • Nodes: Processing units with specific functions")
	fmt.Println("  • Edges: Conditional connections between nodes")
	fmt.Println("  • State: Data that flows through the workflow")
	fmt.Println()
	fmt.Println("🤖 ReAct Agent Integration:")
	fmt.Println("  • Reasoning: Analyzes tasks and creates plans")
	fmt.Println("  • Acting: Uses tools to accomplish goals")
	fmt.Println("  • Tools: Calculator, web search, data analysis, planner")
	fmt.Println()
	fmt.Println("📊 Workflow Features:")
	fmt.Println("  • Dynamic routing based on task analysis")
	fmt.Println("  • Parallel execution paths")
	fmt.Println("  • State management and history tracking")
	fmt.Println("  • Result aggregation and synthesis")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /graph    - Show detailed graph structure")
	fmt.Println("  /state    - Show current workflow state")
	fmt.Println("  /history  - Show execution history")
	fmt.Println("  /reset    - Reset workflow state")
	fmt.Println("  /help     - Show this help message")
	fmt.Println()
	fmt.Println("Example tasks to try:")
	fmt.Println("  • 'Calculate the compound interest on $1000 at 5% for 3 years'")
	fmt.Println("  • 'Research the latest developments in quantum computing'")
	fmt.Println("  • 'Analyze the pros and cons of remote work'")
	fmt.Println("  • 'Plan a machine learning project for customer segmentation'")
	fmt.Println()
}

func (g *WorkflowGraph) showGraphStructure() {
	fmt.Println("\n🔄 Workflow Graph Structure")
	fmt.Println("===========================")
	fmt.Println()

	fmt.Printf("📊 Graph Statistics:\n")
	fmt.Printf("   Nodes: %d\n", len(g.Nodes))
	fmt.Printf("   Edges: %d\n", len(g.Edges))
	fmt.Printf("   Start Node: %s\n", g.StartNode)
	fmt.Printf("   End Nodes: %s\n", strings.Join(g.EndNodes, ", "))
	fmt.Println()

	fmt.Println("🔗 Node Definitions:")
	nodeOrder := []string{"input", "analysis", "react", "decision", "task_math", "task_research", "task_analysis", "aggregation", "output"}
	for _, nodeID := range nodeOrder {
		if node, exists := g.Nodes[nodeID]; exists {
			fmt.Printf("   %s (%s): %s\n", nodeID, node.GetType(), node.GetDescription())
		}
	}
	fmt.Println()

	fmt.Println("➡️  Edge Connections:")
	for _, edge := range g.Edges {
		condition := ""
		if edge.Condition != nil {
			condition = " [conditional]"
		}
		fmt.Printf("   %s → %s (%s)%s\n", edge.From, edge.To, edge.Label, condition)
	}
	fmt.Println()

	fmt.Println("🛠️  ReAct Agent Tools:")
	if reactNode, exists := g.Nodes["react"].(*ReActNode); exists {
		for name, tool := range reactNode.Agent.Tools {
			fmt.Printf("   %s: %s\n", name, tool.GetDescription())
		}
	}
	fmt.Println()

	fmt.Println("🔄 Workflow Flow:")
	fmt.Println("   1. Input → Analysis (classify task)")
	fmt.Println("   2. Analysis → ReAct (reasoning & planning)")
	fmt.Println("   3. ReAct → Decision (route to appropriate task)")
	fmt.Println("   4. Decision → Task Execution (math/research/analysis)")
	fmt.Println("   5. Task → Aggregation (combine results)")
	fmt.Println("   6. Aggregation → Output (format final response)")
	fmt.Println()
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
