// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package backends provides filesystem backend implementations

package backends

import (
	"context"
	"strings"
)

// CompositeBackend routes different paths to different backends
// This matches LangGraph's CompositeBackend pattern for mixed storage
// Example: /memories/* -> StoreBackend (persistent), everything else -> StateBackend (ephemeral)
type CompositeBackend struct {
	defaultBackend BackendProtocol
	routes         []route
}

type route struct {
	prefix  string
	backend BackendProtocol
}

// NewCompositeBackend creates a new composite backend with a default
func NewCompositeBackend(defaultBackend BackendProtocol) *CompositeBackend {
	return &CompositeBackend{
		defaultBackend: defaultBackend,
		routes:         make([]route, 0),
	}
}

// AddRoute adds a path prefix route to a specific backend
// Longer prefixes are matched first (most specific wins)
func (b *CompositeBackend) AddRoute(prefix string, backend BackendProtocol) {
	// Normalize prefix
	if prefix != "/" && !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// Insert in order by prefix length (longest first for most specific matching)
	newRoute := route{prefix: prefix, backend: backend}
	inserted := false

	for i, r := range b.routes {
		if len(prefix) > len(r.prefix) {
			// Insert before this route
			b.routes = append(b.routes[:i], append([]route{newRoute}, b.routes[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		b.routes = append(b.routes, newRoute)
	}
}

// getBackend returns the appropriate backend for a path
func (b *CompositeBackend) getBackend(path string) BackendProtocol {
	// Check routes in order (longest prefix first)
	for _, r := range b.routes {
		if strings.HasPrefix(path, r.prefix) {
			return r.backend
		}
	}
	return b.defaultBackend
}

// Read delegates to the appropriate backend
func (b *CompositeBackend) Read(ctx context.Context, path string, offset, limit int) (string, error) {
	backend := b.getBackend(path)
	return backend.Read(ctx, path, offset, limit)
}

// Write delegates to the appropriate backend
func (b *CompositeBackend) Write(ctx context.Context, path string, content string) (*WriteResult, error) {
	backend := b.getBackend(path)
	return backend.Write(ctx, path, content)
}

// Edit delegates to the appropriate backend
func (b *CompositeBackend) Edit(ctx context.Context, path string, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	backend := b.getBackend(path)
	return backend.Edit(ctx, path, oldStr, newStr, replaceAll)
}

// List lists files, potentially across multiple backends
func (b *CompositeBackend) List(ctx context.Context, path string) ([]FileInfo, error) {
	backend := b.getBackend(path)
	return backend.List(ctx, path)
}

// Glob finds files, potentially across multiple backends
func (b *CompositeBackend) Glob(ctx context.Context, pattern string, basePath string) ([]string, error) {
	// For glob, we might need to check multiple backends
	// For now, delegate to the backend for the base path
	backend := b.getBackend(basePath)
	return backend.Glob(ctx, pattern, basePath)
}

// Grep searches across files, potentially in multiple backends
func (b *CompositeBackend) Grep(ctx context.Context, pattern string, path string, globFilter string) ([]GrepMatch, error) {
	// For grep, delegate to the backend for the path
	backend := b.getBackend(path)
	return backend.Grep(ctx, pattern, path, globFilter)
}

// Execute implements SandboxBackendProtocol if default backend supports it
func (b *CompositeBackend) Execute(ctx context.Context, command string) (*ExecuteResult, error) {
	if sandbox, ok := b.defaultBackend.(SandboxBackendProtocol); ok {
		return sandbox.Execute(ctx, command)
	}
	return &ExecuteResult{
		Output:   "Error: Execution not supported by this backend configuration",
		ExitCode: 1,
	}, nil
}

// SupportsExecution checks if the composite backend supports execution
func (b *CompositeBackend) SupportsExecution() bool {
	_, ok := b.defaultBackend.(SandboxBackendProtocol)
	return ok
}
