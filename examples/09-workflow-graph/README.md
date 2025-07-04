# 🔄 Complex Workflow Graph with ReAct Agent

This example demonstrates advanced workflow orchestration using a graph-based architecture with nodes, edges, conditional routing, and ReAct (Reasoning and Acting) agent integration - inspired by modern AI agent frameworks like LangGraph.

## 🎯 What You'll Learn

- **Graph-Based Workflows**: Multi-node workflows with conditional edges
- **ReAct Agent Pattern**: Reasoning and Acting with tool integration
- **State Management**: Data flow and state tracking across nodes
- **Dynamic Routing**: Conditional workflow paths based on analysis
- **Tool Integration**: Advanced tool usage within workflow context
- **Result Aggregation**: Combining outputs from parallel execution paths

## 🏗️ Architecture

### Workflow Graph Structure

```
Input → Analysis → ReAct → Decision
                              ├─→ Math Task ──┐
                              ├─→ Research Task ─→ Aggregation → Output
                              └─→ Analysis Task ─┘
```

### Node Types

1. **Input Node**: Entry point, initializes workflow state
2. **Analysis Node**: Classifies task type and complexity
3. **ReAct Node**: Implements reasoning and planning with tools
4. **Decision Node**: Routes to appropriate task execution path
5. **Task Execution Nodes**: Specialized processing (math/research/analysis)
6. **Aggregation Node**: Combines and synthesizes results
7. **Output Node**: Formats final response

### Edge Types

- **Sequential Edges**: Direct node-to-node connections
- **Conditional Edges**: Route based on state conditions
- **Parallel Edges**: Multiple execution paths from decision points

## 🤖 ReAct Agent Integration

The ReAct agent implements the **Reasoning and Acting** pattern:

### Reasoning Phase

- Analyzes the task and context
- Creates step-by-step plans
- Identifies required tools and capabilities

### Acting Phase

- Executes planned actions
- Uses available tools
- Adapts based on intermediate results

### Available Tools

- **Calculator**: Mathematical operations and computations
- **Web Search**: Information retrieval (simulated)
- **Data Analysis**: Statistical analysis and insights
- **Planner**: Task planning and strategy creation

## 🚀 Running the Example

### Prerequisites

1. **Ollama Installation**: Download from [ollama.com](https://ollama.com)
2. **Start Ollama**: `ollama serve`
3. **Pull Model**: `ollama pull gemma3:1b`

### Execution

```bash
cd examples/09-workflow-graph
go run main.go
```

### Example Tasks

Try these tasks to see different workflow paths:

**Mathematical Tasks** (routes to math execution):

```
Calculate the compound interest on $1000 at 5% for 3 years
Solve the quadratic equation x² + 5x + 6 = 0
What is the derivative of x³ + 2x² - 5x + 1?
```

**Research Tasks** (routes to research execution):

```
Research the latest developments in quantum computing
What are the current trends in artificial intelligence?
Explain the benefits of renewable energy sources
```

**Analysis Tasks** (routes to analysis execution):

```
Analyze the pros and cons of remote work
Compare different machine learning algorithms
Evaluate the impact of social media on society
```

## 📊 Workflow Features

### State Management

- **Persistent State**: Data flows through all nodes
- **Context Tracking**: Maintains execution context
- **History Recording**: Tracks all node executions
- **Metadata**: Additional workflow information

### Dynamic Routing

- **Task Classification**: Automatic task type detection
- **Conditional Edges**: Route based on analysis results
- **Parallel Execution**: Multiple specialized processing paths
- **Result Synthesis**: Intelligent aggregation of outputs

### Monitoring & Debugging

- **Execution Tracking**: Real-time workflow progress
- **Performance Metrics**: Node execution times
- **Error Handling**: Graceful failure management
- **State Inspection**: View workflow state at any point

## 🛠️ Interactive Commands

- `/graph` - Show detailed workflow graph structure
- `/state` - Display current workflow state
- `/history` - View execution history
- `/reset` - Reset workflow state
- `/help` - Show comprehensive help

## 🔧 Advanced Features

### Conditional Edge Logic

```go
{
    From: "decision",
    To: "task_math",
    Label: "mathematical_task",
    Condition: func(state *WorkflowState) bool {
        taskType, exists := state.Context["task_type"].(string)
        return exists && strings.Contains(strings.ToLower(taskType), "math")
    },
}
```

### State Transformation

```go
type WorkflowState struct {
    ID          string                 `json:"id"`
    Input       string                 `json:"input"`
    CurrentNode string                 `json:"current_node"`
    Context     map[string]interface{} `json:"context"`
    History     []NodeExecution        `json:"history"`
    Result      string                 `json:"result"`
    Metadata    map[string]string      `json:"metadata"`
}
```

### ReAct Implementation

```go
func (agent *ReActAgent) Execute(state *WorkflowState) (*WorkflowState, error) {
    // 1. Analyze the task
    // 2. Create reasoning plan
    // 3. Identify required tools
    // 4. Execute action sequence
    // 5. Update state with results
}
```

## 📈 Performance Characteristics

- **Average Execution Time**: 5-15 seconds (depends on task complexity)
- **Memory Usage**: ~200-400MB (includes full state tracking)
- **Scalability**: Supports complex multi-step workflows
- **Reliability**: Built-in error handling and recovery

## 🎓 Learning Outcomes

After running this example, you'll understand:

1. **Graph-Based Architecture**: How to design workflows as directed graphs
2. **ReAct Pattern**: Implementing reasoning and acting in AI agents
3. **State Management**: Managing data flow in complex workflows
4. **Conditional Routing**: Dynamic workflow paths based on conditions
5. **Tool Integration**: Using tools within workflow contexts
6. **Result Synthesis**: Combining outputs from multiple execution paths

## 🔄 Workflow Execution Flow

1. **Input Processing**: Task received and initial state created
2. **Analysis Phase**: Task classification and complexity assessment
3. **ReAct Planning**: Reasoning about approach and tool requirements
4. **Decision Routing**: Conditional routing to specialized execution paths
5. **Task Execution**: Specialized processing based on task type
6. **Result Aggregation**: Synthesis of results from execution paths
7. **Output Formatting**: Final response preparation and delivery

## 🌟 Key Concepts Demonstrated

### Graph Theory in AI Workflows

- **Nodes**: Processing units with specific capabilities
- **Edges**: Connections defining possible transitions
- **State**: Data that flows through the graph
- **Routing**: Dynamic path selection based on conditions

### ReAct Agent Pattern

- **Observation**: Understanding current state and context
- **Thought**: Reasoning about next actions
- **Action**: Executing planned steps with tools
- **Reflection**: Evaluating results and planning next steps

### Advanced Workflow Patterns

- **Conditional Branching**: Different paths based on analysis
- **Parallel Processing**: Multiple simultaneous execution paths
- **State Aggregation**: Combining results from parallel paths
- **Error Recovery**: Handling failures gracefully

## 🚀 Next Steps

This example provides a foundation for building sophisticated AI agent workflows. You can extend it by:

1. **Adding More Node Types**: Create specialized processing nodes
2. **Enhanced Tool Integration**: Implement real tool connections
3. **Complex Routing Logic**: More sophisticated conditional edges
4. **Persistence Layer**: Save and restore workflow states
5. **Distributed Execution**: Scale across multiple instances
6. **Visual Workflow Designer**: GUI for workflow creation

This represents the cutting edge of AI agent architecture, combining the power of graph-based workflows with intelligent reasoning and tool usage!
