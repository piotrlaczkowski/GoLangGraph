package core

import (
	"context"
	"fmt"
)

// ConditionalEdge represents a conditional edge that can route to different nodes
type ConditionalEdge struct {
	ID        string                 `json:"id"`
	From      string                 `json:"from"`
	Condition EdgeCondition          `json:"-"`
	Routes    map[string]string      `json:"routes"` // condition result -> target node
	Metadata  map[string]interface{} `json:"metadata"`
}

// RouterFunction represents a function that determines the next node based on state
type RouterFunction func(ctx context.Context, state *BaseState) (string, error)

// ConditionalRouter manages conditional routing logic
type ConditionalRouter struct {
	routes   map[string]RouterFunction
	fallback string
}

// NewConditionalRouter creates a new conditional router
func NewConditionalRouter(fallback string) *ConditionalRouter {
	return &ConditionalRouter{
		routes:   make(map[string]RouterFunction),
		fallback: fallback,
	}
}

// AddRoute adds a route with a condition
func (cr *ConditionalRouter) AddRoute(condition string, router RouterFunction) {
	cr.routes[condition] = router
}

// Route determines the next node based on state
func (cr *ConditionalRouter) Route(ctx context.Context, state *BaseState) (string, error) {
	for _, router := range cr.routes {
		result, err := router(ctx, state)
		if err != nil {
			continue // Try next condition
		}
		if result != "" {
			return result, nil
		}
	}

	// Return fallback if no conditions match
	return cr.fallback, nil
}

// AddConditionalEdges adds conditional edges to the graph
func (g *Graph) AddConditionalEdges(from string, condition EdgeCondition, routes map[string]string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Verify source node exists
	if _, exists := g.Nodes[from]; !exists {
		return fmt.Errorf("source node %s does not exist", from)
	}

	// Verify target nodes exist
	for _, to := range routes {
		if to != END && to != "__end__" {
			if _, exists := g.Nodes[to]; !exists {
				return fmt.Errorf("target node %s does not exist", to)
			}
		}
	}

	// Create conditional edge
	edge := &ConditionalEdge{
		ID:        fmt.Sprintf("conditional_%s", from),
		From:      from,
		Condition: condition,
		Routes:    routes,
		Metadata:  make(map[string]interface{}),
	}

	// Store the conditional edge (we'll handle this in graph execution)
	if g.Metadata == nil {
		g.Metadata = make(map[string]interface{})
	}

	conditionalEdges, exists := g.Metadata["conditional_edges"]
	if !exists {
		conditionalEdges = make(map[string]*ConditionalEdge)
		g.Metadata["conditional_edges"] = conditionalEdges
	}

	conditionalEdges.(map[string]*ConditionalEdge)[from] = edge

	return nil
}

// GetConditionalEdge retrieves a conditional edge for a node
func (g *Graph) GetConditionalEdge(nodeID string) (*ConditionalEdge, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.Metadata == nil {
		return nil, false
	}

	conditionalEdges, exists := g.Metadata["conditional_edges"]
	if !exists {
		return nil, false
	}

	edge, exists := conditionalEdges.(map[string]*ConditionalEdge)[nodeID]
	return edge, exists
}

// Common routing functions

// RouteByMessageType routes based on the type of the last message
func RouteByMessageType(ctx context.Context, state *BaseState) (string, error) {
	messages, exists := state.Get("messages")
	if !exists {
		return "", fmt.Errorf("no messages in state")
	}

	messageList, ok := messages.([]interface{})
	if !ok || len(messageList) == 0 {
		return "", fmt.Errorf("invalid or empty message list")
	}

	lastMessage := messageList[len(messageList)-1]
	messageMap, ok := lastMessage.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	messageType, exists := messageMap["type"]
	if !exists {
		return "", fmt.Errorf("message has no type")
	}

	switch messageType {
	case "human":
		return "process_human_message", nil
	case "ai":
		return "process_ai_message", nil
	case "tool":
		return "process_tool_message", nil
	default:
		return "default_processor", nil
	}
}

// RouteByToolCalls routes based on whether tool calls are present
func RouteByToolCalls(ctx context.Context, state *BaseState) (string, error) {
	toolCalls, exists := state.Get("tool_calls")
	if !exists {
		return "no_tools", nil
	}

	toolCallList, ok := toolCalls.([]interface{})
	if !ok {
		return "no_tools", nil
	}

	if len(toolCallList) > 0 {
		return "execute_tools", nil
	}

	return "no_tools", nil
}

// RouteByCondition routes based on a boolean condition in state
func RouteByCondition(conditionKey string, trueRoute string, falseRoute string) RouterFunction {
	return func(ctx context.Context, state *BaseState) (string, error) {
		condition, exists := state.Get(conditionKey)
		if !exists {
			return falseRoute, nil
		}

		conditionBool, ok := condition.(bool)
		if !ok {
			return falseRoute, nil
		}

		if conditionBool {
			return trueRoute, nil
		}

		return falseRoute, nil
	}
}

// RouteByCounter routes based on a counter value
func RouteByCounter(counterKey string, maxCount int, continueRoute string, exitRoute string) RouterFunction {
	return func(ctx context.Context, state *BaseState) (string, error) {
		counter, exists := state.Get(counterKey)
		if !exists {
			return continueRoute, nil
		}

		counterInt, ok := counter.(int)
		if !ok {
			return continueRoute, nil
		}

		if counterInt >= maxCount {
			return exitRoute, nil
		}

		return continueRoute, nil
	}
}

// RouteByStateValue routes based on a specific state value
func RouteByStateValue(key string, routes map[interface{}]string, defaultRoute string) RouterFunction {
	return func(ctx context.Context, state *BaseState) (string, error) {
		value, exists := state.Get(key)
		if !exists {
			return defaultRoute, nil
		}

		if route, exists := routes[value]; exists {
			return route, nil
		}

		return defaultRoute, nil
	}
}

// START and END constants for graph flow control
const (
	START = "__start__"
	END   = "__end__"
)

// IsStartNode checks if a node is the start node
func (g *Graph) IsStartNode(nodeID string) bool {
	return nodeID == START || nodeID == g.StartNode
}

// IsEndNode checks if a node is an end node
func (g *Graph) IsEndNode(nodeID string) bool {
	if nodeID == END || nodeID == "__end__" {
		return true
	}

	for _, endNode := range g.EndNodes {
		if endNode == nodeID {
			return true
		}
	}

	return false
}

// GetNextNodes determines the next nodes to execute based on current node and state
func (g *Graph) GetNextNodes(ctx context.Context, currentNodeID string, state *BaseState) ([]string, error) {
	// Check for conditional edges first
	if conditionalEdge, exists := g.GetConditionalEdge(currentNodeID); exists {
		nextNode, err := conditionalEdge.Condition(ctx, state)
		if err != nil {
			return nil, fmt.Errorf("conditional edge evaluation failed: %w", err)
		}

		// Map the condition result to actual node
		if targetNode, exists := conditionalEdge.Routes[nextNode]; exists {
			return []string{targetNode}, nil
		}

		// If no mapping found, use the result directly
		return []string{nextNode}, nil
	}

	// Check regular edges
	var nextNodes []string
	for _, edge := range g.Edges {
		if edge.From == currentNodeID {
			nextNodes = append(nextNodes, edge.To)
		}
	}

	return nextNodes, nil
}
