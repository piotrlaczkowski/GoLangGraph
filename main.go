// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ConversationMessage represents a message in conversation history
type ConversationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentRequest represents a generic request to any agent
type AgentRequest struct {
	Message             string                `json:"message,omitempty"`
	CurrentMessage      string                `json:"current_message,omitempty"`
	ConversationHistory []ConversationMessage `json:"conversation_history,omitempty"`
}

// Server represents the HTTP server for the multi-agent system
type Server struct {
	port           int
	ollamaEndpoint string
}

// NewServer creates a new server instance
func NewServer() (*Server, error) {
	// Get configuration from environment
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	ollamaEndpoint := os.Getenv("OLLAMA_ENDPOINT")
	if ollamaEndpoint == "" {
		ollamaEndpoint = "http://localhost:11434"
	}

	return &Server{
		port:           port,
		ollamaEndpoint: ollamaEndpoint,
	}, nil
}

// handleAgentRequest handles requests to specific agents
func (s *Server) handleAgentRequest(agentID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request
		var req AgentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Create response based on agent type (mock implementation)
		var response map[string]interface{}

		switch agentID {
		case "designer":
			response = map[string]interface{}{
				"image":       "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==",
				"description": fmt.Sprintf("🎨 Creative visual design for: %s", req.Message),
			}
		case "interviewer":
			response = map[string]interface{}{
				"response":  fmt.Sprintf("🗣️ Tell me more about your vision for '%s' in your ideal habitat of 2035?", req.Message),
				"completed": false,
			}
		case "highlighter":
			response = map[string]interface{}{
				"highlight": fmt.Sprintf("💡 Key insight: '%s' represents an innovative approach to future living", req.Message),
			}
		case "storymaker":
			response = map[string]interface{}{
				"story": fmt.Sprintf("📖 In 2035, I wake up in my habitat where '%s' has transformed daily life in unexpected ways...", req.Message),
			}
		default:
			http.Error(w, "Unknown agent", http.StatusNotFound)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"agent_id":  agentID,
			"output":    response,
			"success":   true,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}); err != nil {
			log.Printf("❌ Failed to encode response: %v", err)
		}

		log.Printf("✅ Agent %s processed request successfully", agentID)
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":          "healthy",
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"agents":          []string{"designer", "interviewer", "highlighter", "storymaker"},
		"architecture":    "clean_modular_structure",
		"ollama_endpoint": s.ollamaEndpoint,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleCapabilities returns agent capabilities
func (s *Server) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities := map[string]map[string]interface{}{
		"designer": {
			"primary_function": "visual_design_generation",
			"input_types":      []string{"text_descriptions", "design_requirements"},
			"output_types":     []string{"visual_descriptions", "placeholder_images"},
			"file_location":    "agents/designer.go",
		},
		"interviewer": {
			"primary_function": "conversation_facilitation",
			"input_types":      []string{"user_responses", "conversation_history"},
			"output_types":     []string{"follow_up_questions", "conversation_summaries"},
			"file_location":    "agents/interviewer.go",
		},
		"highlighter": {
			"primary_function": "insight_extraction",
			"input_types":      []string{"conversation_history", "text_content"},
			"output_types":     []string{"key_insights", "structured_highlights"},
			"file_location":    "agents/highlighter.go",
		},
		"storymaker": {
			"primary_function": "narrative_creation",
			"input_types":      []string{"conversation_history", "ideation_content"},
			"output_types":     []string{"engaging_stories", "immersive_narratives"},
			"file_location":    "agents/storymaker.go",
		},
	}

	response := map[string]interface{}{
		"system_info": map[string]interface{}{
			"name":           "habitat-2035-ideation",
			"version":        "1.0.0",
			"architecture":   "clean_modular_structure",
			"agents_count":   len(capabilities),
			"structure_note": "Each agent is in its own file for better maintainability",
		},
		"agents": capabilities,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// enableCORS adds CORS headers to responses
func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("📡 %s %s - %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// setupRoutes sets up HTTP routes
func (s *Server) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Apply middleware
	router.Use(s.enableCORS)
	router.Use(s.loggingMiddleware)

	// Health and info endpoints
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
	router.HandleFunc("/capabilities", s.handleCapabilities).Methods("GET")

	// Agent endpoints
	router.HandleFunc("/api/designer", s.handleAgentRequest("designer")).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/interviewer", s.handleAgentRequest("interviewer")).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/highlighter", s.handleAgentRequest("highlighter")).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/storymaker", s.handleAgentRequest("storymaker")).Methods("POST", "OPTIONS")

	return router
}

// Start starts the HTTP server
func (s *Server) Start() error {
	router := s.setupRoutes()

	log.Printf("🚀 Starting server on port %d", s.port)
	log.Printf("🔗 Ollama endpoint: %s", s.ollamaEndpoint)
	log.Println("📍 Available endpoints:")
	log.Println("  • GET  /health - Health check")
	log.Println("  • GET  /capabilities - Agent capabilities")
	log.Println("  • POST /api/designer - Designer agent")
	log.Println("  • POST /api/interviewer - Interviewer agent")
	log.Println("  • POST /api/highlighter - Highlighter agent")
	log.Println("  • POST /api/storymaker - Storymaker agent")

	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), router)
}

func main() {
	log.Println("🏗️  Initializing Habitat 2035 Ideation Multi-Agent System")
	log.Println("📁 Clean Architecture: Each agent in its own file for better maintenance")

	// Create and initialize server
	server, err := NewServer()
	if err != nil {
		log.Fatalf("❌ Failed to create server: %v", err)
	}

	// Start server
	log.Println("🎯 System ready for habitat ideation!")
	if err := server.Start(); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
