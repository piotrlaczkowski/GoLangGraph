package core

import (
	"context"
	"fmt"
	"testing"
)

func TestNewGraph(t *testing.T) {
	graph := NewGraph("test_graph")
	if graph == nil {
		t.Fatal("NewGraph() returned nil")
	}

	if graph.Nodes == nil {
		t.Error("Graph nodes should be initialized")
	}

	if graph.Edges == nil {
		t.Error("Graph edges should be initialized")
	}

	if graph.Name != "test_graph" {
		t.Error("Graph name should be set correctly")
	}
}

func TestGraph_AddNode(t *testing.T) {
	graph := NewGraph("test_graph")

	// Create a simple node function
	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("processed", true)
		return state, nil
	}

	// Test adding a node
	node := graph.AddNode("test_node", "Test Node", nodeFunc)
	if node == nil {
		t.Error("AddNode() should return a node")
	}

	if node.ID != "test_node" {
		t.Error("Node ID should be set correctly")
	}

	if node.Name != "Test Node" {
		t.Error("Node name should be set correctly")
	}

	// Test that node was added to graph
	if _, exists := graph.Nodes["test_node"]; !exists {
		t.Error("Node should be added to graph")
	}
}

func TestGraph_AddEdge(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	// Add nodes first
	graph.AddNode("node1", "Node 1", nodeFunc)
	graph.AddNode("node2", "Node 2", nodeFunc)

	// Test adding edge without condition
	edge := graph.AddEdge("node1", "node2", nil)
	if edge == nil {
		t.Error("AddEdge() should return an edge")
	}

	if edge.From != "node1" {
		t.Error("Edge from should be set correctly")
	}

	if edge.To != "node2" {
		t.Error("Edge to should be set correctly")
	}

	// Test adding edge with condition
	condition := func(ctx context.Context, state *BaseState) (string, error) {
		return "node2", nil
	}

	edge2 := graph.AddEdge("node1", "node2", condition)
	if edge2 == nil {
		t.Error("AddEdge() with condition should return an edge")
	}

	if edge2.Condition == nil {
		t.Error("Edge condition should be set")
	}
}

func TestGraph_SetStartNode(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	// Test setting start node for non-existent node
	err := graph.SetStartNode("nonexistent")
	if err == nil {
		t.Error("SetStartNode() should return error for non-existent node")
	}

	// Add node and set start node
	graph.AddNode("start", "Start Node", nodeFunc)
	err = graph.SetStartNode("start")
	if err != nil {
		t.Errorf("SetStartNode() failed: %v", err)
	}

	if graph.StartNode != "start" {
		t.Error("SetStartNode() failed to set start node")
	}
}

func TestGraph_AddEndNode(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	// Test adding end node for non-existent node
	err := graph.AddEndNode("nonexistent")
	if err == nil {
		t.Error("AddEndNode() should return error for non-existent node")
	}

	// Add node and set as end node
	graph.AddNode("end", "End Node", nodeFunc)
	err = graph.AddEndNode("end")
	if err != nil {
		t.Errorf("AddEndNode() failed: %v", err)
	}

	found := false
	for _, endNode := range graph.EndNodes {
		if endNode == "end" {
			found = true
			break
		}
	}

	if !found {
		t.Error("AddEndNode() failed to add end node")
	}
}

func TestGraph_Validate(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	// Test validation without start node
	err := graph.Validate()
	if err == nil {
		t.Error("Validate() should return error when start node is not set")
	}

	// Add nodes and edges
	graph.AddNode("start", "Start Node", nodeFunc)
	graph.AddNode("end", "End Node", nodeFunc)
	graph.AddEdge("start", "end", nil)
	graph.SetStartNode("start")
	graph.AddEndNode("end")

	// Test successful validation
	err = graph.Validate()
	if err != nil {
		t.Errorf("Validate() failed: %v", err)
	}
}

func TestGraph_Execute(t *testing.T) {
	graph := NewGraph("test_graph")

	// Create test nodes
	node1 := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("node1_executed", true)
		return state, nil
	}

	node2 := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("node2_executed", true)
		return state, nil
	}

	// Build graph
	graph.AddNode("node1", "Node 1", node1)
	graph.AddNode("node2", "Node 2", node2)
	graph.AddEdge("node1", "node2", nil)
	graph.SetStartNode("node1")
	graph.AddEndNode("node2")

	// Execute graph
	state := NewBaseState()
	ctx := context.Background()

	result, err := graph.Execute(ctx, state)
	if err != nil {
		t.Errorf("Execute() failed: %v", err)
	}

	// Check that both nodes were executed
	val1, exists1 := result.Get("node1_executed")
	val2, exists2 := result.Get("node2_executed")

	if !exists1 || val1 != true {
		t.Error("Execute() failed to execute node1")
	}
	if !exists2 || val2 != true {
		t.Error("Execute() failed to execute node2")
	}
}

func TestGraph_ExecuteWithConditionalEdge(t *testing.T) {
	graph := NewGraph("test_graph")

	// Create test nodes
	node1 := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("node1_executed", true)
		state.Set("condition_value", "go_to_node2")
		return state, nil
	}

	node2 := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("node2_executed", true)
		return state, nil
	}

	node3 := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("node3_executed", true)
		return state, nil
	}

	// Conditional function
	condition := func(ctx context.Context, state *BaseState) (string, error) {
		val, _ := state.Get("condition_value")
		if val == "go_to_node2" {
			return "node2", nil
		}
		return "node3", nil
	}

	// Build graph
	graph.AddNode("node1", "Node 1", node1)
	graph.AddNode("node2", "Node 2", node2)
	graph.AddNode("node3", "Node 3", node3)
	graph.AddEdge("node1", "node2", condition)
	graph.AddEdge("node1", "node3", condition)
	graph.SetStartNode("node1")
	graph.AddEndNode("node2")
	graph.AddEndNode("node3")

	// Execute graph
	state := NewBaseState()
	ctx := context.Background()

	result, err := graph.Execute(ctx, state)
	if err != nil {
		t.Errorf("Execute() failed: %v", err)
	}

	// Check that node1 and node2 were executed
	val1, exists1 := result.Get("node1_executed")
	val2, exists2 := result.Get("node2_executed")
	_, exists3 := result.Get("node3_executed")

	if !exists1 || val1 != true {
		t.Error("Execute() failed to execute node1")
	}
	if !exists2 || val2 != true {
		t.Error("Execute() failed to execute node2")
	}
	if exists3 {
		t.Error("Execute() should not have executed node3")
	}
}

func TestGraph_ExecuteWithoutStartNode(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	graph.AddNode("node1", "Node 1", nodeFunc)

	state := NewBaseState()
	ctx := context.Background()

	_, err := graph.Execute(ctx, state)
	if err == nil {
		t.Error("Execute() should return error when start node is not set")
	}
}

func TestGraph_GetTopology(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	// Add nodes and edges
	graph.AddNode("node1", "Node 1", nodeFunc)
	graph.AddNode("node2", "Node 2", nodeFunc)
	graph.AddNode("node3", "Node 3", nodeFunc)
	graph.AddEdge("node1", "node2", nil)
	graph.AddEdge("node2", "node3", nil)

	topology := graph.GetTopology()

	if len(topology["node1"]) != 1 || topology["node1"][0] != "node2" {
		t.Error("GetTopology() should return correct topology for node1")
	}

	if len(topology["node2"]) != 1 || topology["node2"][0] != "node3" {
		t.Error("GetTopology() should return correct topology for node2")
	}
}

func TestGraph_StreamAndInterrupt(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("processed", true)
		return state, nil
	}

	graph.AddNode("node1", "Node 1", nodeFunc)
	graph.SetStartNode("node1")
	graph.AddEndNode("node1")

	// Test streaming
	streamChan := graph.Stream()
	if streamChan == nil {
		t.Error("Stream() should return a channel")
	}

	// Test interrupt
	graph.Interrupt()

	// Test IsRunning
	if graph.IsRunning() {
		t.Error("IsRunning() should return false after interrupt")
	}
}

func TestGraph_Reset(t *testing.T) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	graph.AddNode("node1", "Node 1", nodeFunc)
	graph.SetStartNode("node1")

	// Execute to create some history
	state := NewBaseState()
	ctx := context.Background()
	graph.Execute(ctx, state)

	// Reset graph
	graph.Reset()

	// Check that execution history is cleared
	history := graph.GetExecutionHistory()
	if len(history) != 0 {
		t.Error("Reset() should clear execution history")
	}
}

// Benchmark tests
func BenchmarkGraph_AddNode(b *testing.B) {
	graph := NewGraph("test_graph")
	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		return state, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nodeID := fmt.Sprintf("node_%d", i)
		graph.AddNode(nodeID, "Test Node", nodeFunc)
	}
}

func BenchmarkGraph_Execute(b *testing.B) {
	graph := NewGraph("test_graph")

	nodeFunc := func(ctx context.Context, state *BaseState) (*BaseState, error) {
		state.Set("processed", true)
		return state, nil
	}

	graph.AddNode("node1", "Node 1", nodeFunc)
	graph.SetStartNode("node1")
	graph.AddEndNode("node1")

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state := NewBaseState()
		graph.Execute(ctx, state)
	}
}
