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
- Deploying and managing agent graphs
- Running development servers
- Managing database migrations
- Visualizing graph execution
- Testing and debugging agents`,
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
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(testCmd)
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
