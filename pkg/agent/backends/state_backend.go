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
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// StateBackend stores files in agent state (ephemeral, session-scoped)
// Files are stored in state["files"] and persist only within the thread/session
// This matches LangGraph's default StateBackend behavior
type StateBackend struct {
	// getState retrieves the current state
	// This is injected at runtime to avoid circular dependencies
	getState func() map[string]interface{}
}

// NewStateBackend creates a new state-based backend
func NewStateBackend(getState func() map[string]interface{}) *StateBackend {
	return &StateBackend{
		getState: getState,
	}
}

// Read reads file content from state
func (b *StateBackend) Read(ctx context.Context, path string, offset, limit int) (string, error) {
	files := b.getFiles()
	fileData, exists := files[path]
	if !exists {
		return "", fmt.Errorf("file '%s' not found", path)
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

	// Format with line numbers (cat -n style)
	var sb strings.Builder
	for i := offset; i < end; i++ {
		line := lines[i]
		// Truncate very long lines
		if len(line) > 2000 {
			line = line[:2000] + "... (truncated)"
		}
		sb.WriteString(fmt.Sprintf("%6d\t%s\n", i+1, line))
	}

	return sb.String(), nil
}

// Write creates or overwrites a file in state
func (b *StateBackend) Write(ctx context.Context, path string, content string) (*WriteResult, error) {
	now := time.Now().Format(time.RFC3339)

	fileData := &FileData{
		Content:    strings.Split(content, "\n"),
		CreatedAt:  now,
		ModifiedAt: now,
	}

	// Return update for state
	return &WriteResult{
		Path: path,
		FilesUpdate: map[string]*FileData{
			path: fileData,
		},
		Error: "",
	}, nil
}

// Edit performs string replacement in a file
func (b *StateBackend) Edit(ctx context.Context, path string, oldStr, newStr string, replaceAll bool) (*EditResult, error) {
	files := b.getFiles()
	fileData, exists := files[path]
	if !exists {
		return &EditResult{Error: fmt.Sprintf("file '%s' not found", path)}, nil
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

	updatedFileData := &FileData{
		Content:    strings.Split(newContent, "\n"),
		CreatedAt:  fileData.CreatedAt,
		ModifiedAt: time.Now().Format(time.RFC3339),
	}

	return &EditResult{
		Path:        path,
		Occurrences: occurrences,
		FilesUpdate: map[string]*FileData{
			path: updatedFileData,
		},
		Error: "",
	}, nil
}

// List lists files in a directory
func (b *StateBackend) List(ctx context.Context, path string) ([]FileInfo, error) {
	files := b.getFiles()
	var infos []FileInfo

	// Normalize path
	if path != "/" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	for filePath, fileData := range files {
		// Check if file is in this directory
		if path == "/" || strings.HasPrefix(filePath, path) {
			// Get relative path
			relPath := strings.TrimPrefix(filePath, path)
			// Skip if it's a subdirectory file (contains /)
			if path != "/" && strings.Contains(relPath, "/") {
				continue
			}

			modTime, _ := time.Parse(time.RFC3339, fileData.ModifiedAt)
			infos = append(infos, FileInfo{
				Path:       filePath,
				IsDir:      false,
				Size:       int64(len(strings.Join(fileData.Content, "\n"))),
				ModifiedAt: modTime,
			})
		}
	}

	return infos, nil
}

// Glob finds files matching a pattern
func (b *StateBackend) Glob(ctx context.Context, pattern string, basePath string) ([]string, error) {
	files := b.getFiles()
	var matches []string

	for filePath := range files {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err != nil {
			return nil, err
		}
		if matched {
			if basePath == "" || strings.HasPrefix(filePath, basePath) {
				matches = append(matches, filePath)
			}
		}
	}

	return matches, nil
}

// Grep searches for text in files
func (b *StateBackend) Grep(ctx context.Context, pattern string, path string, globFilter string) ([]GrepMatch, error) {
	files := b.getFiles()
	var matches []GrepMatch

	for filePath, fileData := range files {
		// Apply path filter
		if path != "" && !strings.HasPrefix(filePath, path) {
			continue
		}

		// Apply glob filter
		if globFilter != "" {
			matched, _ := filepath.Match(globFilter, filepath.Base(filePath))
			if !matched {
				continue
			}
		}

		// Search in content
		for i, line := range fileData.Content {
			if strings.Contains(line, pattern) {
				matches = append(matches, GrepMatch{
					Path:       filePath,
					LineNumber: i + 1,
					Line:       line,
				})
			}
		}
	}

	return matches, nil
}

// getFiles retrieves the files map from state
func (b *StateBackend) getFiles() map[string]*FileData {
	if b.getState == nil {
		return make(map[string]*FileData)
	}

	state := b.getState()
	if files, ok := state["files"].(map[string]*FileData); ok {
		return files
	}

	return make(map[string]*FileData)
}
