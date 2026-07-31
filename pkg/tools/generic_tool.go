// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

// ToolExecutor is a function that executes a tool
type ToolExecutor func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// GenericTool is a generic implementation of the Tool interface
type GenericTool struct {
	name        string
	description string
	executor    ToolExecutor
	parameters  map[string]interface{}
	config      map[string]interface{}
}

// NewGenericTool creates a new generic tool
func NewGenericTool(name, description string, executor ToolExecutor, parameters map[string]interface{}) *GenericTool {
	return &GenericTool{
		name:        name,
		description: description,
		executor:    executor,
		parameters:  parameters,
		config:      make(map[string]interface{}),
	}
}

func (t *GenericTool) GetName() string {
	return t.name
}

func (t *GenericTool) GetDescription() string {
	return t.description
}

func (t *GenericTool) GetDefinition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.Function{
			Name:        t.name,
			Description: t.description,
			Parameters:  t.parameters,
		},
	}
}

func (t *GenericTool) Execute(ctx context.Context, args string) (interface{}, error) {
	parsedArgs, err := parseToolArgs(args)
	if err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	return t.executor(ctx, parsedArgs)
}

func (t *GenericTool) Validate(args string) error {
	_, err := parseToolArgs(args)
	if err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	return nil
}

// parseToolArgs unmarshals tool arguments with light SLM-oriented repairs
// (trailing commas, single quotes, missing closing braces).
func parseToolArgs(args string) (map[string]interface{}, error) {
	args = strings.TrimSpace(args)
	if args == "" {
		return map[string]interface{}{}, nil
	}
	var parsedArgs map[string]interface{}
	if err := json.Unmarshal([]byte(args), &parsedArgs); err == nil {
		return parsedArgs, nil
	}
	for _, candidate := range repairJSONCandidates(args) {
		if err := json.Unmarshal([]byte(candidate), &parsedArgs); err == nil {
			return parsedArgs, nil
		}
	}
	return nil, fmt.Errorf("could not parse tool args as JSON")
}

func repairJSONCandidates(raw string) []string {
	raw = strings.TrimSpace(raw)
	out := []string{raw}
	// Strip trailing commas before } or ]
	stripped := raw
	for {
		next := strings.ReplaceAll(stripped, ",}", "}")
		next = strings.ReplaceAll(next, ", }", "}")
		next = strings.ReplaceAll(next, ",]", "]")
		next = strings.ReplaceAll(next, ", ]", "]")
		if next == stripped {
			break
		}
		stripped = next
	}
	out = append(out, stripped)
	if !strings.Contains(raw, `"`) && strings.Contains(raw, `'`) {
		out = append(out, strings.ReplaceAll(raw, `'`, `"`))
		out = append(out, strings.ReplaceAll(stripped, `'`, `"`))
	}
	// Close truncated objects/arrays
	opens := strings.Count(stripped, "{") - strings.Count(stripped, "}")
	opensArr := strings.Count(stripped, "[") - strings.Count(stripped, "]")
	closed := stripped
	for opensArr > 0 {
		closed += "]"
		opensArr--
	}
	for opens > 0 {
		closed += "}"
		opens--
	}
	if closed != stripped {
		out = append(out, closed)
	}
	return out
}

func (t *GenericTool) GetConfig() map[string]interface{} {
	return t.config
}

func (t *GenericTool) SetConfig(config map[string]interface{}) error {
	t.config = config
	return nil
}
