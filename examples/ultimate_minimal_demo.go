// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package examples

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/builder"
)

// This demonstrates the ULTIMATE minimal code experience
// Create any type of agent in just 1 line of code!
// Run with: go run ultimate_minimal_demo.go

func RunUltimateMinimalDemo() {
	UltimateMinimalDemo()
}

func UltimateMinimalDemo() {
	fmt.Println("🚀 GoLangGraph: Ultimate Minimal Code Experience")
	fmt.Println("===============================================")
	fmt.Println()

	ctx := context.Background()

	// ========== ONE-LINE AGENT CREATION ==========

	fmt.Println("💫 ONE-LINE AGENT CREATION")
	fmt.Println("===========================")

	// 1. One-line Chat Agent
	fmt.Println("📝 1-Line Chat Agent:")
	chatAgent := builder.OneLineChat("UltraChat")
	response, _ := chatAgent.Execute(ctx, "Hello! Tell me about Go programming.")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// 2. One-line ReAct Agent
	fmt.Println("🧠 1-Line ReAct Agent:")
	reactAgent := builder.OneLineReAct("UltraReAct")
	response, _ = reactAgent.Execute(ctx, "Calculate the square root of 144")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// 3. One-line Tool Agent
	fmt.Println("🔧 1-Line Tool Agent:")
	toolAgent := builder.OneLineTool("UltraTool")
	response, _ = toolAgent.Execute(ctx, "What's the current time?")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// 4. One-line RAG Agent
	fmt.Println("📚 1-Line RAG Agent:")
	ragAgent := builder.OneLineRAG("UltraRAG")
	response, _ = ragAgent.Execute(ctx, "Search for information about Go concurrency")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// ========== SPECIALIZED AGENTS ==========

	fmt.Println("🎯 SPECIALIZED AGENTS (1-Line Each)")
	fmt.Println("====================================")

	// Quick builder instance
	quick := builder.Quick()

	// 5. Research Agent
	fmt.Println("🔍 Research Agent:")
	researcher := quick.Researcher("UltraResearcher")
	response, _ = researcher.Execute(ctx, "Research the benefits of Go programming")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// 6. Writer Agent
	fmt.Println("✍️ Writer Agent:")
	writer := quick.Writer("UltraWriter")
	response, _ = writer.Execute(ctx, "Write a brief introduction to Go programming")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// 7. Analyst Agent
	fmt.Println("📊 Analyst Agent:")
	analyst := quick.Analyst("UltraAnalyst")
	response, _ = analyst.Execute(ctx, "Analyze the performance characteristics of Go")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// 8. Coder Agent
	fmt.Println("💻 Coder Agent:")
	coder := quick.Coder("UltraCoder")
	response, _ = coder.Execute(ctx, "Write a simple Go function to calculate factorial")
	fmt.Printf("   Response: %s\n\n", truncate(response.Output, 100))

	// ========== MULTI-AGENT WORKFLOWS ==========

	fmt.Println("👥 MULTI-AGENT WORKFLOWS")
	fmt.Println("=========================")

	// 9. One-line Pipeline
	fmt.Println("🔄 Sequential Pipeline:")
	pipeline := builder.OneLinePipeline(researcher, writer)
	results, _ := pipeline.Execute(ctx, "Research Go benefits and write a summary")
	fmt.Printf("   Pipeline Results: %d agents executed\n", len(results))
	for i, result := range results {
		outputStr, _ := result.Output.(string)
		fmt.Printf("   Agent %d: %s\n", i+1, truncate(outputStr, 80))
	}
	fmt.Println()

	// 10. One-line Swarm
	fmt.Println("🐝 Parallel Swarm:")
	swarm := builder.OneLineSwarm(analyst, coder)
	results, _ = swarm.Execute(ctx, "Analyze Go performance and write example code")
	fmt.Printf("   Swarm Results: %d agents executed in parallel\n", len(results))
	for i, result := range results {
		outputStr, _ := result.Output.(string)
		fmt.Printf("   Agent %d: %s\n", i+1, truncate(outputStr, 80))
	}
	fmt.Println()

	// ========== ADVANCED BUILDER PATTERNS ==========

	fmt.Println("🏗️ ADVANCED BUILDER PATTERNS")
	fmt.Println("=============================")

	// 11. Custom Configuration
	fmt.Println("⚙️ Custom Configuration:")
	customAgent := builder.Quick().
		WithConfig(&builder.QuickConfig{
			DefaultModel: "gpt-4",
			Temperature:  0.2,
			MaxTokens:    500,
			SystemPrompt: "You are a precise, concise AI assistant.",
		}).
		Chat("CustomChat")
	response, _ = customAgent.Execute(ctx, "Explain Go interfaces briefly")
	fmt.Printf("   Custom Response: %s\n\n", truncate(response.Output, 100))

	// 12. Multi-Agent Coordinator
	fmt.Println("🎯 Multi-Agent Coordinator:")
	coordinator := builder.Quick().Multi()
	coordinator.RegisterAgent("researcher", researcher)
	coordinator.RegisterAgent("writer", writer)
	coordinator.RegisterAgent("analyst", analyst)

	coordResults, _ := coordinator.ExecuteSequential(ctx, []string{"researcher", "analyst", "writer"}, "Create a comprehensive Go programming guide")
	fmt.Printf("   Coordination Results: %d agents in sequence\n", len(coordResults))
	for i, result := range coordResults {
		outputStr, _ := result.Output.(string)
		fmt.Printf("   Step %d: %s\n", i+1, truncate(outputStr, 80))
	}
	fmt.Println()

	// ========== SERVER DEPLOYMENT ==========

	fmt.Println("🚀 SERVER DEPLOYMENT")
	fmt.Println("====================")

	// 13. One-line Server
	fmt.Println("🌐 One-Line Server Creation:")
	server := builder.OneLineServer(8080)
	fmt.Printf("   Server created and ready to start on port 8080\n")
	fmt.Printf("   Features: REST API, WebSocket streaming, agent management\n\n")

	// Note: We're not actually starting the server in this demo
	_ = server

	// ========== SUMMARY ==========

	fmt.Println("✨ SUMMARY: ULTIMATE MINIMAL CODE EXPERIENCE")
	fmt.Println("=============================================")
	fmt.Println()
	fmt.Println("🎯 What you just saw:")
	fmt.Println("• 1-line agent creation for any type (Chat, ReAct, Tool, RAG)")
	fmt.Println("• Specialized agents (Researcher, Writer, Analyst, Coder)")
	fmt.Println("• Multi-agent workflows (Pipeline, Swarm, Coordinator)")
	fmt.Println("• Advanced builder patterns with custom configuration")
	fmt.Println("• One-line server deployment")
	fmt.Println()
	fmt.Println("🚀 Key Benefits:")
	fmt.Println("• Minimal code: 1-line agent creation")
	fmt.Println("• Full LangGraph compatibility")
	fmt.Println("• Production-ready features")
	fmt.Println("• Auto-configured LLM providers")
	fmt.Println("• Built-in tools and persistence")
	fmt.Println("• Visual debugging and monitoring")
	fmt.Println()
	fmt.Println("💡 Next Steps:")
	fmt.Println("• Set OPENAI_API_KEY environment variable for real LLM")
	fmt.Println("• Run with Ollama for local LLM inference")
	fmt.Println("• Deploy to production with the one-line server")
	fmt.Println("• Build complex multi-agent workflows")
	fmt.Println()
	fmt.Println("🎉 GoLangGraph: The most minimal way to build AI agents!")
}

// Helper function to truncate long strings
func truncate(v interface{}, maxLen int) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ========== BONUS: REAL-WORLD EXAMPLES ==========

// Example: Customer Support System
func CustomerSupportExample() {
	fmt.Println("🎯 REAL-WORLD EXAMPLE: Customer Support System")
	fmt.Println("===============================================")

	// Create specialized agents for customer support
	classifier := builder.Quick().Chat("TicketClassifier")
	resolver := builder.Quick().ReAct("IssueResolver")
	escalator := builder.Quick().Tool("EscalationAgent")

	// Create support pipeline
	supportPipeline := builder.OneLinePipeline(classifier, resolver, escalator)

	ctx := context.Background()
	results, err := supportPipeline.Execute(ctx, "Customer complaint: My order hasn't arrived and it's been 2 weeks")

	if err != nil {
		log.Printf("Support pipeline error: %v", err)
		return
	}

	fmt.Printf("Support Pipeline Results: %d steps\n", len(results))
	for i, result := range results {
		outputStr, _ := result.Output.(string)
		fmt.Printf("Step %d: %s\n", i+1, truncate(outputStr, 100))
	}
}

// Example: Content Creation Workflow
func ContentCreationExample() {
	fmt.Println("🎯 REAL-WORLD EXAMPLE: Content Creation Workflow")
	fmt.Println("================================================")

	// Create content creation team
	researcher := builder.Quick().Researcher("ContentResearcher")
	writer := builder.Quick().Writer("ContentWriter")
	editor := builder.Quick().Chat("ContentEditor")

	// Create content swarm for parallel processing
	contentSwarm := builder.OneLineSwarm(researcher, writer, editor)

	ctx := context.Background()
	results, err := contentSwarm.Execute(ctx, "Create a comprehensive blog post about Go programming best practices")

	if err != nil {
		log.Printf("Content creation error: %v", err)
		return
	}

	fmt.Printf("Content Creation Results: %d agents worked in parallel\n", len(results))
	for i, result := range results {
		outputStr, _ := result.Output.(string)
		fmt.Printf("Agent %d: %s\n", i+1, truncate(outputStr, 100))
	}
}

// Example: Data Analysis Pipeline
func DataAnalysisExample() {
	fmt.Println("🎯 REAL-WORLD EXAMPLE: Data Analysis Pipeline")
	fmt.Println("==============================================")

	// Create data analysis team
	dataCollector := builder.Quick().Tool("DataCollector")
	analyst := builder.Quick().Analyst("DataAnalyst")
	reporter := builder.Quick().Writer("ReportWriter")

	// Create analysis pipeline
	analysisPipeline := builder.OneLinePipeline(dataCollector, analyst, reporter)

	ctx := context.Background()
	results, err := analysisPipeline.Execute(ctx, "Analyze Go programming language adoption trends and create a report")

	if err != nil {
		log.Printf("Data analysis error: %v", err)
		return
	}

	fmt.Printf("Data Analysis Results: %d sequential steps\n", len(results))
	for i, result := range results {
		outputStr, _ := result.Output.(string)
		fmt.Printf("Step %d: %s\n", i+1, truncate(outputStr, 100))
	}
}

// Example: Code Review System
func CodeReviewExample() {
	fmt.Println("🎯 REAL-WORLD EXAMPLE: Code Review System")
	fmt.Println("==========================================")

	// Create code review team
	codeAnalyzer := builder.Quick().Coder("CodeAnalyzer")
	securityChecker := builder.Quick().Tool("SecurityChecker")
	reviewer := builder.Quick().Chat("CodeReviewer")

	// Create review coordinator
	coordinator := builder.Quick().Multi()
	coordinator.RegisterAgent("analyzer", codeAnalyzer)
	coordinator.RegisterAgent("security", securityChecker)
	coordinator.RegisterAgent("reviewer", reviewer)

	ctx := context.Background()
	results, err := coordinator.ExecuteSequential(ctx, []string{"analyzer", "security", "reviewer"}, "Review this Go code for best practices and security issues")

	if err != nil {
		log.Printf("Code review error: %v", err)
		return
	}

	fmt.Printf("Code Review Results: %d review steps\n", len(results))
	for i, result := range results {
		outputStr, _ := result.Output.(string)
		fmt.Printf("Review Step %d: %s\n", i+1, truncate(outputStr, 100))
	}
}

// Example: One-Line Production Deployment
func ProductionDeploymentExample() {
	fmt.Println("🎯 REAL-WORLD EXAMPLE: Production Deployment")
	fmt.Println("=============================================")

	// Create production-ready server with all features
	server := builder.Quick().
		WithConfig(&builder.QuickConfig{
			DefaultModel:   "gpt-4",
			Temperature:    0.7,
			MaxTokens:      2000,
			EnableAllTools: true,
			UseMemory:      true,
		}).
		Server(8080)

	fmt.Println("Production server created with:")
	fmt.Println("• REST API endpoints")
	fmt.Println("• WebSocket streaming")
	fmt.Println("• Agent management")
	fmt.Println("• Session persistence")
	fmt.Println("• Health monitoring")
	fmt.Println("• Visual debugging")
	fmt.Println("• All built-in tools")
	fmt.Println("• Memory checkpointing")
	fmt.Println()
	fmt.Println("Ready to serve thousands of concurrent requests!")

	// In production, you would start the server:
	// go server.Start()

	_ = server
}

// Example: Enterprise Multi-Agent System
func EnterpriseExample() {
	fmt.Println("🎯 REAL-WORLD EXAMPLE: Enterprise Multi-Agent System")
	fmt.Println("====================================================")

	// Create enterprise-grade agents
	quick := builder.Quick().WithConfig(&builder.QuickConfig{
		DefaultModel:   "gpt-4",
		Temperature:    0.3,
		MaxTokens:      4000,
		EnableAllTools: true,
		UseMemory:      true,
	})

	// Create department agents
	salesAgent := quick.Chat("SalesAssistant")
	supportAgent := quick.ReAct("SupportAgent")
	devAgent := quick.Coder("DevAssistant")
	analyticsAgent := quick.Analyst("AnalyticsAgent")

	// Create department coordinators
	salesCoord := builder.OneLinePipeline(salesAgent)
	supportCoord := builder.OneLinePipeline(supportAgent)
	devCoord := builder.OneLinePipeline(devAgent)
	analyticsCoord := builder.OneLinePipeline(analyticsAgent)

	fmt.Println("Enterprise system created with:")
	fmt.Println("• Sales department agent")
	fmt.Println("• Customer support agent")
	fmt.Println("• Development assistant")
	fmt.Println("• Analytics agent")
	fmt.Println("• Departmental coordinators")
	fmt.Println("• Enterprise-grade configuration")
	fmt.Println()
	fmt.Println("Each department can handle hundreds of concurrent requests!")

	// Use the coordinators
	_ = salesCoord
	_ = supportCoord
	_ = devCoord
	_ = analyticsCoord
}

// Example: AI-Powered Development Team
func AIDevTeamExample() {
	fmt.Println("🎯 REAL-WORLD EXAMPLE: AI-Powered Development Team")
	fmt.Println("===================================================")

	// Create AI development team
	architect := builder.Quick().Coder("SoftwareArchitect")
	developer := builder.Quick().Coder("Developer")
	tester := builder.Quick().Tool("QATester")
	reviewer := builder.Quick().Chat("CodeReviewer")
	deployer := builder.Quick().Tool("DeploymentAgent")

	// Create development pipeline
	devPipeline := builder.OneLinePipeline(architect, developer, tester, reviewer, deployer)

	ctx := context.Background()
	results, err := devPipeline.Execute(ctx, "Create a microservice for user authentication in Go")

	if err != nil {
		log.Printf("Development pipeline error: %v", err)
		return
	}

	fmt.Printf("AI Development Team Results: %d development phases\n", len(results))
	phases := []string{"Architecture", "Development", "Testing", "Review", "Deployment"}
	for i, result := range results {
		if i < len(phases) {
			outputStr, _ := result.Output.(string)
			fmt.Printf("%s: %s\n", phases[i], truncate(outputStr, 100))
		}
	}

	fmt.Println()
	fmt.Println("Complete software development lifecycle in one pipeline!")
}

// Run all examples
func RunAllExamples() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🌟 REAL-WORLD EXAMPLES SHOWCASE")
	fmt.Println(strings.Repeat("=", 60))

	time.Sleep(1 * time.Second)
	CustomerSupportExample()

	time.Sleep(1 * time.Second)
	ContentCreationExample()

	time.Sleep(1 * time.Second)
	DataAnalysisExample()

	time.Sleep(1 * time.Second)
	CodeReviewExample()

	time.Sleep(1 * time.Second)
	ProductionDeploymentExample()

	time.Sleep(1 * time.Second)
	EnterpriseExample()

	time.Sleep(1 * time.Second)
	AIDevTeamExample()

	fmt.Println("\n🎉 All examples completed successfully!")
	fmt.Println("GoLangGraph: From 1-line agents to enterprise systems!")
}
