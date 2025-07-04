// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package core

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StateValue represents any value that can be stored in state
type StateValue interface{}

// StateSnapshot represents a snapshot of the state at a specific point in time
type StateSnapshot struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]StateValue  `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// StateHistory manages the history of state changes
type StateHistory struct {
	snapshots []StateSnapshot
	maxSize   int
	mu        sync.RWMutex
}

// NewStateHistory creates a new state history with a maximum size
func NewStateHistory(maxSize int) *StateHistory {
	return &StateHistory{
		snapshots: make([]StateSnapshot, 0),
		maxSize:   maxSize,
	}
}

// AddSnapshot adds a new snapshot to the history
func (sh *StateHistory) AddSnapshot(snapshot StateSnapshot) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.snapshots = append(sh.snapshots, snapshot)
	if len(sh.snapshots) > sh.maxSize {
		sh.snapshots = sh.snapshots[1:]
	}
}

// GetSnapshots returns all snapshots in the history
func (sh *StateHistory) GetSnapshots() []StateSnapshot {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	result := make([]StateSnapshot, len(sh.snapshots))
	copy(result, sh.snapshots)
	return result
}

// GetSnapshot returns a specific snapshot by ID
func (sh *StateHistory) GetSnapshot(id string) (*StateSnapshot, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	for _, snapshot := range sh.snapshots {
		if snapshot.ID == id {
			return &snapshot, nil
		}
	}
	return nil, fmt.Errorf("snapshot with ID %s not found", id)
}

// BaseState represents the base state structure
type BaseState struct {
	data     map[string]StateValue
	metadata map[string]interface{}
	history  *StateHistory
	mu       sync.RWMutex
}

// NewBaseState creates a new base state
func NewBaseState() *BaseState {
	return &BaseState{
		data:     make(map[string]StateValue),
		metadata: make(map[string]interface{}),
		history:  NewStateHistory(100), // Keep last 100 snapshots
	}
}

// Get retrieves a value from the state
func (bs *BaseState) Get(key string) (StateValue, bool) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	value, exists := bs.data[key]
	return value, exists
}

// Set sets a value in the state
func (bs *BaseState) Set(key string, value StateValue) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.data[key] = value
}

// Delete removes a key from the state
func (bs *BaseState) Delete(key string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delete(bs.data, key)
}

// Keys returns all keys in the state
func (bs *BaseState) Keys() []string {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	keys := make([]string, 0, len(bs.data))
	for k := range bs.data {
		keys = append(keys, k)
	}
	return keys
}

// GetAll returns a copy of all data in the state
func (bs *BaseState) GetAll() map[string]StateValue {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	result := make(map[string]StateValue)
	for k, v := range bs.data {
		result[k] = deepCopy(v)
	}
	return result
}

// SetMetadata sets metadata for the state
func (bs *BaseState) SetMetadata(key string, value interface{}) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.metadata[key] = value
}

// GetMetadata retrieves metadata from the state
func (bs *BaseState) GetMetadata(key string) (interface{}, bool) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	value, exists := bs.metadata[key]
	return value, exists
}

// CreateSnapshot creates a snapshot of the current state
func (bs *BaseState) CreateSnapshot() StateSnapshot {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	snapshot := StateSnapshot{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data:      make(map[string]StateValue),
		Metadata:  make(map[string]interface{}),
	}

	// Deep copy data
	for k, v := range bs.data {
		snapshot.Data[k] = deepCopy(v)
	}

	// Deep copy metadata
	for k, v := range bs.metadata {
		snapshot.Metadata[k] = deepCopy(v)
	}

	// Add to history
	bs.history.AddSnapshot(snapshot)

	return snapshot
}

// RestoreFromSnapshot restores the state from a snapshot
func (bs *BaseState) RestoreFromSnapshot(snapshot StateSnapshot) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	// Clear current data
	bs.data = make(map[string]StateValue)
	bs.metadata = make(map[string]interface{})

	// Restore data
	for k, v := range snapshot.Data {
		bs.data[k] = deepCopy(v)
	}

	// Restore metadata
	for k, v := range snapshot.Metadata {
		bs.metadata[k] = deepCopy(v)
	}
}

// GetHistory returns the state history
func (bs *BaseState) GetHistory() *StateHistory {
	return bs.history
}

// Merge merges another state into this state
func (bs *BaseState) Merge(other *BaseState) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	otherData := other.GetAll()
	for k, v := range otherData {
		bs.data[k] = v
	}
}

// Clone creates a deep copy of the state
func (bs *BaseState) Clone() *BaseState {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	clone := NewBaseState()

	// Deep copy data
	for k, v := range bs.data {
		clone.data[k] = deepCopy(v)
	}

	// Deep copy metadata
	for k, v := range bs.metadata {
		clone.metadata[k] = deepCopy(v)
	}

	return clone
}

// ToJSON converts the state to JSON
func (bs *BaseState) ToJSON() ([]byte, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	stateData := struct {
		Data     map[string]StateValue  `json:"data"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		Data:     bs.data,
		Metadata: bs.metadata,
	}

	return json.Marshal(stateData)
}

// FromJSON loads the state from JSON
func (bs *BaseState) FromJSON(data []byte) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	var stateData struct {
		Data     map[string]StateValue  `json:"data"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := json.Unmarshal(data, &stateData); err != nil {
		return err
	}

	bs.data = stateData.Data
	bs.metadata = stateData.Metadata

	return nil
}

// deepCopy creates a deep copy of a value
func deepCopy(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	// Handle basic types
	switch v := src.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string:
		return v
	case []byte:
		dst := make([]byte, len(v))
		copy(dst, v)
		return dst
	}

	// Handle complex types using reflection
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.New(srcVal.Type()).Elem()

	deepCopyRecursive(srcVal, dstVal)
	return dstVal.Interface()
}

// deepCopyRecursive performs recursive deep copying
func deepCopyRecursive(src, dst reflect.Value) {
	switch src.Kind() {
	case reflect.Ptr:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.New(src.Type().Elem()))
		deepCopyRecursive(src.Elem(), dst.Elem())
	case reflect.Interface:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.ValueOf(deepCopy(src.Interface())))
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			deepCopyRecursive(src.Field(i), dst.Field(i))
		}
	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			deepCopyRecursive(src.Index(i), dst.Index(i))
		}
	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			srcVal := src.MapIndex(key)
			dstVal := reflect.New(srcVal.Type()).Elem()
			deepCopyRecursive(srcVal, dstVal)
			dst.SetMapIndex(key, dstVal)
		}
	default:
		dst.Set(src)
	}
}

// StateManager manages multiple states and provides advanced operations
type StateManager struct {
	states map[string]*BaseState
	mu     sync.RWMutex
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]*BaseState),
	}
}

// CreateState creates a new state with the given ID
func (sm *StateManager) CreateState(id string) *BaseState {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state := NewBaseState()
	sm.states[id] = state
	return state
}

// GetState retrieves a state by ID
func (sm *StateManager) GetState(id string) (*BaseState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, exists := sm.states[id]
	return state, exists
}

// DeleteState removes a state by ID
func (sm *StateManager) DeleteState(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.states, id)
}

// ListStates returns all state IDs
func (sm *StateManager) ListStates() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.states))
	for id := range sm.states {
		ids = append(ids, id)
	}
	return ids
}
