// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package persistence provides storage interfaces and implementations

package persistence

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Store provides persistent key-value storage across sessions
// This interface enables long-term memory and persistent file storage
type Store interface {
	// Get retrieves a value by key
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value by key
	Set(ctx context.Context, key string, value string) error

	// Delete removes a value by key
	Delete(ctx context.Context, key string) error

	// Has checks if a key exists
	Has(ctx context.Context, key string) (bool, error)

	// List returns all keys matching a prefix (optional, may return empty for simple implementations)
	List(ctx context.Context, prefix string) ([]string, error)
}

// InMemoryStore provides a simple in-memory implementation of Store
type InMemoryStore struct {
	data map[string]string
	mu   sync.RWMutex
}

// NewInMemoryStore creates a new in-memory store
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

func (s *InMemoryStore) Get(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.data[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("key '%s' not found", key)
}

func (s *InMemoryStore) Set(ctx context.Context, key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	return nil
}

func (s *InMemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}

func (s *InMemoryStore) Has(ctx context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	return ok, nil
}

func (s *InMemoryStore) List(ctx context.Context, prefix string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var keys []string
	for key := range s.data {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys, nil
}
