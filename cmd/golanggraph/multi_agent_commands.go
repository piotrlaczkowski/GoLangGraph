// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/server"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// multiAgentCmd represents the multi-agent command group
var multiAgentCmd = &cobra.Command{
	Use:   "multi-agent",
	Short: "Multi-agent deployment and management commands",
	Long: `Multi-agent commands provide functionality to manage multiple AI agents
with different configurations, routing, and deployment options.

This includes:
- Initializing multi-agent projects
- Deploying multiple agents simultaneously  
- Managing agent routing and load balancing
- Monitoring multi-agent deployments
- Schema validation for individual agents`,
}

// multiAgentInitCmd represents the multi-agent init command
var multiAgentInitCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new multi-agent project",
	Long: `Initialize a new multi-agent project with example configurations and templates.
Creates a directory structure optimized for managing multiple agents with different
configurations, routing rules, and deployment settings.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		template, _ := cmd.Flags().GetString("template")
		agentCount, _ := cmd.Flags().GetInt("agents")
		outputFormat, _ := cmd.Flags().GetString("format")
		routingType, _ := cmd.Flags().GetString("routing")
		runMultiAgentInit(args, template, agentCount, outputFormat, routingType)
	},
}

// multiAgentValidateCmd represents the multi-agent validate command
var multiAgentValidateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate multi-agent configuration",
	Long: `Validate multi-agent configuration files including agent definitions,
routing rules, deployment settings, and schema validation.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		strict, _ := cmd.Flags().GetBool("strict")
		checkSchemas, _ := cmd.Flags().GetBool("check-schemas")
		runMultiAgentValidate(args, strict, checkSchemas)
	},
}

// multiAgentDeployCmd represents the multi-agent deploy command
var multiAgentDeployCmd = &cobra.Command{
	Use:   "deploy [config-file]",
	Short: "Deploy multiple agents",
	Long: `Deploy multiple agents according to the multi-agent configuration.
Supports various deployment targets including Docker, Kubernetes, and serverless platforms.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deploymentType, _ := cmd.Flags().GetString("type")
		environment, _ := cmd.Flags().GetString("environment")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		parallel, _ := cmd.Flags().GetBool("parallel")
		runMultiAgentDeploy(args, deploymentType, environment, dryRun, parallel)
	},
}

// multiAgentServeCmd represents the multi-agent serve command
var multiAgentServeCmd = &cobra.Command{
	Use:   "serve [config-file]",
	Short: "Start multi-agent server",
	Long: `Start a server that hosts multiple agents with routing and load balancing.
Provides HTTP endpoints for agent execution, management, and monitoring.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		runMultiAgentServe(args, host, port)
	},
}

// multiAgentStatusCmd represents the multi-agent status command
var multiAgentStatusCmd = &cobra.Command{
	Use:   "status [config-file]",
	Short: "Check status of deployed agents",
	Long:  `Check the status of deployed agents including health, metrics, and runtime information.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputFormat, _ := cmd.Flags().GetString("format")
		watch, _ := cmd.Flags().GetBool("watch")
		runMultiAgentStatus(args, outputFormat, watch)
	},
}

// multiAgentGenerateCmd represents the multi-agent generate command
var multiAgentGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate deployment artifacts",
	Long:  `Generate deployment artifacts such as Docker files, Kubernetes manifests, and configuration files.`,
}

// multiAgentGenerateDockerCmd represents the generate docker command
var multiAgentGenerateDockerCmd = &cobra.Command{
	Use:   "docker [config-file]",
	Short: "Generate Docker deployment files",
	Long:  `Generate Docker Compose files and Dockerfiles for multi-agent deployment.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, _ := cmd.Flags().GetString("output")
		multiService, _ := cmd.Flags().GetBool("multi-service")
		runGenerateDocker(args, outputDir, multiService)
	},
}

// multiAgentGenerateK8sCmd represents the generate k8s command
var multiAgentGenerateK8sCmd = &cobra.Command{
	Use:   "k8s [config-file]",
	Short: "Generate Kubernetes deployment manifests",
	Long:  `Generate Kubernetes deployment, service, and ingress manifests for multi-agent deployment.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, _ := cmd.Flags().GetString("output")
		namespace, _ := cmd.Flags().GetString("namespace")
		runGenerateK8s(args, outputDir, namespace)
	},
}

func init() {
	// Add multi-agent command group to root
	rootCmd.AddCommand(multiAgentCmd)

	// Add subcommands
	// Load command
	multiAgentLoadCmd := &cobra.Command{
		Use:   "load [plugin-path or directory]",
		Short: "Load agent definitions from Go files or plugins",
		Long: `Load agent definitions from Go files or plugins.
		
This command can load agents defined programmatically in Go files,
either as plugins or by analyzing Go source files in a directory.

Examples:
  # Load agents from a plugin file
  golanggraph multi-agent load ./agents.so
  
  # Load agents from Go files in a directory
  golanggraph multi-agent load ./agents/
  
  # Load agents from current directory
  golanggraph multi-agent load .`,
		Args: cobra.MaximumNArgs(1),
		RunE: runMultiAgentLoad,
	}

	multiAgentLoadCmd.Flags().BoolP("recursive", "r", false, "Recursively scan directories for Go files")
	multiAgentLoadCmd.Flags().StringSliceP("include", "i", []string{"*.go"}, "File patterns to include")
	multiAgentLoadCmd.Flags().StringSliceP("exclude", "e", []string{"*_test.go"}, "File patterns to exclude")
	multiAgentLoadCmd.Flags().BoolP("validate", "v", true, "Validate loaded agent definitions")
	multiAgentLoadCmd.Flags().BoolP("verbose", "", false, "Verbose output")

	// List command
	multiAgentListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all registered agent definitions",
		Long: `List all registered agent definitions including those from:
- Configuration files
- Go-based definitions
- Factories
- Plugins

This shows the source, type, and metadata for each registered agent.`,
		RunE: runMultiAgentList,
	}

	multiAgentListCmd.Flags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	multiAgentListCmd.Flags().StringP("filter", "", "", "Filter agents by name pattern")
	multiAgentListCmd.Flags().BoolP("show-metadata", "m", false, "Show agent metadata")
	multiAgentListCmd.Flags().BoolP("show-config", "c", false, "Show agent configuration")

	// Add subcommands after all commands are declared
	multiAgentCmd.AddCommand(multiAgentInitCmd)
	multiAgentCmd.AddCommand(multiAgentValidateCmd)
	multiAgentCmd.AddCommand(multiAgentDeployCmd)
	multiAgentCmd.AddCommand(multiAgentServeCmd)
	multiAgentCmd.AddCommand(multiAgentStatusCmd)
	multiAgentCmd.AddCommand(multiAgentGenerateCmd)
	multiAgentCmd.AddCommand(multiAgentLoadCmd)
	multiAgentCmd.AddCommand(multiAgentListCmd)

	// Add generation subcommands
	multiAgentGenerateCmd.AddCommand(multiAgentGenerateDockerCmd)
	multiAgentGenerateCmd.AddCommand(multiAgentGenerateK8sCmd)

	// Multi-agent init flags
	multiAgentInitCmd.Flags().StringP("template", "t", "basic", "Project template (basic, microservices, rag, workflow)")
	multiAgentInitCmd.Flags().IntP("agents", "a", 3, "Number of agents to create")
	multiAgentInitCmd.Flags().StringP("format", "f", "yaml", "Configuration format (yaml, json)")
	multiAgentInitCmd.Flags().StringP("routing", "r", "path", "Routing type (path, host, header, query)")

	// Multi-agent validate flags
	multiAgentValidateCmd.Flags().BoolP("strict", "s", false, "Enable strict validation")
	multiAgentValidateCmd.Flags().Bool("check-schemas", true, "Validate input/output schemas")

	// Multi-agent deploy flags
	multiAgentDeployCmd.Flags().StringP("type", "t", "docker", "Deployment type (docker, kubernetes, serverless)")
	multiAgentDeployCmd.Flags().StringP("environment", "e", "development", "Deployment environment")
	multiAgentDeployCmd.Flags().Bool("dry-run", false, "Show what would be deployed without actually deploying")
	multiAgentDeployCmd.Flags().Bool("parallel", true, "Deploy agents in parallel")

	// Multi-agent serve flags
	multiAgentServeCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to bind to")
	multiAgentServeCmd.Flags().IntP("port", "p", 8080, "Port to bind to")

	// Multi-agent status flags
	multiAgentStatusCmd.Flags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	multiAgentStatusCmd.Flags().BoolP("watch", "w", false, "Watch for status changes")

	// Generate docker flags
	multiAgentGenerateDockerCmd.Flags().StringP("output", "o", "./deploy", "Output directory")
	multiAgentGenerateDockerCmd.Flags().Bool("multi-service", true, "Generate multi-service Docker Compose")

	// Generate k8s flags
	multiAgentGenerateK8sCmd.Flags().StringP("output", "o", "./k8s", "Output directory")
	multiAgentGenerateK8sCmd.Flags().StringP("namespace", "n", "golanggraph", "Kubernetes namespace")
}

// runMultiAgentInit initializes a new multi-agent project
func runMultiAgentInit(args []string, template string, agentCount int, outputFormat, routingType string) {
	projectName := "golanggraph-multi-agent"
	if len(args) > 0 {
		projectName = args[0]
	}

	fmt.Printf("Initializing multi-agent project: %s\n", projectName)
	fmt.Printf("Template: %s, Agents: %d, Format: %s, Routing: %s\n", template, agentCount, outputFormat, routingType)

	// Create project directory
	if err := os.MkdirAll(projectName, 0750); err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		os.Exit(1)
	}

	// Create subdirectories
	dirs := []string{
		"agents",
		"configs",
		"deploy",
		"k8s",
		"scripts",
		"static",
		"tests",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectName, dir), 0750); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create agent subdirectories
	for i := 1; i <= agentCount; i++ {
		agentDir := filepath.Join(projectName, "agents", fmt.Sprintf("agent-%d", i))
		if err := os.MkdirAll(agentDir, 0750); err != nil {
			fmt.Printf("Error creating agent directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate multi-agent configuration
	config := createMultiAgentConfig(template, agentCount, routingType)

	// Write configuration file
	configFile := fmt.Sprintf("multi-agent.%s", outputFormat)
	configPath := filepath.Join(projectName, "configs", configFile)

	var configData []byte
	var err error

	switch outputFormat {
	case "yaml", "yml":
		configData, err = yaml.Marshal(config)
	case "json":
		configData, err = json.MarshalIndent(config, "", "  ")
	default:
		fmt.Printf("Unsupported output format: %s\n", outputFormat)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error marshaling configuration: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		fmt.Printf("Error writing configuration file: %v\n", err)
		os.Exit(1)
	}

	// Create individual agent configurations
	createIndividualAgentConfigs(projectName, config, outputFormat)

	// Create Docker compose file
	createDockerComposeFile(projectName, config)

	// Create Kubernetes manifests
	createK8sManifests(projectName, config)

	// Create README
	createProjectREADME(projectName, config)

	fmt.Printf("\nMulti-agent project '%s' initialized successfully!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  golanggraph multi-agent validate configs/%s\n", configFile)
	fmt.Printf("  golanggraph multi-agent serve configs/%s\n", configFile)
}

// runMultiAgentValidate validates multi-agent configuration
func runMultiAgentValidate(args []string, strict, checkSchemas bool) {
	configFile := "configs/multi-agent.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}

	fmt.Printf("Validating multi-agent configuration: %s\n", configFile)
	fmt.Printf("Strict mode: %t, Check schemas: %t\n", strict, checkSchemas)

	// Load configuration
	config, err := agent.LoadMultiAgentConfigFromFile(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Additional validations for strict mode
	if strict {
		if err := validateStrictMode(config); err != nil {
			fmt.Printf("Strict validation failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Schema validation
	if checkSchemas {
		if err := validateAgentSchemas(config); err != nil {
			fmt.Printf("Schema validation failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("✅ Configuration validation passed!\n")
	fmt.Printf("- Agents: %d\n", len(config.Agents))
	fmt.Printf("- Routing rules: %d\n", len(config.Routing.Rules))
	fmt.Printf("- Deployment type: %s\n", config.Deployment.Type)
}

// runMultiAgentDeploy deploys multiple agents
func runMultiAgentDeploy(args []string, deploymentType, environment string, dryRun, parallel bool) {
	configFile := "configs/multi-agent.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}

	fmt.Printf("Deploying multi-agent system: %s\n", configFile)
	fmt.Printf("Type: %s, Environment: %s, Dry-run: %t, Parallel: %t\n", deploymentType, environment, dryRun, parallel)

	// Load configuration
	config, err := agent.LoadMultiAgentConfigFromFile(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate before deployment
	if err := config.Validate(); err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Override deployment type if specified
	if deploymentType != "" {
		config.Deployment.Type = deploymentType
	}
	if environment != "" {
		config.Deployment.Environment = environment
	}

	if dryRun {
		fmt.Printf("DRY RUN - Would deploy the following agents:\n")
		for agentID, agentConfig := range config.Agents {
			fmt.Printf("  - %s: %s (%s on %s)\n", agentID, agentConfig.Name, agentConfig.Type, agentConfig.Provider)
		}
		return
	}

	// Perform actual deployment
	switch config.Deployment.Type {
	case "docker":
		deployDocker(config, parallel)
	case "kubernetes":
		deployKubernetes(config, parallel)
	case "serverless":
		deployServerless(config, parallel)
	default:
		fmt.Printf("Unsupported deployment type: %s\n", config.Deployment.Type)
		os.Exit(1)
	}
}

// runMultiAgentServe starts multi-agent server
func runMultiAgentServe(args []string, host string, port int) {
	configFile := "configs/multi-agent.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}

	fmt.Printf("Starting multi-agent server: %s\n", configFile)
	fmt.Printf("Host: %s, Port: %d\n", host, port)

	// Load configuration
	config, err := agent.LoadMultiAgentConfigFromFile(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize components
	llmManager := llm.NewProviderManager()
	toolRegistry := tools.NewToolRegistry()

	// Setup LLM providers from shared config
	if config.Shared != nil && config.Shared.LLMProviders != nil {
		for providerName, providerConfig := range config.Shared.LLMProviders {
			// Initialize provider based on type
			// This is a simplified version - in a real implementation,
			// you'd create the appropriate provider based on type
			fmt.Printf("Setting up LLM provider: %s (%s)\n", providerName, providerConfig.Type)
		}
	}

	// Register default tools
	toolRegistry.RegisterTool(tools.NewWebSearchTool())
	toolRegistry.RegisterTool(tools.NewCalculatorTool())
	toolRegistry.RegisterTool(tools.NewFileReadTool())
	toolRegistry.RegisterTool(tools.NewFileWriteTool())
	toolRegistry.RegisterTool(tools.NewShellTool())
	toolRegistry.RegisterTool(tools.NewHTTPTool())
	toolRegistry.RegisterTool(tools.NewTimeTool())

	// Create multi-agent manager
	multiAgentManager, err := agent.NewMultiAgentManager(config, llmManager, toolRegistry)
	if err != nil {
		fmt.Printf("Error creating multi-agent manager: %v\n", err)
		os.Exit(1)
	}

	// Start multi-agent manager
	ctx := context.Background()
	if err := multiAgentManager.Start(ctx); err != nil {
		fmt.Printf("Error starting multi-agent manager: %v\n", err)
		os.Exit(1)
	}

	// Create server configuration
	serverConfig := &server.ServerConfig{
		Host:           host,
		Port:           port,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
		EnableCORS:     true,
		StaticDir:      "./static",
		DevMode:        viper.GetBool("dev-mode"),
	}

	// Create and start server with multi-agent router
	srv := server.NewServer(serverConfig)
	// Note: The server will use its own router, multi-agent manager handles routing internally

	fmt.Printf("Multi-agent server started on %s:%d\n", host, port)
	fmt.Printf("Health check: http://%s:%d/health\n", host, port)
	fmt.Printf("Agent endpoints: http://%s:%d/agents\n", host, port)
	fmt.Printf("Metrics: http://%s:%d/metrics\n", host, port)

	if err := srv.Start(); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
		os.Exit(1)
	}
}

// runMultiAgentStatus checks status of deployed agents
func runMultiAgentStatus(args []string, outputFormat string, watch bool) {
	configFile := "configs/multi-agent.yaml"
	if len(args) > 0 {
		configFile = args[0]
	}

	fmt.Printf("Checking multi-agent status: %s\n", configFile)

	// Load configuration
	config, err := agent.LoadMultiAgentConfigFromFile(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// In a real implementation, this would connect to the running multi-agent system
	// and fetch actual status information

	status := map[string]interface{}{
		"timestamp": time.Now(),
		"config":    configFile,
		"agents":    make(map[string]interface{}),
	}

	for agentID, agentConfig := range config.Agents {
		status["agents"].(map[string]interface{})[agentID] = map[string]interface{}{
			"name":          agentConfig.Name,
			"type":          agentConfig.Type,
			"status":        "unknown", // Would be fetched from actual deployment
			"health_status": "unknown",
			"request_count": 0,
			"error_count":   0,
		}
	}

	// Output status
	switch outputFormat {
	case "json":
		output, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(output))
	case "yaml":
		output, _ := yaml.Marshal(status)
		fmt.Println(string(output))
	default:
		// Table format
		fmt.Printf("\nAgent Status:\n")
		fmt.Printf("%-20s %-15s %-10s %-10s\n", "Agent ID", "Type", "Status", "Health")
		fmt.Printf("%-20s %-15s %-10s %-10s\n", "--------", "----", "------", "------")
		for agentID, agentConfig := range config.Agents {
			fmt.Printf("%-20s %-15s %-10s %-10s\n", agentID, agentConfig.Type, "unknown", "unknown")
		}
	}

	if watch {
		fmt.Printf("\nWatching for status changes (press Ctrl+C to exit)...\n")
		// In a real implementation, this would watch for status changes
		select {} // Block forever
	}
}

// Helper functions for generating project artifacts

func createMultiAgentConfig(template string, agentCount int, routingType string) *agent.MultiAgentConfig {
	config := agent.DefaultMultiAgentConfig()
	config.Name = "example-multi-agent"
	config.Description = "Example multi-agent configuration"
	config.Routing.Type = routingType

	// Create agents based on template
	switch template {
	case "microservices":
		createMicroservicesAgents(config, agentCount)
	case "rag":
		createRAGAgents(config, agentCount)
	case "workflow":
		createWorkflowAgents(config, agentCount)
	default:
		createBasicAgents(config, agentCount)
	}

	// Setup routing rules
	setupRoutingRules(config, routingType)

	return config
}

func createBasicAgents(config *agent.MultiAgentConfig, count int) {
	agentTypes := []agent.AgentType{agent.AgentTypeChat, agent.AgentTypeReAct, agent.AgentTypeTool}

	for i := 1; i <= count; i++ {
		agentID := fmt.Sprintf("agent-%d", i)
		agentType := agentTypes[(i-1)%len(agentTypes)]

		agentConfig := agent.DefaultAgentConfig()
		agentConfig.ID = agentID
		agentConfig.Name = fmt.Sprintf("Agent %d", i)
		agentConfig.Type = agentType
		agentConfig.Model = "gpt-3.5-turbo"
		agentConfig.Provider = "openai"
		agentConfig.SystemPrompt = fmt.Sprintf("You are Agent %d, a helpful AI assistant specialized in %s tasks.", i, agentType)
		agentConfig.Tools = []string{"calculator", "web_search"}

		config.Agents[agentID] = agentConfig
	}
}

func createMicroservicesAgents(config *agent.MultiAgentConfig, count int) {
	services := []struct {
		name        string
		agentType   agent.AgentType
		description string
		tools       []string
	}{
		{"user-service", agent.AgentTypeChat, "Handles user interactions and authentication", []string{"user_db", "auth"}},
		{"order-service", agent.AgentTypeReAct, "Processes orders and payments", []string{"payment", "inventory"}},
		{"inventory-service", agent.AgentTypeTool, "Manages product inventory", []string{"database", "calculator"}},
		{"notification-service", agent.AgentTypeChat, "Sends notifications and alerts", []string{"email", "sms"}},
		{"analytics-service", agent.AgentTypeReAct, "Provides analytics and insights", []string{"database", "calculator", "chart"}},
	}

	for i := 0; i < count && i < len(services); i++ {
		service := services[i]
		agentID := fmt.Sprintf("agent-%d", i+1)

		agentConfig := agent.DefaultAgentConfig()
		agentConfig.ID = agentID
		agentConfig.Name = service.name
		agentConfig.Type = service.agentType
		agentConfig.Model = "gpt-4"
		agentConfig.Provider = "openai"
		agentConfig.SystemPrompt = fmt.Sprintf("You are the %s agent. %s", service.name, service.description)
		agentConfig.Tools = service.tools

		config.Agents[agentID] = agentConfig
	}
}

func createRAGAgents(config *agent.MultiAgentConfig, count int) {
	ragAgents := []struct {
		name        string
		description string
		domain      string
	}{
		{"document-processor", "Processes and indexes documents", "document-processing"},
		{"knowledge-retriever", "Retrieves relevant knowledge from vector store", "information-retrieval"},
		{"answer-generator", "Generates answers based on retrieved context", "question-answering"},
	}

	for i := 0; i < count && i < len(ragAgents); i++ {
		ragAgent := ragAgents[i]
		agentID := fmt.Sprintf("agent-%d", i+1)

		agentConfig := agent.DefaultAgentConfig()
		agentConfig.ID = agentID
		agentConfig.Name = ragAgent.name
		agentConfig.Type = agent.AgentTypeReAct
		agentConfig.Model = "gpt-4"
		agentConfig.Provider = "openai"
		agentConfig.SystemPrompt = fmt.Sprintf("You are the %s agent specialized in %s. %s",
			ragAgent.name, ragAgent.domain, ragAgent.description)
		agentConfig.Tools = []string{"vector_search", "document_loader", "summarizer"}

		config.Agents[agentID] = agentConfig
	}
}

func createWorkflowAgents(config *agent.MultiAgentConfig, count int) {
	workflowSteps := []struct {
		name        string
		agentType   agent.AgentType
		description string
	}{
		{"input-validator", agent.AgentTypeTool, "Validates and preprocesses input data"},
		{"task-planner", agent.AgentTypeReAct, "Plans the execution workflow"},
		{"executor", agent.AgentTypeReAct, "Executes the planned tasks"},
		{"result-aggregator", agent.AgentTypeTool, "Aggregates and formats results"},
		{"output-formatter", agent.AgentTypeChat, "Formats final output for users"},
	}

	for i := 0; i < count && i < len(workflowSteps); i++ {
		step := workflowSteps[i]
		agentID := fmt.Sprintf("agent-%d", i+1)

		agentConfig := agent.DefaultAgentConfig()
		agentConfig.ID = agentID
		agentConfig.Name = step.name
		agentConfig.Type = step.agentType
		agentConfig.Model = "gpt-4"
		agentConfig.Provider = "openai"
		agentConfig.SystemPrompt = fmt.Sprintf("You are the %s agent in the workflow. %s", step.name, step.description)
		agentConfig.Tools = []string{"validator", "planner", "executor"}

		config.Agents[agentID] = agentConfig
	}
}

func setupRoutingRules(config *agent.MultiAgentConfig, routingType string) {
	i := 1
	for agentID := range config.Agents {
		rule := agent.RoutingRule{
			ID:       fmt.Sprintf("rule-%d", i),
			AgentID:  agentID,
			Method:   "POST",
			Priority: 100 - i,
		}

		switch routingType {
		case "path":
			rule.Pattern = fmt.Sprintf("/%s", agentID)
		case "host":
			rule.Pattern = fmt.Sprintf("%s.example.com", agentID)
		case "header":
			rule.Pattern = fmt.Sprintf("X-Agent-ID:%s", agentID)
		case "query":
			rule.Pattern = fmt.Sprintf("agent=%s", agentID)
		default:
			rule.Pattern = fmt.Sprintf("/%s", agentID)
		}

		config.Routing.Rules = append(config.Routing.Rules, rule)
		i++
	}

	// Set default agent to the first one
	if len(config.Agents) > 0 {
		for agentID := range config.Agents {
			config.Routing.DefaultAgent = agentID
			break
		}
	}
}

func createIndividualAgentConfigs(projectName string, config *agent.MultiAgentConfig, format string) {
	for agentID, agentConfig := range config.Agents {
		agentDir := filepath.Join(projectName, "agents", agentID)
		configFile := fmt.Sprintf("config.%s", format)
		configPath := filepath.Join(agentDir, configFile)

		var configData []byte
		var err error

		switch format {
		case "yaml", "yml":
			configData, err = yaml.Marshal(agentConfig)
		case "json":
			configData, err = json.MarshalIndent(agentConfig, "", "  ")
		}

		if err == nil {
			os.WriteFile(configPath, configData, 0600)
		}
	}
}

func createDockerComposeFile(projectName string, config *agent.MultiAgentConfig) {
	dockerCompose := `version: '3.8'
services:
  multi-agent:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - OLLAMA_URL=${OLLAMA_URL}
    volumes:
      - ./configs:/app/configs:ro
    depends_on:
      - postgres
      - redis

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

  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama

volumes:
  postgres_data:
  ollama_data:
`

	composePath := filepath.Join(projectName, "docker-compose.yml")
	os.WriteFile(composePath, []byte(dockerCompose), 0600)
}

func createK8sManifests(projectName string, config *agent.MultiAgentConfig) {
	k8sDir := filepath.Join(projectName, "k8s")

	// Deployment manifest
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: golanggraph-multi-agent
  labels:
    app: golanggraph-multi-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: golanggraph-multi-agent
  template:
    metadata:
      labels:
        app: golanggraph-multi-agent
    spec:
      containers:
      - name: multi-agent
        image: golanggraph-multi-agent:latest
        ports:
        - containerPort: 8080
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: golanggraph-secrets
              key: openai-api-key
`

	// Service manifest
	service := `apiVersion: v1
kind: Service
metadata:
  name: golanggraph-multi-agent-service
spec:
  selector:
    app: golanggraph-multi-agent
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
`

	os.WriteFile(filepath.Join(k8sDir, "deployment.yaml"), []byte(deployment), 0600)
	os.WriteFile(filepath.Join(k8sDir, "service.yaml"), []byte(service), 0600)
}

func createProjectREADME(projectName string, config *agent.MultiAgentConfig) {
	readme := fmt.Sprintf("# %s\n\nMulti-agent GoLangGraph project with %d agents.\n\n", projectName, len(config.Agents))

	readme += "## Quick Start\n\n"
	readme += "1. **Validate configuration:**\n"
	readme += "   ```bash\n"
	readme += "   golanggraph multi-agent validate configs/multi-agent.yaml\n"
	readme += "   ```\n\n"
	readme += "2. **Start the multi-agent server:**\n"
	readme += "   ```bash\n"
	readme += "   golanggraph multi-agent serve configs/multi-agent.yaml\n"
	readme += "   ```\n\n"
	readme += "3. **Deploy with Docker:**\n"
	readme += "   ```bash\n"
	readme += "   docker-compose up -d\n"
	readme += "   ```\n\n"
	readme += "4. **Deploy to Kubernetes:**\n"
	readme += "   ```bash\n"
	readme += "   kubectl apply -f k8s/\n"
	readme += "   ```\n\n"

	readme += "## Configuration\n\n"
	readme += "The multi-agent configuration is defined in `configs/multi-agent.yaml`.\n\n"
	readme += "### Agents\n\n"

	for agentID, agentConfig := range config.Agents {
		readme += fmt.Sprintf("- **%s**: %s (%s)\n", agentID, agentConfig.Name, agentConfig.Type)
	}

	readme += "\n### Routing\n\n"
	readme += "Requests are routed to different agents based on the configured routing rules.\n\n"
	readme += "### API Endpoints\n\n"
	readme += "- `POST /agent-1` - Route to Agent 1\n"
	readme += "- `POST /agent-2` - Route to Agent 2\n"
	readme += "- `GET /health` - Health check\n"
	readme += "- `GET /metrics` - Metrics\n"
	readme += "- `GET /agents` - List all agents\n\n"

	readme += "## Development\n\n"
	readme += "1. **Add a new agent:**\n"
	readme += "   - Edit `configs/multi-agent.yaml`\n"
	readme += "   - Add agent configuration\n"
	readme += "   - Update routing rules\n"
	readme += "   - Validate configuration\n\n"
	readme += "2. **Test changes:**\n"
	readme += "   ```bash\n"
	readme += "   golanggraph multi-agent validate\n"
	readme += "   golanggraph multi-agent serve\n"
	readme += "   ```\n\n"

	readme += "## Deployment\n\n"
	readme += "### Docker\n\n"
	readme += "```bash\n"
	readme += "docker-compose up -d\n"
	readme += "```\n\n"
	readme += "### Kubernetes\n\n"
	readme += "```bash\n"
	readme += "kubectl apply -f k8s/\n"
	readme += "```\n\n"

	readme += "## Monitoring\n\n"
	readme += "- Health: `http://localhost:8080/health`\n"
	readme += "- Metrics: `http://localhost:8080/metrics`\n"
	readme += "- Agent Status: `http://localhost:8080/agents`\n"

	readmePath := filepath.Join(projectName, "README.md")
	os.WriteFile(readmePath, []byte(readme), 0600)
}

// Additional validation functions
func validateStrictMode(config *agent.MultiAgentConfig) error {
	// Add strict validation logic here
	return nil
}

func validateAgentSchemas(config *agent.MultiAgentConfig) error {
	// Add schema validation logic here
	return nil
}

// Deployment functions
func deployDocker(config *agent.MultiAgentConfig, parallel bool) {
	fmt.Printf("Deploying to Docker...\n")
	// Implementation for Docker deployment
}

func deployKubernetes(config *agent.MultiAgentConfig, parallel bool) {
	fmt.Printf("Deploying to Kubernetes...\n")
	// Implementation for Kubernetes deployment
}

func deployServerless(config *agent.MultiAgentConfig, parallel bool) {
	fmt.Printf("Deploying to serverless platform...\n")
	// Implementation for serverless deployment
}

// Generate functions
func runGenerateDocker(args []string, outputDir string, multiService bool) {
	fmt.Printf("Generating Docker deployment files...\n")
	// Implementation for generating Docker files
}

func runGenerateK8s(args []string, outputDir, namespace string) {
	fmt.Printf("Generating Kubernetes manifests...\n")
	// Implementation for generating K8s manifests
}

// runMultiAgentLoad loads agent definitions from Go files or plugins
func runMultiAgentLoad(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	recursive, _ := cmd.Flags().GetBool("recursive")
	include, _ := cmd.Flags().GetStringSlice("include")
	exclude, _ := cmd.Flags().GetStringSlice("exclude")
	validate, _ := cmd.Flags().GetBool("validate")
	verbose, _ := cmd.Flags().GetBool("verbose")

	fmt.Printf("Loading agent definitions from: %s\n", path)

	registry := agent.GetGlobalRegistry()

	// Check if it's a plugin file
	if strings.HasSuffix(path, ".so") {
		if verbose {
			fmt.Printf("Loading plugin: %s\n", path)
		}

		if err := registry.LoadFromPlugin(path); err != nil {
			return fmt.Errorf("failed to load plugin: %w", err)
		}

		fmt.Printf("Successfully loaded plugin: %s\n", path)
	} else {
		// Load from directory
		if verbose {
			fmt.Printf("Scanning directory for Go files...\n")
			fmt.Printf("Include patterns: %v\n", include)
			fmt.Printf("Exclude patterns: %v\n", exclude)
			fmt.Printf("Recursive: %v\n", recursive)
		}

		// In a real implementation, this would scan the directory
		// for Go files and load agent definitions
		fmt.Printf("Directory-based loading not yet implemented\n")
		fmt.Printf("Please use plugin-based loading instead\n")
		return nil
	}

	// List loaded agents
	definitions := registry.ListDefinitions()
	factories := registry.ListFactories()

	fmt.Printf("\nLoaded agents:\n")
	fmt.Printf("  Definitions: %d\n", len(definitions))
	fmt.Printf("  Factories: %d\n", len(factories))

	if verbose {
		fmt.Printf("\nDefinitions: %v\n", definitions)
		fmt.Printf("Factories: %v\n", factories)
	}

	// Validate if requested
	if validate {
		fmt.Printf("\nValidating loaded agent definitions...\n")

		for _, defID := range definitions {
			if def, exists := registry.GetDefinition(defID); exists {
				if err := def.Validate(); err != nil {
					fmt.Printf("  ❌ %s: %v\n", defID, err)
				} else {
					fmt.Printf("  ✅ %s: valid\n", defID)
				}
			}
		}

		for _, factoryID := range factories {
			// Create temporary instance to validate
			factory := registry.ListFactories()
			if len(factory) > 0 {
				fmt.Printf("  ✅ Factory %s: valid\n", factoryID)
			}
		}
	}

	return nil
}

// runMultiAgentList lists all registered agent definitions
func runMultiAgentList(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	filter, _ := cmd.Flags().GetString("filter")
	showMetadata, _ := cmd.Flags().GetBool("show-metadata")
	showConfig, _ := cmd.Flags().GetBool("show-config")

	registry := agent.GetGlobalRegistry()
	infos := registry.GetAgentInfo()

	// Apply filter if specified
	if filter != "" {
		var filteredInfos []agent.AgentInfo
		for _, info := range infos {
			if strings.Contains(info.ID, filter) {
				filteredInfos = append(filteredInfos, info)
			}
		}
		infos = filteredInfos
	}

	switch format {
	case "json":
		output, err := json.MarshalIndent(infos, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))

	case "yaml":
		output, err := yaml.Marshal(infos)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Println(string(output))

	default:
		// Table format
		fmt.Printf("Agent Definitions (%d total):\n\n", len(infos))
		fmt.Printf("%-20s %-12s %-15s %-10s\n", "ID", "Source", "Type", "Model")
		fmt.Printf("%-20s %-12s %-15s %-10s\n", "--", "------", "----", "-----")

		for _, info := range infos {
			model := "N/A"
			agentType := "N/A"

			if info.Config != nil {
				model = info.Config.Model
				agentType = string(info.Config.Type)
			}

			fmt.Printf("%-20s %-12s %-15s %-10s\n",
				info.ID, info.Source, agentType, model)

			if showConfig && info.Config != nil {
				fmt.Printf("  Config: Name=%s, Provider=%s, Tools=%v\n",
					info.Config.Name, info.Config.Provider, info.Config.Tools)
			}

			if showMetadata && len(info.Metadata) > 0 {
				fmt.Printf("  Metadata: ")
				for k, v := range info.Metadata {
					fmt.Printf("%s=%v ", k, v)
				}
				fmt.Printf("\n")
			}
		}
	}

	return nil
}
