// Package core provides the fundamental building blocks for creating and executing graph-based workflows in GoLangGraph.
//
// The core package implements a graph execution engine that allows you to define workflows as directed graphs
// where nodes represent computational units and edges define the flow of execution. This package is the
// foundation of the GoLangGraph framework and provides the essential abstractions for building AI agent workflows.
//
// # Graph Execution Model
//
// The core execution model revolves around two main concepts:
//
//   - Graph: A directed graph structure containing nodes and edges that defines the workflow topology
//   - BaseState: A thread-safe state container that carries data between nodes during execution
//
// # Basic Usage
//
// Creating and executing a simple graph:
//
//	graph := core.NewGraph("my-workflow")
//
//	// Add nodes with processing functions
//	graph.AddNode("start", "Start Node", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
//		state.Set("message", "Hello, World!")
//		return state, nil
//	})
//
//	graph.AddNode("end", "End Node", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
//		message, _ := state.Get("message")
//		fmt.Println(message)
//		return state, nil
//	})
//
//	// Connect nodes
//	graph.AddEdge("start", "end", nil)
//	graph.SetStartNode("start")
//	graph.AddEndNode("end")
//
//	// Execute the graph
//	initialState := core.NewBaseState()
//	ctx := context.Background()
//	finalState, err := graph.Execute(ctx, initialState)
//
// # Conditional Execution
//
// Graphs support conditional edges that determine the next node based on the current state:
//
//	graph.AddEdge("decision", "path_a", func(ctx context.Context, state *core.BaseState) (string, error) {
//		if condition, _ := state.Get("condition"); condition == "A" {
//			return "path_a", nil
//		}
//		return "path_b", nil
//	})
//
// # State Management
//
// The BaseState provides thread-safe access to workflow data:
//
//	state := core.NewBaseState()
//	state.Set("key", "value")
//	value, exists := state.Get("key")
//	state.SetMetadata("execution_id", "12345")
//
//	// Clone state for parallel processing
//	clonedState := state.Clone()
//
//	// Merge states from parallel branches
//	state.Merge(otherState)
//
// # Streaming Execution
//
// For long-running workflows, use streaming execution to receive intermediate results:
//
//	resultChan := make(chan *core.BaseState, 10)
//	go func() {
//		err := graph.Stream(ctx, initialState, resultChan)
//		close(resultChan)
//	}()
//
//	for state := range resultChan {
//		// Process intermediate state
//	}
//
// # Error Handling
//
// The package provides comprehensive error handling with automatic retries and graceful degradation:
//
//   - Node execution errors are wrapped with context information
//   - Validation errors prevent invalid graph configurations
//   - Timeout handling for long-running operations
//   - Interrupt support for graceful cancellation
//
// # Thread Safety
//
// All core types are designed to be thread-safe:
//
//   - BaseState uses read-write mutexes for concurrent access
//   - Graph execution supports parallel node processing
//   - State cloning enables safe parallel branches
//
// # Performance Considerations
//
// The core package is optimized for performance:
//
//   - Minimal memory allocation during execution
//   - Efficient state management with copy-on-write semantics
//   - Lazy evaluation of conditional edges
//   - Configurable retry policies and timeouts
//
// For more advanced usage patterns and integration with other GoLangGraph packages,
// see the examples in the examples/ directory and the comprehensive documentation
// in the docs/ directory.
package core
