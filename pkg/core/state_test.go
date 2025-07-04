package core

import (
	"testing"
)

func TestNewBaseState(t *testing.T) {
	state := NewBaseState()
	if state == nil {
		t.Fatal("NewBaseState() returned nil")
	}

	// Test that state is properly initialized
	keys := state.Keys()
	if len(keys) != 0 {
		t.Error("BaseState should be empty initially")
	}
}

func TestBaseState_Set(t *testing.T) {
	state := NewBaseState()

	// Test setting a value
	state.Set("key1", "value1")
	if val, exists := state.Get("key1"); !exists || val != "value1" {
		t.Error("Set() failed to set value")
	}

	// Test overwriting a value
	state.Set("key1", "value2")
	if val, exists := state.Get("key1"); !exists || val != "value2" {
		t.Error("Set() failed to overwrite value")
	}

	// Test setting different types
	state.Set("int", 42)
	state.Set("bool", true)
	state.Set("slice", []string{"a", "b", "c"})

	if val, _ := state.Get("int"); val != 42 {
		t.Error("Set() failed to set int value")
	}
	if val, _ := state.Get("bool"); val != true {
		t.Error("Set() failed to set bool value")
	}
	if val, _ := state.Get("slice"); len(val.([]string)) != 3 {
		t.Error("Set() failed to set slice value")
	}
}

func TestBaseState_Get(t *testing.T) {
	state := NewBaseState()

	// Test getting non-existent key
	val, exists := state.Get("nonexistent")
	if exists {
		t.Error("Get() should return false for non-existent key")
	}
	if val != nil {
		t.Error("Get() should return nil for non-existent key")
	}

	// Test getting existing key
	state.Set("key1", "value1")
	val, exists = state.Get("key1")
	if !exists {
		t.Error("Get() should return true for existing key")
	}
	if val != "value1" {
		t.Error("Get() should return correct value")
	}
}

func TestBaseState_Delete(t *testing.T) {
	state := NewBaseState()

	// Test deleting non-existent key
	state.Delete("nonexistent") // Should not panic

	// Test deleting existing key
	state.Set("key1", "value1")
	state.Delete("key1")

	_, exists := state.Get("key1")
	if exists {
		t.Error("Delete() failed to remove key")
	}
}

func TestBaseState_Keys(t *testing.T) {
	state := NewBaseState()

	// Test empty state
	keys := state.Keys()
	if len(keys) != 0 {
		t.Error("Keys() should return empty slice for empty state")
	}

	// Test with values
	state.Set("key1", "value1")
	state.Set("key2", "value2")
	state.Set("key3", "value3")

	keys = state.Keys()
	if len(keys) != 3 {
		t.Error("Keys() should return correct number of keys")
	}

	// Check all keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	if !keyMap["key1"] || !keyMap["key2"] || !keyMap["key3"] {
		t.Error("Keys() should return all keys")
	}
}

func TestBaseState_Clone(t *testing.T) {
	state := NewBaseState()
	state.Set("key1", "value1")
	state.Set("key2", 42)
	state.Set("key3", []string{"a", "b"})

	clone := state.Clone()

	// Test that clone has same values
	val1, exists1 := clone.Get("key1")
	val2, exists2 := clone.Get("key2")
	val3, exists3 := clone.Get("key3")

	if !exists1 || val1 != "value1" {
		t.Error("Clone() failed to clone string value")
	}
	if !exists2 || val2 != 42 {
		t.Error("Clone() failed to clone int value")
	}
	if !exists3 || len(val3.([]string)) != 2 {
		t.Error("Clone() failed to clone slice value")
	}

	// Test that modifying clone doesn't affect original
	clone.Set("key1", "modified")
	originalVal, _ := state.Get("key1")
	if originalVal != "value1" {
		t.Error("Clone() should create independent copy")
	}

	// Test that modifying original doesn't affect clone
	state.Set("key4", "new")
	_, exists := clone.Get("key4")
	if exists {
		t.Error("Clone() should create independent copy")
	}
}

func TestBaseState_Merge(t *testing.T) {
	state1 := NewBaseState()
	state1.Set("key1", "value1")
	state1.Set("key2", "value2")

	state2 := NewBaseState()
	state2.Set("key2", "overwritten")
	state2.Set("key3", "value3")

	state1.Merge(state2)

	// Test that key1 is preserved
	val1, exists1 := state1.Get("key1")
	if !exists1 || val1 != "value1" {
		t.Error("Merge() should preserve non-conflicting keys")
	}

	// Test that key2 is overwritten
	val2, exists2 := state1.Get("key2")
	if !exists2 || val2 != "overwritten" {
		t.Error("Merge() should overwrite conflicting keys")
	}

	// Test that key3 is added
	val3, exists3 := state1.Get("key3")
	if !exists3 || val3 != "value3" {
		t.Error("Merge() should add new keys")
	}
}

func TestBaseState_GetAll(t *testing.T) {
	state := NewBaseState()
	state.Set("key1", "value1")
	state.Set("key2", 42)

	all := state.GetAll()

	if len(all) != 2 {
		t.Error("GetAll() should return all key-value pairs")
	}

	if all["key1"] != "value1" || all["key2"] != 42 {
		t.Error("GetAll() should return correct values")
	}
}

func TestBaseState_Metadata(t *testing.T) {
	state := NewBaseState()

	// Test setting metadata
	state.SetMetadata("meta1", "metavalue1")
	val, exists := state.GetMetadata("meta1")
	if !exists || val != "metavalue1" {
		t.Error("SetMetadata/GetMetadata failed")
	}

	// Test non-existent metadata
	_, exists = state.GetMetadata("nonexistent")
	if exists {
		t.Error("GetMetadata should return false for non-existent metadata")
	}
}

func TestBaseState_Snapshot(t *testing.T) {
	state := NewBaseState()
	state.Set("key1", "value1")
	state.SetMetadata("meta1", "metavalue1")

	snapshot := state.CreateSnapshot()

	// Test snapshot contains data
	if snapshot.Data["key1"] != "value1" {
		t.Error("Snapshot should contain state data")
	}

	if snapshot.Metadata["meta1"] != "metavalue1" {
		t.Error("Snapshot should contain metadata")
	}

	// Test restore from snapshot
	state.Set("key1", "modified")
	state.RestoreFromSnapshot(snapshot)

	val, _ := state.Get("key1")
	if val != "value1" {
		t.Error("RestoreFromSnapshot should restore original value")
	}
}

func TestBaseState_JSON(t *testing.T) {
	state := NewBaseState()
	state.Set("key1", "value1")
	state.Set("key2", 42)

	// Test ToJSON
	jsonData, err := state.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() failed: %v", err)
	}

	// Test FromJSON
	newState := NewBaseState()
	err = newState.FromJSON(jsonData)
	if err != nil {
		t.Errorf("FromJSON() failed: %v", err)
	}

	// Verify data was restored
	val1, exists1 := newState.Get("key1")
	val2, exists2 := newState.Get("key2")

	if !exists1 || val1 != "value1" {
		t.Error("FromJSON() failed to restore string value")
	}
	if !exists2 || val2.(float64) != 42 { // JSON numbers are float64
		t.Error("FromJSON() failed to restore numeric value")
	}
}

// Benchmark tests
func BenchmarkBaseState_Set(b *testing.B) {
	state := NewBaseState()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state.Set("key", i)
	}
}

func BenchmarkBaseState_Get(b *testing.B) {
	state := NewBaseState()
	state.Set("key", "value")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state.Get("key")
	}
}

func BenchmarkBaseState_Clone(b *testing.B) {
	state := NewBaseState()
	for i := 0; i < 100; i++ {
		state.Set(string(rune('a'+i%26)), i)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state.Clone()
	}
}
