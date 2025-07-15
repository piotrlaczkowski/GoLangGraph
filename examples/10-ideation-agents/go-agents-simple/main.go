// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package main

import (
	"context"
	"log"
	"os"

	"go-agents-simple/agents"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/server"
)

func main() {
	// 🚀 GoLangGraph Auto-Server: From Zero to Production in ~30 Lines!
	log.Println("🚀 GoLangGraph Auto-Server: 4 agents → Full production system!")
	log.Println("📦 Code Reduction: 2000+ lines → ~100 lines (95% reduction!)")

	// Configure auto-server with all production features
	config := server.DefaultAutoServerConfig()
	config.EnableWebUI = true      // Auto-generates chat interface
	config.EnablePlayground = true // Auto-generates API playground
	config.EnableMetricsAPI = true // Auto-generates metrics
	config.SchemaValidation = true // Auto-validates input/output
	config.EnableCORS = true       // Production CORS support
	// Get Ollama endpoint from environment variable or use default
	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://localhost:11434"
	}
	config.OllamaEndpoint = ollamaEndpoint

	// Create auto-server - this replaces 2000+ lines of custom infrastructure!
	autoServer := server.NewAutoServer(config)

	log.Println("📝 Registering agents using existing agent definitions...")

	// Use the existing agent definitions from the agents package
	agentDefinitions := map[string]agent.AgentDefinition{
		"designer":    agents.NewDesignerDefinition(),
		"interviewer": agents.NewInterviewerDefinition(),
		"highlighter": agents.NewHighlighterDefinition(),
		"storymaker":  agents.NewStorymakerDefinition(),
	}

	// Register each agent with the auto-server
	for id, definition := range agentDefinitions {
		if err := autoServer.RegisterAgent(id, definition); err != nil {
			log.Fatalf("Failed to register agent %s: %v", id, err)
		}
		log.Printf("✅ Registered %s agent with comprehensive schema validation", id)
	}

	// Start the server - auto-generates ALL production features:
	// ✅ REST endpoints for all agents with schema validation
	// ✅ Web chat interface with agent switching
	// ✅ API playground with live documentation
	// ✅ Health checks and system monitoring
	// ✅ Metrics collection and reporting
	// ✅ CORS support for web integration
	// ✅ Request/response logging
	// ✅ Error handling and recovery
	// ✅ Conversation management
	// ✅ Streaming response support

	log.Println("🚀 Starting auto-server with full production infrastructure...")
	log.Println("🌐 Web UI: http://localhost:8080/")
	log.Println("🛝 API Playground: http://localhost:8080/playground")
	log.Println("❤️  Health: http://localhost:8080/health")
	log.Println("📊 Metrics: http://localhost:8080/metrics")

	// Start the server (this one line replaces 2000+ lines of infrastructure!)
	// Use background context to run indefinitely
	if err := autoServer.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start auto-server: %v", err)
	}
}
