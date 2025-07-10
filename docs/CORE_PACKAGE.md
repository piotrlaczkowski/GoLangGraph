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

// Configure the graph (optional)
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

The `BaseState` struct provides thread-safe state management.

```go
type BaseState struct {
    data     map[string]interface{}
    metadata map[string]interface{}
    mu       sync.RWMutex
    // ... internal fields
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

#### Error Handling Node

```go
errorHandlingNode := func(ctx context.Context, state *BaseState) (*BaseState, error) {
    input, exists := state.Get("input")
    if !exists {
        return nil, fmt.Errorf("required input not found")
    }
    
    // Validate input
    if input == "" {
        state.Set("error", "empty input")
        return state, nil
    }
    
    // Process
    state.Set("processed", input)
    return state, nil
}
```

#### Async Node with Context

```go
asyncNode := func(ctx context.Context, state *BaseState) (*core.BaseState, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Simulate async work
    timer := time.NewTimer(2 * time.Second)
    defer timer.Stop()
    
    select {
    case <-timer.C:
        state.Set("async_result", "completed")
        return state, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Edge Conditions

Edge conditions determine the flow of execution based on the current state.

```go
type EdgeCondition func(ctx context.Context, state *BaseState) (string, error)
```

#### Simple Condition

```go
condition := func(ctx context.Context, state *BaseState) (string, error) {
    value, _ := state.Get("score")
    score := value.(int)
    
    if score > 80 {
        return "success_node", nil
    } else if score > 50 {
        return "retry_node", nil
    }
    return "failure_node", nil
}
```

#### Complex Condition

```go
complexCondition := func(ctx context.Context, state *BaseState) (string, error) {
    // Multiple criteria
    status, _ := state.Get("status")
    attempts, _ := state.Get("attempts")
    
    if status == "error" && attempts.(int) < 3 {
        return "retry_node", nil
    } else if status == "success" {
        return "success_node", nil
    }
    return "failure_node", nil
}
```

## Agent Integration

The core package integrates seamlessly with the agent package:

```go
// Create an agent that uses a custom graph
config := &agent.AgentConfig{
    Name: "custom-agent",
    Type: agent.AgentTypeReAct,
    // ... other config
}

// Create agent
agentInstance := agent.NewAgent(config, llmManager, toolRegistry)

// Get the agent's graph for customization
graph := agentInstance.GetGraph()

// Add custom nodes
graph.AddNode("custom", "Custom Processing", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    // Custom logic here
    return state, nil
})
```

## Persistence Integration

State can be persisted using the persistence package:

```go
// Create checkpointer
checkpointer := persistence.NewMemoryCheckpointer()

// Save state as checkpoint
checkpoint := &persistence.Checkpoint{
    ID:       "checkpoint_1",
    ThreadID: "thread_1",
    State:    state,
    Metadata: map[string]interface{}{
        "step": "processing",
    },
    CreatedAt: time.Now(),
}

err := checkpointer.Save(context.Background(), checkpoint)
if err != nil {
    log.Printf("Failed to save checkpoint: %v", err)
}

// Load checkpoint
loaded, err := checkpointer.Load(context.Background(), "thread_1", "checkpoint_1")
if err != nil {
    log.Printf("Failed to load checkpoint: %v", err)
}
```

## Best Practices

### 1. State Management

- Keep state data simple and serializable
- Use metadata for non-critical information
- Clone state when passing between functions
- Validate state data in node functions

### 2. Error Handling

- Always handle errors in node functions
- Use context for cancellation and timeouts
- Set error information in state when appropriate
- Fail fast for critical errors

### 3. Graph Design

- Keep graphs simple and focused
- Use descriptive node and edge names
- Design for reusability
- Test graph execution paths

### 4. Performance

- Minimize state data size
- Use efficient data structures
- Implement proper timeouts
- Monitor execution metrics

## Examples

### Simple Linear Workflow

```go
graph := core.NewGraph("linear_workflow")

// Add nodes
graph.AddNode("input", "Input", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    state.Set("data", "initial_data")
    return state, nil
})

graph.AddNode("process", "Process", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    data, _ := state.Get("data")
    processed := strings.ToUpper(data.(string))
    state.Set("processed", processed)
    return state, nil
})

graph.AddNode("output", "Output", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    processed, _ := state.Get("processed")
    fmt.Printf("Final result: %s\n", processed)
    return state, nil
})

// Connect nodes
graph.AddEdge("input", "process", nil)
graph.AddEdge("process", "output", nil)

// Configure
graph.SetStartNode("input")
graph.AddEndNode("output")

// Execute
result, err := graph.Execute(context.Background(), core.NewBaseState())
```

### Conditional Workflow

```go
graph := core.NewGraph("conditional_workflow")

// Decision node
graph.AddNode("decision", "Decision", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    // Simulate decision logic
    state.Set("choice", "path_a")
    return state, nil
})

// Path A
graph.AddNode("path_a", "Path A", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    state.Set("result", "Took path A")
    return state, nil
})

// Path B
graph.AddNode("path_b", "Path B", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
    state.Set("result", "Took path B")
    return state, nil
})

// Conditional edges
choiceCondition := func(ctx context.Context, state *core.BaseState) (string, error) {
    choice, _ := state.Get("choice")
    return choice.(string), nil
}

graph.AddEdge("decision", "path_a", func(ctx context.Context, state *core.BaseState) (string, error) {
    choice, _ := state.Get("choice")
    if choice == "path_a" {
        return "path_a", nil
    }
    return "", nil
})

graph.AddEdge("decision", "path_b", func(ctx context.Context, state *core.BaseState) (string, error) {
    choice, _ := state.Get("choice")
    if choice == "path_b" {
        return "path_b", nil
    }
    return "", nil
})

// Configure
graph.SetStartNode("decision")
graph.AddEndNode("path_a")
graph.AddEndNode("path_b")
```

This core package provides the foundation for building complex AI workflows with GoLangGraph, offering flexibility, performance, and reliability for production use. 
