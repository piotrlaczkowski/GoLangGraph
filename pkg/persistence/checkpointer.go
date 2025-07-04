// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
)

// Checkpointer defines the interface for state persistence
type Checkpointer interface {
	// Save saves a state checkpoint
	Save(ctx context.Context, checkpoint *Checkpoint) error

	// Load loads a state checkpoint
	Load(ctx context.Context, threadID, checkpointID string) (*Checkpoint, error)

	// List lists checkpoints for a thread
	List(ctx context.Context, threadID string) ([]*CheckpointMetadata, error)

	// Delete deletes a checkpoint
	Delete(ctx context.Context, threadID, checkpointID string) error

	// Close closes the checkpointer
	Close() error
}

// Checkpoint represents a saved state
type Checkpoint struct {
	ID        string                 `json:"id"`
	ThreadID  string                 `json:"thread_id"`
	State     *core.BaseState        `json:"state"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	NodeID    string                 `json:"node_id"`
	StepID    int                    `json:"step_id"`
}

// CheckpointMetadata represents checkpoint metadata without the full state
type CheckpointMetadata struct {
	ID        string                 `json:"id"`
	ThreadID  string                 `json:"thread_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	NodeID    string                 `json:"node_id"`
	StepID    int                    `json:"step_id"`
}

// MemoryCheckpointer implements in-memory checkpointing
type MemoryCheckpointer struct {
	mu          sync.RWMutex
	checkpoints map[string]map[string]*Checkpoint // threadID -> checkpointID -> checkpoint
}

// NewMemoryCheckpointer creates a new memory checkpointer
func NewMemoryCheckpointer() *MemoryCheckpointer {
	return &MemoryCheckpointer{
		checkpoints: make(map[string]map[string]*Checkpoint),
	}
}

// Save saves a checkpoint to memory
func (c *MemoryCheckpointer) Save(ctx context.Context, checkpoint *Checkpoint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.checkpoints[checkpoint.ThreadID] == nil {
		c.checkpoints[checkpoint.ThreadID] = make(map[string]*Checkpoint)
	}

	// Clone the checkpoint to avoid mutations
	cloned := &Checkpoint{
		ID:        checkpoint.ID,
		ThreadID:  checkpoint.ThreadID,
		State:     checkpoint.State.Clone(),
		Metadata:  make(map[string]interface{}),
		CreatedAt: checkpoint.CreatedAt,
		NodeID:    checkpoint.NodeID,
		StepID:    checkpoint.StepID,
	}

	// Clone metadata
	for k, v := range checkpoint.Metadata {
		cloned.Metadata[k] = v
	}

	c.checkpoints[checkpoint.ThreadID][checkpoint.ID] = cloned

	return nil
}

// Load loads a checkpoint from memory
func (c *MemoryCheckpointer) Load(ctx context.Context, threadID, checkpointID string) (*Checkpoint, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	threadCheckpoints, exists := c.checkpoints[threadID]
	if !exists {
		return nil, fmt.Errorf("thread %s not found", threadID)
	}

	checkpoint, exists := threadCheckpoints[checkpointID]
	if !exists {
		return nil, fmt.Errorf("checkpoint %s not found in thread %s", checkpointID, threadID)
	}

	// Clone the checkpoint to avoid mutations
	cloned := &Checkpoint{
		ID:        checkpoint.ID,
		ThreadID:  checkpoint.ThreadID,
		State:     checkpoint.State.Clone(),
		Metadata:  make(map[string]interface{}),
		CreatedAt: checkpoint.CreatedAt,
		NodeID:    checkpoint.NodeID,
		StepID:    checkpoint.StepID,
	}

	// Clone metadata
	for k, v := range checkpoint.Metadata {
		cloned.Metadata[k] = v
	}

	return cloned, nil
}

// List lists checkpoints for a thread
func (c *MemoryCheckpointer) List(ctx context.Context, threadID string) ([]*CheckpointMetadata, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	threadCheckpoints, exists := c.checkpoints[threadID]
	if !exists {
		return []*CheckpointMetadata{}, nil
	}

	var metadata []*CheckpointMetadata
	for _, checkpoint := range threadCheckpoints {
		meta := &CheckpointMetadata{
			ID:        checkpoint.ID,
			ThreadID:  checkpoint.ThreadID,
			Metadata:  make(map[string]interface{}),
			CreatedAt: checkpoint.CreatedAt,
			NodeID:    checkpoint.NodeID,
			StepID:    checkpoint.StepID,
		}

		// Clone metadata
		for k, v := range checkpoint.Metadata {
			meta.Metadata[k] = v
		}

		metadata = append(metadata, meta)
	}

	return metadata, nil
}

// Delete deletes a checkpoint
func (c *MemoryCheckpointer) Delete(ctx context.Context, threadID, checkpointID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	threadCheckpoints, exists := c.checkpoints[threadID]
	if !exists {
		return fmt.Errorf("thread %s not found", threadID)
	}

	if _, exists := threadCheckpoints[checkpointID]; !exists {
		return fmt.Errorf("checkpoint %s not found in thread %s", checkpointID, threadID)
	}

	delete(threadCheckpoints, checkpointID)

	// Clean up empty thread
	if len(threadCheckpoints) == 0 {
		delete(c.checkpoints, threadID)
	}

	return nil
}

// Close closes the memory checkpointer
func (c *MemoryCheckpointer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checkpoints = make(map[string]map[string]*Checkpoint)
	return nil
}

// GetThreadIDs returns all thread IDs
func (c *MemoryCheckpointer) GetThreadIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var threadIDs []string
	for threadID := range c.checkpoints {
		threadIDs = append(threadIDs, threadID)
	}

	return threadIDs
}

// GetCheckpointCount returns the total number of checkpoints
func (c *MemoryCheckpointer) GetCheckpointCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, threadCheckpoints := range c.checkpoints {
		count += len(threadCheckpoints)
	}

	return count
}

// FileCheckpointer implements file-based checkpointing
type FileCheckpointer struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileCheckpointer creates a new file checkpointer
func NewFileCheckpointer(basePath string) *FileCheckpointer {
	return &FileCheckpointer{
		basePath: basePath,
	}
}

// Save saves a checkpoint to file
func (c *FileCheckpointer) Save(ctx context.Context, checkpoint *Checkpoint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create directory structure
	threadDir := fmt.Sprintf("%s/%s", c.basePath, checkpoint.ThreadID)
	if err := ensureDir(threadDir); err != nil {
		return fmt.Errorf("failed to create thread directory: %w", err)
	}

	// Serialize checkpoint
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	// Write to file
	filePath := fmt.Sprintf("%s/%s.json", threadDir, checkpoint.ID)
	if err := writeFile(filePath, data); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return nil
}

// Load loads a checkpoint from file
func (c *FileCheckpointer) Load(ctx context.Context, threadID, checkpointID string) (*Checkpoint, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filePath := fmt.Sprintf("%s/%s/%s.json", c.basePath, threadID, checkpointID)

	data, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// List lists checkpoints for a thread
func (c *FileCheckpointer) List(ctx context.Context, threadID string) ([]*CheckpointMetadata, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	threadDir := fmt.Sprintf("%s/%s", c.basePath, threadID)

	files, err := listFiles(threadDir, ".json")
	if err != nil {
		return []*CheckpointMetadata{}, nil // Return empty list if directory doesn't exist
	}

	var metadata []*CheckpointMetadata
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s", threadDir, file)

		data, err := readFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var checkpoint Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue // Skip files that can't be unmarshaled
		}

		meta := &CheckpointMetadata{
			ID:        checkpoint.ID,
			ThreadID:  checkpoint.ThreadID,
			Metadata:  checkpoint.Metadata,
			CreatedAt: checkpoint.CreatedAt,
			NodeID:    checkpoint.NodeID,
			StepID:    checkpoint.StepID,
		}

		metadata = append(metadata, meta)
	}

	return metadata, nil
}

// Delete deletes a checkpoint
func (c *FileCheckpointer) Delete(ctx context.Context, threadID, checkpointID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := fmt.Sprintf("%s/%s/%s.json", c.basePath, threadID, checkpointID)

	if err := deleteFile(filePath); err != nil {
		return fmt.Errorf("failed to delete checkpoint file: %w", err)
	}

	return nil
}

// Close closes the file checkpointer
func (c *FileCheckpointer) Close() error {
	// Nothing to close for file checkpointer
	return nil
}

// CheckpointManager manages checkpointing for graph execution
type CheckpointManager struct {
	checkpointer Checkpointer
	enabled      bool
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(checkpointer Checkpointer) *CheckpointManager {
	return &CheckpointManager{
		checkpointer: checkpointer,
		enabled:      checkpointer != nil,
	}
}

// SaveCheckpoint saves a checkpoint
func (cm *CheckpointManager) SaveCheckpoint(ctx context.Context, threadID, nodeID string, stepID int, state *core.BaseState) error {
	if !cm.enabled {
		return nil
	}

	checkpoint := &Checkpoint{
		ID:        fmt.Sprintf("%s-%d", nodeID, stepID),
		ThreadID:  threadID,
		State:     state,
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		NodeID:    nodeID,
		StepID:    stepID,
	}

	return cm.checkpointer.Save(ctx, checkpoint)
}

// LoadCheckpoint loads a checkpoint
func (cm *CheckpointManager) LoadCheckpoint(ctx context.Context, threadID, checkpointID string) (*Checkpoint, error) {
	if !cm.enabled {
		return nil, fmt.Errorf("checkpointing is not enabled")
	}

	return cm.checkpointer.Load(ctx, threadID, checkpointID)
}

// ListCheckpoints lists checkpoints for a thread
func (cm *CheckpointManager) ListCheckpoints(ctx context.Context, threadID string) ([]*CheckpointMetadata, error) {
	if !cm.enabled {
		return []*CheckpointMetadata{}, nil
	}

	return cm.checkpointer.List(ctx, threadID)
}

// DeleteCheckpoint deletes a checkpoint
func (cm *CheckpointManager) DeleteCheckpoint(ctx context.Context, threadID, checkpointID string) error {
	if !cm.enabled {
		return nil
	}

	return cm.checkpointer.Delete(ctx, threadID, checkpointID)
}

// IsEnabled returns true if checkpointing is enabled
func (cm *CheckpointManager) IsEnabled() bool {
	return cm.enabled
}

// Close closes the checkpoint manager
func (cm *CheckpointManager) Close() error {
	if cm.enabled {
		return cm.checkpointer.Close()
	}
	return nil
}

// TimeTravel provides time travel functionality
type TimeTravel struct {
	checkpointManager *CheckpointManager
}

// NewTimeTravel creates a new time travel instance
func NewTimeTravel(checkpointManager *CheckpointManager) *TimeTravel {
	return &TimeTravel{
		checkpointManager: checkpointManager,
	}
}

// RewindTo rewinds execution to a specific checkpoint
func (tt *TimeTravel) RewindTo(ctx context.Context, threadID, checkpointID string) (*core.BaseState, error) {
	checkpoint, err := tt.checkpointManager.LoadCheckpoint(ctx, threadID, checkpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	return checkpoint.State, nil
}

// GetHistory returns the execution history for a thread
func (tt *TimeTravel) GetHistory(ctx context.Context, threadID string) ([]*CheckpointMetadata, error) {
	return tt.checkpointManager.ListCheckpoints(ctx, threadID)
}

// FindCheckpointByStep finds a checkpoint by step ID
func (tt *TimeTravel) FindCheckpointByStep(ctx context.Context, threadID string, stepID int) (*CheckpointMetadata, error) {
	checkpoints, err := tt.checkpointManager.ListCheckpoints(ctx, threadID)
	if err != nil {
		return nil, err
	}

	for _, checkpoint := range checkpoints {
		if checkpoint.StepID == stepID {
			return checkpoint, nil
		}
	}

	return nil, fmt.Errorf("checkpoint with step ID %d not found", stepID)
}

// FindCheckpointByNode finds the latest checkpoint for a specific node
func (tt *TimeTravel) FindCheckpointByNode(ctx context.Context, threadID, nodeID string) (*CheckpointMetadata, error) {
	checkpoints, err := tt.checkpointManager.ListCheckpoints(ctx, threadID)
	if err != nil {
		return nil, err
	}

	var latest *CheckpointMetadata
	for _, checkpoint := range checkpoints {
		if checkpoint.NodeID == nodeID {
			if latest == nil || checkpoint.CreatedAt.After(latest.CreatedAt) {
				latest = checkpoint
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no checkpoint found for node %s", nodeID)
	}

	return latest, nil
}

// Placeholder functions for file operations (would be implemented with actual file I/O)
func ensureDir(path string) error {
	// Implementation would create directory if it doesn't exist
	return nil
}

func writeFile(path string, data []byte) error {
	// Implementation would write data to file
	return nil
}

func readFile(path string) ([]byte, error) {
	// Implementation would read data from file
	return []byte{}, nil
}

func listFiles(dir, extension string) ([]string, error) {
	// Implementation would list files with given extension in directory
	return []string{}, nil
}

func deleteFile(path string) error {
	// Implementation would delete file
	return nil
}
