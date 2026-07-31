// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package agent

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/agent/backends"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/tools"
)

// Middleware defines the interface for agent middleware
type Middleware interface {
	// Name returns the name of the middleware
	Name() string
	// BeforeRun is called before the agent execution starts
	BeforeRun(ctx context.Context, agent *BaseAgent, input string) (string, error)
	// AfterRun is called after the agent execution finishes
	AfterRun(ctx context.Context, agent *BaseAgent, result *AgentExecution) (*AgentExecution, error)
}

// BaseMiddleware provides default no-op implementations
type BaseMiddleware struct{}

func (m *BaseMiddleware) Name() string {
	return "base"
}

func (m *BaseMiddleware) BeforeRun(ctx context.Context, agent *BaseAgent, input string) (string, error) {
	return input, nil
}

func (m *BaseMiddleware) AfterRun(ctx context.Context, agent *BaseAgent, result *AgentExecution) (*AgentExecution, error) {
	return result, nil
}

// TodoListMiddleware implements planning capabilities
type TodoListMiddleware struct {
	BaseMiddleware
}

func NewTodoListMiddleware() *TodoListMiddleware {
	return &TodoListMiddleware{}
}

func (m *TodoListMiddleware) Name() string {
	return "todo_list"
}

func (m *TodoListMiddleware) BeforeRun(ctx context.Context, agent *BaseAgent, input string) (string, error) {
	// Inject planning instructions
	planningPrompt := "\n\nPLANNING: Before solving a complex task, you should break it down into steps using the 'write_todos' tool. This helps you stay organized."
	agent.config.SystemPrompt += planningPrompt

	// Register write_todos tool
	writeTodosTool := tools.NewGenericTool(
		"write_todos",
		"Create or update a list of todo items to plan your work.",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			todos, ok := args["todos"].([]interface{})
			if !ok {
				return "", fmt.Errorf("todos argument must be a list")
			}

			var sb strings.Builder
			sb.WriteString("TODO List:\n")
			for i, todo := range todos {
				sb.WriteString(fmt.Sprintf("%d. %v\n", i+1, todo))
			}
			return sb.String(), nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"todos": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
					"description": "List of todo items",
				},
			},
			"required": []string{"todos"},
		},
	)

	agent.toolRegistry.RegisterTool(writeTodosTool)
	agent.config.Tools = append(agent.config.Tools, "write_todos")

	return input, nil
}

// FilesystemMiddleware implements virtual filesystem capabilities with pluggable backends
type FilesystemMiddleware struct {
	BaseMiddleware
	backend backends.BackendProtocol
	// Legacy support - will be removed
	files map[string]string
	mu    sync.RWMutex
}

// NewFilesystemMiddleware creates filesystem middleware with a backend
// For backwards compatibility, if backend is nil, uses in-memory storage
func NewFilesystemMiddleware(backend backends.BackendProtocol) *FilesystemMiddleware {
	return &FilesystemMiddleware{
		backend: backend,
		files:   make(map[string]string), // Legacy fallback
	}
}

func (m *FilesystemMiddleware) Name() string {
	return "filesystem"
}

func (m *FilesystemMiddleware) BeforeRun(ctx context.Context, agent *BaseAgent, input string) (string, error) {
	// Initialize backend if needed
	if m.backend == nil {
		// Legacy mode: use in-memory backend linked to agent state
		m.backend = backends.NewStateBackend(func() map[string]interface{} {
			// Access agent state through a closure
			// This will be properly set up when we have ToolRuntime
			return make(map[string]interface{})
		})
	}

	// Inject filesystem instructions
	fsPrompt := "\n\nFILESYSTEM: You have access to a virtual filesystem with tools: ls, read_file, write_file, edit_file, glob, grep. Use these to manage intermediate results and large context."
	agent.config.SystemPrompt += fsPrompt

	// Register filesystem tools
	m.registerTools(agent)

	return input, nil
}

func (m *FilesystemMiddleware) registerTools(agent *BaseAgent) {
	// ls tool
	lsTool := tools.NewGenericTool(
		"ls",
		"List files in the virtual filesystem.",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			m.mu.RLock()
			defer m.mu.RUnlock()

			if len(m.files) == 0 {
				return "No files found.", nil
			}

			var files []string
			for name := range m.files {
				files = append(files, name)
			}
			return strings.Join(files, "\n"), nil
		},
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	)
	agent.toolRegistry.RegisterTool(lsTool)
	agent.config.Tools = append(agent.config.Tools, "ls")

	// write_file tool
	writeFileTool := tools.NewGenericTool(
		"write_file",
		"Write content to a file.",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			filename, ok := args["filename"].(string)
			if !ok {
				return "", fmt.Errorf("filename must be a string")
			}
			content, ok := args["content"].(string)
			if !ok {
				return "", fmt.Errorf("content must be a string")
			}

			m.mu.Lock()
			defer m.mu.Unlock()
			m.files[filename] = content
			return fmt.Sprintf("File '%s' written successfully.", filename), nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filename": map[string]interface{}{
					"type":        "string",
					"description": "Name of the file",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write",
				},
			},
			"required": []string{"filename", "content"},
		},
	)
	agent.toolRegistry.RegisterTool(writeFileTool)
	agent.config.Tools = append(agent.config.Tools, "write_file")

	// read_file tool
	readFileTool := tools.NewGenericTool(
		"read_file",
		"Read content from a file. Supports pagination.",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			filename, ok := args["filename"].(string)
			if !ok {
				return "", fmt.Errorf("filename must be a string")
			}

			offset := 0
			if val, ok := args["offset"].(float64); ok {
				offset = int(val)
			}
			limit := -1
			if val, ok := args["limit"].(float64); ok {
				limit = int(val)
			}

			m.mu.RLock()
			defer m.mu.RUnlock()
			content, exists := m.files[filename]
			if !exists {
				return "", fmt.Errorf("file '%s' not found", filename)
			}

			lines := strings.Split(content, "\n")
			if offset >= len(lines) {
				return "", fmt.Errorf("offset %d is out of bounds (file has %d lines)", offset, len(lines))
			}

			end := len(lines)
			if limit > 0 {
				end = offset + limit
				if end > len(lines) {
					end = len(lines)
				}
			}

			var sb strings.Builder
			for i := offset; i < end; i++ {
				sb.WriteString(fmt.Sprintf("%d\t%s\n", i+1, lines[i]))
			}

			return sb.String(), nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filename": map[string]interface{}{
					"type":        "string",
					"description": "Name of the file",
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Line number to start reading from (0-indexed)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Number of lines to read",
				},
			},
			"required": []string{"filename"},
		},
	)
	agent.toolRegistry.RegisterTool(readFileTool)
	agent.config.Tools = append(agent.config.Tools, "read_file")

	// edit_file tool
	editFileTool := tools.NewGenericTool(
		"edit_file",
		"Edit a file by replacing a string with a new string.",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			filename, ok := args["filename"].(string)
			if !ok {
				return "", fmt.Errorf("filename must be a string")
			}
			oldString, ok := args["old_string"].(string)
			if !ok {
				return "", fmt.Errorf("old_string must be a string")
			}
			newString, ok := args["new_string"].(string)
			if !ok {
				return "", fmt.Errorf("new_string must be a string")
			}
			replaceAll, _ := args["replace_all"].(bool)

			m.mu.Lock()
			defer m.mu.Unlock()
			content, exists := m.files[filename]
			if !exists {
				return "", fmt.Errorf("file '%s' not found", filename)
			}

			if !strings.Contains(content, oldString) {
				return "", fmt.Errorf("old_string not found in file")
			}

			if replaceAll {
				content = strings.ReplaceAll(content, oldString, newString)
			} else {
				if strings.Count(content, oldString) > 1 {
					return "", fmt.Errorf("old_string found multiple times, use replace_all=true or provide more context")
				}
				content = strings.Replace(content, oldString, newString, 1)
			}

			m.files[filename] = content
			return fmt.Sprintf("File '%s' edited successfully.", filename), nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filename": map[string]interface{}{
					"type":        "string",
					"description": "Name of the file",
				},
				"old_string": map[string]interface{}{
					"type":        "string",
					"description": "String to replace",
				},
				"new_string": map[string]interface{}{
					"type":        "string",
					"description": "New string",
				},
				"replace_all": map[string]interface{}{
					"type":        "boolean",
					"description": "Replace all occurrences",
				},
			},
			"required": []string{"filename", "old_string", "new_string"},
		},
	)
	agent.toolRegistry.RegisterTool(editFileTool)
	agent.config.Tools = append(agent.config.Tools, "edit_file")

	// glob tool
	globTool := tools.NewGenericTool(
		"glob",
		"Find files matching a pattern.",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			pattern, ok := args["pattern"].(string)
			if !ok {
				return "", fmt.Errorf("pattern must be a string")
			}

			m.mu.RLock()
			defer m.mu.RUnlock()

			var matches []string
			for name := range m.files {
				matched, err := filepath.Match(pattern, name)
				if err == nil && matched {
					matches = append(matches, name)
				}
			}
			return strings.Join(matches, "\n"), nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern",
				},
			},
			"required": []string{"pattern"},
		},
	)
	agent.toolRegistry.RegisterTool(globTool)
	agent.config.Tools = append(agent.config.Tools, "glob")

	// grep tool
	grepTool := tools.NewGenericTool(
		"grep",
		"Search for a pattern in files.",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			pattern, ok := args["pattern"].(string)
			if !ok {
				return "", fmt.Errorf("pattern must be a string")
			}
			globPattern, _ := args["glob"].(string)

			m.mu.RLock()
			defer m.mu.RUnlock()

			var sb strings.Builder
			for name, content := range m.files {
				if globPattern != "" {
					matched, _ := filepath.Match(globPattern, name)
					if !matched {
						continue
					}
				}

				if strings.Contains(content, pattern) {
					sb.WriteString(fmt.Sprintf("%s:\n", name))
					lines := strings.Split(content, "\n")
					for i, line := range lines {
						if strings.Contains(line, pattern) {
							sb.WriteString(fmt.Sprintf("%d: %s\n", i+1, line))
						}
					}
					sb.WriteString("\n")
				}
			}
			return sb.String(), nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Text pattern to search for",
				},
				"glob": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern to filter files",
				},
			},
			"required": []string{"pattern"},
		},
	)
	agent.toolRegistry.RegisterTool(grepTool)
	agent.config.Tools = append(agent.config.Tools, "grep")

	// execute tool
	executeTool := tools.NewGenericTool(
		"execute",
		"Execute a shell command (Virtual Mode: Limited support).",
		func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			command, ok := args["command"].(string)
			if !ok {
				return "", fmt.Errorf("command must be a string")
			}
			return fmt.Sprintf("Virtual Filesystem: Command execution is not supported in this mode. Command '%s' was not run.", command), nil
		},
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "Shell command to execute",
				},
			},
			"required": []string{"command"},
		},
	)
	agent.toolRegistry.RegisterTool(executeTool)
	agent.config.Tools = append(agent.config.Tools, "execute")
}

// SubAgentMiddleware enables task delegation to other agents
type SubAgentMiddleware struct {
	BaseMiddleware
	subAgents map[string]Agent
	executor  *SubAgentExecutor
	mu        sync.RWMutex
}

// NewSubAgentMiddleware creates a new SubAgent middleware
func NewSubAgentMiddleware() *SubAgentMiddleware {
	return &SubAgentMiddleware{
		subAgents: make(map[string]Agent),
		executor:  NewSubAgentExecutor(GetGlobalRegistry()),
	}
}

// RegisterSubAgent registers a sub-agent for delegation
func (m *SubAgentMiddleware) RegisterSubAgent(name string, agent Agent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subAgents[name] = agent
}

func (m *SubAgentMiddleware) Name() string {
	return "sub_agent"
}

func (m *SubAgentMiddleware) BeforeRun(ctx context.Context, agent *BaseAgent, input string) (string, error) {
	// Inject sub-agent instructions
	if len(m.subAgents) > 0 {
		var names []string
		for name := range m.subAgents {
			names = append(names, name)
		}
		subAgentPrompt := fmt.Sprintf("\n\nSUB-AGENTS: You can delegate tasks to the following sub-agents: %s. Use the 'delegate_task' tool.", strings.Join(names, ", "))
		agent.config.SystemPrompt += subAgentPrompt

		// Register delegate_task tool
		delegateTool := tools.NewGenericTool(
			"delegate_task",
			"Launch an ephemeral subagent to handle complex, multi-step independent tasks. "+
				"Available agents: "+strings.Join(names, ", ")+". "+
				"Use 'general-purpose' for general tasks. "+
				"Returns a single result.",
			func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				agentName, ok := args["agent_name"].(string)
				if !ok {
					return "", fmt.Errorf("agent_name must be a string")
				}
				task, ok := args["task"].(string)
				if !ok {
					return "", fmt.Errorf("task must be a string")
				}

				subAgent, exists := m.subAgents[agentName]
				if !exists {
					return "", fmt.Errorf("sub-agent '%s' not found", agentName)
				}

				// Execute sub-agent
				result, err := subAgent.Execute(ctx, task)
				if err != nil {
					return "", fmt.Errorf("sub-agent execution failed: %v", err)
				}

				return fmt.Sprintf("Sub-agent '%s' result: %s", agentName, result.Output), nil
			},
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the sub-agent",
					},
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Task description for the sub-agent",
					},
				},
				"required": []string{"agent_name", "task"},
			},
		)
		agent.toolRegistry.RegisterTool(delegateTool)
		agent.config.Tools = append(agent.config.Tools, "delegate_task")
	}

	return input, nil
}

// SummarizationMiddleware implements conversation summarization
type SummarizationMiddleware struct {
	BaseMiddleware
	maxMessages int
}

func NewSummarizationMiddleware(maxMessages int) *SummarizationMiddleware {
	if maxMessages <= 0 {
		maxMessages = 10 // Default
	}
	return &SummarizationMiddleware{
		maxMessages: maxMessages,
	}
}

func (m *SummarizationMiddleware) Name() string {
	return "summarization"
}

func (m *SummarizationMiddleware) BeforeRun(ctx context.Context, agent *BaseAgent, input string) (string, error) {
	// Check conversation history length
	history := agent.conversation.GetMessages()
	if len(history) <= m.maxMessages {
		return input, nil
	}

	// Trigger summarization
	keepCount := m.maxMessages / 2
	if keepCount < 1 {
		keepCount = 1
	}

	messagesToSummarize := history[:len(history)-keepCount]
	messagesToKeep := history[len(history)-keepCount:]

	// Construct prompt for summarization
	var sb strings.Builder
	sb.WriteString("Please summarize the following conversation history concisely, preserving key information and decisions:\n\n")
	for _, msg := range messagesToSummarize {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}

	// Call LLM to summarize
	req := llm.CompletionRequest{
		Model:     agent.config.Model,
		Messages:  []llm.Message{{Role: "user", Content: sb.String()}},
		MaxTokens: 500,
	}

	resp, err := agent.llmManager.Complete(ctx, agent.config.Provider, req)
	if err != nil {
		return input, nil
	}

	if len(resp.Choices) == 0 {
		return input, nil
	}
	summary := resp.Choices[0].Message.Content

	// Update conversation history
	newHistory := []llm.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("Previous conversation summary: %s", summary),
		},
	}
	newHistory = append(newHistory, messagesToKeep...)

	agent.conversation.Clear()
	for _, msg := range newHistory {
		agent.conversation.AddMessage(msg)
	}

	return input, nil
}
