// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// handleHealth handles health check requests
func (as *AutoServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":            "healthy",
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
		"agents":            as.registry.ListDefinitions(),
		"agent_count":       len(as.agentInstances),
		"schema_validation": as.config.SchemaValidation,
		"ollama_endpoint":   as.config.OllamaEndpoint,
		"response_time":     "fast",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleCapabilities handles system capabilities requests
func (as *AutoServer) handleCapabilities(w http.ResponseWriter, r *http.Request) {
	capabilities := map[string]interface{}{
		"agents":        as.getAgentCapabilities(),
		"llm_providers": as.llmManager.ListProviders(),
		"tools":         as.toolRegistry.ListTools(),
		"features": map[string]bool{
			"web_ui":            as.config.EnableWebUI,
			"playground":        as.config.EnablePlayground,
			"schema_api":        as.config.EnableSchemaAPI,
			"metrics_api":       as.config.EnableMetricsAPI,
			"streaming":         true,
			"conversations":     true,
			"schema_validation": as.config.SchemaValidation,
		},
		"system_info": map[string]interface{}{
			"version":     "1.0.0",
			"framework":   "GoLangGraph",
			"server_type": "auto-generated",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(capabilities)
}

// handleListAgents handles agent listing requests
func (as *AutoServer) handleListAgents(w http.ResponseWriter, r *http.Request) {
	agents := make([]map[string]interface{}, 0)

	for agentID, metadata := range as.agentMetadata {
		agent := map[string]interface{}{
			"id":          agentID,
			"name":        metadata["name"],
			"type":        metadata["type"],
			"description": fmt.Sprintf("AI agent for %s tasks", agentID),
			"endpoint":    fmt.Sprintf("%s/%s", as.config.BasePath, agentID),
			"schema":      fmt.Sprintf("/schemas/%s", agentID),
			"status":      "active",
			"metadata":    metadata,
		}
		agents = append(agents, agent)
	}

	response := map[string]interface{}{
		"agents":      agents,
		"total_count": len(agents),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAgentInfo handles individual agent information requests
func (as *AutoServer) handleAgentInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agentId"]

	metadata, exists := as.agentMetadata[agentID]
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	agent := as.agentInstances[agentID]
	config := agent.GetConfig()

	// Create description from system prompt or fallback
	description := config.SystemPrompt
	if description == "" {
		description = fmt.Sprintf("%s agent using %s model", config.Type, config.Model)
	}

	info := map[string]interface{}{
		"id":          agentID,
		"name":        config.Name,
		"description": description,
		"endpoint":    fmt.Sprintf("%s/%s", as.config.BasePath, agentID),
		"schema":      config,
		"config":      config,
		"metadata":    metadata,
		"endpoints": map[string]string{
			"execute":      fmt.Sprintf("%s/%s", as.config.BasePath, agentID),
			"stream":       fmt.Sprintf("%s/%s/stream", as.config.BasePath, agentID),
			"conversation": fmt.Sprintf("%s/%s/conversation", as.config.BasePath, agentID),
			"status":       fmt.Sprintf("%s/%s/status", as.config.BasePath, agentID),
		},
		"capabilities": map[string]interface{}{
			"streaming":     true,
			"conversations": true,
			"tools":         config.Tools,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// createAgentHandler creates a handler for agent execution
func (as *AutoServer) createAgentHandler(agentID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, exists := as.agentInstances[agentID]
		if !exists {
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}

		// Parse request body
		var requestData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Execute agent
		start := time.Now()
		ctx := context.Background()

		// Convert requestData to string
		var input string
		if message, ok := requestData["message"].(string); ok {
			input = message
		} else if inputStr, ok := requestData["input"].(string); ok {
			input = inputStr
		} else {
			// Convert entire request to JSON string as fallback
			if inputBytes, err := json.Marshal(requestData); err == nil {
				input = string(inputBytes)
			} else {
				input = "No input provided"
			}
		}

		result, err := agent.Execute(ctx, input)
		if err != nil {
			response := map[string]interface{}{
				"success":         false,
				"agent_id":        agentID,
				"error":           err.Error(),
				"processing_time": time.Since(start).String(),
				"timestamp":       time.Now().UTC().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Get output
		output := result.Output

		response := map[string]interface{}{
			"success":         true,
			"agent_id":        agentID,
			"output":          output,
			"schema_valid":    true, // TODO: Implement schema validation
			"processing_time": time.Since(start).String(),
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// createAgentStreamHandler creates a handler for streaming agent responses
func (as *AutoServer) createAgentStreamHandler(agentID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, exists := as.agentInstances[agentID]
		if !exists {
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Parse request body
		var requestData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Stream response
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Execute agent (simplified streaming implementation)
		ctx := context.Background()

		// Convert requestData to string
		var input string
		if message, ok := requestData["message"].(string); ok {
			input = message
		} else if inputStr, ok := requestData["input"].(string); ok {
			input = inputStr
		} else {
			// Convert entire request to JSON string as fallback
			if inputBytes, err := json.Marshal(requestData); err == nil {
				input = string(inputBytes)
			} else {
				input = "No input provided"
			}
		}

		result, err := agent.Execute(ctx, input)
		if err != nil {
			fmt.Fprintf(w, "data: {\"error\": \"%s\"}\n\n", err.Error())
			flusher.Flush()
			return
		}

		output := result.Output
		responseData, _ := json.Marshal(map[string]interface{}{
			"success":  true,
			"agent_id": agentID,
			"output":   output,
			"complete": true,
		})

		fmt.Fprintf(w, "data: %s\n\n", responseData)
		flusher.Flush()
	}
}

// createConversationHandler creates a handler for conversation management
func (as *AutoServer) createConversationHandler(agentID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, exists := as.agentInstances[agentID]
		if !exists {
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case "GET":
			// Get conversation history
			conversation := agent.GetConversation()
			response := map[string]interface{}{
				"agent_id":      agentID,
				"conversation":  conversation,
				"message_count": len(conversation),
				"timestamp":     time.Now().UTC().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case "POST":
			// Add to conversation
			var requestData map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// This would add to conversation - simplified implementation
			response := map[string]interface{}{
				"success":   true,
				"agent_id":  agentID,
				"added":     true,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		case "DELETE":
			// Clear conversation
			agent.ClearConversation()
			response := map[string]interface{}{
				"success":   true,
				"agent_id":  agentID,
				"cleared":   true,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// createStatusHandler creates a handler for agent status
func (as *AutoServer) createStatusHandler(agentID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		agent, exists := as.agentInstances[agentID]
		if !exists {
			http.Error(w, "Agent not found", http.StatusNotFound)
			return
		}

		status := map[string]interface{}{
			"agent_id":    agentID,
			"status":      "healthy",
			"is_running":  agent.IsRunning(),
			"config":      agent.GetConfig(),
			"uptime":      "unknown", // TODO: Track uptime
			"last_active": time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}

// handleChatInterface serves the web chat interface
func (as *AutoServer) handleChatInterface(w http.ResponseWriter, r *http.Request) {
	// Generate dynamic chat interface based on available agents
	agents := make([]map[string]string, 0)
	for agentID, metadata := range as.agentMetadata {
		agents = append(agents, map[string]string{
			"id":   agentID,
			"name": fmt.Sprintf("%v", metadata["name"]),
		})
	}

	chatHTML := as.generateChatHTML(agents)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(chatHTML)); err != nil {
		as.logger.WithError(err).Error("Failed to write chat HTML response")
	}
}

// handlePlayground serves the API playground interface
func (as *AutoServer) handlePlayground(w http.ResponseWriter, r *http.Request) {
	playgroundHTML := as.generatePlaygroundHTML()
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(playgroundHTML)); err != nil {
		as.logger.WithError(err).Error("Failed to write playground HTML response")
	}
}

// handleDebug serves the debug interface
func (as *AutoServer) handleDebug(w http.ResponseWriter, r *http.Request) {
	debugHTML := as.generateDebugHTML()
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(debugHTML)); err != nil {
		as.logger.WithError(err).Error("Failed to write debug HTML response")
	}
}

// handleSchemas handles schema-related requests
func (as *AutoServer) handleSchemas(w http.ResponseWriter, r *http.Request) {
	schemas := make(map[string]interface{})

	// Generate schemas for each agent
	for agentID, metadata := range as.agentMetadata {
		schemas[agentID] = as.generateAgentSchema(agentID, metadata)
	}

	response := map[string]interface{}{
		"schemas":   schemas,
		"count":     len(schemas),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAgentSchema handles individual agent schema requests
func (as *AutoServer) handleAgentSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agentId"]

	metadata, exists := as.agentMetadata[agentID]
	if !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	schema := as.generateAgentSchema(agentID, metadata)
	response := map[string]interface{}{
		"agent_id":  agentID,
		"schema":    schema,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleValidateSchema handles schema validation requests
func (as *AutoServer) handleValidateSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agentId"]

	if _, exists := as.agentMetadata[agentID]; !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simplified validation - in real implementation would use JSON schema
	validationType := r.URL.Query().Get("type")
	if validationType == "" {
		validationType = "input"
	}

	response := map[string]interface{}{
		"valid":     true,
		"agent_id":  agentID,
		"type":      validationType,
		"errors":    []string{},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMetrics handles system metrics requests
func (as *AutoServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(as.startTime)

	metrics := map[string]interface{}{
		"total_agents":  len(as.agentInstances),
		"active_agents": len(as.agentInstances),
		"requests":      as.requestCount,
		"uptime":        uptime.String(),
		"system": map[string]interface{}{
			"total_agents":   len(as.agentInstances),
			"active_agents":  len(as.agentInstances),
			"total_requests": as.requestCount,
			"uptime":         uptime.String(),
		},
		"agents":    make(map[string]interface{}),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleAgentMetrics handles agent-specific metrics requests
func (as *AutoServer) handleAgentMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agentId"]

	if _, exists := as.agentInstances[agentID]; !exists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	metrics := map[string]interface{}{
		"agent_id":    agentID,
		"requests":    0, // TODO: Track metrics
		"errors":      0,
		"avg_latency": "0ms",
		"last_active": time.Now().UTC().Format(time.RFC3339),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// Helper methods

func (as *AutoServer) getAgentCapabilities() []map[string]interface{} {
	capabilities := make([]map[string]interface{}, 0)

	for agentID, metadata := range as.agentMetadata {
		capability := map[string]interface{}{
			"id":           agentID,
			"name":         metadata["name"],
			"type":         metadata["type"],
			"tools":        metadata["tools"],
			"streaming":    true,
			"conversation": true,
		}
		capabilities = append(capabilities, capability)
	}

	return capabilities
}

func (as *AutoServer) generateAgentSchema(agentID string, metadata map[string]interface{}) map[string]interface{} {
	// Generate a basic schema based on agent type
	agentType := fmt.Sprintf("%v", metadata["type"])

	schema := map[string]interface{}{
		"input": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Input message for the agent",
					"minLength":   1,
					"maxLength":   1000,
				},
			},
			"required": []string{"message"},
		},
		"output": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"response": map[string]interface{}{
					"type":        "string",
					"description": "Agent response",
				},
			},
			"required": []string{"response"},
		},
	}

	// Customize schema based on agent type
	switch agentType {
	case "chat":
		// Chat agents might support conversation history
		inputProps := schema["input"].(map[string]interface{})["properties"].(map[string]interface{})
		inputProps["conversation_history"] = map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"role":    map[string]interface{}{"type": "string"},
					"content": map[string]interface{}{"type": "string"},
				},
			},
		}
	case "react":
		// ReAct agents might support tools
		outputProps := schema["output"].(map[string]interface{})["properties"].(map[string]interface{})
		outputProps["tool_calls"] = map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
			},
		}
	}

	return schema
}

func (as *AutoServer) generateChatHTML(agents []map[string]string) string {
	agentsJSON, _ := json.Marshal(agents)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoLangGraph Multi-Agent Chat</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #f0f2f5; }
        .chat-container { max-width: 900px; margin: 20px auto; background: white; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 20px; border-radius: 12px 12px 0 0; }
        .agent-selector { display: flex; gap: 10px; margin-top: 15px; }
        .agent-btn { padding: 8px 16px; background: rgba(255,255,255,0.2); border: none; border-radius: 20px; color: white; cursor: pointer; transition: all 0.3s; }
        .agent-btn.active { background: rgba(255,255,255,0.3); }
        .chat-messages { height: 400px; overflow-y: auto; padding: 20px; }
        .message { margin-bottom: 15px; display: flex; align-items: flex-start; gap: 10px; }
        .message.user { flex-direction: row-reverse; }
        .message-avatar { width: 40px; height: 40px; border-radius: 50%%; display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 12px; }
        .message.user .message-avatar { background: #667eea; }
        .message.agent .message-avatar { background: #764ba2; }
        .message-bubble { max-width: 70%%; padding: 12px 16px; border-radius: 18px; word-wrap: break-word; }
        .message.user .message-bubble { background: #667eea; color: white; }
        .message.agent .message-bubble { background: #e9ecef; color: #333; }
        .input-container { padding: 20px; border-top: 1px solid #e9ecef; }
        .chat-form { display: flex; gap: 10px; }
        .chat-input { flex: 1; padding: 12px 16px; border: 1px solid #ddd; border-radius: 25px; outline: none; }
        .send-btn { padding: 12px 24px; background: #667eea; color: white; border: none; border-radius: 25px; cursor: pointer; }
        .status-bar { background: #e9ecef; padding: 8px 20px; font-size: 12px; color: #6c757d; text-align: center; }
    </style>
</head>
<body>
    <div class="chat-container">
        <div class="header">
            <h1>🤖 GoLangGraph Multi-Agent System</h1>
            <p>Auto-generated chat interface with dynamic agent discovery</p>
            <div class="agent-selector" id="agentSelector">
                <!-- Agents will be populated dynamically -->
            </div>
        </div>
        <div class="chat-messages" id="chatMessages">
            <div class="message agent">
                <div class="message-avatar">AI</div>
                <div class="message-bubble">Welcome! Select an agent and start chatting. This interface was auto-generated by GoLangGraph!</div>
            </div>
        </div>
        <div class="input-container">
            <form class="chat-form" onsubmit="sendMessage(event)">
                <input type="text" class="chat-input" id="messageInput" placeholder="Type your message..." autocomplete="off">
                <button type="submit" class="send-btn" id="sendBtn">Send</button>
            </form>
        </div>
        <div class="status-bar" id="statusBar">System loading...</div>
    </div>

    <script>
        const agents = %s;
        let currentAgent = null;
        let conversationHistory = [];

        // Initialize
        document.addEventListener('DOMContentLoaded', function() {
            initializeAgents();
            checkSystemStatus();
        });

        function initializeAgents() {
            const selector = document.getElementById('agentSelector');
            agents.forEach((agent, index) => {
                const btn = document.createElement('button');
                btn.className = 'agent-btn' + (index === 0 ? ' active' : '');
                btn.textContent = agent.name;
                btn.onclick = () => selectAgent(agent.id, btn);
                selector.appendChild(btn);

                if (index === 0) {
                    currentAgent = agent.id;
                }
            });
        }

        function selectAgent(agentId, btn) {
            document.querySelectorAll('.agent-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            currentAgent = agentId;
            addSystemMessage('Switched to ' + agentId + ' agent');
        }

        async function sendMessage(event) {
            event.preventDefault();

            const input = document.getElementById('messageInput');
            const message = input.value.trim();

            if (!message || !currentAgent) return;

            addUserMessage(message);
            input.value = '';

            try {
                const response = await fetch('%s/' + currentAgent, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ message: message })
                });

                const data = await response.json();

                if (data.success) {
                    addAgentMessage(data.output.response || JSON.stringify(data.output));
                } else {
                    addErrorMessage('Error: ' + (data.error || 'Unknown error'));
                }
            } catch (error) {
                addErrorMessage('Connection error: ' + error.message);
            }
        }

        function addUserMessage(message) {
            addMessage('user', 'YOU', message);
        }

        function addAgentMessage(message) {
            addMessage('agent', currentAgent.toUpperCase().substr(0,3), message);
        }

        function addSystemMessage(message) {
            addMessage('agent', 'SYS', message);
        }

        function addErrorMessage(message) {
            addMessage('agent', 'ERR', message);
        }

        function addMessage(type, avatar, text) {
            const container = document.getElementById('chatMessages');
            const div = document.createElement('div');
            div.className = 'message ' + type;
            div.innerHTML = '<div class="message-avatar">' + avatar + '</div><div class="message-bubble">' + text + '</div>';
            container.appendChild(div);
            container.scrollTop = container.scrollHeight;
        }

        async function checkSystemStatus() {
            try {
                const response = await fetch('/health');
                const data = await response.json();
                document.getElementById('statusBar').textContent = '✅ System online • Agents: ' + data.agent_count;
            } catch (error) {
                document.getElementById('statusBar').textContent = '❌ Connection failed';
            }
        }
    </script>
</body>
</html>`, string(agentsJSON), as.config.BasePath)
}

func (as *AutoServer) generatePlaygroundHTML() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoLangGraph API Playground</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; }
        .endpoint { margin-bottom: 20px; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .method { display: inline-block; padding: 4px 8px; border-radius: 3px; color: white; font-weight: bold; }
        .post { background: #28a745; }
        .get { background: #007bff; }
        textarea { width: 100%; height: 100px; margin: 10px 0; }
        button { padding: 10px 20px; background: #007bff; color: white; border: none; border-radius: 5px; cursor: pointer; }
        .response { background: #f8f9fa; padding: 10px; border-radius: 5px; margin-top: 10px; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 GoLangGraph API Playground</h1>
        <p>Auto-generated API testing interface</p>

        <div class="endpoint">
            <h3><span class="method get">GET</span> /health</h3>
            <button onclick="testEndpoint('/health', 'GET')">Test Health Check</button>
            <div id="health-response" class="response"></div>
        </div>

        <div class="endpoint">
            <h3><span class="method get">GET</span> /agents</h3>
            <button onclick="testEndpoint('/agents', 'GET')">List Agents</button>
            <div id="agents-response" class="response"></div>
        </div>

        <div class="endpoint">
            <h3><span class="method post">POST</span> /api/{agentId}</h3>
            <input type="text" id="agent-id" placeholder="Agent ID" value="">
            <textarea id="agent-input" placeholder='{"message": "Hello, world!"}'></textarea>
            <button onclick="testAgentEndpoint()">Test Agent</button>
            <div id="agent-response" class="response"></div>
        </div>
    </div>

    <script>
        async function testEndpoint(path, method, body) {
            const options = { method };
            if (body) {
                options.headers = { 'Content-Type': 'application/json' };
                options.body = body;
            }

            try {
                const response = await fetch(path, options);
                const data = await response.json();
                return JSON.stringify(data, null, 2);
            } catch (error) {
                return 'Error: ' + error.message;
            }
        }

        async function testAgentEndpoint() {
            const agentId = document.getElementById('agent-id').value;
            const input = document.getElementById('agent-input').value;

            if (!agentId) {
                document.getElementById('agent-response').textContent = 'Please enter an agent ID';
                return;
            }

            const result = await testEndpoint('/api/' + agentId, 'POST', input);
            document.getElementById('agent-response').textContent = result;
        }

        // Auto-populate available agents
        fetch('/agents').then(r => r.json()).then(data => {
            if (data.agents && data.agents.length > 0) {
                document.getElementById('agent-id').value = data.agents[0].id;
            }
        });
    </script>
</body>
</html>`
}

func (as *AutoServer) generateDebugHTML() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoLangGraph Debug Interface</title>
    <style>
        body { font-family: monospace; margin: 20px; background: #1e1e1e; color: #d4d4d4; }
        .container { max-width: 1200px; margin: 0 auto; }
        .section { margin-bottom: 30px; }
        .logs { background: #252526; padding: 15px; border-radius: 5px; height: 300px; overflow-y: auto; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background: #2d2d30; border-radius: 5px; }
        button { padding: 8px 16px; background: #0e639c; color: white; border: none; border-radius: 3px; cursor: pointer; margin: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔧 GoLangGraph Debug Interface</h1>

        <div class="section">
            <h2>System Metrics</h2>
            <div id="metrics">
                <div class="metric">Agents: <span id="agent-count">-</span></div>
                <div class="metric">Uptime: <span id="uptime">-</span></div>
                <div class="metric">Requests: <span id="requests">-</span></div>
            </div>
        </div>

        <div class="section">
            <h2>System Logs</h2>
            <button onclick="refreshLogs()">Refresh</button>
            <button onclick="clearLogs()">Clear</button>
            <div class="logs" id="logs">System debug interface ready...</div>
        </div>

        <div class="section">
            <h2>Agent Status</h2>
            <div id="agent-status"></div>
        </div>
    </div>

    <script>
        function refreshLogs() {
            document.getElementById('logs').innerHTML += '[' + new Date().toISOString() + '] Debug refresh triggered\n';
        }

        function clearLogs() {
            document.getElementById('logs').innerHTML = '';
        }

        // Update metrics periodically
        setInterval(async () => {
            try {
                const health = await fetch('/health').then(r => r.json());
                document.getElementById('agent-count').textContent = health.agent_count || 0;

                const agents = await fetch('/agents').then(r => r.json());
                const statusHTML = agents.agents.map(a =>
                    '<div class="metric">' + a.id + ': ' + a.status + '</div>'
                ).join('');
                document.getElementById('agent-status').innerHTML = statusHTML;
            } catch (error) {
                console.error('Failed to update metrics:', error);
            }
        }, 5000);
    </script>
</body>
</html>`
}
