# Core Package Documentation

The `pkg/core` package provides the foundational components for GoLangGraph, including the graph execution engine and state management system.

## Overview

The core package implements a directed graph execution model where:
- **Nodes** represent computational units (functions)
- **Edges** define execution flow between nodes
- **State** carries data throughout the execution
- **Conditions** enable dynamic routing decisions

## Key Components

### Graph

The `Graph` struct is the main execution engine that manages workflow execution.

```go
type Graph struct {
    ID        string
    Name      string
    Nodes     map[string]*Node
    Edges     map[string]*Edge
    StartNode string
    EndNodes  []string
    Config    *GraphConfig
    // ... internal fields
}
```

#### Creating a Graph

```go
// Create a new graph
graph := core.NewGraph("my_workflow")

// Configure the graph
graph.Config.MaxIterations = 50
graph.Config.Timeout = 5 * time.Minute
graph.Config.EnableStreaming = true
```

#### Adding Nodes

```go
// Define a node function
nodeFunc := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    // Process the state
    state.Set("processed", true)
    return state, nil
}

// Add the node to the graph
graph.AddNode("process_node", "Process Data", nodeFunc)
```

#### Adding Edges

```go
// Simple edge (unconditional)
graph.AddEdge("node1", "node2", nil)

// Conditional edge
condition := func(ctx context.Context, state *core.BaseState) (string, error) {
    value, _ := state.Get("decision")
    if value == "continue" {
        return "node2", nil
    }
    return "", nil // Don't take this edge
}
graph.AddEdge("node1", "node2", condition)
```

#### Executing the Graph

```go
// Set start and end nodes
graph.SetStartNode("start_node")
graph.AddEndNode("end_node")

// Execute the workflow
initialState := core.NewBaseState()
initialState.Set("input", "Hello, World!")

result, err := graph.Execute(context.Background(), initialState)
if err != nil {
    log.Fatal(err)
}

// Access the final state
output, _ := result.Get("output")
fmt.Printf("Result: %s\n", output)
```

### State Management

The `BaseState` struct provides thread-safe state management with history tracking.

```go
type BaseState struct {
    data     map[string]interface{}
    metadata map[string]interface{}
    history  []StateSnapshot
    mu       sync.RWMutex
}
```

#### State Operations

```go
// Create a new state
state := core.NewBaseState()

// Set values
state.Set("key", "value")
state.Set("number", 42)
state.Set("data", map[string]interface{}{
    "nested": "value",
})

// Get values
value, exists := state.Get("key")
if exists {
    fmt.Printf("Value: %s\n", value)
}

// Delete values
state.Delete("key")

// Get all keys
keys := state.Keys()

// Clone state (deep copy)
cloned := state.Clone()

// Merge states
other := core.NewBaseState()
other.Set("new_key", "new_value")
state.Merge(other)
```

#### State Metadata

```go
// Set metadata
state.SetMetadata("version", "1.0")
state.SetMetadata("timestamp", time.Now())

// Get metadata
version, _ := state.GetMetadata("version")

// Get all metadata
metadata := state.GetAllMetadata()
```

#### State History

```go
// Create a snapshot
snapshot := state.CreateSnapshot("checkpoint_1")

// Restore from snapshot
err := state.RestoreFromSnapshot(snapshot)
if err != nil {
    log.Printf("Failed to restore: %v", err)
}

// Get history
history := state.GetHistory()
for _, snapshot := range history {
    fmt.Printf("Snapshot: %s at %v\n", snapshot.Name, snapshot.Timestamp)
}
```

### Node Functions

Node functions are the building blocks of your workflow. They receive a context and state, and return a modified state.

```go
type NodeFunc func(ctx context.Context, state *BaseState) (*BaseState, error)
```

#### Simple Node

```go
simpleNode := func(ctx context.Context, state *BaseState) (*BaseState, error) {
    // Get input
    input, _ := state.Get("input")
    
    // Process
    result := strings.ToUpper(input.(string))
    
    // Set output
    state.Set("output", result)
    
    return state, nil
}
```

#### Node with Error Handling

```go
validationNode := func(ctx context.Context, state *BaseState) (*BaseState, error) {
    input, exists := state.Get("input")
    if !exists {
        return nil, fmt.Errorf("input is required")
    }
    
    if input == "" {
        return nil, fmt.Errorf("input cannot be empty")
    }
    
    state.Set("validated", true)
    return state, nil
}
```

#### Async Node

```go
asyncNode := func(ctx context.Context, state *BaseState) (*BaseState, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Simulate async work
    time.Sleep(100 * time.Millisecond)
    
    state.Set("async_result", "completed")
    return state, nil
}
```

### Edge Conditions

Edge conditions enable dynamic routing based on state values.

```go
type EdgeCondition func(ctx context.Context, state *BaseState) (string, error)
```

#### Simple Condition

```go
condition := func(ctx context.Context, state *BaseState) (string, error) {
    value, _ := state.Get("decision")
    if value == "yes" {
        return "yes_node", nil
    }
    return "no_node", nil
}
```

#### Complex Condition

```go
complexCondition := func(ctx context.Context, state *BaseState) (string, error) {
    score, exists := state.Get("score")
    if !exists {
        return "", fmt.Errorf("score is required for routing")
    }
    
    scoreValue := score.(float64)
    switch {
    case scoreValue >= 90:
        return "excellent_node", nil
    case scoreValue >= 70:
        return "good_node", nil
    case scoreValue >= 50:
        return "average_node", nil
    default:
        return "poor_node", nil
    }
}
```

### Graph Configuration

The `GraphConfig` struct provides configuration options for graph execution.

```go
type GraphConfig struct {
    MaxIterations     int           // Maximum number of iterations
    Timeout           time.Duration // Execution timeout
    EnableStreaming   bool          // Enable real-time streaming
    EnableCheckpoints bool          // Enable checkpointing
    ParallelExecution bool          // Enable parallel execution
    RetryAttempts     int           // Number of retry attempts
    RetryDelay        time.Duration // Delay between retries
}
```

#### Custom Configuration

```go
config := &core.GraphConfig{
    MaxIterations:     100,
    Timeout:           10 * time.Minute,
    EnableStreaming:   true,
    EnableCheckpoints: true,
    ParallelExecution: true,
    RetryAttempts:     3,
    RetryDelay:        1 * time.Second,
}

graph := core.NewGraph("configured_graph")
graph.Config = config
```

### Streaming Execution

Enable real-time monitoring of graph execution.

```go
// Enable streaming
graph.Config.EnableStreaming = true

// Get the streaming channel
streamChan := graph.Stream()

// Execute in a goroutine
go func() {
    result, err := graph.Execute(context.Background(), initialState)
    if err != nil {
        log.Printf("Execution failed: %v", err)
    }
}()

// Listen for execution updates
for result := range streamChan {
    fmt.Printf("Node %s completed in %v\n", result.NodeID, result.Duration)
    if result.Error != nil {
        fmt.Printf("Error in node %s: %v\n", result.NodeID, result.Error)
    }
}
```

### Parallel Execution

Execute multiple nodes in parallel for improved performance.

```go
// Enable parallel execution
graph.Config.ParallelExecution = true

// Execute multiple nodes in parallel
nodeIDs := []string{"node1", "node2", "node3"}
results, err := graph.ExecuteParallel(context.Background(), nodeIDs, state)
if err != nil {
    log.Fatal(err)
}

// Process results
for nodeID, result := range results {
    fmt.Printf("Node %s: Success=%v, Duration=%v\n", 
        nodeID, result.Success, result.Duration)
}
```

### Graph Introspection

Get information about the graph structure and execution.

```go
// Get topology
topology := graph.GetTopology()
for from, targets := range topology {
    fmt.Printf("Node %s connects to: %v\n", from, targets)
}

// Get execution history
history := graph.GetExecutionHistory()
for _, result := range history {
    fmt.Printf("Executed %s at %v (Duration: %v)\n", 
        result.NodeID, result.Timestamp, result.Duration)
}

// Get current state
currentState := graph.GetCurrentState()
if currentState != nil {
    fmt.Printf("Current state has %d keys\n", len(currentState.Keys()))
}

// Check if running
if graph.IsRunning() {
    fmt.Println("Graph is currently executing")
}
```

### Error Handling

Comprehensive error handling throughout the execution lifecycle.

```go
// Node with error handling
errorNode := func(ctx context.Context, state *BaseState) (*BaseState, error) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic: %v", r)
        }
    }()
    
    // Validate inputs
    if err := validateInputs(state); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Process with error handling
    result, err := processData(state)
    if err != nil {
        return nil, fmt.Errorf("processing failed: %w", err)
    }
    
    state.Set("result", result)
    return state, nil
}

// Execute with error handling
result, err := graph.Execute(context.Background(), initialState)
if err != nil {
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        log.Printf("Execution timed out")
    case errors.Is(err, context.Canceled):
        log.Printf("Execution was canceled")
    default:
        log.Printf("Execution failed: %v", err)
    }
    return
}
```

### Best Practices

#### 1. State Management

```go
// ✅ Good: Use typed access with validation
func safeGetString(state *core.BaseState, key string) (string, error) {
    value, exists := state.Get(key)
    if !exists {
        return "", fmt.Errorf("key %s not found", key)
    }
    
    str, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("key %s is not a string", key)
    }
    
    return str, nil
}

// ❌ Bad: Direct type assertion without checking
func unsafeGetString(state *core.BaseState, key string) string {
    value, _ := state.Get(key)
    return value.(string) // Panic if not string or doesn't exist
}
```

#### 2. Node Design

```go
// ✅ Good: Focused, single-responsibility nodes
func validateInputNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    // Only validate inputs
    return validateInputs(state)
}

func processDataNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    // Only process data
    return processData(state)
}

// ❌ Bad: Monolithic node doing everything
func monolithicNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    // Validate, process, format, save - too many responsibilities
    // ...
}
```

#### 3. Error Handling

```go
// ✅ Good: Descriptive error messages with context
func processNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    input, exists := state.Get("input")
    if !exists {
        return nil, fmt.Errorf("processNode: input is required")
    }
    
    result, err := processInput(input)
    if err != nil {
        return nil, fmt.Errorf("processNode: failed to process input %v: %w", input, err)
    }
    
    state.Set("result", result)
    return state, nil
}
```

#### 4. Context Usage

```go
// ✅ Good: Respect context cancellation
func longRunningNode(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    for i := 0; i < 1000; i++ {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        
        // Do work
        time.Sleep(10 * time.Millisecond)
    }
    
    return state, nil
}
```

## Testing

The core package includes comprehensive tests. Run them with:

```bash
go test ./pkg/core -v
```

### Example Test

```go
func TestGraph_Execute(t *testing.T) {
    graph := core.NewGraph("test_graph")
    
    node1 := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        state.Set("node1_executed", true)
        return state, nil
    }
    
    node2 := func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
        state.Set("node2_executed", true)
        return state, nil
    }
    
    graph.AddNode("node1", "Node 1", node1)
    graph.AddNode("node2", "Node 2", node2)
    graph.AddEdge("node1", "node2", nil)
    graph.SetStartNode("node1")
    graph.AddEndNode("node2")
    
    state := core.NewBaseState()
    result, err := graph.Execute(context.Background(), state)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    val1, _ := result.Get("node1_executed")
    val2, _ := result.Get("node2_executed")
    assert.True(t, val1.(bool))
    assert.True(t, val2.(bool))
}
```

## Performance Considerations

- **State Cloning**: State is cloned at each node execution. For large states, consider using references where appropriate.
- **Concurrent Access**: All state operations are thread-safe, but excessive concurrent access may impact performance.
- **Memory Usage**: Large execution histories can consume significant memory. Consider periodic cleanup.
- **Streaming**: Streaming adds minimal overhead but requires proper channel management.

## Conclusion

The core package provides a solid foundation for building complex workflow systems. Its graph-based execution model, combined with flexible state management and comprehensive error handling, makes it suitable for a wide range of applications from simple data processing pipelines to complex AI agent workflows. 
