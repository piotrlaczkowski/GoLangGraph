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
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

func TestNewGraphVisualizer(t *testing.T) {
	config := DefaultVisualizerConfig()
	visualizer := NewGraphVisualizer(config, nil)

	if visualizer == nil {
		t.Fatal("NewGraphVisualizer returned nil")
	}
}

func TestGraphVisualizer_GetGraphTopology(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)
	graph := createTestGraph()

	topology := visualizer.GetGraphTopology(graph)
	if topology == nil {
		t.Fatal("GetGraphTopology returned nil")
	}

	if len(topology.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(topology.Nodes))
	}

	if len(topology.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(topology.Edges))
	}

	// Check node details
	node1Found := false
	node2Found := false
	for _, node := range topology.Nodes {
		if node.ID == "node1" {
			node1Found = true
			if node.Name != "Node 1" {
				t.Errorf("Expected node1 name to be 'Node 1', got %s", node.Name)
			}
			if !node.IsStartNode {
				t.Error("Node1 should be start node")
			}
		}
		if node.ID == "node2" {
			node2Found = true
			if node.Name != "Node 2" {
				t.Errorf("Expected node2 name to be 'Node 2', got %s", node.Name)
			}
			if !node.IsEndNode {
				t.Error("Node2 should be end node")
			}
		}
	}

	if !node1Found {
		t.Error("Node1 not found in topology")
	}
	if !node2Found {
		t.Error("Node2 not found in topology")
	}

	// Check edge details
	if topology.Edges[0].From != "node1" {
		t.Errorf("Expected edge from 'node1', got %s", topology.Edges[0].From)
	}
	if topology.Edges[0].To != "node2" {
		t.Errorf("Expected edge to 'node2', got %s", topology.Edges[0].To)
	}
}

func TestGraphVisualizer_GenerateMermaidDiagram(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)
	graph := createTestGraph()
	topology := visualizer.GetGraphTopology(graph)

	mermaidOutput := visualizer.GenerateMermaidDiagram(topology)
	if mermaidOutput == "" {
		t.Error("Mermaid output should not be empty")
	}

	// Check for basic Mermaid structure
	if !strings.Contains(mermaidOutput, "graph") {
		t.Error("Mermaid output should contain 'graph'")
	}

	if !strings.Contains(mermaidOutput, "node1") {
		t.Error("Mermaid output should contain node1")
	}

	if !strings.Contains(mermaidOutput, "node2") {
		t.Error("Mermaid output should contain node2")
	}

	if !strings.Contains(mermaidOutput, "-->") {
		t.Error("Mermaid output should contain edges")
	}
}

func TestGraphVisualizer_GenerateDotDiagram(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)
	graph := createTestGraph()
	topology := visualizer.GetGraphTopology(graph)

	dotOutput := visualizer.GenerateDotDiagram(topology)
	if dotOutput == "" {
		t.Error("DOT output should not be empty")
	}

	// Check for basic DOT structure
	if !strings.Contains(dotOutput, "digraph") {
		t.Error("DOT output should contain 'digraph'")
	}

	if !strings.Contains(dotOutput, "node1") {
		t.Error("DOT output should contain node1")
	}

	if !strings.Contains(dotOutput, "node2") {
		t.Error("DOT output should contain node2")
	}

	if !strings.Contains(dotOutput, "->") {
		t.Error("DOT output should contain edges")
	}
}

func TestGraphVisualizer_RecordStep(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)

	step := &ExecutionStep{
		ID:        "step1",
		ThreadID:  "thread1",
		NodeID:    "node1",
		StepType:  "enter",
		Timestamp: time.Now(),
		Input:     "test input",
		Output:    "test output",
		Metadata:  map[string]interface{}{"key": "value"},
	}

	visualizer.RecordStep(step)

	history := visualizer.GetExecutionHistory("")
	if len(history) != 1 {
		t.Errorf("Expected 1 step in history, got %d", len(history))
	}

	if history[0].ID != "step1" {
		t.Errorf("Expected step ID 'step1', got %s", history[0].ID)
	}

	if history[0].ThreadID != "thread1" {
		t.Errorf("Expected thread ID 'thread1', got %s", history[0].ThreadID)
	}
}

func TestGraphVisualizer_GetExecutionHistory(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)

	// Add multiple steps
	step1 := &ExecutionStep{
		ID:       "step1",
		ThreadID: "thread1",
		NodeID:   "node1",
		StepType: "enter",
	}
	step2 := &ExecutionStep{
		ID:       "step2",
		ThreadID: "thread2",
		NodeID:   "node2",
		StepType: "enter",
	}
	step3 := &ExecutionStep{
		ID:       "step3",
		ThreadID: "thread1",
		NodeID:   "node2",
		StepType: "exit",
	}

	visualizer.RecordStep(step1)
	visualizer.RecordStep(step2)
	visualizer.RecordStep(step3)

	// Test getting all history
	allHistory := visualizer.GetExecutionHistory("")
	if len(allHistory) != 3 {
		t.Errorf("Expected 3 steps in total history, got %d", len(allHistory))
	}

	// Test filtering by thread ID
	thread1History := visualizer.GetExecutionHistory("thread1")
	if len(thread1History) != 2 {
		t.Errorf("Expected 2 steps for thread1, got %d", len(thread1History))
	}

	thread2History := visualizer.GetExecutionHistory("thread2")
	if len(thread2History) != 1 {
		t.Errorf("Expected 1 step for thread2, got %d", len(thread2History))
	}
}

func TestGraphVisualizer_GenerateExecutionTrace(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)

	// Add some steps
	step1 := &ExecutionStep{
		ID:       "step1",
		ThreadID: "thread1",
		NodeID:   "node1",
		StepType: "enter",
	}
	step2 := &ExecutionStep{
		ID:       "step2",
		ThreadID: "thread1",
		NodeID:   "node1",
		StepType: "exit",
	}

	visualizer.RecordStep(step1)
	visualizer.RecordStep(step2)

	trace := visualizer.GenerateExecutionTrace("thread1")
	if trace == "" {
		t.Error("Execution trace should not be empty")
	}

	// Check for trace content (it's a Mermaid diagram)
	if !strings.Contains(trace, "graph TD") {
		t.Error("Trace should contain Mermaid graph declaration")
	}

	if !strings.Contains(trace, "node1") {
		t.Error("Trace should contain node ID")
	}
}

func TestGraphVisualizer_GetDebugInfo(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)

	// Add some steps
	step1 := &ExecutionStep{
		ID:       "step1",
		ThreadID: "thread1",
		NodeID:   "node1",
		StepType: "enter",
		Duration: 100 * time.Millisecond,
	}

	visualizer.RecordStep(step1)

	debugInfo := visualizer.GetDebugInfo("thread1")
	if debugInfo == nil {
		t.Fatal("Debug info should not be nil")
	}

	if debugInfo["thread_id"] != "thread1" {
		t.Errorf("Expected thread_id 'thread1', got %v", debugInfo["thread_id"])
	}

	if debugInfo["total_steps"] != 1 {
		t.Errorf("Expected total_steps 1, got %v", debugInfo["total_steps"])
	}
}

func TestVisualizerConfig(t *testing.T) {
	config := &VisualizerConfig{
		EnableRealTimeUpdates: true,
		MaxHistorySize:        50,
		OutputFormat:          "dot",
		IncludeMetadata:       false,
		IncludeTimestamps:     false,
	}

	visualizer := NewGraphVisualizer(config, nil)

	if visualizer == nil {
		t.Fatal("NewGraphVisualizer returned nil")
	}

	// Test that config is applied
	if !visualizer.config.EnableRealTimeUpdates {
		t.Error("EnableRealTimeUpdates should be true")
	}

	if visualizer.config.MaxHistorySize != 50 {
		t.Errorf("MaxHistorySize should be 50, got %d", visualizer.config.MaxHistorySize)
	}

	if visualizer.config.OutputFormat != "dot" {
		t.Errorf("OutputFormat should be 'dot', got %s", visualizer.config.OutputFormat)
	}

	if visualizer.config.IncludeMetadata {
		t.Error("IncludeMetadata should be false")
	}

	if visualizer.config.IncludeTimestamps {
		t.Error("IncludeTimestamps should be false")
	}
}

func TestDefaultVisualizerConfig(t *testing.T) {
	config := DefaultVisualizerConfig()

	if config == nil {
		t.Fatal("DefaultVisualizerConfig returned nil")
	}

	if !config.EnableRealTimeUpdates {
		t.Error("Default EnableRealTimeUpdates should be true")
	}

	if config.MaxHistorySize != 100 {
		t.Errorf("Default MaxHistorySize should be 100, got %d", config.MaxHistorySize)
	}

	if config.OutputFormat != "mermaid" {
		t.Errorf("Default OutputFormat should be 'mermaid', got %s", config.OutputFormat)
	}

	if !config.IncludeMetadata {
		t.Error("Default IncludeMetadata should be true")
	}

	if !config.IncludeTimestamps {
		t.Error("Default IncludeTimestamps should be true")
	}
}

func TestGraphVisualizer_Subscribe(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)
	subscriber := &MockSubscriber{}

	visualizer.Subscribe(subscriber)

	// Record a step to trigger subscriber
	step := &ExecutionStep{
		ID:       "step1",
		ThreadID: "thread1",
		NodeID:   "node1",
		StepType: "enter",
	}

	visualizer.RecordStep(step)

	// Check that subscriber was called
	if !subscriber.StepExecutedCalled {
		t.Error("Subscriber OnStepExecuted should have been called")
	}
}

func TestGraphVisualizer_MaxHistorySize(t *testing.T) {
	config := &VisualizerConfig{
		MaxHistorySize: 2,
	}
	visualizer := NewGraphVisualizer(config, nil)

	// Add 3 steps (more than max history size)
	for i := 0; i < 3; i++ {
		step := &ExecutionStep{
			ID:       fmt.Sprintf("step%d", i+1),
			ThreadID: "thread1",
			NodeID:   "node1",
			StepType: "enter",
		}
		visualizer.RecordStep(step)
	}

	history := visualizer.GetExecutionHistory("")
	if len(history) != 2 {
		t.Errorf("Expected history size to be limited to 2, got %d", len(history))
	}

	// Check that the oldest step was removed
	if history[0].ID == "step1" {
		t.Error("Oldest step should have been removed")
	}
}

func TestGraphVisualizer_ConcurrentAccess(t *testing.T) {
	visualizer := NewGraphVisualizer(nil, nil)
	graph := createTestGraph()

	// Test concurrent access to visualizer methods
	done := make(chan bool, 10)

	// Concurrent topology generation
	for i := 0; i < 5; i++ {
		go func() {
			topology := visualizer.GetGraphTopology(graph)
			if topology == nil {
				t.Error("Concurrent topology generation failed")
			}
			done <- true
		}()
	}

	// Concurrent step recording
	for i := 0; i < 5; i++ {
		go func(index int) {
			step := &ExecutionStep{
				ID:       fmt.Sprintf("step%d", index),
				ThreadID: "thread1",
				NodeID:   "node1",
				StepType: "enter",
			}
			visualizer.RecordStep(step)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper functions and types

func createTestGraph() *core.Graph {
	graph := core.NewGraph("test-graph")

	// Add nodes
	graph.AddNode("node1", "Node 1", testNodeFunction)
	graph.AddNode("node2", "Node 2", testNodeFunction)

	// Add edge
	graph.AddEdge("node1", "node2", nil)

	// Set start and end nodes
	graph.SetStartNode("node1")
	graph.AddEndNode("node2")

	return graph
}

func testNodeFunction(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
	// Simple test node function
	state.Set("processed", true)
	return state, nil
}

// MockSubscriber implements VisualizationSubscriber for testing
type MockSubscriber struct {
	StepExecutedCalled   bool
	GraphCompletedCalled bool
	ErrorCalled          bool
	LastStep             *ExecutionStep
	LastThreadID         string
	LastSteps            []ExecutionStep
	LastError            error
}

func (m *MockSubscriber) OnStepExecuted(step *ExecutionStep) {
	m.StepExecutedCalled = true
	m.LastStep = step
}

func (m *MockSubscriber) OnGraphCompleted(threadID string, steps []ExecutionStep) {
	m.GraphCompletedCalled = true
	m.LastThreadID = threadID
	m.LastSteps = steps
}

func (m *MockSubscriber) OnError(err error) {
	m.ErrorCalled = true
	m.LastError = err
}

// Benchmark tests
func BenchmarkGraphVisualizer_GetGraphTopology(b *testing.B) {
	visualizer := NewGraphVisualizer(nil, nil)
	graph := createTestGraph()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		visualizer.GetGraphTopology(graph)
	}
}

func BenchmarkGraphVisualizer_GenerateMermaidDiagram(b *testing.B) {
	visualizer := NewGraphVisualizer(nil, nil)
	graph := createTestGraph()
	topology := visualizer.GetGraphTopology(graph)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		visualizer.GenerateMermaidDiagram(topology)
	}
}

func BenchmarkGraphVisualizer_GenerateDotDiagram(b *testing.B) {
	visualizer := NewGraphVisualizer(nil, nil)
	graph := createTestGraph()
	topology := visualizer.GetGraphTopology(graph)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		visualizer.GenerateDotDiagram(topology)
	}
}

func BenchmarkGraphVisualizer_RecordStep(b *testing.B) {
	visualizer := NewGraphVisualizer(nil, nil)
	step := &ExecutionStep{
		ID:       "step1",
		ThreadID: "thread1",
		NodeID:   "node1",
		StepType: "enter",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		visualizer.RecordStep(step)
	}
}
