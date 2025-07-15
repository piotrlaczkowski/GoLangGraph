// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/server"
)

// autoServeCmd represents the auto-serve command
var autoServeCmd = &cobra.Command{
	Use:   "auto-serve [config-file-or-directory]",
	Short: "Auto-generate and serve a multi-agent system with minimal configuration",
	Long: `Auto-serve automatically discovers agent definitions and generates a complete
multi-agent system with REST endpoints, web interfaces, and schema validation.

This command embodies the GoLangGraph vision: define your agents with minimal code
and get a production-ready deployment automatically.

Features:
- Automatic endpoint generation for each agent
- Dynamic web chat interface
- Schema validation and API documentation
- Metrics and monitoring endpoints
- Hot-reload during development
- Production-ready deployment

Examples:
  # Serve agents from current directory
  golanggraph auto-serve

  # Serve agents from a config file
  golanggraph auto-serve agents.yaml

  # Serve agents from a directory with custom port
  golanggraph auto-serve ./agents --port 3000

  # Enable development mode with hot-reload
  golanggraph auto-serve --dev --watch

  # Deploy to production
  golanggraph auto-serve --env production --host 0.0.0.0`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAutoServe,
}

func init() {
	rootCmd.AddCommand(autoServeCmd)

	// Server configuration
	autoServeCmd.Flags().StringP("host", "H", "0.0.0.0", "Host to bind to")
	autoServeCmd.Flags().IntP("port", "p", 8080, "Port to bind to")
	autoServeCmd.Flags().String("base-path", "/api", "Base path for agent endpoints")

	// Feature toggles
	autoServeCmd.Flags().Bool("web-ui", true, "Enable web chat interface")
	autoServeCmd.Flags().Bool("playground", true, "Enable API playground")
	autoServeCmd.Flags().Bool("schema-api", true, "Enable schema API endpoints")
	autoServeCmd.Flags().Bool("metrics", true, "Enable metrics endpoints")
	autoServeCmd.Flags().Bool("cors", true, "Enable CORS support")
	autoServeCmd.Flags().Bool("schema-validation", true, "Enable schema validation")

	// LLM configuration
	autoServeCmd.Flags().String("ollama-endpoint", "http://localhost:11434", "Ollama endpoint URL")
	autoServeCmd.Flags().String("openai-api-key", "", "OpenAI API key")
	autoServeCmd.Flags().String("anthropic-api-key", "", "Anthropic API key")

	// Development features
	autoServeCmd.Flags().Bool("dev", false, "Enable development mode")
	autoServeCmd.Flags().Bool("watch", false, "Watch for file changes and hot-reload")
	autoServeCmd.Flags().Bool("verbose", false, "Enable verbose logging")

	// Production features
	autoServeCmd.Flags().String("env", "development", "Environment (development, staging, production)")
	autoServeCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")
	autoServeCmd.Flags().Duration("timeout", 0, "Request timeout")
	autoServeCmd.Flags().Int64("max-request-size", 0, "Maximum request size in bytes")

	// Docker and deployment
	autoServeCmd.Flags().Bool("generate-dockerfile", false, "Generate Dockerfile for deployment")
	autoServeCmd.Flags().Bool("generate-docker-compose", false, "Generate docker-compose.yml")
	autoServeCmd.Flags().Bool("generate-k8s", false, "Generate Kubernetes manifests")

	// Plugin support
	autoServeCmd.Flags().StringSlice("plugins", []string{}, "Load agent plugins")
	autoServeCmd.Flags().StringSlice("agent-dirs", []string{}, "Additional agent directories")
}

func runAutoServe(cmd *cobra.Command, args []string) error {
	// Parse flags
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	basePath, _ := cmd.Flags().GetString("base-path")
	webUI, _ := cmd.Flags().GetBool("web-ui")
	playground, _ := cmd.Flags().GetBool("playground")
	schemaAPI, _ := cmd.Flags().GetBool("schema-api")
	metrics, _ := cmd.Flags().GetBool("metrics")
	cors, _ := cmd.Flags().GetBool("cors")
	schemaValidation, _ := cmd.Flags().GetBool("schema-validation")
	ollamaEndpoint, _ := cmd.Flags().GetString("ollama-endpoint")
	dev, _ := cmd.Flags().GetBool("dev")
	watch, _ := cmd.Flags().GetBool("watch")
	_, _ = cmd.Flags().GetBool("verbose") // verbose flag parsed but not used in current implementation
	env, _ := cmd.Flags().GetString("env")
	generateDockerfile, _ := cmd.Flags().GetBool("generate-dockerfile")
	generateDockerCompose, _ := cmd.Flags().GetBool("generate-docker-compose")
	generateK8s, _ := cmd.Flags().GetBool("generate-k8s")
	plugins, _ := cmd.Flags().GetStringSlice("plugins")
	agentDirs, _ := cmd.Flags().GetStringSlice("agent-dirs")

	fmt.Printf("🚀 GoLangGraph Auto-Serve - Minimal Code, Maximum Power!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n\n")

	// Determine source path
	sourcePath := "."
	if len(args) > 0 {
		sourcePath = args[0]
	}

	// Create auto-server configuration
	config := &server.AutoServerConfig{
		Host:             host,
		Port:             port,
		BasePath:         basePath,
		EnableWebUI:      webUI,
		EnablePlayground: playground,
		EnableSchemaAPI:  schemaAPI,
		EnableMetricsAPI: metrics,
		EnableCORS:       cors,
		SchemaValidation: schemaValidation,
		OllamaEndpoint:   ollamaEndpoint,
		LLMProviders:     make(map[string]interface{}),
		Middleware:       []string{"cors", "logging", "recovery"},
	}

	// Add LLM providers based on flags
	if openaiKey, _ := cmd.Flags().GetString("openai-api-key"); openaiKey != "" {
		config.LLMProviders["openai"] = map[string]string{"api_key": openaiKey}
	}
	if anthropicKey, _ := cmd.Flags().GetString("anthropic-api-key"); anthropicKey != "" {
		config.LLMProviders["anthropic"] = map[string]string{"api_key": anthropicKey}
	}

	// Adjust config for environment
	if env == "production" {
		config.EnablePlayground = false // Disable playground in production
		fmt.Printf("🔒 Production mode enabled - some debug features disabled\n")
	}

	if dev {
		fmt.Printf("🛠️  Development mode enabled\n")
		if watch {
			fmt.Printf("👀 File watching enabled (hot-reload)\n")
		}
	}

	// Create auto-server
	autoServer := server.NewAutoServer(config)

	// Load agents from various sources
	fmt.Printf("📁 Loading agents from: %s\n", sourcePath)

	// Check if source is a file or directory
	if info, err := os.Stat(sourcePath); err == nil {
		if info.IsDir() {
			// Load from directory
			if err := autoServer.LoadAgentsFromDirectory(sourcePath); err != nil {
				fmt.Printf("⚠️  Warning: %v\n", err)
			}
		} else if filepath.Ext(sourcePath) == ".yaml" || filepath.Ext(sourcePath) == ".yml" {
			// Load from config file
			if err := autoServer.LoadAgentsFromConfig(sourcePath); err != nil {
				return fmt.Errorf("failed to load agents from config: %w", err)
			}
		}
	}

	// Load additional agent directories
	for _, dir := range agentDirs {
		fmt.Printf("📁 Loading additional agents from: %s\n", dir)
		if err := autoServer.LoadAgentsFromDirectory(dir); err != nil {
			fmt.Printf("⚠️  Warning: %v\n", err)
		}
	}

	// Load plugins
	registry := agent.GetGlobalRegistry()
	for _, pluginPath := range plugins {
		fmt.Printf("🔌 Loading plugin: %s\n", pluginPath)
		if err := registry.LoadFromPlugin(pluginPath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to load plugin %s: %v\n", pluginPath, err)
		}
	}

	// Register example agents if none found
	if len(registry.ListDefinitions()) == 0 {
		fmt.Printf("📝 No agents found, creating example agents...\n")
		createExampleAgents(autoServer)
	}

	// Generate deployment files if requested
	if generateDockerfile {
		if err := generateDockerfileForProject(sourcePath); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate Dockerfile: %v\n", err)
		} else {
			fmt.Printf("🐳 Generated Dockerfile\n")
		}
	}

	if generateDockerCompose {
		if err := generateDockerComposeForProject(sourcePath, config); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate docker-compose.yml: %v\n", err)
		} else {
			fmt.Printf("🐳 Generated docker-compose.yml\n")
		}
	}

	if generateK8s {
		if err := generateKubernetesManifests(sourcePath, config); err != nil {
			fmt.Printf("⚠️  Warning: Failed to generate Kubernetes manifests: %v\n", err)
		} else {
			fmt.Printf("☸️  Generated Kubernetes manifests\n")
		}
	}

	// Print configuration summary
	fmt.Printf("\n🔧 Configuration:\n")
	fmt.Printf("   Host: %s\n", host)
	fmt.Printf("   Port: %d\n", port)
	fmt.Printf("   Base Path: %s\n", basePath)
	fmt.Printf("   Environment: %s\n", env)
	fmt.Printf("   Ollama: %s\n", ollamaEndpoint)
	fmt.Printf("   Features: ")
	features := []string{}
	if webUI {
		features = append(features, "WebUI")
	}
	if playground {
		features = append(features, "Playground")
	}
	if schemaAPI {
		features = append(features, "Schema API")
	}
	if metrics {
		features = append(features, "Metrics")
	}
	if schemaValidation {
		features = append(features, "Validation")
	}
	fmt.Printf("%v\n\n", features)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Printf("\n🛑 Shutdown signal received, stopping server...\n")
		cancel()
	}()

	// Start the server
	fmt.Printf("🎉 Starting auto-generated multi-agent system!\n")
	fmt.Printf("═══════════════════════════════════════════════════════════\n")

	// Print quick access URLs
	baseURL := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf("\n🌐 Quick Access URLs:\n")
	if webUI {
		fmt.Printf("   💬 Chat Interface: %s/chat\n", baseURL)
	}
	if playground {
		fmt.Printf("   🏗️  API Playground: %s/playground\n", baseURL)
	}
	fmt.Printf("   📋 System Health: %s/health\n", baseURL)
	fmt.Printf("   🤖 List Agents: %s/agents\n", baseURL)
	if schemaAPI {
		fmt.Printf("   📄 API Schemas: %s/schemas\n", baseURL)
	}
	if metrics {
		fmt.Printf("   📊 Metrics: %s/metrics\n", baseURL)
	}
	fmt.Printf("   🔧 Debug: %s/debug\n", baseURL)
	fmt.Printf("\n")

	if err := autoServer.Start(ctx); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	fmt.Printf("✅ Server stopped gracefully\n")
	return nil
}

// createExampleAgents creates example agents if none are found
func createExampleAgents(autoServer *server.AutoServer) {
	// Create a simple chat agent
	chatConfig := agent.DefaultAgentConfig()
	chatConfig.ID = "chat"
	chatConfig.Name = "Chat Agent"
	chatConfig.Type = agent.AgentTypeChat
	chatConfig.SystemPrompt = "You are a helpful AI assistant. Provide clear and concise responses."
	chatDefinition := agent.NewBaseAgentDefinition(chatConfig)

	autoServer.RegisterAgent("chat", chatDefinition)

	// Create a ReAct agent
	reactConfig := agent.DefaultAgentConfig()
	reactConfig.ID = "react"
	reactConfig.Name = "ReAct Agent"
	reactConfig.Type = agent.AgentTypeReAct
	reactConfig.SystemPrompt = "You are a reasoning agent that can think and act. Break down complex problems step by step."
	reactConfig.Tools = []string{"calculator", "web_search"}
	reactDefinition := agent.NewBaseAgentDefinition(reactConfig)

	autoServer.RegisterAgent("react", reactDefinition)

	// Create a tool agent
	toolConfig := agent.DefaultAgentConfig()
	toolConfig.ID = "tools"
	toolConfig.Name = "Tool Agent"
	toolConfig.Type = agent.AgentTypeTool
	toolConfig.SystemPrompt = "You are a specialized agent that excels at using tools to accomplish tasks."
	toolConfig.Tools = []string{"file_read", "file_write", "shell", "http"}
	toolDefinition := agent.NewBaseAgentDefinition(toolConfig)

	autoServer.RegisterAgent("tools", toolDefinition)

	fmt.Printf("   ✅ Created 3 example agents: chat, react, tools\n")
}

// generateDockerfileForProject generates a Dockerfile for the project
func generateDockerfileForProject(projectPath string) error {
	dockerfileContent := `# Auto-generated Dockerfile by GoLangGraph
FROM golang:1.23-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git ca-certificates

COPY . .
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main", "auto-serve", "--host", "0.0.0.0", "--port", "8080"]
`

	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	return os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)
}

// generateDockerComposeForProject generates a docker-compose.yml
func generateDockerComposeForProject(projectPath string, config *server.AutoServerConfig) error {
	composeContent := fmt.Sprintf(`# Auto-generated docker-compose.yml by GoLangGraph
version: '3.8'

services:
  golanggraph-agents:
    build: .
    ports:
      - "%d:8080"
    environment:
      - OLLAMA_ENDPOINT=%s
    volumes:
      - ./data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    networks:
      - golanggraph-network

  # Optional: Include Ollama service
  # ollama:
  #   image: ollama/ollama:latest
  #   ports:
  #     - "11434:11434"
  #   volumes:
  #     - ollama-data:/root/.ollama
  #   restart: unless-stopped
  #   networks:
  #     - golanggraph-network

networks:
  golanggraph-network:
    driver: bridge

volumes:
  ollama-data:
`, config.Port, config.OllamaEndpoint)

	composePath := filepath.Join(projectPath, "docker-compose.yml")
	return os.WriteFile(composePath, []byte(composeContent), 0644)
}

// generateKubernetesManifests generates Kubernetes deployment manifests
func generateKubernetesManifests(projectPath string, config *server.AutoServerConfig) error {
	manifestContent := fmt.Sprintf(`# Auto-generated Kubernetes manifests by GoLangGraph
apiVersion: apps/v1
kind: Deployment
metadata:
  name: golanggraph-agents
  labels:
    app: golanggraph-agents
spec:
  replicas: 2
  selector:
    matchLabels:
      app: golanggraph-agents
  template:
    metadata:
      labels:
        app: golanggraph-agents
    spec:
      containers:
      - name: golanggraph-agents
        image: golanggraph-agents:latest
        ports:
        - containerPort: 8080
        env:
        - name: OLLAMA_ENDPOINT
          value: "%s"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: golanggraph-agents-service
spec:
  selector:
    app: golanggraph-agents
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: golanggraph-agents-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: agents.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: golanggraph-agents-service
            port:
              number: 80
`, config.OllamaEndpoint)

	manifestPath := filepath.Join(projectPath, "k8s-manifests.yaml")
	return os.WriteFile(manifestPath, []byte(manifestContent), 0644)
}
