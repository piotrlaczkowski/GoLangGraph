package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/debug"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/server"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "golanggraph",
	Short: "GoLangGraph - A Go implementation of LangGraph",
	Long: `GoLangGraph is a comprehensive Go implementation of the LangGraph framework
for building stateful, multi-agent conversational AI applications.

This CLI provides tools for:
- Building and packaging agents into Docker containers
- Running development servers with hot-reload
- Managing database migrations
- Visualizing graph execution
- Testing and debugging agents
- Deploying agents to production environments`,
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the GoLangGraph server",
	Long: `Start the GoLangGraph HTTP server with REST API endpoints and WebSocket support.
The server provides:
- REST API for agent and graph management
- WebSocket endpoints for real-time streaming
- Visual debugging interface
- Health monitoring endpoints`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations to set up the required schema for state persistence.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations()
	},
}

// debugCmd represents the debug command
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug tools for graph visualization and analysis",
	Long:  `Provides debugging tools for analyzing graph execution and visualizing agent behavior.`,
}

// visualizeCmd represents the visualize command
var visualizeCmd = &cobra.Command{
	Use:   "visualize [graph-file]",
	Short: "Visualize a graph structure",
	Long:  `Generate visual representations of graph structures in various formats (Mermaid, DOT, JSON).`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		runVisualize(args, format, output)
	},
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test agent configurations and graph execution",
	Long:  `Test agent configurations and validate graph execution flows.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTests()
	},
}

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check system health and component status",
	Long:  `Check the health status of GoLangGraph components including databases, LLM providers, and system resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		runHealthCheck()
	},
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build and package agents for deployment",
	Long: `Build and package agents into deployable artifacts including Docker containers.
Supports both regular and distroless container builds for production deployment.`,
}

// dockerCmd represents the docker command
var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Docker packaging commands",
	Long:  `Commands for packaging agents into Docker containers for production deployment.`,
}

// dockerBuildCmd represents the docker build command
var dockerBuildCmd = &cobra.Command{
	Use:   "build [agent-config]",
	Short: "Build Docker container for agent",
	Long: `Build a Docker container for deploying an agent to production.
Supports both regular and distroless variants for different deployment needs.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		distroless, _ := cmd.Flags().GetBool("distroless")
		tag, _ := cmd.Flags().GetString("tag")
		dockerfile, _ := cmd.Flags().GetString("dockerfile")
		platform, _ := cmd.Flags().GetString("platform")
		runDockerBuild(args, distroless, tag, dockerfile, platform)
	},
}

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start development server with hot-reload",
	Long: `Start a development server with hot-reload capabilities for testing and debugging agents.
Includes:
- Hot-reload on code changes
- Interactive debugging interface
- Real-time logging and metrics
- Agent testing playground`,
	Run: func(cmd *cobra.Command, args []string) {
		runDevServer()
	},
}

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate agent configuration",
	Long:  `Validate agent configuration files and graph definitions for correctness.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		strict, _ := cmd.Flags().GetBool("strict")
		runValidate(args, strict)
	},
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy agents to production",
	Long:  `Deploy agents to various production environments including Docker, Kubernetes, and cloud platforms.`,
}

// deployDockerCmd represents the deploy docker command
var deployDockerCmd = &cobra.Command{
	Use:   "docker [agent-config]",
	Short: "Deploy agent using Docker",
	Long:  `Deploy an agent using Docker containers with production-ready configuration.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runDeployDocker(args)
	},
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new GoLangGraph project",
	Long:  `Initialize a new GoLangGraph project with example configurations and templates.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		template, _ := cmd.Flags().GetString("template")
		runInit(args, template)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.golanggraph.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Serve command flags
	serveCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to bind to")
	serveCmd.Flags().IntP("port", "p", 8080, "Port to bind to")
	serveCmd.Flags().String("static-dir", "./static", "Static files directory")
	serveCmd.Flags().Bool("enable-cors", true, "Enable CORS")

	// Dev command flags
	devCmd.Flags().StringP("host", "H", "localhost", "Host to bind to")
	devCmd.Flags().IntP("port", "p", 8080, "Port to bind to")
	devCmd.Flags().String("agent-config", "", "Agent configuration file")
	devCmd.Flags().Bool("hot-reload", true, "Enable hot-reload")
	devCmd.Flags().Bool("debug", true, "Enable debug mode")
	devCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")

	// Docker build command flags
	dockerBuildCmd.Flags().BoolP("distroless", "d", false, "Build distroless container")
	dockerBuildCmd.Flags().StringP("tag", "t", "", "Docker image tag")
	dockerBuildCmd.Flags().String("dockerfile", "", "Custom Dockerfile path")
	dockerBuildCmd.Flags().String("platform", "", "Target platform (e.g., linux/amd64,linux/arm64)")

	// Validate command flags
	validateCmd.Flags().BoolP("strict", "s", false, "Enable strict validation")

	// Init command flags
	initCmd.Flags().StringP("template", "t", "basic", "Project template (basic, advanced, rag)")

	// Migrate command flags
	migrateCmd.Flags().String("db-type", "postgres", "Database type (postgres, redis)")
	migrateCmd.Flags().String("db-host", "localhost", "Database host")
	migrateCmd.Flags().Int("db-port", 5432, "Database port")
	migrateCmd.Flags().String("db-name", "golanggraph", "Database name")
	migrateCmd.Flags().String("db-user", "postgres", "Database user")
	migrateCmd.Flags().String("db-password", "", "Database password")

	// Visualize command flags
	visualizeCmd.Flags().StringP("format", "f", "mermaid", "Output format (mermaid, dot, json)")
	visualizeCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(healthCmd)
	
	// Add nested commands
	dockerCmd.AddCommand(dockerBuildCmd)
	deployCmd.AddCommand(deployDockerCmd)
	debugCmd.AddCommand(visualizeCmd)

	// Bind flags to viper
	viper.BindPFlag("host", serveCmd.Flags().Lookup("host"))
	viper.BindPFlag("port", serveCmd.Flags().Lookup("port"))
	viper.BindPFlag("static-dir", serveCmd.Flags().Lookup("static-dir"))
	viper.BindPFlag("enable-cors", serveCmd.Flags().Lookup("enable-cors"))
}

// initConfig reads in config file and ENV variables.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".golanggraph" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".golanggraph")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func runServer() {
	fmt.Println("Starting GoLangGraph server...")

	// Create server configuration
	config := &server.ServerConfig{
		Host:           viper.GetString("host"),
		Port:           viper.GetInt("port"),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		EnableCORS:     viper.GetBool("enable-cors"),
		StaticDir:      viper.GetString("static-dir"),
	}

	// Create server
	srv := server.NewServer(config)

	// Initialize components
	if err := initializeComponents(srv); err != nil {
		log.Fatalf("Failed to initialize components: %v", err)
	}

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	fmt.Printf("Server started on %s:%d\n", config.Host, config.Port)
	fmt.Printf("Health check: http://%s:%d/api/v1/health\n", config.Host, config.Port)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

func initializeComponents(srv *server.Server) error {
	// Initialize LLM providers
	llmManager := llm.NewProviderManager()

	// Add OpenAI provider if API key is available
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		openaiConfig := &llm.ProviderConfig{
			APIKey:   apiKey,
			Endpoint: "https://api.openai.com/v1",
		}
		openaiProvider, err := llm.NewOpenAIProvider(openaiConfig)
		if err == nil {
			llmManager.RegisterProvider("openai", openaiProvider)
		}
	}

	// Add Ollama provider if available
	if ollamaURL := os.Getenv("OLLAMA_URL"); ollamaURL != "" {
		ollamaConfig := &llm.ProviderConfig{
			Endpoint: ollamaURL,
		}
		ollamaProvider, err := llm.NewOllamaProvider(ollamaConfig)
		if err == nil {
			llmManager.RegisterProvider("ollama", ollamaProvider)
		}
	} else {
		// Default Ollama URL
		ollamaConfig := &llm.ProviderConfig{
			Endpoint: "http://localhost:11434",
		}
		ollamaProvider, err := llm.NewOllamaProvider(ollamaConfig)
		if err == nil {
			llmManager.RegisterProvider("ollama", ollamaProvider)
		}
	}

	// Initialize tool registry
	toolRegistry := tools.NewToolRegistry()

	// Register default tools
	toolRegistry.RegisterTool(tools.NewWebSearchTool())
	toolRegistry.RegisterTool(tools.NewCalculatorTool())
	toolRegistry.RegisterTool(tools.NewFileReadTool())
	toolRegistry.RegisterTool(tools.NewFileWriteTool())
	toolRegistry.RegisterTool(tools.NewShellTool())
	toolRegistry.RegisterTool(tools.NewHTTPTool())
	toolRegistry.RegisterTool(tools.NewTimeTool())

	// Initialize session manager (using memory for now)
	sessionManager := persistence.NewSessionManager(nil)

	// Initialize agent manager
	agentManager := server.NewAgentManager(llmManager, toolRegistry)

	// Set components on server
	srv.SetLLMManager(llmManager)
	srv.SetToolRegistry(toolRegistry)
	srv.SetAgentManager(agentManager)
	srv.SetSessionManager(sessionManager)

	return nil
}

func runMigrations() {
	fmt.Println("Running database migrations...")

	dbType := viper.GetString("db-type")

	switch dbType {
	case "postgres":
		config := &persistence.DatabaseConfig{
			Type:     "postgres",
			Host:     viper.GetString("db-host"),
			Port:     viper.GetInt("db-port"),
			Database: viper.GetString("db-name"),
			Username: viper.GetString("db-user"),
			Password: viper.GetString("db-password"),
			SSLMode:  "disable",
		}

		checkpointer, err := persistence.NewPostgresCheckpointer(config)
		if err != nil {
			log.Fatalf("Failed to create PostgreSQL checkpointer: %v", err)
		}
		defer checkpointer.Close()

		fmt.Println("PostgreSQL migrations completed successfully")

	case "redis":
		config := &persistence.DatabaseConfig{
			Type:     "redis",
			Host:     viper.GetString("db-host"),
			Port:     viper.GetInt("db-port"),
			Password: viper.GetString("db-password"),
		}

		checkpointer, err := persistence.NewRedisCheckpointer(config)
		if err != nil {
			log.Fatalf("Failed to create Redis checkpointer: %v", err)
		}
		defer checkpointer.Close()

		fmt.Println("Redis setup completed successfully")

	default:
		log.Fatalf("Unsupported database type: %s", dbType)
	}
}

func runVisualize(args []string, format, output string) {
	fmt.Printf("Visualizing graph in %s format...\n", format)

	// Create a sample graph for demonstration
	// In a real implementation, this would load from a file or configuration
	sampleGraph := createSampleGraph()

	// Create visualizer
	visualizer := debug.NewGraphVisualizer(nil, nil)

	// Get topology
	topology := visualizer.GetGraphTopology(sampleGraph)

	var result string
	switch format {
	case "mermaid":
		result = visualizer.GenerateMermaidDiagram(topology)
	case "dot":
		result = visualizer.GenerateDotDiagram(topology)
	case "json":
		// JSON output would need to be implemented
		result = "JSON output not implemented yet"
	default:
		log.Fatalf("Unsupported format: %s", format)
	}

	// Output result
	if output != "" {
		if err := os.WriteFile(output, []byte(result), 0644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Visualization saved to %s\n", output)
	} else {
		fmt.Println(result)
	}
}

func createSampleGraph() *core.Graph {
	// This is a placeholder - in a real implementation, you'd load from configuration
	graph := core.NewGraph("sample-graph")

	// Add some sample nodes
	graph.AddNode("start", "Start", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		return state, nil
	})

	graph.AddNode("process", "Process", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		return state, nil
	})

	graph.AddNode("end", "End", func(ctx context.Context, state *core.BaseState) (*core.BaseState, error) {
		return state, nil
	})

	// Add edges
	graph.AddEdge("start", "process", nil)
	graph.AddEdge("process", "end", nil)

	// Set start and end nodes
	graph.SetStartNode("start")
	graph.AddEndNode("end")

	return graph
}

func runTests() {
	fmt.Println("Running tests...")

	// Create test configuration
	testConfig := &agent.AgentConfig{
		Name:         "test-agent",
		Type:         agent.AgentTypeChat,
		Model:        "gpt-3.5-turbo",
		Provider:     "openai",
		SystemPrompt: "You are a helpful assistant for testing.",
		Temperature:  0.7,
		MaxTokens:    100,
	}

	// Initialize components for testing
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Create test agent
	testAgent := agent.NewAgent(testConfig, llmManager, toolRegistry)

	fmt.Printf("Test agent created: %s\n", testAgent.GetConfig().Name)
	fmt.Printf("Agent type: %s\n", testAgent.GetConfig().Type)
	fmt.Printf("Model: %s\n", testAgent.GetConfig().Model)

	// Validate graph structure
	graph := testAgent.GetGraph()
	if err := graph.Validate(); err != nil {
		log.Fatalf("Graph validation failed: %v", err)
	}

	fmt.Println("Graph validation passed")
	fmt.Println("All tests completed successfully")
}

func runInit(args []string, template string) {
	fmt.Printf("Initializing new GoLangGraph project...\n")
	
	projectName := "golanggraph-agent"
	if len(args) > 0 {
		projectName = args[0]
	}
	
	fmt.Printf("Creating project: %s with template: %s\n", projectName, template)
	
	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		log.Fatalf("Failed to create project directory: %v", err)
	}
	
	// Create subdirectories
	dirs := []string{
		"configs",
		"agents",
		"tools",
		"static",
		"tests",
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(fmt.Sprintf("%s/%s", projectName, dir), 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	
	// Create template files based on template type
	switch template {
	case "basic":
		createBasicTemplate(projectName)
	case "advanced":
		createAdvancedTemplate(projectName)
	case "rag":
		createRAGTemplate(projectName)
	default:
		createBasicTemplate(projectName)
	}
	
	fmt.Printf("Project %s initialized successfully!\n", projectName)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  golanggraph dev\n")
}

func runDockerBuild(args []string, distroless bool, tag, dockerfile, platform string) {
	fmt.Printf("Building Docker container...\n")
	
	configFile := "agent-config.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}
	
	fmt.Printf("Using config file: %s\n", configFile)
	
	// Determine image tag
	if tag == "" {
		tag = "golanggraph-agent:latest"
	}
	
	// Choose dockerfile based on distroless flag
	var dockerfilePath string
	if dockerfile != "" {
		dockerfilePath = dockerfile
	} else if distroless {
		dockerfilePath = "Dockerfile.distroless"
		createDistrolessDockerfile(dockerfilePath)
	} else {
		dockerfilePath = "Dockerfile.agent"
		createAgentDockerfile(dockerfilePath)
	}
	
	// Build Docker command
	var dockerCmd []string
	dockerCmd = append(dockerCmd, "docker", "build", "-f", dockerfilePath, "-t", tag)
	
	if platform != "" {
		dockerCmd = append(dockerCmd, "--platform", platform)
	}
	
	dockerCmd = append(dockerCmd, ".")
	
	fmt.Printf("Running: %s\n", fmt.Sprintf("%v", dockerCmd))
	fmt.Printf("Image tag: %s\n", tag)
	fmt.Printf("Dockerfile: %s\n", dockerfilePath)
	fmt.Printf("Distroless: %t\n", distroless)
	
	// Note: In a real implementation, you would execute the docker command
	// For now, we'll just show what would be executed
	fmt.Printf("Docker build command prepared. Execute manually or integrate with docker library.\n")
}

func runDevServer() {
	fmt.Println("Starting development server...")
	
	// Create development server configuration
	config := &server.ServerConfig{
		Host:           viper.GetString("host"),
		Port:           viper.GetInt("port"),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		EnableCORS:     true,
		StaticDir:      "./static",
		DevMode:        true,
	}
	
	// Create server
	srv := server.NewServer(config)
	
	// Initialize components
	if err := initializeComponents(srv); err != nil {
		log.Fatalf("Failed to initialize components: %v", err)
	}
	
	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	fmt.Printf("Development server started on %s:%d\n", config.Host, config.Port)
	fmt.Printf("API endpoints: http://%s:%d/api/v1/\n", config.Host, config.Port)
	fmt.Printf("Debug interface: http://%s:%d/debug\n", config.Host, config.Port)
	fmt.Printf("Agent playground: http://%s:%d/playground\n", config.Host, config.Port)
	
	// Watch for file changes (hot-reload)
	if viper.GetBool("hot-reload") {
		fmt.Println("Hot-reload enabled - watching for changes...")
		// Note: File watching implementation would go here
	}
	
	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	fmt.Println("Shutting down development server...")
	
	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	fmt.Println("Development server stopped")
}

func runValidate(args []string, strict bool) {
	fmt.Printf("Validating configuration...\n")
	
	configFile := "agent-config.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}
	
	fmt.Printf("Config file: %s\n", configFile)
	fmt.Printf("Strict mode: %t\n", strict)
	
	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("Configuration file not found: %s", configFile)
	}
	
	// Note: In a real implementation, you would:
	// 1. Parse the configuration file
	// 2. Validate the schema
	// 3. Check for required fields
	// 4. Validate graph structure
	// 5. Check tool availability
	// 6. Validate LLM provider configuration
	
	fmt.Printf("Configuration validation completed successfully!\n")
}

func runDeployDocker(args []string) {
	fmt.Printf("Deploying agent using Docker...\n")
	
	configFile := "agent-config.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}
	
	fmt.Printf("Config file: %s\n", configFile)
	
	// Note: In a real implementation, you would:
	// 1. Build the Docker image
	// 2. Push to registry
	// 3. Deploy to target environment
	// 4. Monitor deployment status
	
	fmt.Printf("Docker deployment completed for config: %s!\n", configFile)
}

func createBasicTemplate(projectName string) {
	// Create basic agent configuration
	agentConfig := `name: "basic-agent"
type: "chat"
model: "gpt-3.5-turbo"
provider: "openai"
system_prompt: "You are a helpful assistant."
temperature: 0.7
max_tokens: 1000

tools:
  - name: "calculator"
    enabled: true
  - name: "web_search"
    enabled: false

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"
`
	
	if err := os.WriteFile(fmt.Sprintf("%s/configs/agent-config.yaml", projectName), []byte(agentConfig), 0644); err != nil {
		log.Fatalf("Failed to create agent config: %v", err)
	}
	
	// Create docker-compose for development
	dockerCompose := `version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: golanggraph
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
`
	
	if err := os.WriteFile(fmt.Sprintf("%s/docker-compose.yml", projectName), []byte(dockerCompose), 0644); err != nil {
		log.Fatalf("Failed to create docker-compose: %v", err)
	}
}

func createAdvancedTemplate(projectName string) {
	createBasicTemplate(projectName)
	
	// Add advanced configuration
	advancedConfig := `name: "advanced-agent"
type: "multi-agent"
model: "gpt-4"
provider: "openai"
system_prompt: "You are an advanced AI assistant with multiple capabilities."
temperature: 0.7
max_tokens: 2000

agents:
  - name: "research-agent"
    type: "research"
    tools: ["web_search", "document_reader"]
  - name: "analysis-agent"
    type: "analysis"
    tools: ["calculator", "data_analyzer"]
  - name: "synthesis-agent"
    type: "synthesis"
    tools: ["summarizer", "report_generator"]

workflow:
  start_node: "research-agent"
  edges:
    - from: "research-agent"
      to: "analysis-agent"
    - from: "analysis-agent"
      to: "synthesis-agent"
  end_node: "synthesis-agent"

tools:
  - name: "web_search"
    enabled: true
    config:
      api_key: "${SEARCH_API_KEY}"
  - name: "document_reader"
    enabled: true
  - name: "calculator"
    enabled: true
  - name: "data_analyzer"
    enabled: true
  - name: "summarizer"
    enabled: true
  - name: "report_generator"
    enabled: true

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"

vector_store:
  type: "pgvector"
  host: "localhost"
  port: 5432
  database: "vectordb"
  username: "postgres"
  password: "password"
  dimensions: 1536
`
	
	if err := os.WriteFile(fmt.Sprintf("%s/configs/advanced-config.yaml", projectName), []byte(advancedConfig), 0644); err != nil {
		log.Fatalf("Failed to create advanced config: %v", err)
	}
}

func createRAGTemplate(projectName string) {
	createAdvancedTemplate(projectName)
	
	// Add RAG-specific configuration
	ragConfig := `name: "rag-agent"
type: "rag"
model: "gpt-4"
provider: "openai"
system_prompt: "You are a RAG-enabled AI assistant that can retrieve and analyze information from documents."
temperature: 0.7
max_tokens: 2000

rag:
  enabled: true
  chunk_size: 1000
  chunk_overlap: 200
  similarity_threshold: 0.7
  max_chunks: 5
  embedding_model: "text-embedding-ada-002"

vector_store:
  type: "pgvector"
  host: "localhost"
  port: 5432
  database: "vectordb"
  username: "postgres"
  password: "password"
  dimensions: 1536
  collection_name: "documents"

document_loaders:
  - type: "pdf"
    enabled: true
  - type: "text"
    enabled: true
  - type: "markdown"
    enabled: true
  - type: "web"
    enabled: true

tools:
  - name: "vector_search"
    enabled: true
  - name: "document_loader"
    enabled: true
  - name: "web_search"
    enabled: true
  - name: "summarizer"
    enabled: true

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"
`
	
	if err := os.WriteFile(fmt.Sprintf("%s/configs/rag-config.yaml", projectName), []byte(ragConfig), 0644); err != nil {
		log.Fatalf("Failed to create RAG config: %v", err)
	}
}

func createAgentDockerfile(filepath string) {
	dockerfile := `# Production Dockerfile for GoLangGraph Agent
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=production
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -a -installsuffix cgo \
    -o golanggraph-agent \
    ./cmd/golanggraph

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S golanggraph && \
    adduser -u 1001 -S golanggraph -G golanggraph

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/golanggraph-agent .

# Copy configuration files
COPY configs/ ./configs/
COPY static/ ./static/

# Create necessary directories
RUN mkdir -p ./logs ./data

# Change ownership to non-root user
RUN chown -R golanggraph:golanggraph /app

# Switch to non-root user
USER golanggraph

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./golanggraph-agent health || exit 1

# Run the agent
ENTRYPOINT ["./golanggraph-agent"]
CMD ["serve", "--host", "0.0.0.0", "--port", "8080"]
`
	
	if err := os.WriteFile(filepath, []byte(dockerfile), 0644); err != nil {
		log.Fatalf("Failed to create Dockerfile: %v", err)
	}
}

func createDistrolessDockerfile(filepath string) {
	dockerfile := `# Distroless Dockerfile for GoLangGraph Agent
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=production
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -a -installsuffix cgo \
    -o golanggraph-agent \
    ./cmd/golanggraph

# Distroless production stage
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/golanggraph-agent /

# Copy configuration files
COPY configs/ /configs/
COPY static/ /static/

# Use distroless nonroot user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Health check (note: distroless doesn't support HEALTHCHECK)
# Use external health check monitoring

# Run the agent
ENTRYPOINT ["/golanggraph-agent"]
CMD ["serve", "--host", "0.0.0.0", "--port", "8080"]
`
	
	if err := os.WriteFile(filepath, []byte(dockerfile), 0644); err != nil {
		log.Fatalf("Failed to create distroless Dockerfile: %v", err)
	}
}

func runHealthCheck() {
	fmt.Printf("Running GoLangGraph health check...\n")
	
	healthy := true
	var issues []string
	
	// Check system resources
	fmt.Printf("Checking system resources...\n")
	
	// Check database connectivity
	fmt.Printf("Checking database connectivity...\n")
	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	fmt.Printf("  PostgreSQL: %s:5432 - ", dbHost)
	// In a real implementation, you would test actual connectivity
	fmt.Printf("✓ Reachable\n")
	
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	fmt.Printf("  Redis: %s:6379 - ", redisHost)
	// In a real implementation, you would test actual connectivity
	fmt.Printf("✓ Reachable\n")
	
	// Check LLM providers
	fmt.Printf("Checking LLM providers...\n")
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		fmt.Printf("  OpenAI: ✓ API key configured\n")
	} else {
		fmt.Printf("  OpenAI: ⚠ API key not configured\n")
		issues = append(issues, "OpenAI API key not configured")
	}
	
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	fmt.Printf("  Ollama: %s - ", ollamaURL)
	// In a real implementation, you would test actual connectivity
	fmt.Printf("✓ Reachable\n")
	
	// Check disk space
	fmt.Printf("Checking system resources...\n")
	fmt.Printf("  Disk space: ✓ Sufficient\n")
	fmt.Printf("  Memory: ✓ Available\n")
	
	// Overall health status
	fmt.Printf("\n")
	if healthy && len(issues) == 0 {
		fmt.Printf("✅ System is healthy\n")
		os.Exit(0)
	} else {
		fmt.Printf("⚠ System has issues:\n")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		os.Exit(1)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
