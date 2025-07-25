// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// AgentType represents the type of agent
type AgentType string

const (
	AgentTypeReAct AgentType = "react"
	AgentTypeChat  AgentType = "chat"
	AgentTypeTool  AgentType = "tool"
)

// AgentConfig represents agent configuration
type AgentConfig struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            AgentType              `json:"type"`
	Model           string                 `json:"model"`
	Provider        string                 `json:"provider"`
	SystemPrompt    string                 `json:"system_prompt"`
	Temperature     float64                `json:"temperature"`
	MaxTokens       int                    `json:"max_tokens"`
	MaxIterations   int                    `json:"max_iterations"`
	Tools           []string               `json:"tools"`
	EnableStreaming bool                   `json:"enable_streaming"`
	StreamingMode   llm.StreamMode         `json:"streaming_mode,omitempty"`
	Timeout         time.Duration          `json:"timeout"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// DefaultAgentConfig returns default agent configuration
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		ID:              uuid.New().String(),
		Type:            AgentTypeChat,
		Temperature:     0.7,
		MaxTokens:       1000,
		MaxIterations:   10,
		Tools:           []string{},
		EnableStreaming: false,
		StreamingMode:   llm.StreamModeAuto,
		Timeout:         30 * time.Second,
		Metadata:        make(map[string]interface{}),
	}
}

// Validate validates the agent configuration
func (config *AgentConfig) Validate() error {
	if config.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if config.Type == "" {
		return fmt.Errorf("agent type is required")
	}

	if config.Model == "" {
		return fmt.Errorf("agent model is required")
	}

	if config.Provider == "" {
		return fmt.Errorf("agent provider is required")
	}

	// Validate MaxTokens - must be reasonable to prevent truncation
	if config.MaxTokens <= 0 {
		return fmt.Errorf("MaxTokens must be greater than 0, got %d", config.MaxTokens)
	}

	// Prevent dangerously low MaxTokens that could cause truncation
	if config.MaxTokens <= 100 {
		return fmt.Errorf("MaxTokens too low (%d), minimum required is 100 to prevent response truncation", config.MaxTokens)
	}

	if config.MaxTokens > 100000 {
		return fmt.Errorf("MaxTokens too large (%d), maximum allowed is 100000", config.MaxTokens)
	}

	// Validate Temperature range
	if config.Temperature < 0 || config.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0 and 2.0, got %f", config.Temperature)
	}

	// Validate MaxIterations
	if config.MaxIterations <= 0 {
		config.MaxIterations = 10 // Set default
	}

	if config.MaxIterations > 100 {
		return fmt.Errorf("MaxIterations too large (%d), maximum allowed is 100", config.MaxIterations)
	}

	return nil
}

// ValidateAndSanitize validates the agent configuration and sanitizes problematic values
func (config *AgentConfig) ValidateAndSanitize() error {
	// First do basic validation for critical fields
	if config.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	if config.Model == "" {
		return fmt.Errorf("agent model is required")
	}

	if config.Provider == "" {
		return fmt.Errorf("agent provider is required")
	}

	// Sanitize MaxTokens - automatically fix low values instead of erroring
	if config.MaxTokens <= 0 {
		config.MaxTokens = 500 // Set to safe default
	} else if config.MaxTokens <= 100 {
		config.MaxTokens = 500 // Sanitize to prevent truncation - any value 100 or below is risky
	}

	if config.MaxTokens > 100000 {
		return fmt.Errorf("MaxTokens too large (%d), maximum allowed is 100000", config.MaxTokens)
	}

	// Sanitize Temperature range
	if config.Temperature < 0 {
		config.Temperature = 0.7 // Set to default
	} else if config.Temperature > 2.0 {
		config.Temperature = 0.7 // Set to default
	}

	// Validate MaxIterations
	if config.MaxIterations <= 0 {
		config.MaxIterations = 10 // Set default
	}

	if config.MaxIterations > 100 {
		return fmt.Errorf("MaxIterations too large (%d), maximum allowed is 100", config.MaxIterations)
	}

	// Set default Type if not specified
	if config.Type == "" {
		config.Type = AgentTypeChat
	}

	// Set default timeout if not specified
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	// Initialize collections if nil
	if config.Tools == nil {
		config.Tools = make([]string, 0)
	}
	if config.Metadata == nil {
		config.Metadata = make(map[string]interface{})
	}

	return nil
}

// Agent represents an AI agent
type Agent struct {
	config       *AgentConfig
	llmManager   *llm.ProviderManager
	toolRegistry *tools.ToolRegistry
	graph        *core.Graph
	conversation *llm.ConversationHistory
	logger       *logrus.Logger
	mu           sync.RWMutex

	// Execution state
	isRunning        bool
	currentIteration int
	executionHistory []AgentExecution
}

// AgentExecution represents an agent execution record
type AgentExecution struct {
	ID               string                 `json:"id"`
	Timestamp        time.Time              `json:"timestamp"`
	Input            string                 `json:"input"`
	Output           string                 `json:"output"`            // Legacy string output for backward compatibility
	StructuredOutput interface{}            `json:"structured_output"` // New structured JSON output
	ToolCalls        []llm.ToolCall         `json:"tool_calls"`
	Duration         time.Duration          `json:"duration"`
	Success          bool                   `json:"success"`
	Error            error                  `json:"error,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
	ExecutionPath    []string               `json:"execution_path"`          // Track which nodes were executed
	StateChanges     []StateChange          `json:"state_changes,omitempty"` // Track state progression
}

// StateChange represents a change in agent state during execution
type StateChange struct {
	NodeID    string                 `json:"node_id"`
	NodeName  string                 `json:"node_name"`
	Timestamp time.Time              `json:"timestamp"`
	Before    map[string]interface{} `json:"before,omitempty"`
	After     map[string]interface{} `json:"after,omitempty"`
	Duration  time.Duration          `json:"duration"`
}

// NewAgent creates a new agent
func NewAgent(config *AgentConfig, llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) *Agent {
	// Create a copy of config to avoid modification of original
	agentConfig := *config

	// Validate and sanitize configuration
	if err := agentConfig.ValidateAndSanitize(); err != nil {
		// Log the error and apply default configuration
		logger := logrus.New()
		logger.WithError(err).Error("Invalid agent configuration, applying defaults")

		// Apply emergency defaults for critical missing fields only
		// MaxTokens is now handled by ValidateAndSanitize
		if agentConfig.Temperature < 0 || agentConfig.Temperature > 2.0 {
			agentConfig.Temperature = 0.7
		}
		if agentConfig.Provider == "" {
			agentConfig.Provider = "ollama"
		}
		if agentConfig.Model == "" {
			agentConfig.Model = "llama2"
		}
	}

	logger := logrus.New()
	logger.WithFields(logrus.Fields{
		"agent_name":  agentConfig.Name,
		"agent_type":  agentConfig.Type,
		"model":       agentConfig.Model,
		"provider":    agentConfig.Provider,
		"max_tokens":  agentConfig.MaxTokens,
		"temperature": agentConfig.Temperature,
	}).Info("Creating agent with validated configuration")

	if agentConfig.ID == "" {
		agentConfig.ID = uuid.New().String()
	}

	agent := &Agent{
		config:           &agentConfig,
		llmManager:       llmManager,
		toolRegistry:     toolRegistry,
		conversation:     llm.NewConversationHistory(),
		logger:           logger,
		executionHistory: make([]AgentExecution, 0),
	}

	// Build the agent's execution graph based on type
	agent.buildGraph()

	return agent
}

// buildGraph builds the execution graph for the agent based on its type
func (a *Agent) buildGraph() {
	a.graph = core.NewGraph(fmt.Sprintf("%s-graph", a.config.Name))

	switch a.config.Type {
	case AgentTypeReAct:
		a.buildReActGraph()
	case AgentTypeChat:
		a.buildChatGraph()
	case AgentTypeTool:
		a.buildToolGraph()
	default:
		a.buildChatGraph() // Default to chat
	}
}

// buildReActGraph builds a ReAct (Reasoning and Acting) graph
func (a *Agent) buildReActGraph() {
	// Define nodes
	reasonNode := a.graph.AddNode("reason", "Reason", a.reasonNode)
	actNode := a.graph.AddNode("act", "Act", a.actNode)
	observeNode := a.graph.AddNode("observe", "Observe", a.observeNode)
	finalizeNode := a.graph.AddNode("finalize", "Finalize", a.finalizeNode)

	// Set metadata
	reasonNode.Metadata["type"] = "reasoning"
	actNode.Metadata["type"] = "action"
	observeNode.Metadata["type"] = "observation"
	finalizeNode.Metadata["type"] = "finalization"

	// Define edges with conditions
	a.graph.AddEdge("reason", "act", a.shouldAct)
	a.graph.AddEdge("reason", "finalize", a.shouldFinalize)
	a.graph.AddEdge("act", "observe", nil) // Always observe after acting
	a.graph.AddEdge("observe", "reason", a.shouldContinueReasoning)
	a.graph.AddEdge("observe", "finalize", a.shouldFinalize)

	// Set start and end nodes
	a.graph.SetStartNode("reason")
	a.graph.AddEndNode("finalize")
}

// buildChatGraph builds a simple chat graph
func (a *Agent) buildChatGraph() {
	// Define nodes
	chatNode := a.graph.AddNode("chat", "Chat", a.chatNode)

	// Set metadata
	chatNode.Metadata["type"] = "chat"

	// Set start and end nodes
	a.graph.SetStartNode("chat")
	a.graph.AddEndNode("chat")
}

// buildToolGraph builds a tool-focused graph
func (a *Agent) buildToolGraph() {
	// Define nodes
	planNode := a.graph.AddNode("plan", "Plan", a.planNode)
	executeNode := a.graph.AddNode("execute", "Execute", a.executeToolsNode)
	reviewNode := a.graph.AddNode("review", "Review", a.reviewNode)

	// Set metadata
	planNode.Metadata["type"] = "planning"
	executeNode.Metadata["type"] = "execution"
	reviewNode.Metadata["type"] = "review"

	// Define edges
	a.graph.AddEdge("plan", "execute", nil)
	a.graph.AddEdge("execute", "review", nil)
	a.graph.AddEdge("review", "plan", a.shouldReplan)

	// Set start and end nodes
	a.graph.SetStartNode("plan")
	a.graph.AddEndNode("review")
}

// Execute executes the agent with the given input
func (a *Agent) Execute(ctx context.Context, input string) (*AgentExecution, error) {
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return nil, fmt.Errorf("agent is already running")
	}
	a.isRunning = true
	a.currentIteration = 0
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.isRunning = false
		a.mu.Unlock()
	}()

	start := time.Now()
	execution := AgentExecution{
		ID:        uuid.New().String(),
		Timestamp: start,
		Input:     input,
		Metadata:  make(map[string]interface{}),
	}

	// Add user message to conversation
	a.conversation.AddMessage(llm.Message{
		Role:    "user",
		Content: input,
	})

	// Prepare initial state
	state := core.NewBaseState()
	state.Set("input", input)
	state.Set("conversation", a.conversation.GetMessages())
	state.Set("iteration", 0)
	state.Set("max_iterations", a.config.MaxIterations)

	// Execute the graph
	finalState, err := a.graph.Execute(ctx, state)
	if err != nil {
		execution.Error = err
		execution.Success = false
	} else {
		execution.Success = true
		if output, exists := finalState.Get("output"); exists {
			// Always store structured output and provide string fallback
			execution.StructuredOutput = output

			switch v := output.(type) {
			case string:
				execution.Output = v
			case map[string]interface{}:
				// Store structured data for proper JSON serialization
				execution.Metadata["structured_output"] = v
				// Extract a meaningful string representation for legacy compatibility
				if response, ok := v["response"].(string); ok {
					execution.Output = response
				} else if description, ok := v["description"].(string); ok {
					execution.Output = description
				} else if story, ok := v["story"].(string); ok {
					execution.Output = story
				} else if summary, ok := v["summary"].(string); ok {
					execution.Output = summary
				} else {
					execution.Output = fmt.Sprintf("%v", v)
				}
			default:
				execution.Output = fmt.Sprintf("%v", v)
			}
		}
		if toolCalls, exists := finalState.Get("tool_calls"); exists {
			if tc, ok := toolCalls.([]llm.ToolCall); ok {
				execution.ToolCalls = tc
			}
		}

		// Track execution path from graph
		if a.graph != nil {
			// This would be populated by the graph execution tracking
			execution.ExecutionPath = []string{} // Placeholder for now
		}
	}

	execution.Duration = time.Since(start)

	// Add execution to history
	a.mu.Lock()
	a.executionHistory = append(a.executionHistory, execution)
	a.mu.Unlock()

	return &execution, err
}

// reasonNode implements the reasoning step in ReAct
func (a *Agent) reasonNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	messages := a.buildReasoningMessages(state)

	req := llm.CompletionRequest{
		Messages:    messages,
		Model:       a.config.Model,
		Temperature: a.config.Temperature,
		MaxTokens:   a.config.MaxTokens,
	}

	resp, err := a.llmManager.Complete(ctx, a.config.Provider, req)
	if err != nil {
		return nil, fmt.Errorf("reasoning failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	reasoning := resp.Choices[0].Message.Content
	state.Set("reasoning", reasoning)

	// Add assistant message to conversation
	a.conversation.AddMessage(resp.Choices[0].Message)

	a.logger.WithField("reasoning", reasoning).Info("Agent reasoning completed")
	return state, nil
}

// actNode implements the action step in ReAct
func (a *Agent) actNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	reasoning, _ := state.Get("reasoning")

	// Parse the reasoning to determine if tool calls are needed
	toolCalls := a.parseToolCalls(fmt.Sprintf("%v", reasoning))

	if len(toolCalls) == 0 {
		// No tools needed, just return the reasoning as action
		state.Set("action", reasoning)
		return state, nil
	}

	// Execute tool calls
	var results []string
	var executedCalls []llm.ToolCall

	for _, toolCall := range toolCalls {
		tool, exists := a.toolRegistry.GetTool(toolCall.Function.Name)
		if !exists {
			results = append(results, fmt.Sprintf("Tool %s not found", toolCall.Function.Name))
			continue
		}

		result, err := tool.Execute(ctx, toolCall.Function.Arguments)
		if err != nil {
			results = append(results, fmt.Sprintf("Tool %s failed: %v", toolCall.Function.Name, err))
		} else {
			results = append(results, result)
		}

		executedCalls = append(executedCalls, toolCall)
	}

	state.Set("action", strings.Join(results, "\n"))
	state.Set("tool_calls", executedCalls)

	a.logger.WithField("tool_calls", len(executedCalls)).Info("Agent action completed")
	return state, nil
}

// observeNode implements the observation step in ReAct
func (a *Agent) observeNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	action, _ := state.Get("action")

	// Create observation based on action results
	observation := fmt.Sprintf("Observation: %v", action)
	state.Set("observation", observation)

	// Add observation to conversation
	a.conversation.AddMessage(llm.Message{
		Role:    "assistant",
		Content: observation,
	})

	// Increment iteration
	iteration, _ := state.Get("iteration")
	if iter, ok := iteration.(int); ok {
		state.Set("iteration", iter+1)
	}

	a.logger.WithField("observation", observation).Info("Agent observation completed")
	return state, nil
}

// finalizeNode implements the finalization step
func (a *Agent) finalizeNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Generate final response
	messages := a.buildFinalizationMessages(state)

	req := llm.CompletionRequest{
		Messages:    messages,
		Model:       a.config.Model,
		Temperature: a.config.Temperature,
		MaxTokens:   a.config.MaxTokens,
	}

	resp, err := a.llmManager.Complete(ctx, a.config.Provider, req)
	if err != nil {
		return nil, fmt.Errorf("finalization failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	output := resp.Choices[0].Message.Content
	state.Set("output", output)

	// Add final message to conversation
	a.conversation.AddMessage(resp.Choices[0].Message)

	a.logger.WithField("output", output).Info("Agent finalization completed")
	return state, nil
}

// chatNode implements simple chat functionality
func (a *Agent) chatNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	messages := a.conversation.GetMessages()

	// Add system prompt if configured
	if a.config.SystemPrompt != "" {
		systemMsg := llm.Message{
			Role:    "system",
			Content: a.config.SystemPrompt,
		}
		messages = append([]llm.Message{systemMsg}, messages...)
	}

	// Add tools if available
	var toolDefs []llm.ToolDefinition
	for _, toolName := range a.config.Tools {
		if tool, exists := a.toolRegistry.GetTool(toolName); exists {
			toolDefs = append(toolDefs, tool.GetDefinition())
		}
	}

	req := llm.CompletionRequest{
		Messages:    messages,
		Model:       a.config.Model,
		Temperature: a.config.Temperature,
		MaxTokens:   a.config.MaxTokens,
		Tools:       toolDefs,
		Stream:      a.config.EnableStreaming,
	}

	var resp *llm.CompletionResponse
	var err error

	// Use streaming mode if enabled
	if a.config.EnableStreaming {
		resp, err = a.llmManager.CompleteWithMode(ctx, a.config.Provider, req, a.config.StreamingMode)
	} else {
		resp, err = a.llmManager.Complete(ctx, a.config.Provider, req)
	}

	if err != nil {
		return nil, fmt.Errorf("chat failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	message := resp.Choices[0].Message

	// Handle tool calls if present
	if len(message.ToolCalls) > 0 {
		var toolResults []string
		for _, toolCall := range message.ToolCalls {
			if tool, exists := a.toolRegistry.GetTool(toolCall.Function.Name); exists {
				result, err := tool.Execute(ctx, toolCall.Function.Arguments)
				if err != nil {
					toolResults = append(toolResults, fmt.Sprintf("Error: %v", err))
				} else {
					toolResults = append(toolResults, result)
				}
			}
		}

		// Add tool results to conversation
		for i, result := range toolResults {
			a.conversation.AddMessage(llm.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: message.ToolCalls[i].ID,
			})
		}

		state.Set("tool_calls", message.ToolCalls)
	}

	output := message.Content
	state.Set("output", output)

	// Add assistant message to conversation
	a.conversation.AddMessage(message)

	a.logger.WithField("output", output).Info("Agent chat completed")
	return state, nil
}

// planNode implements planning for tool agents
func (a *Agent) planNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	input, _ := state.Get("input")

	planPrompt := fmt.Sprintf(`Plan how to accomplish the following task using available tools:
Task: %v

Available tools: %s

Create a step-by-step plan.`, input, strings.Join(a.config.Tools, ", "))

	messages := []llm.Message{
		{Role: "system", Content: "You are a planning agent. Create detailed plans to accomplish tasks using available tools."},
		{Role: "user", Content: planPrompt},
	}

	req := llm.CompletionRequest{
		Messages:    messages,
		Model:       a.config.Model,
		Temperature: a.config.Temperature,
		MaxTokens:   a.config.MaxTokens,
	}

	resp, err := a.llmManager.Complete(ctx, a.config.Provider, req)
	if err != nil {
		return nil, fmt.Errorf("planning failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	plan := resp.Choices[0].Message.Content
	state.Set("plan", plan)

	a.logger.WithField("plan", plan).Info("Agent planning completed")
	return state, nil
}

// executeToolsNode executes tools based on the plan
func (a *Agent) executeToolsNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	plan, _ := state.Get("plan")

	// Parse the plan to extract tool calls
	toolCalls := a.parseToolCalls(fmt.Sprintf("%v", plan))

	var results []string
	var executedCalls []llm.ToolCall

	for _, toolCall := range toolCalls {
		tool, exists := a.toolRegistry.GetTool(toolCall.Function.Name)
		if !exists {
			results = append(results, fmt.Sprintf("Tool %s not found", toolCall.Function.Name))
			continue
		}

		result, err := tool.Execute(ctx, toolCall.Function.Arguments)
		if err != nil {
			results = append(results, fmt.Sprintf("Tool %s failed: %v", toolCall.Function.Name, err))
		} else {
			results = append(results, result)
		}

		executedCalls = append(executedCalls, toolCall)
	}

	state.Set("execution_results", results)
	state.Set("tool_calls", executedCalls)

	a.logger.WithField("tool_calls", len(executedCalls)).Info("Agent tool execution completed")
	return state, nil
}

// reviewNode reviews the execution results
func (a *Agent) reviewNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	results, _ := state.Get("execution_results")
	input, _ := state.Get("input")

	reviewPrompt := fmt.Sprintf(`Review the execution results for the task:
Task: %v
Results: %v

Determine if the task is complete or if more actions are needed.`, input, results)

	messages := []llm.Message{
		{Role: "system", Content: "You are a review agent. Assess if tasks have been completed successfully."},
		{Role: "user", Content: reviewPrompt},
	}

	req := llm.CompletionRequest{
		Messages:    messages,
		Model:       a.config.Model,
		Temperature: a.config.Temperature,
		MaxTokens:   a.config.MaxTokens,
	}

	resp, err := a.llmManager.Complete(ctx, a.config.Provider, req)
	if err != nil {
		return nil, fmt.Errorf("review failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	review := resp.Choices[0].Message.Content
	state.Set("review", review)
	state.Set("output", review)

	a.logger.WithField("review", review).Info("Agent review completed")
	return state, nil
}

// Edge condition functions

func (a *Agent) shouldAct(ctx context.Context, state *core.BaseState) (string, error) {
	reasoning, _ := state.Get("reasoning")
	reasoningStr := fmt.Sprintf("%v", reasoning)

	// Check if the reasoning indicates an action should be taken
	if strings.Contains(strings.ToLower(reasoningStr), "action:") ||
		strings.Contains(strings.ToLower(reasoningStr), "tool:") ||
		strings.Contains(strings.ToLower(reasoningStr), "execute") {
		return "act", nil
	}

	return "", nil
}

func (a *Agent) shouldFinalize(ctx context.Context, state *core.BaseState) (string, error) {
	iteration, _ := state.Get("iteration")
	maxIterations, _ := state.Get("max_iterations")

	if iter, ok := iteration.(int); ok {
		if maxIter, ok := maxIterations.(int); ok {
			if iter >= maxIter {
				return "finalize", nil
			}
		}
	}

	reasoning, _ := state.Get("reasoning")
	reasoningStr := fmt.Sprintf("%v", reasoning)

	// Check if the reasoning indicates completion
	if strings.Contains(strings.ToLower(reasoningStr), "final answer:") ||
		strings.Contains(strings.ToLower(reasoningStr), "conclusion:") ||
		strings.Contains(strings.ToLower(reasoningStr), "complete") {
		return "finalize", nil
	}

	return "", nil
}

func (a *Agent) shouldContinueReasoning(ctx context.Context, state *core.BaseState) (string, error) {
	iteration, _ := state.Get("iteration")
	maxIterations, _ := state.Get("max_iterations")

	if iter, ok := iteration.(int); ok {
		if maxIter, ok := maxIterations.(int); ok {
			if iter < maxIter {
				return "reason", nil
			}
		}
	}

	return "", nil
}

func (a *Agent) shouldReplan(ctx context.Context, state *core.BaseState) (string, error) {
	review, _ := state.Get("review")
	reviewStr := fmt.Sprintf("%v", review)

	// Check if the review indicates more planning is needed
	if strings.Contains(strings.ToLower(reviewStr), "incomplete") ||
		strings.Contains(strings.ToLower(reviewStr), "more actions needed") ||
		strings.Contains(strings.ToLower(reviewStr), "replan") {
		return "plan", nil
	}

	return "", nil
}

// Helper functions

func (a *Agent) buildReasoningMessages(state *core.BaseState) []llm.Message {
	messages := []llm.Message{}

	if a.config.SystemPrompt != "" {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: a.config.SystemPrompt,
		})
	} else {
		messages = append(messages, llm.Message{
			Role: "system",
			Content: `You are a ReAct agent. Think step by step about the problem and decide what action to take.

Format your response as:
Thought: [your reasoning]
Action: [action to take or tool to use]
Action Input: [input for the action]

Or if you have enough information:
Thought: [your reasoning]
Final Answer: [your final response]`,
		})
	}

	// Add conversation history
	messages = append(messages, a.conversation.GetMessages()...)

	return messages
}

func (a *Agent) buildFinalizationMessages(state *core.BaseState) []llm.Message {
	messages := []llm.Message{
		{
			Role:    "system",
			Content: "Provide a final, comprehensive answer based on the reasoning and observations.",
		},
	}

	// Add conversation history
	messages = append(messages, a.conversation.GetMessages()...)

	return messages
}

func (a *Agent) parseToolCalls(text string) []llm.ToolCall {
	var toolCalls []llm.ToolCall

	// Simple parsing - look for tool usage patterns
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for "Action: tool_name" or "Tool: tool_name"
		if strings.HasPrefix(strings.ToLower(line), "action:") ||
			strings.HasPrefix(strings.ToLower(line), "tool:") {

			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				toolName := strings.TrimSpace(parts[1])

				// Create a basic tool call
				toolCall := llm.ToolCall{
					ID:   uuid.New().String(),
					Type: "function",
					Function: llm.FunctionCall{
						Name:      toolName,
						Arguments: "{}",
					},
				}

				toolCalls = append(toolCalls, toolCall)
			}
		}
	}

	return toolCalls
}

// Public methods

// GetConfig returns the agent configuration
func (a *Agent) GetConfig() *AgentConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy
	config := *a.config
	return &config
}

// UpdateConfig updates the agent configuration
func (a *Agent) UpdateConfig(config *AgentConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config = config
	a.buildGraph() // Rebuild graph with new config
}

// GetConversation returns the conversation history
func (a *Agent) GetConversation() []llm.Message {
	return a.conversation.GetMessages()
}

// ClearConversation clears the conversation history
func (a *Agent) ClearConversation() {
	a.conversation.Clear()
}

// GetExecutionHistory returns the execution history
func (a *Agent) GetExecutionHistory() []AgentExecution {
	a.mu.RLock()
	defer a.mu.RUnlock()

	history := make([]AgentExecution, len(a.executionHistory))
	copy(history, a.executionHistory)
	return history
}

// IsRunning returns whether the agent is currently running
func (a *Agent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}

// GetGraph returns the agent's execution graph
func (a *Agent) GetGraph() *core.Graph {
	return a.graph
}

// SetGraph sets the agent's execution graph
func (a *Agent) SetGraph(graph *core.Graph) {
	a.graph = graph
}

// EnableStreaming enables streaming mode for the agent
func (a *Agent) EnableStreaming() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.EnableStreaming = true
	if a.config.StreamingMode == llm.StreamModeNone {
		a.config.StreamingMode = llm.StreamModeAuto
	}

	// Enable streaming on the provider as well
	return a.llmManager.EnableStreaming(a.config.Provider)
}

// DisableStreaming disables streaming mode for the agent
func (a *Agent) DisableStreaming() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.EnableStreaming = false
	a.config.StreamingMode = llm.StreamModeNone

	// Disable streaming on the provider as well
	return a.llmManager.DisableStreaming(a.config.Provider)
}

// SetStreamingMode sets the streaming mode for the agent
func (a *Agent) SetStreamingMode(mode llm.StreamMode) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.StreamingMode = mode
	a.config.EnableStreaming = mode != llm.StreamModeNone

	// Set streaming mode on the provider as well
	return a.llmManager.SetStreamingMode(a.config.Provider, mode)
}

// GetStreamingMode returns the current streaming mode
func (a *Agent) GetStreamingMode() llm.StreamMode {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.StreamingMode
}

// IsStreamingEnabled returns whether streaming is enabled
func (a *Agent) IsStreamingEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.EnableStreaming
}

// MultiAgentCoordinator manages multiple agents working together
type MultiAgentCoordinator struct {
	agents map[string]*Agent
	logger *logrus.Logger
	mu     sync.RWMutex
}

// NewMultiAgentCoordinator creates a new multi-agent coordinator
func NewMultiAgentCoordinator() *MultiAgentCoordinator {
	return &MultiAgentCoordinator{
		agents: make(map[string]*Agent),
		logger: logrus.New(),
	}
}

// AddAgent adds an agent to the coordinator
func (mac *MultiAgentCoordinator) AddAgent(id string, agent *Agent) {
	mac.mu.Lock()
	defer mac.mu.Unlock()

	mac.agents[id] = agent
}

// RemoveAgent removes an agent from the coordinator
func (mac *MultiAgentCoordinator) RemoveAgent(id string) {
	mac.mu.Lock()
	defer mac.mu.Unlock()

	delete(mac.agents, id)
}

// GetAgent returns an agent by ID
func (mac *MultiAgentCoordinator) GetAgent(id string) (*Agent, bool) {
	mac.mu.RLock()
	defer mac.mu.RUnlock()

	agent, exists := mac.agents[id]
	return agent, exists
}

// ListAgents returns all agent IDs
func (mac *MultiAgentCoordinator) ListAgents() []string {
	mac.mu.RLock()
	defer mac.mu.RUnlock()

	ids := make([]string, 0, len(mac.agents))
	for id := range mac.agents {
		ids = append(ids, id)
	}
	return ids
}

// ExecuteSequential executes agents sequentially, passing output to the next
func (mac *MultiAgentCoordinator) ExecuteSequential(ctx context.Context, agentIDs []string, initialInput string) ([]AgentExecution, error) {
	var executions []AgentExecution
	currentInput := initialInput

	for _, agentID := range agentIDs {
		agent, exists := mac.GetAgent(agentID)
		if !exists {
			return executions, fmt.Errorf("agent %s not found", agentID)
		}

		execution, err := agent.Execute(ctx, currentInput)
		if err != nil {
			return executions, fmt.Errorf("agent %s failed: %w", agentID, err)
		}

		executions = append(executions, *execution)
		currentInput = execution.Output // Use output as input for next agent
	}

	return executions, nil
}

// ExecuteParallel executes agents in parallel with the same input
func (mac *MultiAgentCoordinator) ExecuteParallel(ctx context.Context, agentIDs []string, input string) (map[string]AgentExecution, error) {
	results := make(map[string]AgentExecution)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(agentIDs))

	for _, agentID := range agentIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			agent, exists := mac.GetAgent(id)
			if !exists {
				errChan <- fmt.Errorf("agent %s not found", id)
				return
			}

			execution, err := agent.Execute(ctx, input)
			if err != nil {
				errChan <- fmt.Errorf("agent %s failed: %w", id, err)
				return
			}

			mu.Lock()
			results[id] = *execution
			mu.Unlock()
		}(agentID)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		return nil, err
	}

	return results, nil
}
