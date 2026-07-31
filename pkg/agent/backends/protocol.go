// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package backends provides filesystem backend abstractions for agent middleware

package backends

import (
	"context"
	"time"
)

// BackendProtocol defines the interface for filesystem backends
// This matches LangGraph's FilesystemMiddleware backend abstraction
type BackendProtocol interface {
	// Read reads file content with optional pagination
	// Returns formatted content (with line numbers in cat -n style)
	Read(ctx context.Context, path string, offset, limit int) (string, error)

	// Write creates or overwrites a file
	Write(ctx context.Context, path string, content string) (*WriteResult, error)

	// Edit performs string replacement in a file
	Edit(ctx context.Context, path string, oldStr, newStr string, replaceAll bool) (*EditResult, error)

	// List lists files in a directory
	List(ctx context.Context, path string) ([]FileInfo, error)

	// Glob finds files matching a pattern
	Glob(ctx context.Context, pattern string, basePath string) ([]string, error)

	// Grep searches for text in files
	Grep(ctx context.Context, pattern string, path string, globFilter string) ([]GrepMatch, error)
}

// SandboxBackendProtocol extends BackendProtocol with command execution
// Only backends implementing this can support the "execute" tool
type SandboxBackendProtocol interface {
	BackendProtocol

	// Execute runs a shell command in a sandboxed environment
	Execute(ctx context.Context, command string) (*ExecuteResult, error)
}

// FileInfo represents file metadata
type FileInfo struct {
	Path       string
	IsDir      bool
	Size       int64
	ModifiedAt time.Time
}

// FileData represents file content with metadata
type FileData struct {
	// Content is the file content split into lines
	Content []string
	// CreatedAt is ISO 8601 timestamp
	CreatedAt string
	// ModifiedAt is ISO 8601 timestamp
	ModifiedAt string
}

// WriteResult contains the result of a write operation
type WriteResult struct {
	// Path is the file path that was written
	Path string
	// FilesUpdate contains state updates for StateBackend
	// nil for other backends
	FilesUpdate map[string]*FileData
	// Error message if operation failed (empty string on success)
	Error string
}

// EditResult contains the result of an edit operation
type EditResult struct {
	// Path is the file path that was edited
	Path string
	// Occurrences is the number of replacements made
	Occurrences int
	// FilesUpdate contains state updates for StateBackend
	// nil for other backends
	FilesUpdate map[string]*FileData
	// Error message if operation failed (empty string on success)
	Error string
}

// GrepMatch represents a grep search result
type GrepMatch struct {
	// Path is the file path
	Path string
	// LineNumber is the line number (1-indexed)
	LineNumber int
	// Line is the matching line content
	Line string
}

// ExecuteResult contains the result of a command execution
type ExecuteResult struct {
	// Output is the combined stdout/stderr
	Output string
	// ExitCode is the command exit code
	ExitCode int
	// Truncated indicates if output was truncated
	Truncated bool
}

// BackendFactory is a function that creates a backend from a ToolRuntime
// This allows backends to access agent state and store at runtime
type BackendFactory func(runtime interface{}) BackendProtocol
