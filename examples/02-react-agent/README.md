# 🧠 ReAct Agent Example

This example demonstrates the ReAct (Reasoning and Acting) pattern with GoLangGraph and Ollama. ReAct agents can reason about problems and take actions using tools to solve complex tasks.

## 📋 Prerequisites

1. **Ollama installed and running**:
   ```bash
   # Install Ollama (if not already installed)
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Start Ollama service
   ollama serve
   ```

2. **Pull the tool-enabled Gemma3 model**:
   ```bash
   # Pull the tool-enabled model for better performance
   ollama pull orieg/gemma3-tools:1b
   
   # Or use the standard model
   ollama pull gemma3:1b
   ```

## 🚀 Running the Example

```bash
# From the project root
cd examples/02-react-agent
go run main.go
```

## 🎯 What This Example Demonstrates

- ✅ **ReAct Pattern Implementation** - Reasoning and Acting cycle
- ✅ **Tool Integration** - Calculator, web search, file operations
- ✅ **Multi-step Problem Solving** - Complex task decomposition
- ✅ **Iterative Reasoning** - Step-by-step problem solving
- ✅ **Error Handling** - Graceful tool failure handling
- ✅ **Performance Monitoring** - Execution tracking and metrics

## 🔧 Key Features

### ReAct Pattern
The ReAct pattern follows this cycle:
1. **Thought**: Reason about the problem
2. **Action**: Choose a tool to use
3. **Observation**: Observe the result
4. **Repeat**: Continue until the task is complete

### Available Tools
- 🧮 **Calculator**: Mathematical operations
- 🔍 **Web Search**: Information retrieval (simulated)
- 📁 **File Operations**: Read and write files
- ⏰ **Time**: Current time and date operations
- 🌐 **HTTP**: Web requests and API calls

### Example Interactions

```
User: What is the square root of 144, and what year was it when this number was first discovered?

🧠 ReAct Agent:
Thought: I need to find the square root of 144 first, then research when this concept was discovered.

Action: calculator
Action Input: sqrt(144)

Observation: The square root of 144 is 12.

Thought: Now I need to research when the concept of square roots was first discovered.

Action: web_search
Action Input: history of square root discovery mathematics

Observation: Square roots were known to ancient civilizations, with evidence from Babylonian mathematics around 1800-1600 BCE.

Final Answer: The square root of 144 is 12. The concept of square roots was first discovered by ancient Babylonians around 1800-1600 BCE.
```

## 📊 Expected Output

```
🧠 GoLangGraph ReAct Agent Example
==================================

✅ Ollama provider initialized with tool-enabled model
✅ Tools registered: calculator, web_search, file_read, file_write, time
✅ ReAct agent created: SmartAgent
✅ Agent ready for reasoning and action

💭 ReAct Session Started (type '/quit' to exit)
───────────────────────────────────────────────

You: Calculate the area of a circle with radius 5, then tell me what that area represents in square meters.

🧠 SmartAgent:
Thought: I need to calculate the area of a circle with radius 5. The formula is π × r².

Action: calculator
Action Input: pi * 5^2

Observation: π × 5² = 78.54 square units

Thought: The area is 78.54 square units. The user asked what this represents in square meters, so I should explain the practical meaning.

Final Answer: The area of a circle with radius 5 is approximately 78.54 square units. If the radius is measured in meters, this represents 78.54 square meters, which is roughly the size of a small apartment or a large classroom.

⏱️  Response time: 3.2s | Iterations: 2 | Tools used: 1

You: /quit
```

## 🛠️ Customization Options

### 1. Configure Available Tools
```go
// Add custom tools
toolRegistry.RegisterTool(tools.NewEmailTool())
toolRegistry.RegisterTool(tools.NewDatabaseTool())

// Configure agent with specific tools
agentConfig.Tools = []string{"calculator", "email", "database"}
```

### 2. Adjust Reasoning Parameters
```go
agentConfig.MaxIterations = 5    // Maximum reasoning steps
agentConfig.Temperature = 0.1    // More focused reasoning
agentConfig.MaxTokens = 1000     // Longer reasoning chains
```

### 3. Custom System Prompt
```go
agentConfig.SystemPrompt = `You are an expert problem solver. 
Use the ReAct pattern: Think, Act, Observe, and repeat until you have a complete solution.
Be thorough in your reasoning and always explain your thought process.`
```

## 🔍 Code Structure

```
02-react-agent/
├── README.md          # This documentation
├── main.go           # Main example code
├── tools.go          # Tool implementations
├── config.go         # Configuration helpers
└── examples.go       # Example scenarios
```

## 🎓 Learning Objectives

After running this example, you'll understand:

1. **ReAct Pattern Implementation** - How reasoning and acting work together
2. **Tool Integration** - How agents use external tools
3. **Multi-step Problem Solving** - Breaking complex tasks into steps
4. **Error Handling** - Managing tool failures gracefully
5. **Performance Optimization** - Monitoring and improving agent performance

## 🧪 Example Scenarios

The example includes several pre-built scenarios:

### Mathematical Problem Solving
```
"Calculate the compound interest on $1000 invested at 5% annual rate for 3 years, 
then convert the result to euros using current exchange rates."
```

### Research and Analysis
```
"Research the population of Tokyo and calculate how many football fields would be 
needed to house everyone if each person needs 10 square meters."
```

### File Operations
```
"Read the data from data.csv, calculate the average of the numbers, 
and write a summary report to results.txt."
```

### Multi-step Reasoning
```
"Plan a trip itinerary: find flights from NYC to London, check the weather forecast, 
and calculate the total cost including accommodation."
```

## 🔗 Next Steps

Once you're comfortable with ReAct agents, try:

- **[03-multi-agent](../03-multi-agent/)** - Multiple agents working together
- **[04-rag-system](../04-rag-system/)** - Retrieval-Augmented Generation
- **[05-streaming](../05-streaming/)** - Real-time streaming responses

## 🐛 Troubleshooting

### Common Issues

1. **Tool not found errors**:
   ```bash
   # Make sure tools are properly registered
   toolRegistry.RegisterTool(tools.NewCalculatorTool())
   ```

2. **Infinite reasoning loops**:
   ```go
   // Set reasonable max iterations
   agentConfig.MaxIterations = 10
   ```

3. **Tool execution failures**:
   - Check tool permissions (file operations)
   - Verify network connectivity (web search)
   - Validate input formats

### Performance Tips

- Use `orieg/gemma3-tools:1b` for better tool integration
- Set appropriate `MaxIterations` to prevent infinite loops
- Monitor tool execution times
- Use focused system prompts for specific domains

## 📚 Additional Resources

- [ReAct Paper](https://arxiv.org/abs/2210.03629) - Original research paper
- [GoLangGraph Tools Documentation](../../docs/tools.md)
- [Ollama Tools Integration](https://ollama.com/orieg/gemma3-tools) 