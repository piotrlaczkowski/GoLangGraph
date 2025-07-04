# 👥 Multi-Agent System Example

This example demonstrates how to create and coordinate multiple AI agents working together using GoLangGraph and Ollama. Multiple agents can collaborate, specialize in different tasks, and share information to solve complex problems.

## 📋 Prerequisites

1. **Ollama installed and running**:
   ```bash
   # Install Ollama (if not already installed)
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Start Ollama service
   ollama serve
   ```

2. **Pull the required models**:
   ```bash
   # Pull the tool-enabled model for better performance
   ollama pull orieg/gemma3-tools:1b
   
   # Or use the standard model
   ollama pull gemma3:1b
   ```

## 🚀 Running the Example

```bash
# From the project root
cd examples/03-multi-agent
go run main.go
```

## 🎯 What This Example Demonstrates

- ✅ **Multiple Agent Creation** - Different agents with specialized roles
- ✅ **Agent Coordination** - Agents working together on tasks
- ✅ **Task Distribution** - Automatic task assignment based on agent capabilities
- ✅ **Information Sharing** - Agents sharing results and context
- ✅ **Workflow Orchestration** - Complex multi-step workflows
- ✅ **Parallel Processing** - Concurrent agent execution

## 🔧 Key Features

### Agent Roles
The system includes several specialized agents:

- 🧮 **Analyst Agent**: Data analysis and mathematical computations
- 🔍 **Researcher Agent**: Information gathering and web research
- 📝 **Writer Agent**: Content creation and documentation
- 🛠️ **Coordinator Agent**: Task orchestration and workflow management
- 🔧 **Specialist Agent**: Domain-specific expertise (configurable)

### Collaboration Patterns

1. **Sequential Workflow**: Agents work in sequence, passing results
2. **Parallel Processing**: Multiple agents work simultaneously
3. **Hierarchical Structure**: Coordinator manages sub-agents
4. **Peer-to-Peer**: Agents communicate directly with each other

### Example Interactions

```
User: Analyze the sales data from Q1, research market trends, and write a comprehensive report.

🎭 Multi-Agent System:

📋 Coordinator Agent:
Task received: Complex analysis and reporting
Breaking down into subtasks:
1. Data analysis → Analyst Agent
2. Market research → Researcher Agent  
3. Report writing → Writer Agent

🧮 Analyst Agent:
Analyzing Q1 sales data...
- Total sales: $2.5M (+15% vs Q4)
- Top product: Widget Pro (35% of sales)
- Growth trend: Consistent 3% monthly increase

🔍 Researcher Agent:
Researching market trends...
- Industry growth rate: 12% annually
- Competitor analysis: Market share stable
- Emerging trends: AI integration, sustainability focus

📝 Writer Agent:
Generating comprehensive report...
Incorporating analysis from Analyst and research from Researcher...

Final Report: "Q1 Sales Analysis and Market Outlook"
[Comprehensive report with data, trends, and recommendations]

⏱️  Total time: 8.5s | Agents used: 4 | Tasks completed: 3
```

## 📊 Expected Output

```
👥 GoLangGraph Multi-Agent System Example
==========================================

✅ Ollama provider initialized
✅ Agent pool created with 5 specialized agents:
   🧮 Analyst Agent (data analysis, calculations)
   🔍 Researcher Agent (information gathering)
   📝 Writer Agent (content creation)
   🛠️ Coordinator Agent (task orchestration)
   🔧 Specialist Agent (domain expertise)

✅ Multi-agent system ready for collaboration

💼 Multi-Agent Session Started (type '/quit' to exit)
─────────────────────────────────────────────────────

You: Calculate the ROI for our marketing campaigns and create a summary report.

🎭 Multi-Agent Workflow:

📋 Coordinator Agent:
Analyzing task: "Calculate ROI and create summary report"
Task breakdown:
1. ROI calculations → Analyst Agent
2. Report generation → Writer Agent

🧮 Analyst Agent (Task 1/2):
Calculating marketing campaign ROI...
- Campaign A: 250% ROI ($50k spent, $125k return)
- Campaign B: 180% ROI ($30k spent, $54k return)  
- Campaign C: 320% ROI ($20k spent, $64k return)
- Overall ROI: 243% ($100k spent, $243k return)

📝 Writer Agent (Task 2/2):
Creating summary report based on analysis...

📄 Marketing Campaign ROI Summary Report

Executive Summary:
Our Q1 marketing campaigns delivered exceptional results with an overall ROI of 243%.

Key Findings:
• Campaign C achieved the highest ROI at 320%
• All campaigns exceeded the 150% ROI target
• Total investment of $100k generated $243k in returns

Recommendations:
• Increase budget allocation to Campaign C model
• Analyze Campaign C's success factors for replication
• Consider expanding similar high-ROI initiatives

⏱️  Workflow completed in 6.2s | Agents: 3 | Success rate: 100%

You: /quit
```

## 🛠️ Customization Options

### 1. Configure Agent Roles
```go
// Create specialized agents
analystAgent := CreateAnalystAgent("DataExpert", analysisTools)
researcherAgent := CreateResearcherAgent("InfoGatherer", researchTools)
writerAgent := CreateWriterAgent("ContentCreator", writingTools)

// Add to agent pool
agentPool.AddAgent(analystAgent)
agentPool.AddAgent(researcherAgent)
agentPool.AddAgent(writerAgent)
```

### 2. Define Workflows
```go
// Sequential workflow
workflow := NewSequentialWorkflow()
workflow.AddStep("analysis", "analyst")
workflow.AddStep("research", "researcher")
workflow.AddStep("writing", "writer")

// Parallel workflow
parallelWorkflow := NewParallelWorkflow()
parallelWorkflow.AddParallelTasks([]string{"analysis", "research"})
parallelWorkflow.AddStep("synthesis", "writer")
```

### 3. Custom Communication Patterns
```go
// Direct agent communication
agentPool.EnableDirectCommunication(true)

// Shared memory/context
agentPool.SetSharedContext(context.WithValue(ctx, "shared_data", data))

// Message passing
agentPool.SetMessageBroker(NewMessageBroker())
```

## 🔍 Code Structure

```
03-multi-agent/
├── README.md              # This documentation
├── main.go               # Main example code
├── agents.go             # Agent implementations
├── coordinator.go        # Workflow coordination
├── communication.go      # Agent communication
├── workflows.go          # Workflow definitions
└── examples.go           # Example scenarios
```

## 🎓 Learning Objectives

After running this example, you'll understand:

1. **Multi-Agent Architecture** - How to design agent systems
2. **Task Decomposition** - Breaking complex tasks into subtasks
3. **Agent Coordination** - Managing multiple agents effectively
4. **Workflow Orchestration** - Designing agent workflows
5. **Communication Patterns** - How agents share information
6. **Performance Optimization** - Parallel vs sequential execution

## 🧪 Example Scenarios

The example includes several pre-built multi-agent scenarios:

### Business Analysis Workflow
```
"Analyze our Q3 financial data, research industry benchmarks, 
and create a board presentation with recommendations."
```

### Content Creation Pipeline
```
"Research the latest AI trends, write a technical blog post, 
and create social media summaries for different platforms."
```

### Customer Support Automation
```
"Analyze customer feedback, identify common issues, 
and draft response templates for the support team."
```

### Product Development Workflow
```
"Research competitor features, analyze user feedback, 
and create a product roadmap with priority rankings."
```

## 🔗 Next Steps

Once you're comfortable with multi-agent systems, try:

- **[04-rag-system](../04-rag-system/)** - Retrieval-Augmented Generation
- **[05-streaming](../05-streaming/)** - Real-time streaming responses
- **[06-persistence](../06-persistence/)** - Data persistence and memory

## 🐛 Troubleshooting

### Common Issues

1. **Agent communication failures**:
   ```go
   // Enable debug logging
   agentPool.SetLogLevel(logrus.DebugLevel)
   
   // Check agent status
   status := agentPool.GetAgentStatus()
   ```

2. **Workflow deadlocks**:
   ```go
   // Set timeouts for workflows
   workflow.SetTimeout(60 * time.Second)
   
   // Enable workflow monitoring
   workflow.EnableMonitoring(true)
   ```

3. **Resource contention**:
   ```go
   // Limit concurrent agents
   agentPool.SetMaxConcurrentAgents(3)
   
   // Use resource pooling
   agentPool.EnableResourcePooling(true)
   ```

### Performance Tips

- Use parallel workflows for independent tasks
- Implement proper error handling and retries
- Monitor agent performance and resource usage
- Use specialized agents for specific domains
- Implement efficient communication patterns

## 📚 Additional Resources

- [Multi-Agent Systems Theory](https://en.wikipedia.org/wiki/Multi-agent_system)
- [GoLangGraph Agent Documentation](../../docs/agents.md)
- [Workflow Patterns](../../docs/workflows.md)
- [Agent Communication Protocols](../../docs/communication.md) 