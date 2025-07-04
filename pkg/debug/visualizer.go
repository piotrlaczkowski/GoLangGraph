// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package debug

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

// VisualizerConfig represents visualizer configuration
type VisualizerConfig struct {
	EnableRealTimeUpdates bool   `json:"enable_real_time_updates"`
	MaxHistorySize        int    `json:"max_history_size"`
	OutputFormat          string `json:"output_format"` // "mermaid", "dot", "json"
	IncludeMetadata       bool   `json:"include_metadata"`
	IncludeTimestamps     bool   `json:"include_timestamps"`
}

// DefaultVisualizerConfig returns default visualizer configuration
func DefaultVisualizerConfig() *VisualizerConfig {
	return &VisualizerConfig{
		EnableRealTimeUpdates: true,
		MaxHistorySize:        100,
		OutputFormat:          "mermaid",
		IncludeMetadata:       true,
		IncludeTimestamps:     true,
	}
}

// GraphVisualizer provides graph visualization and debugging capabilities
type GraphVisualizer struct {
	config           *VisualizerConfig
	logger           *logrus.Logger
	executionHistory []ExecutionStep
	subscribers      []VisualizationSubscriber
	checkpointer     persistence.Checkpointer
	mu               sync.RWMutex // Added mutex for thread safety
}

// ExecutionStep represents a single step in graph execution
type ExecutionStep struct {
	ID            string                 `json:"id"`
	ThreadID      string                 `json:"thread_id"`
	NodeID        string                 `json:"node_id"`
	StepType      string                 `json:"step_type"` // "enter", "exit", "error"
	Timestamp     time.Time              `json:"timestamp"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Input         interface{}            `json:"input,omitempty"`
	Output        interface{}            `json:"output,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
	StateSnapshot *core.BaseState        `json:"state_snapshot,omitempty"`
}

// VisualizationSubscriber defines interface for visualization subscribers
type VisualizationSubscriber interface {
	OnStepExecuted(step *ExecutionStep)
	OnGraphCompleted(threadID string, steps []ExecutionStep)
	OnError(err error)
}

// GraphTopology represents the structure of a graph
type GraphTopology struct {
	Nodes []NodeInfo `json:"nodes"`
	Edges []EdgeInfo `json:"edges"`
}

// NodeInfo represents information about a graph node
type NodeInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsStartNode bool                   `json:"is_start_node"`
	IsEndNode   bool                   `json:"is_end_node"`
	Position    *Position              `json:"position,omitempty"`
}

// EdgeInfo represents information about a graph edge
type EdgeInfo struct {
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Condition string                 `json:"condition,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Position represents node position for visualization
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// NewGraphVisualizer creates a new graph visualizer
func NewGraphVisualizer(config *VisualizerConfig, checkpointer persistence.Checkpointer) *GraphVisualizer {
	if config == nil {
		config = DefaultVisualizerConfig()
	}

	return &GraphVisualizer{
		config:           config,
		logger:           logrus.New(),
		executionHistory: make([]ExecutionStep, 0),
		subscribers:      make([]VisualizationSubscriber, 0),
		checkpointer:     checkpointer,
	}
}

// Subscribe adds a visualization subscriber
func (gv *GraphVisualizer) Subscribe(subscriber VisualizationSubscriber) {
	gv.mu.Lock()
	defer gv.mu.Unlock()
	gv.subscribers = append(gv.subscribers, subscriber)
}

// RecordStep records an execution step
func (gv *GraphVisualizer) RecordStep(step *ExecutionStep) {
	gv.mu.Lock()
	defer gv.mu.Unlock()

	// Add to history
	gv.executionHistory = append(gv.executionHistory, *step)

	// Trim history if needed
	if len(gv.executionHistory) > gv.config.MaxHistorySize {
		gv.executionHistory = gv.executionHistory[1:]
	}

	// Notify subscribers
	for _, subscriber := range gv.subscribers {
		subscriber.OnStepExecuted(step)
	}

	// Save checkpoint if available
	if gv.checkpointer != nil && step.StateSnapshot != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		checkpoint := &persistence.Checkpoint{
			ID:        step.ID,
			ThreadID:  step.ThreadID,
			State:     step.StateSnapshot,
			Metadata:  step.Metadata,
			CreatedAt: step.Timestamp,
			NodeID:    step.NodeID,
			StepID:    0, // Use numeric step ID
		}

		if err := gv.checkpointer.Save(ctx, checkpoint); err != nil {
			gv.logger.WithError(err).Error("Failed to save checkpoint")
		}
	}
}

// GetExecutionHistory returns the execution history
func (gv *GraphVisualizer) GetExecutionHistory(threadID string) []ExecutionStep {
	gv.mu.RLock()
	defer gv.mu.RUnlock()

	if threadID == "" {
		// Return a copy to avoid race conditions
		result := make([]ExecutionStep, len(gv.executionHistory))
		copy(result, gv.executionHistory)
		return result
	}

	var filtered []ExecutionStep
	for _, step := range gv.executionHistory {
		if step.ThreadID == threadID {
			filtered = append(filtered, step)
		}
	}
	return filtered
}

// GetGraphTopology extracts topology information from a graph
func (gv *GraphVisualizer) GetGraphTopology(graph *core.Graph) *GraphTopology {
	topology := &GraphTopology{
		Nodes: make([]NodeInfo, 0),
		Edges: make([]EdgeInfo, 0),
	}

	// Extract nodes
	for _, node := range graph.Nodes {
		nodeInfo := NodeInfo{
			ID:          node.ID,
			Name:        node.Name,
			Type:        gv.getNodeType(node),
			Metadata:    node.Metadata,
			IsStartNode: node.ID == graph.StartNode,
			IsEndNode:   gv.isEndNode(graph, node.ID),
		}
		topology.Nodes = append(topology.Nodes, nodeInfo)
	}

	// Extract edges
	for _, edge := range graph.Edges {
		edgeInfo := EdgeInfo{
			From:      edge.From,
			To:        edge.To,
			Condition: gv.getConditionName(edge),
			Metadata:  make(map[string]interface{}),
		}
		topology.Edges = append(topology.Edges, edgeInfo)
	}

	return topology
}

// GenerateMermaidDiagram generates a Mermaid diagram from graph topology
func (gv *GraphVisualizer) GenerateMermaidDiagram(topology *GraphTopology) string {
	var builder strings.Builder
	builder.WriteString("graph TD\n")

	// Add nodes
	for _, node := range topology.Nodes {
		nodeShape := gv.getMermaidNodeShape(node)
		builder.WriteString(fmt.Sprintf("    %s%s\n", node.ID, nodeShape))

		// Add styling for special nodes
		if node.IsStartNode {
			builder.WriteString("    classDef startNode fill:#90EE90\n")
			builder.WriteString(fmt.Sprintf("    class %s startNode\n", node.ID))
		}
		if node.IsEndNode {
			builder.WriteString("    classDef endNode fill:#FFB6C1\n")
			builder.WriteString(fmt.Sprintf("    class %s endNode\n", node.ID))
		}
	}

	// Add edges
	for _, edge := range topology.Edges {
		edgeLabel := ""
		if edge.Condition != "" {
			edgeLabel = fmt.Sprintf("|%s|", edge.Condition)
		}
		builder.WriteString(fmt.Sprintf("    %s --> %s%s\n", edge.From, edge.To, edgeLabel))
	}

	return builder.String()
}

// GenerateDotDiagram generates a DOT diagram from graph topology
func (gv *GraphVisualizer) GenerateDotDiagram(topology *GraphTopology) string {
	var builder strings.Builder
	builder.WriteString("digraph G {\n")
	builder.WriteString("    rankdir=TD;\n")
	builder.WriteString("    node [shape=box];\n")

	// Add nodes
	for _, node := range topology.Nodes {
		attrs := []string{fmt.Sprintf("label=\"%s\"", node.Name)}

		if node.IsStartNode {
			attrs = append(attrs, "style=filled", "fillcolor=lightgreen")
		}
		if node.IsEndNode {
			attrs = append(attrs, "style=filled", "fillcolor=lightcoral")
		}

		builder.WriteString(fmt.Sprintf("    %s [%s];\n", node.ID, strings.Join(attrs, ", ")))
	}

	// Add edges
	for _, edge := range topology.Edges {
		attrs := []string{}
		if edge.Condition != "" {
			attrs = append(attrs, fmt.Sprintf("label=\"%s\"", edge.Condition))
		}

		edgeAttrs := ""
		if len(attrs) > 0 {
			edgeAttrs = fmt.Sprintf(" [%s]", strings.Join(attrs, ", "))
		}

		builder.WriteString(fmt.Sprintf("    %s -> %s%s;\n", edge.From, edge.To, edgeAttrs))
	}

	builder.WriteString("}\n")
	return builder.String()
}

// GenerateExecutionTrace generates an execution trace visualization
func (gv *GraphVisualizer) GenerateExecutionTrace(threadID string) string {
	steps := gv.GetExecutionHistory(threadID)

	var builder strings.Builder
	builder.WriteString("graph TD\n")

	for i, step := range steps {
		nodeID := fmt.Sprintf("step%d", i)
		label := fmt.Sprintf("%s\\n%s", step.NodeID, step.StepType)

		if gv.config.IncludeTimestamps {
			label += fmt.Sprintf("\\n%s", step.Timestamp.Format("15:04:05"))
		}

		// Color based on step type
		var color string
		switch step.StepType {
		case "enter":
			color = "#87CEEB"
		case "exit":
			color = "#90EE90"
		case "error":
			color = "#FFB6C1"
		default:
			color = "#F0F0F0"
		}

		builder.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeID, label))
		builder.WriteString(fmt.Sprintf("    style %s fill:%s\n", nodeID, color))

		// Connect to next step
		if i < len(steps)-1 {
			builder.WriteString(fmt.Sprintf("    %s --> step%d\n", nodeID, i+1))
		}
	}

	return builder.String()
}

// GetDebugInfo returns debug information for a specific execution
func (gv *GraphVisualizer) GetDebugInfo(threadID string) map[string]interface{} {
	steps := gv.GetExecutionHistory(threadID)

	debugInfo := map[string]interface{}{
		"thread_id":      threadID,
		"total_steps":    len(steps),
		"execution_time": gv.calculateTotalExecutionTime(steps),
		"steps":          steps,
	}

	if len(steps) > 0 {
		debugInfo["start_time"] = steps[0].Timestamp
		debugInfo["end_time"] = steps[len(steps)-1].Timestamp
	}

	// Add performance metrics
	debugInfo["performance"] = gv.calculatePerformanceMetrics(steps)

	return debugInfo
}

// Helper methods

func (gv *GraphVisualizer) getNodeType(node *core.Node) string {
	if nodeType, exists := node.Metadata["type"]; exists {
		if str, ok := nodeType.(string); ok {
			return str
		}
	}
	return "default"
}

func (gv *GraphVisualizer) isEndNode(graph *core.Graph, nodeID string) bool {
	for _, endNode := range graph.EndNodes {
		if endNode == nodeID {
			return true
		}
	}
	return false
}

func (gv *GraphVisualizer) getConditionName(edge *core.Edge) string {
	if edge.Condition == nil {
		return ""
	}
	// This is a placeholder - in a real implementation, you'd want to
	// extract meaningful condition names from the condition function
	return "condition"
}

func (gv *GraphVisualizer) getMermaidNodeShape(node NodeInfo) string {
	switch node.Type {
	case "start":
		return fmt.Sprintf("((%s))", node.Name)
	case "end":
		return fmt.Sprintf("((%s))", node.Name)
	case "decision":
		return fmt.Sprintf("{%s}", node.Name)
	case "process":
		return fmt.Sprintf("[%s]", node.Name)
	default:
		return fmt.Sprintf("[%s]", node.Name)
	}
}

func (gv *GraphVisualizer) calculateTotalExecutionTime(steps []ExecutionStep) time.Duration {
	if len(steps) == 0 {
		return 0
	}

	var total time.Duration
	for _, step := range steps {
		total += step.Duration
	}
	return total
}

func (gv *GraphVisualizer) calculatePerformanceMetrics(steps []ExecutionStep) map[string]interface{} {
	if len(steps) == 0 {
		return map[string]interface{}{}
	}

	nodeStats := make(map[string][]time.Duration)
	var totalDuration time.Duration

	for _, step := range steps {
		if step.Duration > 0 {
			nodeStats[step.NodeID] = append(nodeStats[step.NodeID], step.Duration)
			totalDuration += step.Duration
		}
	}

	avgDurations := make(map[string]time.Duration)
	for nodeID, durations := range nodeStats {
		var sum time.Duration
		for _, d := range durations {
			sum += d
		}
		avgDurations[nodeID] = sum / time.Duration(len(durations))
	}

	return map[string]interface{}{
		"total_duration":    totalDuration,
		"average_durations": avgDurations,
		"node_execution_counts": func() map[string]int {
			counts := make(map[string]int)
			for _, step := range steps {
				counts[step.NodeID]++
			}
			return counts
		}(),
	}
}

// WebSocketSubscriber implements VisualizationSubscriber for WebSocket updates
type WebSocketSubscriber struct {
	SendMessage func(message interface{}) error
	logger      *logrus.Logger
}

// NewWebSocketSubscriber creates a new WebSocket subscriber
func NewWebSocketSubscriber(sendMessage func(message interface{}) error) *WebSocketSubscriber {
	return &WebSocketSubscriber{
		SendMessage: sendMessage,
		logger:      logrus.New(),
	}
}

// OnStepExecuted handles step execution events
func (ws *WebSocketSubscriber) OnStepExecuted(step *ExecutionStep) {
	message := map[string]interface{}{
		"type": "step_executed",
		"data": step,
	}

	if err := ws.SendMessage(message); err != nil {
		ws.logger.WithError(err).Error("Failed to send step executed message")
	}
}

// OnGraphCompleted handles graph completion events
func (ws *WebSocketSubscriber) OnGraphCompleted(threadID string, steps []ExecutionStep) {
	message := map[string]interface{}{
		"type": "graph_completed",
		"data": map[string]interface{}{
			"thread_id": threadID,
			"steps":     steps,
		},
	}

	if err := ws.SendMessage(message); err != nil {
		ws.logger.WithError(err).Error("Failed to send graph completed message")
	}
}

// OnError handles error events
func (ws *WebSocketSubscriber) OnError(err error) {
	message := map[string]interface{}{
		"type": "error",
		"data": map[string]interface{}{
			"error": err.Error(),
		},
	}

	if err := ws.SendMessage(message); err != nil {
		ws.logger.WithError(err).Error("Failed to send error message")
	}
}
