// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package agent

import (
	"sync"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

// SharedState manages state sharing across multiple subagents
// Provides both global state (shared across all) and local state (per-subagent)
type SharedState struct {
	mu     sync.RWMutex
	global map[string]interface{}     // Shared across all subagents
	local  map[string]*core.BaseState // Per-subagent state
}

// NewSharedState creates a new shared state manager
func NewSharedState() *SharedState {
	return &SharedState{
		global: make(map[string]interface{}),
		local:  make(map[string]*core.BaseState),
	}
}

// SetGlobal sets a value in the global shared state (thread-safe)
func (ss *SharedState) SetGlobal(key string, value interface{}) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.global[key] = value
}

// GetGlobal retrieves a value from the global shared state (thread-safe)
func (ss *SharedState) GetGlobal(key string) (interface{}, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	val, exists := ss.global[key]
	return val, exists
}

// GetAllGlobal returns a copy of all global state (thread-safe)
func (ss *SharedState) GetAllGlobal() map[string]interface{} {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	copy := make(map[string]interface{}, len(ss.global))
	for k, v := range ss.global {
		copy[k] = v
	}
	return copy
}

// GetLocalState gets or creates a local state for a specific agent
func (ss *SharedState) GetLocalState(agentID string) *core.BaseState {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if state, exists := ss.local[agentID]; exists {
		return state
	}

	// Create new local state
	state := core.NewBaseState()
	ss.local[agentID] = state
	return state
}

// SetLocalState sets the local state for a specific agent
func (ss *SharedState) SetLocalState(agentID string, state *core.BaseState) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.local[agentID] = state
}

// MergeFromLocal merges values from a local state into global state
func (ss *SharedState) MergeFromLocal(agentID string, keys []string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	localState, exists := ss.local[agentID]
	if !exists {
		return
	}

	for _, key := range keys {
		if val, ok := localState.Get(key); ok {
			ss.global[key] = val
		}
	}
}

// Clear removes all state (useful for testing)
func (ss *SharedState) Clear() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.global = make(map[string]interface{})
	ss.local = make(map[string]*core.BaseState)
}
