# Tools Integration Example

This example demonstrates **advanced tool integration** with GoLangGraph. Learn how to create, integrate, and manage sophisticated tools that extend agent capabilities with real-world functionality.

## 🎯 What You'll Learn

- **Advanced Tool Development**: Create complex, multi-functional tools
- **Tool Orchestration**: Coordinate multiple tools for complex tasks
- **External API Integration**: Connect to real-world services and APIs
- **Tool State Management**: Maintain tool state across interactions
- **Error Handling**: Robust error handling for tool operations
- **Tool Security**: Implement secure tool execution patterns

## 🛠️ Advanced Tool Features

- **File System Operations**: Read, write, and manipulate files
- **Web Scraping**: Extract information from websites
- **API Integrations**: Connect to REST APIs and services
- **Database Operations**: Query and manipulate databases
- **Image Processing**: Basic image manipulation and analysis
- **Code Execution**: Safe code execution in sandboxed environments

## 🚀 Features

- **20+ Production-Ready Tools**: Comprehensive tool library
- **Tool Chaining**: Combine tools for complex workflows
- **Async Tool Execution**: Non-blocking tool operations
- **Tool Validation**: Input validation and sanitization
- **Tool Monitoring**: Performance and usage tracking
- **Tool Documentation**: Auto-generated tool documentation

## 📋 Prerequisites

1. **Ollama Installation**:

   ```bash
   # Install Ollama
   curl -fsSL https://ollama.com/install.sh | sh
   
   # Pull tool-optimized models
   ollama pull gemma3:1b
   ollama pull orieg/gemma3-tools:1b  # Better tool integration
   ```

2. **External Dependencies** (Optional):

   ```bash
   # For image processing
   sudo apt-get install imagemagick
   
   # For web scraping
   go get github.com/gocolly/colly/v2
   
   # For database operations
   go get github.com/lib/pq  # PostgreSQL
   ```

## 🔧 Tool Categories

### 1. File System Tools

```go
// File operations
tools.RegisterTool("file_read", NewFileReadTool())
tools.RegisterTool("file_write", NewFileWriteTool())
tools.RegisterTool("directory_list", NewDirectoryListTool())
tools.RegisterTool("file_search", NewFileSearchTool())
```

### 2. Web Tools

```go
// Web operations
tools.RegisterTool("web_scrape", NewWebScrapeTool())
tools.RegisterTool("http_request", NewHTTPRequestTool())
tools.RegisterTool("url_analyze", NewURLAnalyzeTool())
```

### 3. Data Tools

```go
// Data processing
tools.RegisterTool("json_parse", NewJSONParseTool())
tools.RegisterTool("csv_process", NewCSVProcessTool())
tools.RegisterTool("data_transform", NewDataTransformTool())
```

### 4. System Tools

```go
// System operations
tools.RegisterTool("shell_execute", NewShellExecuteTool())
tools.RegisterTool("process_monitor", NewProcessMonitorTool())
tools.RegisterTool("system_info", NewSystemInfoTool())
```

## 💻 Usage

### Basic Tool Integration

```bash
cd examples/07-tools-integration
go run main.go

# Use tools in conversation
> Can you read the contents of README.md?
> [Agent uses file_read tool to read the file]

> Search for all .go files in the current directory
> [Agent uses file_search tool to find Go files]
```

### Advanced Tool Workflows

```bash
# Complex multi-tool workflow
> Download the content from https://example.com, extract the title, 
  and save it to a file called title.txt

# Agent orchestrates:
# 1. http_request tool to fetch content
# 2. web_scrape tool to extract title
# 3. file_write tool to save result
```

## 📁 Project Structure

```
07-tools-integration/
├── main.go                  # Main application
├── tools/                   # Tool implementations
│   ├── file_tools.go       # File system tools
│   ├── web_tools.go        # Web-related tools
│   ├── data_tools.go       # Data processing tools
│   ├── system_tools.go     # System operation tools
│   ├── api_tools.go        # External API tools
│   └── security_tools.go   # Security utilities
├── config/                  # Tool configurations
│   ├── tool_config.yaml   # Tool settings
│   └── security_config.yaml # Security policies
├── sandbox/                # Sandboxed execution environment
├── cache/                  # Tool result caching
└── README.md              # This file
```

## 🔍 Example Interactions

### File System Operations

```
You: List all Python files in the current directory and show their sizes

🤖 Assistant: I'll search for Python files and get their information.

[Using file_search tool with pattern "*.py"]
[Using file_info tool for each found file]

Found 3 Python files:
1. script.py (2.1 KB) - Modified 2 hours ago
2. utils.py (1.5 KB) - Modified yesterday  
3. main.py (3.2 KB) - Modified 1 week ago

Total: 6.8 KB across 3 files
```

### Web Scraping and Analysis

```
You: Get the latest news headlines from https://news.ycombinator.com

🤖 Assistant: I'll scrape the Hacker News homepage for headlines.

[Using web_scrape tool with CSS selector for headlines]

Latest Hacker News Headlines:
1. "New AI Model Achieves Breakthrough in Code Generation"
2. "Rust 1.75 Released with Performance Improvements" 
3. "The Future of Quantum Computing in 2024"
4. "Open Source Database Reaches 1M Downloads"
5. "WebAssembly Usage Grows 300% Year Over Year"

[Retrieved 30 headlines total - showing top 5]
```

### API Integration

```
You: Check the weather in New York and create a summary report

🤖 Assistant: I'll get the current weather data and create a report.

[Using weather_api tool for New York]
[Using file_write tool to create report]

Weather Report for New York:
- Temperature: 72°F (22°C)
- Condition: Partly Cloudy
- Humidity: 65%
- Wind: 8 mph NW
- UV Index: 6 (High)

Report saved to: weather_report_2024-01-15.txt
```

## ⚙️ Tool Development

### Custom Tool Template

```go
type CustomTool struct {
    name        string
    description string
    config      *ToolConfig
}

func NewCustomTool() *CustomTool {
    return &CustomTool{
        name:        "custom_tool",
        description: "Description of what this tool does",
        config:      DefaultToolConfig(),
    }
}

func (t *CustomTool) GetName() string {
    return t.name
}

func (t *CustomTool) GetDescription() string {
    return t.description
}

func (t *CustomTool) GetDefinition() llm.ToolDefinition {
    return llm.ToolDefinition{
        Type: "function",
        Function: llm.Function{
            Name:        t.name,
            Description: t.description,
            Parameters: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "input": map[string]interface{}{
                        "type":        "string",
                        "description": "Input parameter description",
                    },
                },
                "required": []string{"input"},
            },
        },
    }
}

func (t *CustomTool) Execute(ctx context.Context, input string) (string, error) {
    // Tool implementation
    return "Tool result", nil
}

func (t *CustomTool) Validate(args string) error {
    // Input validation
    return nil
}

func (t *CustomTool) GetConfig() map[string]interface{} {
    return map[string]interface{}{
        "name":        t.name,
        "description": t.description,
    }
}

func (t *CustomTool) SetConfig(config map[string]interface{}) error {
    // Configuration updates
    return nil
}
```

### Tool Registration

```go
// Register tools with the agent
func RegisterAllTools(registry *tools.ToolRegistry) error {
    toolList := []tools.Tool{
        NewFileReadTool(),
        NewFileWriteTool(),
        NewWebScrapeTool(),
        NewHTTPRequestTool(),
        NewJSONParseTool(),
        NewShellExecuteTool(),
        // ... more tools
    }
    
    for _, tool := range toolList {
        if err := registry.RegisterTool(tool); err != nil {
            return fmt.Errorf("failed to register tool %s: %w", 
                tool.GetName(), err)
        }
    }
    
    return nil
}
```

## 🔐 Security Considerations

### Sandboxed Execution

```go
type SandboxConfig struct {
    AllowedPaths    []string      `json:"allowed_paths"`
    AllowedCommands []string      `json:"allowed_commands"`
    TimeoutDuration time.Duration `json:"timeout_duration"`
    MaxMemoryMB     int           `json:"max_memory_mb"`
    NetworkAccess   bool          `json:"network_access"`
}

func (t *ShellExecuteTool) Execute(ctx context.Context, input string) (string, error) {
    // Validate command against whitelist
    if !t.isCommandAllowed(input) {
        return "", fmt.Errorf("command not allowed: %s", input)
    }
    
    // Execute in sandbox
    return t.executeInSandbox(ctx, input)
}
```

### Input Validation

```go
func (t *FileReadTool) Validate(args string) error {
    var params struct {
        FilePath string `json:"file_path"`
    }
    
    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }
    
    // Validate file path
    if !t.isPathAllowed(params.FilePath) {
        return fmt.Errorf("access denied to path: %s", params.FilePath)
    }
    
    // Check for path traversal
    if strings.Contains(params.FilePath, "..") {
        return fmt.Errorf("path traversal not allowed")
    }
    
    return nil
}
```

### Rate Limiting

```go
type RateLimiter struct {
    requests map[string][]time.Time
    maxPerMinute int
    mutex sync.RWMutex
}

func (rl *RateLimiter) Allow(toolName string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    now := time.Now()
    requests := rl.requests[toolName]
    
    // Remove old requests
    cutoff := now.Add(-time.Minute)
    var recent []time.Time
    for _, req := range requests {
        if req.After(cutoff) {
            recent = append(recent, req)
        }
    }
    
    if len(recent) >= rl.maxPerMinute {
        return false
    }
    
    recent = append(recent, now)
    rl.requests[toolName] = recent
    return true
}
```

## 📊 Tool Monitoring

### Performance Metrics

```go
type ToolMetrics struct {
    Name           string        `json:"name"`
    ExecutionCount int           `json:"execution_count"`
    TotalDuration  time.Duration `json:"total_duration"`
    AverageDuration time.Duration `json:"average_duration"`
    ErrorCount     int           `json:"error_count"`
    LastUsed       time.Time     `json:"last_used"`
}

func (tm *ToolMonitor) RecordExecution(toolName string, duration time.Duration, err error) {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    
    metrics := tm.metrics[toolName]
    metrics.ExecutionCount++
    metrics.TotalDuration += duration
    metrics.AverageDuration = metrics.TotalDuration / time.Duration(metrics.ExecutionCount)
    metrics.LastUsed = time.Now()
    
    if err != nil {
        metrics.ErrorCount++
    }
    
    tm.metrics[toolName] = metrics
}
```

### Usage Analytics

```go
type ToolAnalytics struct {
    MostUsedTools    []string          `json:"most_used_tools"`
    ToolCombinations [][]string        `json:"tool_combinations"`
    ErrorPatterns    map[string]int    `json:"error_patterns"`
    PerformanceData  []PerformancePoint `json:"performance_data"`
}

func (ta *ToolAnalytics) GenerateReport() *AnalyticsReport {
    return &AnalyticsReport{
        TopTools:        ta.getTopTools(10),
        CommonWorkflows: ta.getCommonWorkflows(),
        ErrorSummary:    ta.getErrorSummary(),
        Recommendations: ta.getRecommendations(),
    }
}
```

## 🔄 Tool Orchestration

### Workflow Engine

```go
type WorkflowStep struct {
    ToolName string                 `json:"tool_name"`
    Input    map[string]interface{} `json:"input"`
    Output   string                 `json:"output"`
}

type Workflow struct {
    Name  string         `json:"name"`
    Steps []WorkflowStep `json:"steps"`
}

func (we *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowResult, error) {
    result := &WorkflowResult{
        WorkflowName: workflow.Name,
        Steps:        make([]StepResult, 0, len(workflow.Steps)),
    }
    
    for i, step := range workflow.Steps {
        stepResult, err := we.executeStep(ctx, step, result)
        if err != nil {
            return nil, fmt.Errorf("step %d failed: %w", i, err)
        }
        result.Steps = append(result.Steps, *stepResult)
    }
    
    return result, nil
}
```

### Tool Dependencies

```go
type ToolDependency struct {
    ToolName     string   `json:"tool_name"`
    Dependencies []string `json:"dependencies"`
    Optional     []string `json:"optional"`
}

func (td *ToolDependencyManager) ResolveDependencies(toolName string) ([]string, error) {
    visited := make(map[string]bool)
    var result []string
    
    if err := td.resolveDependenciesRecursive(toolName, visited, &result); err != nil {
        return nil, err
    }
    
    return result, nil
}
```

## 🧪 Testing Tools

### Tool Testing Framework

```go
type ToolTest struct {
    Name           string                 `json:"name"`
    Tool           string                 `json:"tool"`
    Input          map[string]interface{} `json:"input"`
    ExpectedOutput string                 `json:"expected_output"`
    ExpectedError  string                 `json:"expected_error"`
}

func (tf *ToolTestFramework) RunTests(tests []ToolTest) *TestResults {
    results := &TestResults{
        Total:  len(tests),
        Passed: 0,
        Failed: 0,
        Tests:  make([]TestResult, 0, len(tests)),
    }
    
    for _, test := range tests {
        result := tf.runSingleTest(test)
        results.Tests = append(results.Tests, result)
        
        if result.Passed {
            results.Passed++
        } else {
            results.Failed++
        }
    }
    
    return results
}
```

### Mock Tools for Testing

```go
type MockTool struct {
    name           string
    responses      map[string]string
    callCount      int
    lastInput      string
}

func NewMockTool(name string, responses map[string]string) *MockTool {
    return &MockTool{
        name:      name,
        responses: responses,
    }
}

func (mt *MockTool) Execute(ctx context.Context, input string) (string, error) {
    mt.callCount++
    mt.lastInput = input
    
    if response, exists := mt.responses[input]; exists {
        return response, nil
    }
    
    return "", fmt.Errorf("no mock response for input: %s", input)
}
```

## 🔗 Integration Examples

### External API Integration

```go
type APITool struct {
    name     string
    baseURL  string
    apiKey   string
    client   *http.Client
}

func (at *APITool) Execute(ctx context.Context, input string) (string, error) {
    var params struct {
        Endpoint string                 `json:"endpoint"`
        Method   string                 `json:"method"`
        Headers  map[string]string      `json:"headers"`
        Body     map[string]interface{} `json:"body"`
    }
    
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", err
    }
    
    return at.makeAPIRequest(ctx, params)
}
```

### Database Integration

```go
type DatabaseTool struct {
    name string
    db   *sql.DB
}

func (dt *DatabaseTool) Execute(ctx context.Context, input string) (string, error) {
    var params struct {
        Query  string        `json:"query"`
        Params []interface{} `json:"params"`
    }
    
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", err
    }
    
    return dt.executeQuery(ctx, params.Query, params.Params...)
}
```

## 📚 Learning Resources

- **Tool Design Patterns**: Best practices for tool development
- **Security**: Secure tool execution and validation
- **Performance**: Optimizing tool performance
- **Integration**: Connecting tools to external services
- **Testing**: Comprehensive tool testing strategies

## 🚀 Next Steps

After mastering tool integration:

1. Explore **08-production-ready** for production tool deployment
2. Build custom tools for your specific use cases
3. Create tool marketplaces and sharing platforms
4. Implement advanced tool orchestration workflows

## 🤝 Contributing

Enhance this example by:

- Adding new tool categories
- Improving security implementations
- Contributing performance optimizations
- Sharing integration patterns

---

**Happy Tool Building!** 🔧

This tools integration example provides a comprehensive foundation for building sophisticated tool ecosystems with GoLangGraph that extend agent capabilities with real-world functionality.
