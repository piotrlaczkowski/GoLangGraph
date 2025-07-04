package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// NodeFunc represents a function that can be executed as a node
type NodeFunc func(ctx context.Context, state *BaseState) (*BaseState, error)

// EdgeCondition represents a condition function for conditional edges
type EdgeCondition func(ctx context.Context, state *BaseState) (string, error)

// Node represents a node in the graph
type Node struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Function NodeFunc               `json:"-"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Edge represents an edge in the graph
type Edge struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Condition EdgeCondition          `json:"-"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ExecutionResult represents the result of node execution
type ExecutionResult struct {
	NodeID    string        `json:"node_id"`
	Success   bool          `json:"success"`
	Error     error         `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	State     *BaseState    `json:"state,omitempty"`
}

// GraphConfig represents configuration for graph execution
type GraphConfig struct {
	MaxIterations     int           `json:"max_iterations"`
	Timeout           time.Duration `json:"timeout"`
	EnableStreaming   bool          `json:"enable_streaming"`
	EnableCheckpoints bool          `json:"enable_checkpoints"`
	ParallelExecution bool          `json:"parallel_execution"`
	RetryAttempts     int           `json:"retry_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`
}

// DefaultGraphConfig returns default configuration
func DefaultGraphConfig() *GraphConfig {
	return &GraphConfig{
		MaxIterations:     100,
		Timeout:           30 * time.Minute,
		EnableStreaming:   true,
		EnableCheckpoints: true,
		ParallelExecution: true,
		RetryAttempts:     3,
		RetryDelay:        1 * time.Second,
	}
}

// Graph represents the execution graph
type Graph struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Nodes     map[string]*Node       `json:"nodes"`
	Edges     map[string]*Edge       `json:"edges"`
	StartNode string                 `json:"start_node"`
	EndNodes  []string               `json:"end_nodes"`
	Config    *GraphConfig           `json:"config"`
	Metadata  map[string]interface{} `json:"metadata"`

	// Execution state
	currentState     *BaseState
	executionHistory []*ExecutionResult
	isRunning        bool
	mu               sync.RWMutex

	// Streaming and interrupts
	streamChan    chan *ExecutionResult
	interruptChan chan struct{}

	// Logger
	logger *logrus.Logger
}

// NewGraph creates a new graph
func NewGraph(name string) *Graph {
	return &Graph{
		ID:               uuid.New().String(),
		Name:             name,
		Nodes:            make(map[string]*Node),
		Edges:            make(map[string]*Edge),
		EndNodes:         make([]string, 0),
		Config:           DefaultGraphConfig(),
		Metadata:         make(map[string]interface{}),
		executionHistory: make([]*ExecutionResult, 0),
		streamChan:       make(chan *ExecutionResult, 100),
		interruptChan:    make(chan struct{}),
		logger:           logrus.New(),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(id, name string, fn NodeFunc) *Node {
	g.mu.Lock()
	defer g.mu.Unlock()

	node := &Node{
		ID:       id,
		Name:     name,
		Function: fn,
		Metadata: make(map[string]interface{}),
	}

	g.Nodes[id] = node
	return node
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(from, to string, condition EdgeCondition) *Edge {
	g.mu.Lock()
	defer g.mu.Unlock()

	edge := &Edge{
		ID:        uuid.New().String(),
		From:      from,
		To:        to,
		Condition: condition,
		Metadata:  make(map[string]interface{}),
	}

	g.Edges[edge.ID] = edge
	return edge
}

// SetStartNode sets the starting node for execution
func (g *Graph) SetStartNode(nodeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[nodeID]; !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}

	g.StartNode = nodeID
	return nil
}

// AddEndNode adds an end node to the graph
func (g *Graph) AddEndNode(nodeID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.Nodes[nodeID]; !exists {
		return fmt.Errorf("node %s does not exist", nodeID)
	}

	g.EndNodes = append(g.EndNodes, nodeID)
	return nil
}

// Validate validates the graph structure
func (g *Graph) Validate() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Check if start node is set
	if g.StartNode == "" {
		return fmt.Errorf("start node is not set")
	}

	// Check if start node exists
	if _, exists := g.Nodes[g.StartNode]; !exists {
		return fmt.Errorf("start node %s does not exist", g.StartNode)
	}

	// Check if end nodes exist
	for _, endNode := range g.EndNodes {
		if _, exists := g.Nodes[endNode]; !exists {
			return fmt.Errorf("end node %s does not exist", endNode)
		}
	}

	// Check if all edges reference existing nodes
	for _, edge := range g.Edges {
		if _, exists := g.Nodes[edge.From]; !exists {
			return fmt.Errorf("edge %s references non-existent from node %s", edge.ID, edge.From)
		}
		if _, exists := g.Nodes[edge.To]; !exists {
			return fmt.Errorf("edge %s references non-existent to node %s", edge.ID, edge.To)
		}
	}

	return nil
}

// Execute executes the graph with the given initial state
func (g *Graph) Execute(ctx context.Context, initialState *BaseState) (*BaseState, error) {
	if err := g.Validate(); err != nil {
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	g.mu.Lock()
	g.isRunning = true
	g.currentState = initialState.Clone()
	g.executionHistory = make([]*ExecutionResult, 0)
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		g.isRunning = false
		g.mu.Unlock()
	}()

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, g.Config.Timeout)
	defer cancel()

	// Start execution from the start node
	currentNode := g.StartNode
	iterations := 0

	for {

		// Check for context cancellation
		select {
		case <-execCtx.Done():
			return nil, fmt.Errorf("execution timeout or cancelled")
		case <-g.interruptChan:
			return g.currentState, fmt.Errorf("execution interrupted")
		default:
		}

		// Check iteration limit
		if iterations >= g.Config.MaxIterations {
			return nil, fmt.Errorf("maximum iterations (%d) exceeded", g.Config.MaxIterations)
		}

		// Execute the current node
		result, err := g.executeNode(execCtx, currentNode)
		if err != nil {
			return nil, fmt.Errorf("node execution failed: %w", err)
		}

		// Update current state
		g.mu.Lock()
		g.currentState = result.State
		g.executionHistory = append(g.executionHistory, result)
		g.mu.Unlock()

		// Stream result if enabled
		if g.Config.EnableStreaming {
			select {
			case g.streamChan <- result:
			default:
				// Channel is full, skip streaming this result
			}
		}

		// Check if we've reached an end node AFTER executing it
		if g.isEndNode(currentNode) {
			break
		}

		// Determine next node
		nextNode, err := g.getNextNode(execCtx, currentNode)
		if err != nil {
			return nil, fmt.Errorf("failed to determine next node: %w", err)
		}

		if nextNode == "" {
			// No next node, end execution
			break
		}

		currentNode = nextNode
		iterations++
	}

	return g.currentState, nil
}

// executeNode executes a single node
func (g *Graph) executeNode(ctx context.Context, nodeID string) (*ExecutionResult, error) {
	g.mu.RLock()
	node, exists := g.Nodes[nodeID]
	state := g.currentState.Clone()
	g.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	g.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"node_name": node.Name,
		"graph_id":  g.ID,
	}).Info("Executing node")

	start := time.Now()

	// Execute the node function with retry logic
	var resultState *BaseState
	var err error

	for attempt := 0; attempt <= g.Config.RetryAttempts; attempt++ {
		resultState, err = node.Function(ctx, state)
		if err == nil {
			break
		}

		if attempt < g.Config.RetryAttempts {
			g.logger.WithFields(logrus.Fields{
				"node_id": nodeID,
				"attempt": attempt + 1,
				"error":   err,
			}).Warn("Node execution failed, retrying")

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(g.Config.RetryDelay):
				// Continue with retry
			}
		}
	}

	duration := time.Since(start)

	result := &ExecutionResult{
		NodeID:    nodeID,
		Success:   err == nil,
		Error:     err,
		Duration:  duration,
		Timestamp: time.Now(),
		State:     resultState,
	}

	if err != nil {
		g.logger.WithFields(logrus.Fields{
			"node_id":  nodeID,
			"duration": duration,
			"error":    err,
		}).Error("Node execution failed")
	} else {
		g.logger.WithFields(logrus.Fields{
			"node_id":  nodeID,
			"duration": duration,
		}).Info("Node execution completed")
	}

	return result, err
}

// getNextNode determines the next node to execute
func (g *Graph) getNextNode(ctx context.Context, currentNodeID string) (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Find all outgoing edges from the current node
	var outgoingEdges []*Edge
	for _, edge := range g.Edges {
		if edge.From == currentNodeID {
			outgoingEdges = append(outgoingEdges, edge)
		}
	}

	// If no outgoing edges, execution ends
	if len(outgoingEdges) == 0 {
		return "", nil
	}

	// If only one edge and no condition, follow it
	if len(outgoingEdges) == 1 && outgoingEdges[0].Condition == nil {
		return outgoingEdges[0].To, nil
	}

	// Evaluate conditions for conditional edges
	for _, edge := range outgoingEdges {
		if edge.Condition != nil {
			nextNodeID, err := edge.Condition(ctx, g.currentState)
			if err != nil {
				return "", fmt.Errorf("edge condition evaluation failed: %w", err)
			}
			// The condition function should return the node ID to go to
			// Check if the returned node ID matches this edge's target
			if nextNodeID == edge.To {
				return edge.To, nil
			}
		}
	}

	// If no condition matched, follow the first unconditional edge
	for _, edge := range outgoingEdges {
		if edge.Condition == nil {
			return edge.To, nil
		}
	}

	return "", fmt.Errorf("no valid next node found from %s", currentNodeID)
}

// isEndNode checks if a node is an end node
func (g *Graph) isEndNode(nodeID string) bool {
	for _, endNode := range g.EndNodes {
		if endNode == nodeID {
			return true
		}
	}
	return false
}

// Stream returns a channel for streaming execution results
func (g *Graph) Stream() <-chan *ExecutionResult {
	return g.streamChan
}

// Interrupt interrupts the current execution
func (g *Graph) Interrupt() {
	select {
	case g.interruptChan <- struct{}{}:
	default:
		// Channel is full or closed
	}
}

// IsRunning returns whether the graph is currently executing
func (g *Graph) IsRunning() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.isRunning
}

// GetExecutionHistory returns the execution history
func (g *Graph) GetExecutionHistory() []*ExecutionResult {
	g.mu.RLock()
	defer g.mu.RUnlock()

	history := make([]*ExecutionResult, len(g.executionHistory))
	copy(history, g.executionHistory)
	return history
}

// GetCurrentState returns the current state
func (g *Graph) GetCurrentState() *BaseState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.currentState == nil {
		return nil
	}
	return g.currentState.Clone()
}

// Reset resets the graph execution state
func (g *Graph) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.currentState = nil
	g.executionHistory = make([]*ExecutionResult, 0)
	g.isRunning = false
}

// ExecuteParallel executes multiple nodes in parallel (for super-step execution)
func (g *Graph) ExecuteParallel(ctx context.Context, nodeIDs []string, state *BaseState) (map[string]*ExecutionResult, error) {
	if len(nodeIDs) == 0 {
		return make(map[string]*ExecutionResult), nil
	}

	results := make(map[string]*ExecutionResult)
	resultsMu := sync.Mutex{}
	errChan := make(chan error, len(nodeIDs))

	var wg sync.WaitGroup

	for _, nodeID := range nodeIDs {
		wg.Add(1)
		go func(nID string) {
			defer wg.Done()

			result, err := g.executeNodeWithState(ctx, nID, state.Clone())
			if err != nil {
				errChan <- fmt.Errorf("node %s failed: %w", nID, err)
				return
			}

			resultsMu.Lock()
			results[nID] = result
			resultsMu.Unlock()
		}(nodeID)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		return nil, err
	}

	return results, nil
}

// executeNodeWithState executes a node with a specific state
func (g *Graph) executeNodeWithState(ctx context.Context, nodeID string, state *BaseState) (*ExecutionResult, error) {
	g.mu.RLock()
	node, exists := g.Nodes[nodeID]
	g.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node %s does not exist", nodeID)
	}

	start := time.Now()
	resultState, err := node.Function(ctx, state)
	duration := time.Since(start)

	return &ExecutionResult{
		NodeID:    nodeID,
		Success:   err == nil,
		Error:     err,
		Duration:  duration,
		Timestamp: time.Now(),
		State:     resultState,
	}, err
}

// GetNodesByType returns nodes filtered by metadata type
func (g *Graph) GetNodesByType(nodeType string) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var nodes []*Node
	for _, node := range g.Nodes {
		if nodeTypeValue, exists := node.Metadata["type"]; exists {
			if nodeTypeValue == nodeType {
				nodes = append(nodes, node)
			}
		}
	}
	return nodes
}

// GetTopology returns the graph topology as adjacency list
func (g *Graph) GetTopology() map[string][]string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	topology := make(map[string][]string)

	// Initialize all nodes
	for nodeID := range g.Nodes {
		topology[nodeID] = make([]string, 0)
	}

	// Add edges
	for _, edge := range g.Edges {
		topology[edge.From] = append(topology[edge.From], edge.To)
	}

	return topology
}

// Close closes the graph and cleans up resources
func (g *Graph) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()

	close(g.streamChan)
	close(g.interruptChan)
}
