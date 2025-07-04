// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - ReAct Agent Implementation

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Tool represents a tool that the agent can use
type Tool struct {
	name        string
	description string
	usage       string
	execute     func(input string) string
}

// Reasoning represents the agent's reasoning process
type Reasoning struct {
	thought     string
	needsAction bool
	action      string
	actionInput string
}

// ReActAgent represents a ReAct (Reasoning and Acting) agent
type ReActAgent struct {
	endpoint string
	model    string
	history  []string
	tools    map[string]*Tool
}

// NewReActAgent creates a new ReAct agent
func NewReActAgent(endpoint, model string) *ReActAgent {
	agent := &ReActAgent{
		endpoint: endpoint,
		model:    model,
		history:  make([]string, 0),
		tools:    make(map[string]*Tool),
	}

	// Initialize tools
	agent.initializeTools()

	return agent
}

// validateConnection checks if Ollama is running and accessible
func (a *ReActAgent) validateConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", a.endpoint+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status: %d", resp.StatusCode)
	}

	return nil
}

// initializeTools sets up the available tools
func (a *ReActAgent) initializeTools() {
	// Calculator tool
	a.tools["calculator"] = &Tool{
		name:        "calculator",
		description: "Performs mathematical calculations including basic arithmetic, square roots, powers, etc.",
		usage:       "calculator: expression (e.g., 'sqrt(144)', '2^3', '10 + 5 * 2')",
		execute:     a.executeCalculator,
	}

	// Unit converter tool
	a.tools["unit_converter"] = &Tool{
		name:        "unit_converter",
		description: "Converts between different units (temperature, length, weight, etc.)",
		usage:       "unit_converter: value unit1 to unit2 (e.g., '100 fahrenheit to celsius')",
		execute:     a.executeUnitConverter,
	}

	// Data analyzer tool
	a.tools["data_analyzer"] = &Tool{
		name:        "data_analyzer",
		description: "Analyzes numerical data (mean, median, mode, standard deviation, etc.)",
		usage:       "data_analyzer: operation on numbers (e.g., 'mean of 1,2,3,4,5')",
		execute:     a.executeDataAnalyzer,
	}
}

// analyzeInput analyzes the user input and determines what action to take
func (a *ReActAgent) analyzeInput(input string) Reasoning {
	input = strings.ToLower(input)

	// Simple rule-based reasoning for demonstration
	// In a real implementation, you'd use the LLM for this

	if strings.Contains(input, "calculate") || strings.Contains(input, "math") ||
		strings.Contains(input, "sqrt") || strings.Contains(input, "square") ||
		strings.Contains(input, "+") || strings.Contains(input, "-") ||
		strings.Contains(input, "*") || strings.Contains(input, "/") {
		return Reasoning{
			thought:     "The user wants to perform a mathematical calculation. I should use the calculator tool.",
			needsAction: true,
			action:      "calculator",
			actionInput: input,
		}
	}

	if strings.Contains(input, "convert") || strings.Contains(input, "fahrenheit") ||
		strings.Contains(input, "celsius") || strings.Contains(input, "to") {
		return Reasoning{
			thought:     "The user wants to convert between units. I should use the unit converter tool.",
			needsAction: true,
			action:      "unit_converter",
			actionInput: input,
		}
	}

	if strings.Contains(input, "mean") || strings.Contains(input, "average") ||
		strings.Contains(input, "median") || strings.Contains(input, "analyze") {
		return Reasoning{
			thought:     "The user wants to analyze some data. I should use the data analyzer tool.",
			needsAction: true,
			action:      "data_analyzer",
			actionInput: input,
		}
	}

	return Reasoning{
		thought:     "This seems like a general question that doesn't require specific tools. I can answer directly.",
		needsAction: false,
		action:      "",
		actionInput: "",
	}
}

// executeAction executes the specified action with the given input
func (a *ReActAgent) executeAction(action, input string) string {
	tool, exists := a.tools[action]
	if !exists {
		return fmt.Sprintf("Error: Tool '%s' not found", action)
	}

	return tool.execute(input)
}

// generateResponse generates the final response based on reasoning and action results
func (a *ReActAgent) generateResponse(input string, reasoning Reasoning, actionResult string) string {
	if !reasoning.needsAction {
		// Use LLM for general questions
		return a.callOllamaForResponse(input)
	}

	// Combine action result with context
	prompt := fmt.Sprintf("The user asked: '%s'\nI used a tool and got this result: %s\nProvide a helpful response that explains the result.", input, actionResult)
	return a.callOllamaForResponse(prompt)
}

// callOllamaForResponse makes a request to Ollama for generating responses
func (a *ReActAgent) callOllamaForResponse(prompt string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	systemPrompt := "You are a helpful AI assistant that can reason and act. Provide clear, concise responses."
	fullPrompt := systemPrompt + "\n\n" + prompt

	reqBody := OllamaRequest{
		Model:  a.model,
		Prompt: fullPrompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Sprintf("Error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("Error calling Ollama: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Ollama returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response: %v", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return fmt.Sprintf("Error unmarshaling response: %v", err)
	}

	return strings.TrimSpace(ollamaResp.Response)
}
