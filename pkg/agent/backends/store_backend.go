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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

// StoreBackend persists files across sessions using persistence.Store
// Files are stored in the persistent store and survive beyond thread lifetime
// This matches LangGraph's StoreBackend for long-term memory
type StoreBackend struct {
	store     persistence.Store
	namespace string // Namespace for isolating file storage (e.g., user ID)
}

// NewStoreBackend creates a new store-based backend
func NewStoreBackend(store persistence.Store, namespace string) *StoreBackend {
	if namespace == "" {
		namespace = "default"
	}
	return &StoreBackend{
		store:     store,
		namespace: namespace,
	}
}

// Read reads file content from persistent store
func (b *StoreBackend) Read(ctx context.Context, path string, offset, limit int) (string, error) {
	fileData, err := b.getFileData(ctx, path)
	if err != nil {
		return "", err
	}

	lines := fileData.Content
	if len(lines) == 0 {
		return "System reminder: File exists but has empty contents\n", nil
	}

	// Apply pagination
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

	// Format with line numbers
	var sb strings.Builder
	for i := offset; i < end; i++ {
		line := lines[i]
		if len(line) > 2000 {
			line = line[:2000] + "... (truncated)"
		}
		sb.WriteString(fmt.Sprintf("%6d\t%s\n", i+1, line))
	}

	return sb.String(), nil
}

// Write creates or overwrites a file in persistent store
func (b *StoreBackend) Write(ctx context.Context, path string, content string) (*WriteResult, error) {
	now := time.Now().Format(time.RFC3339)

	fileData := &FileData{
		Content:    strings.Split(content, "\n"),
		CreatedAt:  now,
		ModifiedAt: now,
	}

	if err := b.saveFileData(ctx, path, fileData); err != nil {
		return &WriteResult{Error: err.Error()}, nil
	}

	return &WriteResult{
		Path:        path,
		FilesUpdate: nil, // Store backend doesn't update state
		Error:       "",
	}, nil
}

// Edit performs string replacement in a file
func (b *StoreBackend) Edit(ctx context.Context, path string, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	fileData, err := b.getFileData(ctx, path)
	if err != nil {
		return &EditResult{Error: err.Error()}, nil
	}

	content := strings.Join(fileData.Content, "\n")

	if !strings.Contains(content, oldStr) {
		return &EditResult{Error: "old_string not found in file"}, nil
	}

	occurrences := strings.Count(content, oldStr)

	if !replaceAll && occurrences > 1 {
		return &EditResult{
			Error: fmt.Sprintf("old_string found %d times, use replace_all=true or provide more context", occurrences),
		}, nil
	}

	var newContent string
	if replaceAll {
		newContent = strings.ReplaceAll(content, oldStr, newStr)
	} else {
		newContent = strings.Replace(content, oldStr, newStr, 1)
		occurrences = 1
	}

	fileData.Content = strings.Split(newContent, "\n")
	fileData.ModifiedAt = time.Now().Format(time.RFC3339)

	if err := b.saveFileData(ctx, path, fileData); err != nil {
		return &EditResult{Error: err.Error()}, nil
	}

	return &EditResult{
		Path:        path,
		Occurrences: occurrences,
		FilesUpdate: nil,
		Error:       "",
	}, nil
}

// List lists files in a directory from persistent store
func (b *StoreBackend) List(ctx context.Context, path string) ([]FileInfo, error) {
	// TODO: This would require the Store interface to support key listing/scanning
	// For now, return empty list (to be enhanced based on Store implementation)

	var infos []FileInfo

	return infos, nil
}

// Glob finds files matching a pattern
func (b *StoreBackend) Glob(ctx context.Context, pattern string, basePath string) ([]string, error) {
	// Simplified implementation - would need Store enhancement for production
	// This would require the Store interface to support key listing/scanning
	return []string{}, nil
}

// Grep searches for text in files
func (b *StoreBackend) Grep(ctx context.Context, pattern string, path string, globFilter string) ([]GrepMatch, error) {
	// Simplified implementation - would need to scan all files
	// This is left as a future enhancement when Store supports key listing
	return []GrepMatch{}, nil
}

// getKey creates a namespaced key for the file
func (b *StoreBackend) getKey(path string) string {
	return fmt.Sprintf("files:%s:%s", b.namespace, path)
}

// getFileData retrieves file data from store
func (b *StoreBackend) getFileData(ctx context.Context, path string) (*FileData, error) {
	key := b.getKey(path)

	data, err := b.store.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("file '%s' not found", path)
	}

	var fileData FileData
	if err := json.Unmarshal([]byte(data), &fileData); err != nil {
		return nil, fmt.Errorf("failed to parse file data: %w", err)
	}

	return &fileData, nil
}

// saveFileData saves file data to store
func (b *StoreBackend) saveFileData(ctx context.Context, path string, fileData *FileData) error {
	key := b.getKey(path)

	data, err := json.Marshal(fileData)
	if err != nil {
		return fmt.Errorf("failed to serialize file data: %w", err)
	}

	return b.store.Set(ctx, key, string(data))
}
