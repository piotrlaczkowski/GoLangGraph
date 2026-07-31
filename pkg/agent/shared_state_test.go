// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

// Copyright (c) 2024 GoLangGraph Team
package agent

import (
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestSharedState_BasicOperations(t *testing.T) {
	ss := NewSharedState()

	// Test global state
	ss.SetGlobal("key1", "value1")
	val, exists := ss.GetGlobal("key1")

	assert.True(t, exists)
	assert.Equal(t, "value1", val)

	// Test non-existent key
	_, exists = ss.GetGlobal("nonexistent")
	assert.False(t, exists)
}

func TestSharedState_LocalState(t *testing.T) {
	ss := NewSharedState()

	// Get local state for agent1
	state1 := ss.GetLocalState("agent1")
	assert.NotNil(t, state1)

	// Modify local state
	state1.Set("local_key", "local_value")

	// Retrieve same state
	state1Again := ss.GetLocalState("agent1")
	val, _ := state1Again.Get("local_key")
	assert.Equal(t, "local_value", val)

	// Different agent gets different state
	state2 := ss.GetLocalState("agent2")
	_, exists := state2.Get("local_key")
	assert.False(t, exists)
}

func TestSharedState_MergeFromLocal(t *testing.T) {
	ss := NewSharedState()

	// Setup local state
	localState := ss.GetLocalState("test-agent")
	localState.Set("result", "computed_value")
	localState.Set("temp", "temp_value")

	// Merge only "result" into global
	ss.MergeFromLocal("test-agent", []string{"result"})

	// Check global state has result
	val, exists := ss.GetGlobal("result")
	assert.True(t, exists)
	assert.Equal(t, "computed_value", val)

	// Check temp was not merged
	_, exists = ss.GetGlobal("temp")
	assert.False(t, exists)
}

func TestSharedState_ConcurrentAccess(t *testing.T) {
	ss := NewSharedState()
	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func(n int) {
			ss.SetGlobal("key"+string(rune(n)), n)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			_ = ss.GetAllGlobal()
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// No panics = success
}

func TestSubAgentRequest_Structure(t *testing.T) {
	req := SubAgentRequest{
		AgentID:    "test-agent",
		Input:      "test input",
		Timeout:    5 * time.Second,
		ShareState: true,
	}

	assert.Equal(t, "test-agent", req.AgentID)
	assert.Equal(t, "test input", req.Input)
	assert.Equal(t, 5*time.Second, req.Timeout)
	assert.True(t, req.ShareState)
}

func TestSubAgentResult_Structure(t *testing.T) {
	result := SubAgentResult{
		AgentID:  "test-agent",
		Output:   "test output",
		State:    core.NewBaseState(),
		Duration: 100 * time.Millisecond,
		Error:    nil,
	}

	assert.Equal(t, "test-agent", result.AgentID)
	assert.Equal(t, "test output", result.Output)
	assert.NotNil(t, result.State)
	assert.Equal(t, 100*time.Millisecond, result.Duration)
	assert.NoError(t, result.Error)
}
