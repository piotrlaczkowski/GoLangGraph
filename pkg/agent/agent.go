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
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
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
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Type            AgentType      `json:"type"`
	Model           string         `json:"model"`
	Provider        string         `json:"provider"`
	SystemPrompt    string         `json:"system_prompt"`
	Temperature     float64        `json:"temperature"`
	MaxTokens       int            `json:"max_tokens"`
	MaxIterations   int            `json:"max_iterations"`
	Tools           []string       `json:"tools"`
	EnableStreaming bool           `json:"enable_streaming"`
	StreamingMode   llm.StreamMode `json:"streaming_mode,omitempty"`
	// EarlyExit cancels remaining stream tokens once a complete JSON/tool-call
	// is formed. Nil disables token-stream early-exit (multipass JSON exit still applies).
	EarlyExit    llm.EarlyExitFunc        `json:"-"`
	Timeout      time.Duration            `json:"timeout"`
	Metadata     map[string]interface{}   `json:"metadata"`
	Middleware   []Middleware             `json:"-"`
	InterruptOn  []string                 `json:"interrupt_on"`
	Checkpointer persistence.Checkpointer `json:"-"`
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

// BaseAgent represents the standard AI agent implementation
type BaseAgent struct {
	config            *AgentConfig
	llmManager        *llm.ProviderManager
	toolRegistry      *tools.ToolRegistry
	toolNode          *ToolNode         // Centralized tool execution with command support
	interruptManager  *InterruptManager // HITL interrupt lifecycle management
	graph             *core.Graph
	conversation      *llm.ConversationHistory
	logger            *logrus.Logger
	middleware        []Middleware
	checkpointManager *persistence.CheckpointManager
	mu                sync.RWMutex

	// Execution state
	isRunning        bool
	currentIteration int
	executionHistory []AgentExecution
	pendingToolCalls []llm.ToolCall // seeded on HITL resume (mid-tool-call)
}

// ... (AgentExecution struct remains same)

// NewAgent creates a new base agent
func NewAgent(config *AgentConfig, llmManager *llm.ProviderManager, toolRegistry *tools.ToolRegistry) *BaseAgent {
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

	agent := &BaseAgent{
		config:            &agentConfig,
		llmManager:        llmManager,
		toolRegistry:      toolRegistry,
		toolNode:          NewToolNode(toolRegistry),
		interruptManager:  NewInterruptManager(),
		conversation:      llm.NewConversationHistory(),
		logger:            logger,
		executionHistory:  make([]AgentExecution, 0),
		middleware:        agentConfig.Middleware,
		checkpointManager: persistence.NewCheckpointManager(agentConfig.Checkpointer),
	}

	// Build the agent's execution graph based on type
	agent.buildGraph()

	return agent
}

// GetConfig returns the agent's configuration
func (a *BaseAgent) GetConfig() *AgentConfig {
	return a.config
}

// GetGraph returns the agent's execution graph
func (a *BaseAgent) GetGraph() *core.Graph {
	return a.graph
}

// IsRunning returns true if the agent is currently executing
func (a *BaseAgent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isRunning
}

// Name returns the name of the agent
func (a *BaseAgent) Name() string {
	return a.config.Name
}

// buildGraph builds the execution graph for the agent based on its type
func (a *BaseAgent) buildGraph() {
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
func (a *BaseAgent) buildReActGraph() {
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

	// Define edges with conditions.
	// IMPORTANT: do not add a parallel always-true finalize edge. Edges live in a
	// map, so iteration order is random — an always-matching finalize raced with
	// act/reason and skipped tool execution (DeepSeek 400: unpaired tool_calls).
	a.graph.AddEdge("reason", "act", a.shouldAct) // returns "act" or "finalize"
	a.graph.AddEdge("reason", "finalize", a.shouldAct)
	a.graph.AddEdge("act", "observe", a.shouldObserve)              // Conditional edge to handle interrupts
	a.graph.AddEdge("observe", "reason", a.shouldContinueReasoning) // "reason" or "finalize"
	a.graph.AddEdge("observe", "finalize", a.shouldContinueReasoning)

	// Set start and end nodes
	a.graph.SetStartNode("reason")
	a.graph.AddEndNode("finalize")
}

// buildChatGraph builds a simple chat graph
func (a *BaseAgent) buildChatGraph() {
	// Define nodes
	chatNode := a.graph.AddNode("chat", "Chat", a.chatNode)

	// Set metadata
	chatNode.Metadata["type"] = "chat"

	// Set start and end nodes
	a.graph.SetStartNode("chat")
	a.graph.AddEndNode("chat")
}

// buildToolGraph builds a tool-focused graph
func (a *BaseAgent) buildToolGraph() {
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
func (a *BaseAgent) Execute(ctx context.Context, input string) (*AgentExecution, error) {
	return a.ExecuteThread(ctx, uuid.New().String(), input)
}

// ExecuteThread executes the agent with the given input in a specific thread
func (a *BaseAgent) ExecuteThread(ctx context.Context, threadID string, input string) (*AgentExecution, error) {
	a.mu.Lock()
	if a.isRunning {
		a.mu.Unlock()
		return nil, fmt.Errorf("agent is already running")
	}
	a.isRunning = true
	resumeIter := a.currentIteration
	if resumeIter < 0 {
		resumeIter = 0
	}
	// Fresh runs reset iteration; seeded resume keeps currentIteration.
	if a.conversation.Size() == 0 {
		a.currentIteration = 0
		resumeIter = 0
	}
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.isRunning = false
		a.mu.Unlock()
	}()

	// Create execution record
	execution := AgentExecution{
		ID:            uuid.New().String(),
		Input:         input,
		StartTime:     time.Now(),
		Status:        "running",
		Metadata:      make(map[string]interface{}),
		ExecutionPath: make([]string, 0),
		ToolCalls:     make([]llm.ToolCall, 0),
	}

	// Run BeforeRun middleware
	for _, m := range a.middleware {
		var err error
		input, err = m.BeforeRun(ctx, a, input)
		if err != nil {
			return nil, fmt.Errorf("middleware BeforeRun failed: %w", err)
		}
	}
	execution.Input = input // Update input in execution record if modified

	resuming := a.conversation.Size() > 0
	if resuming {
		execution.Metadata["resumed"] = true
		execution.Metadata["resume_iteration"] = resumeIter
	} else if strings.TrimSpace(input) != "" {
		// Add user message to conversation (cold start only).
		a.conversation.AddMessage(llm.Message{
			Role:    "user",
			Content: input,
		})
	}

	// Prepare initial state
	state := core.NewBaseState()
	state.Set("input", input)
	state.Set("conversation", a.conversation.GetMessages())
	state.Set("iteration", resumeIter)
	state.Set("max_iterations", a.config.MaxIterations)
	state.Set("thread_id", threadID)
	if pending := a.pendingToolCalls; len(pending) > 0 {
		state.Set("pending_tool_calls", pending)
		a.pendingToolCalls = nil
	}

	// Load from checkpoint if available
	if a.checkpointManager.IsEnabled() {
		// Try to load latest checkpoint for this thread
		// Note: This logic assumes we want to resume.
		// If input is provided, maybe we are starting a new turn?
		// For now, let's just load history if needed, but we are creating a new state for this run.
		// If we want to resume, we should probably merge state?
		// Let's assume 'state' is fresh for this turn.
	}

	// Execute the graph
	finalState, err := a.graph.Execute(ctx, state)
	if err != nil {
		execution.Error = err
		execution.Success = false
	} else {
		execution.Success = true

		// Check for interrupt
		if interrupt, exists := finalState.Get("__interrupt__"); exists {
			execution.Metadata["interrupt"] = interrupt
			execution.Success = false // It's not fully successful yet
			// We could return a specific error or status here
		}

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
			// This would be populated by the graph
			if executionPathVal, ok := finalState.Get("execution_path"); ok {
				if executionPath, ok := executionPathVal.([]string); ok {
					execution.ExecutionPath = append(execution.ExecutionPath, executionPath...)
				}
			}
		}
	}

	// Update execution record
	if finalState != nil {
		if v, ok := finalState.Get("usage"); ok {
			execution.Metadata["usage"] = v
		}
		if v, ok := finalState.Get("usage_estimated"); ok {
			execution.Metadata["usage_estimated"] = v
		}
	}
	execution.EndTime = time.Now()
	execution.Duration = time.Since(execution.StartTime)

	// Run AfterRun middleware
	for _, m := range a.middleware {
		res, afterErr := m.AfterRun(ctx, a, &execution)
		if afterErr != nil {
			return nil, fmt.Errorf("middleware AfterRun failed: %w", afterErr)
		}
		execution = *res
	}

	// Add execution to history
	a.mu.Lock()
	a.executionHistory = append(a.executionHistory, execution)
	a.mu.Unlock()

	// Save checkpoint
	if a.checkpointManager.IsEnabled() && execution.Success {
		// Save final state
		// We use "final" as nodeID for the end of execution
		_ = a.checkpointManager.SaveCheckpoint(ctx, threadID, "final", 0, finalState)
	}

	return &execution, err
}

// reasonNode implements the reasoning step in ReAct
func (a *BaseAgent) reasonNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	messages := a.buildReasoningMessages(state)

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
		EarlyExit:   a.config.EarlyExit,
	}

	var resp *llm.CompletionResponse
	var err error
	if a.config.EnableStreaming {
		mode := a.config.StreamingMode
		if mode == "" {
			mode = llm.StreamModeForced
		}
		resp, err = a.llmManager.CompleteWithMode(ctx, a.config.Provider, req, mode)
	} else {
		resp, err = a.llmManager.Complete(ctx, a.config.Provider, req)
	}
	if err != nil {
		return nil, fmt.Errorf("reasoning failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	message := resp.Choices[0].Message
	reasoning := message.Content
	state.Set("reasoning", reasoning)

	// Capture native tool calls
	if len(message.ToolCalls) > 0 {
		state.Set("pending_tool_calls", message.ToolCalls)
	} else {
		state.Set("pending_tool_calls", []llm.ToolCall{})
	}

	// Add assistant message to conversation
	a.conversation.AddMessage(message)

	a.accumulateUsage(state, resp)
	a.logger.WithField("reasoning", reasoning).
		WithField("tool_calls", len(message.ToolCalls)).
		WithField("finish_reason", resp.Choices[0].FinishReason).
		Info("Agent reasoning completed")
	return state, nil
}

// actNode implements the action step in ReAct
func (a *BaseAgent) actNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	reasoning, _ := state.Get("reasoning")

	// Check for native tool calls first
	var toolCalls []llm.ToolCall
	if val, ok := state.Get("pending_tool_calls"); ok {
		if val != nil {
			if tc, ok := val.([]llm.ToolCall); ok {
				toolCalls = tc
			}
		}
	} else {
		// Fallback to parsing reasoning text
		toolCalls = a.parseToolCalls(fmt.Sprintf("%v", reasoning))
	}

	if len(toolCalls) == 0 {
		// No tools needed, just return the reasoning as action
		state.Set("action", reasoning)
		return state, nil
	}

	// Check for interrupts before execution
	for _, toolCall := range toolCalls {
		for _, interruptTool := range a.config.InterruptOn {
			if toolCall.Function.Name == interruptTool {
				// Trigger interrupt
				state.Set("__interrupt__", map[string]interface{}{
					"tool_call": toolCall,
					"type":      "approval_required",
				})
				// We return early. The edge condition 'shouldObserve' will handle stopping the graph.
				return state, nil
			}
		}
	}

	// Create ToolRuntime for state injection
	runtime := &ToolRuntime{
		State:    state,
		ThreadID: a.config.ID,
		// Store will be set when we have persistence layer ready
	}

	// Execute tools using ToolNode (supports parallel execution and commands)
	messages, cmd, err := a.toolNode.ExecuteTools(ctx, toolCalls, runtime)
	if err != nil {
		a.logger.WithError(err).Error("Tool execution failed")
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Handle command if returned
	if cmd != nil {
		// Commands can update state, navigate to different nodes, or resume from interrupts
		if cmd.Update != nil {
			for key, value := range cmd.Update {
				state.Set(key, value)
			}
		}
		if cmd.Goto != "" {
			// Store the goto target for the router to handle
			state.Set("__goto__", cmd.Goto)
		}
		// Resume commands would be handled by HITL system
	}

	// Ensure tool call IDs exist and mirror them onto the last assistant message.
	// Strict OpenAI-compat APIs (DeepSeek) reject empty/mismatched tool_call_id pairs.
	for i := range toolCalls {
		if strings.TrimSpace(toolCalls[i].ID) == "" {
			toolCalls[i].ID = uuid.NewString()
		}
		if toolCalls[i].Type == "" {
			toolCalls[i].Type = "function"
		}
	}
	a.ensureAssistantToolCallIDs(toolCalls)

	// Convert tool messages to action results and append role=tool replies so the
	// next reason step has a valid OpenAI tool-call conversation shape.
	var results []string
	for i, msg := range messages {
		if strings.TrimSpace(msg.ToolCallID) == "" && i < len(toolCalls) {
			msg.ToolCallID = toolCalls[i].ID
		}
		if msg.Role == "" {
			msg.Role = "tool"
		}
		results = append(results, msg.Content)
		a.conversation.AddMessage(msg)
	}

	state.Set("action", strings.Join(results, "\n"))
	state.Set("tool_calls", toolCalls)
	state.Set("tool_messages_added", true)

	a.logger.WithField("tool_calls", len(toolCalls)).Info("Agent action completed")
	return state, nil
}

// ensureAssistantToolCallIDs rewrites the most recent assistant tool_calls to match
// the IDs used for subsequent role=tool messages.
func (a *BaseAgent) ensureAssistantToolCallIDs(toolCalls []llm.ToolCall) {
	if a.conversation == nil || len(toolCalls) == 0 {
		return
	}
	msgs := a.conversation.GetMessages()
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != "assistant" || len(msgs[i].ToolCalls) == 0 {
			continue
		}
		updated := msgs[i]
		calls := make([]llm.ToolCall, len(updated.ToolCalls))
		copy(calls, updated.ToolCalls)
		for j := range calls {
			if j < len(toolCalls) {
				calls[j].ID = toolCalls[j].ID
				if calls[j].Type == "" {
					calls[j].Type = toolCalls[j].Type
				}
			}
		}
		updated.ToolCalls = calls
		a.conversation.ReplaceMessage(i, updated)
		return
	}
}

// observeNode implements the observation step in ReAct
func (a *BaseAgent) observeNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	action, _ := state.Get("action")

	// Create observation based on action results
	observation := fmt.Sprintf("Observation: %v", action)
	state.Set("observation", observation)

	// When actNode already appended role=tool messages, do NOT add another
	// assistant "Observation:" turn — that leaves orphan tool_calls and breaks
	// strict providers (DeepSeek 400: insufficient tool messages).
	toolMsgsAdded, _ := state.Get("tool_messages_added")
	if toolMsgsAdded != true {
		a.conversation.AddMessage(llm.Message{
			Role:    "user",
			Content: observation,
		})
	}
	state.Set("tool_messages_added", false)

	// Increment iteration
	iteration, _ := state.Get("iteration")
	if iter, ok := iteration.(int); ok {
		state.Set("iteration", iter+1)
	}

	a.logger.WithField("observation", observation).Info("Agent observation completed")
	return state, nil
}

// finalizeNode implements the finalization step
func (a *BaseAgent) finalizeNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
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
	// SLMs often emit empty replies or raw tool-call XML as "final" — recover from
	// the last observation when possible instead of leaving an unusable finish.
	if strings.TrimSpace(output) == "" || looksLikeToolCallJunk(output) {
		if obs, ok := state.Get("observation"); ok {
			if s, ok := obs.(string); ok && strings.TrimSpace(s) != "" {
				output = fmt.Sprintf(`{"status":"done","summary":%q,"notes":"recovered from incomplete finalize"}`, truncateStr(s, 500))
			}
		}
		if strings.TrimSpace(output) == "" {
			output = `{"status":"blocked","summary":"empty finalize","notes":"retry with clearer finish instruction"}`
		} else if looksLikeToolCallJunk(output) {
			output = `{"status":"blocked","summary":"model ended on a tool call","notes":"retry with clearer finish instruction"}`
		}
	}
	state.Set("output", output)

	// Add final message to conversation
	a.conversation.AddMessage(resp.Choices[0].Message)

	a.logger.WithField("output", output).Info("Agent finalization completed")
	return state, nil
}

func looksLikeToolCallJunk(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return false
	}
	return strings.Contains(lower, "<function") ||
		strings.Contains(lower, "</tool_call>") ||
		strings.Contains(lower, "<tool_call>") ||
		(strings.HasPrefix(lower, "action:") && !strings.Contains(lower, "{"))
}

func truncateStr(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// chatNode implements simple chat functionality
func (a *BaseAgent) chatNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
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
		EarlyExit:   a.config.EarlyExit,
	}

	var resp *llm.CompletionResponse
	var err error

	// Use streaming mode if enabled (supports token-stream early-exit via EarlyExit)
	if a.config.EnableStreaming {
		mode := a.config.StreamingMode
		if mode == "" {
			mode = llm.StreamModeForced
		}
		resp, err = a.llmManager.CompleteWithMode(ctx, a.config.Provider, req, mode)
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
					toolResults = append(toolResults, fmt.Sprintf("%v", result))
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
func (a *BaseAgent) planNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
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
func (a *BaseAgent) executeToolsNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	plan, _ := state.Get("plan")

	// Parse the plan to extract tool calls
	toolCalls := a.parseToolCalls(fmt.Sprintf("%v", plan))

	var results []string
	var executedCalls []llm.ToolCall

	// Check for interrupts
	for _, toolCall := range toolCalls {
		for _, interruptTool := range a.config.InterruptOn {
			if toolCall.Function.Name == interruptTool {
				// Trigger interrupt
				state.Set("__interrupt__", map[string]interface{}{
					"tool_call": toolCall,
					"type":      "approval_required",
				})
				return state, nil
			}
		}
	}

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
			results = append(results, fmt.Sprintf("%v", result))
		}

		executedCalls = append(executedCalls, toolCall)
	}

	state.Set("execution_results", results)
	state.Set("tool_calls", executedCalls)

	a.logger.WithField("tool_calls", len(executedCalls)).Info("Agent tool execution completed")
	return state, nil
}

// reviewNode reviews the execution results
func (a *BaseAgent) reviewNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
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

func (a *BaseAgent) shouldAct(ctx context.Context, state *core.BaseState) (string, error) {
	// Check if reasoning produced tool calls
	_, hasToolCalls := state.Get("pending_tool_calls")
	if hasToolCalls {
		calls, ok := state.Get("pending_tool_calls")
		if ok {
			if toolCalls, ok := calls.([]llm.ToolCall); ok && len(toolCalls) > 0 {
				return "act", nil
			}
		}
	}
	return "finalize", nil
}

// shouldObserve determines if the agent should observe action results
func (a *BaseAgent) shouldObserve(ctx context.Context, state *core.BaseState) (string, error) {
	// If we have an interrupt, we don't proceed to observe yet
	if _, hasInterrupt := state.Get("__interrupt__"); hasInterrupt {
		return "", nil // Stay at current node (act) - execution will pause
	}
	// Always observe after acting (if no interrupt)
	return "observe", nil
}

// shouldContinueReasoning determines if the agent should continue reasoning loop
func (a *BaseAgent) shouldContinueReasoning(ctx context.Context, state *core.BaseState) (string, error) {
	iteration, ok := state.Get("iteration")
	if !ok {
		return "finalize", nil
	}

	iter, ok := iteration.(int)
	if !ok {
		return "finalize", nil
	}

	if iter >= a.config.MaxIterations {
		return "finalize", nil
	}

	return "reason", nil
}

// shouldFinalize always returns finalize (used for unconditional routing)
func (a *BaseAgent) shouldFinalize(ctx context.Context, state *core.BaseState) (string, error) {
	return "finalize", nil
}

// shouldReplan determines if the agent should replan
func (a *BaseAgent) shouldReplan(ctx context.Context, state *core.BaseState) (string, error) {
	return "plan", nil
}

// Helper methods

func (a *BaseAgent) buildReasoningMessages(state *core.BaseState) []llm.Message {
	messages := sanitizeToolPairing(a.conversation.GetMessages())

	// Add system prompt
	if a.config.SystemPrompt != "" {
		systemMsg := llm.Message{
			Role:    "system",
			Content: a.config.SystemPrompt,
		}
		messages = append([]llm.Message{systemMsg}, messages...)
	}

	return messages
}

func (a *BaseAgent) buildFinalizationMessages(state *core.BaseState) []llm.Message {
	messages := a.buildReasoningMessages(state)
	messages = sanitizeToolPairing(messages)
	// Explicit finish steer — SLMs otherwise often emit another tool call as "final".
	messages = append(messages, llm.Message{
		Role: "user",
		Content: "Finalize now. Do NOT emit tool calls or <tool_call> XML. " +
			`Reply with STRICT JSON only: {"status":"done|blocked","summary":"...","files_changed":[],"notes":""}`,
	})
	return messages
}

// sanitizeToolPairing drops trailing unpaired assistant tool_calls (or injects
// stub tool errors) so strict OpenAI-compat APIs accept the history.
func sanitizeToolPairing(messages []llm.Message) []llm.Message {
	if len(messages) == 0 {
		return messages
	}
	out := make([]llm.Message, 0, len(messages)+4)
	for i := 0; i < len(messages); i++ {
		m := messages[i]
		if m.Role != "assistant" || len(m.ToolCalls) == 0 {
			out = append(out, m)
			continue
		}
		needed := make(map[string]bool, len(m.ToolCalls))
		for _, tc := range m.ToolCalls {
			if id := strings.TrimSpace(tc.ID); id != "" {
				needed[id] = true
			}
		}
		j := i + 1
		for j < len(messages) && messages[j].Role == "tool" {
			delete(needed, messages[j].ToolCallID)
			j++
		}
		out = append(out, m)
		for ; i+1 < j; i++ {
			out = append(out, messages[i+1])
		}
		for id := range needed {
			out = append(out, llm.Message{
				Role:       "tool",
				ToolCallID: id,
				Content:    "Error: tool call was not executed (routing skipped act)",
			})
		}
	}
	return out
}

func (a *BaseAgent) parseToolCalls(content string) []llm.ToolCall {
	// Simple parsing - look for tool usage patterns
	lines := strings.Split(content, "\n") // Changed 'text' to 'content'
	var toolCalls []llm.ToolCall          // Declare toolCalls slice
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

// UpdateConfig updates the agent configuration
func (a *BaseAgent) UpdateConfig(config *AgentConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config = config
	a.buildGraph() // Rebuild graph with new config
}

// GetExecutionHistory returns the agent's execution history
func (a *BaseAgent) GetExecutionHistory() []AgentExecution {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Return a copy to avoid race conditions
	history := make([]AgentExecution, len(a.executionHistory))
	copy(history, a.executionHistory)
	return history
}

// ClearHistory clears the agent's execution history
func (a *BaseAgent) ClearHistory() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.executionHistory = make([]AgentExecution, 0)
}

// GetConversation returns the agent's conversation history
func (a *BaseAgent) GetConversation() []llm.Message {
	return a.conversation.GetMessages()
}

// ClearConversation clears the agent's conversation history
func (a *BaseAgent) ClearConversation() {
	a.conversation.Clear()
}

func (a *BaseAgent) accumulateUsage(state *core.BaseState, resp *llm.CompletionResponse) {
	if resp == nil {
		return
	}
	_ = llm.EnsureUsage(resp, llm.CompletionRequest{})
	add := resp.Usage
	var cur llm.Usage
	if v, ok := state.Get("usage"); ok {
		if u, ok := v.(llm.Usage); ok {
			cur = u
		}
	}
	cur.PromptTokens += add.PromptTokens
	cur.CompletionTokens += add.CompletionTokens
	cur.TotalTokens += add.TotalTokens
	if cur.TotalTokens == 0 {
		cur.TotalTokens = cur.PromptTokens + cur.CompletionTokens
	}
	state.Set("usage", cur)
	if est, _ := resp.Metadata["usage_estimated"].(bool); est {
		state.Set("usage_estimated", true)
	}
}

// SeedConversation restores prior ReAct messages for mid-run HITL resume.
func (a *BaseAgent) SeedConversation(messages []llm.Message) {
	a.conversation.Clear()
	for _, m := range messages {
		a.conversation.AddMessage(m)
	}
}

// SeedResumeState restores conversation, iteration, and pending tool calls.
func (a *BaseAgent) SeedResumeState(messages []llm.Message, iteration int, pending []llm.ToolCall) {
	a.SeedConversation(messages)
	a.mu.Lock()
	defer a.mu.Unlock()
	if iteration < 0 {
		iteration = 0
	}
	a.currentIteration = iteration
	a.pendingToolCalls = append([]llm.ToolCall(nil), pending...)
}

// SetGraph sets the agent's execution graph
func (a *BaseAgent) SetGraph(graph *core.Graph) {
	a.graph = graph
}

// EnableStreaming enables streaming mode for the agent
func (a *BaseAgent) EnableStreaming() error {
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
func (a *BaseAgent) DisableStreaming() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.EnableStreaming = false
	a.config.StreamingMode = llm.StreamModeNone

	// Disable streaming on the provider as well
	return a.llmManager.DisableStreaming(a.config.Provider)
}

// SetStreamingMode sets the streaming mode for the agent
func (a *BaseAgent) SetStreamingMode(mode llm.StreamMode) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config.StreamingMode = mode
	a.config.EnableStreaming = mode != llm.StreamModeNone

	// Set streaming mode on the provider as well
	return a.llmManager.SetStreamingMode(a.config.Provider, mode)
}

// GetStreamingMode returns the current streaming mode
func (a *BaseAgent) GetStreamingMode() llm.StreamMode {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.StreamingMode
}

// IsStreamingEnabled returns whether streaming is enabled
func (a *BaseAgent) IsStreamingEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config.EnableStreaming
}

// Multi-Agent Coordination

// MultiAgentCoordinator manages multiple agents for complex workflows
type MultiAgentCoordinator struct {
	agents map[string]Agent
	config *MultiAgentConfig
	mu     sync.RWMutex
}

// NewMultiAgentCoordinator creates a new coordinator
// If config is nil, creates a default configuration
func NewMultiAgentCoordinator(config *MultiAgentConfig) *MultiAgentCoordinator {
	if config == nil {
		config = &MultiAgentConfig{
			Agents: make(map[string]*AgentConfig),
		}
	}

	return &MultiAgentCoordinator{
		agents: make(map[string]Agent),
		config: config,
	}
}

// RegisterAgent adds an agent to the coordinator
func (mac *MultiAgentCoordinator) RegisterAgent(id string, agent Agent) {
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

// GetAgent retrieves an agent by ID
func (mac *MultiAgentCoordinator) GetAgent(id string) (Agent, bool) {
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
		// Use output as input for next agent - handle type assertion
		if outputStr, ok := execution.Output.(string); ok {
			currentInput = outputStr
		} else {
			currentInput = fmt.Sprintf("%v", execution.Output)
		}
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
