package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/llm"
)

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	if registry == nil {
		t.Fatal("NewToolRegistry returned nil")
	}

	// Should have default tools registered
	tools := registry.ListTools()
	if len(tools) == 0 {
		t.Error("Registry should have default tools")
	}

	expectedTools := []string{
		"web_search", "file_read", "file_write", "file_list",
		"shell", "http_request", "calculator", "time",
	}

	for _, expectedTool := range expectedTools {
		if !contains(tools, expectedTool) {
			t.Errorf("Expected tool %s not found in registry", expectedTool)
		}
	}
}

func TestToolRegistry_RegisterTool(t *testing.T) {
	registry := NewToolRegistry()

	// Create a mock tool
	mockTool := &MockTool{
		name:        "mock_tool",
		description: "A mock tool for testing",
	}

	// Register the tool
	registry.RegisterTool(mockTool)

	// Check if tool is registered
	tool, exists := registry.GetTool("mock_tool")
	if !exists {
		t.Error("Tool should be registered")
	}

	if tool != mockTool {
		t.Error("Registered tool should be the same instance")
	}

	// Check if tool appears in list
	tools := registry.ListTools()
	if !contains(tools, "mock_tool") {
		t.Error("Tool should appear in tools list")
	}
}

func TestToolRegistry_GetTool(t *testing.T) {
	registry := NewToolRegistry()

	// Test getting existing tool
	tool, exists := registry.GetTool("calculator")
	if !exists {
		t.Error("Calculator tool should exist")
	}

	if tool.GetName() != "calculator" {
		t.Error("Tool name should be calculator")
	}

	// Test getting non-existing tool
	_, exists = registry.GetTool("non_existing_tool")
	if exists {
		t.Error("Non-existing tool should not exist")
	}
}

func TestToolRegistry_GetDefinitions(t *testing.T) {
	registry := NewToolRegistry()

	// Test getting all definitions
	allDefs := registry.GetAllDefinitions()
	if len(allDefs) == 0 {
		t.Error("Should have tool definitions")
	}

	// Test getting specific definitions
	specificDefs := registry.GetDefinitions([]string{"calculator", "time"})
	if len(specificDefs) != 2 {
		t.Error("Should have 2 specific definitions")
	}

	// Test getting non-existing tool definitions
	emptyDefs := registry.GetDefinitions([]string{"non_existing_tool"})
	if len(emptyDefs) != 0 {
		t.Error("Should have 0 definitions for non-existing tools")
	}
}

func TestCalculatorTool(t *testing.T) {
	tool := NewCalculatorTool()

	if tool.GetName() != "calculator" {
		t.Error("Calculator tool name should be 'calculator'")
	}

	if tool.GetDescription() == "" {
		t.Error("Calculator tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "calculator" {
		t.Error("Definition function name should be 'calculator'")
	}

	// Test execution
	ctx := context.Background()
	args := `{"expression": "2 + 3"}`

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Calculator execution failed: %v", err)
	}

	if !strings.Contains(result, "5") {
		t.Error("Calculator should return 5 for 2 + 3")
	}

	// Test validation
	validArgs := `{"expression": "10 * 5"}`
	if err := tool.Validate(validArgs); err != nil {
		t.Errorf("Valid args should not fail validation: %v", err)
	}

	invalidArgs := `{"expression": ""}`
	if err := tool.Validate(invalidArgs); err == nil {
		t.Error("Empty expression should fail validation")
	}
}

func TestTimeTool(t *testing.T) {
	tool := NewTimeTool()

	if tool.GetName() != "time" {
		t.Error("Time tool name should be 'time'")
	}

	if tool.GetDescription() == "" {
		t.Error("Time tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "time" {
		t.Error("Definition function name should be 'time'")
	}

	// Test execution
	ctx := context.Background()
	args := `{"format": "RFC3339", "timezone": "UTC"}`

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Time execution failed: %v", err)
	}

	if result == "" {
		t.Error("Time tool should return a result")
	}

	// Test with empty args
	result2, err := tool.Execute(ctx, "{}")
	if err != nil {
		t.Fatalf("Time execution with empty args failed: %v", err)
	}

	if result2 == "" {
		t.Error("Time tool should return a result with empty args")
	}
}

func TestWebSearchTool(t *testing.T) {
	tool := NewWebSearchTool()

	if tool.GetName() != "web_search" {
		t.Error("Web search tool name should be 'web_search'")
	}

	if tool.GetDescription() == "" {
		t.Error("Web search tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "web_search" {
		t.Error("Definition function name should be 'web_search'")
	}

	// Test validation
	validArgs := `{"query": "golang programming"}`
	if err := tool.Validate(validArgs); err != nil {
		t.Errorf("Valid args should not fail validation: %v", err)
	}

	invalidArgs := `{"query": ""}`
	if err := tool.Validate(invalidArgs); err == nil {
		t.Error("Empty query should fail validation")
	}

	// Test config
	config := tool.GetConfig()
	if config == nil {
		t.Error("Tool should have config")
	}

	newConfig := map[string]interface{}{
		"max_results": 10,
	}
	if err := tool.SetConfig(newConfig); err != nil {
		t.Errorf("Setting config should not fail: %v", err)
	}
}

func TestFileReadTool(t *testing.T) {
	tool := NewFileReadTool()

	if tool.GetName() != "file_read" {
		t.Error("File read tool name should be 'file_read'")
	}

	if tool.GetDescription() == "" {
		t.Error("File read tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "file_read" {
		t.Error("Definition function name should be 'file_read'")
	}

	// Test validation
	validArgs := `{"file_path": "/tmp/test.txt"}`
	if err := tool.Validate(validArgs); err != nil {
		t.Errorf("Valid args should not fail validation: %v", err)
	}

	invalidArgs := `{"file_path": ""}`
	if err := tool.Validate(invalidArgs); err == nil {
		t.Error("Empty path should fail validation")
	}
}

func TestFileWriteTool(t *testing.T) {
	tool := NewFileWriteTool()

	if tool.GetName() != "file_write" {
		t.Error("File write tool name should be 'file_write'")
	}

	if tool.GetDescription() == "" {
		t.Error("File write tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "file_write" {
		t.Error("Definition function name should be 'file_write'")
	}

	// Test validation
	validArgs := `{"file_path": "/tmp/test.txt", "content": "hello world"}`
	if err := tool.Validate(validArgs); err != nil {
		t.Errorf("Valid args should not fail validation: %v", err)
	}

	invalidArgs := `{"file_path": "", "content": "hello"}`
	if err := tool.Validate(invalidArgs); err == nil {
		t.Error("Empty path should fail validation")
	}
}

func TestFileListTool(t *testing.T) {
	tool := NewFileListTool()

	if tool.GetName() != "file_list" {
		t.Error("File list tool name should be 'file_list'")
	}

	if tool.GetDescription() == "" {
		t.Error("File list tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "file_list" {
		t.Error("Definition function name should be 'file_list'")
	}

	// Test execution with current directory
	ctx := context.Background()
	args := `{"path": ".", "recursive": false}`

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("File list execution failed: %v", err)
	}

	if result == "" {
		t.Error("File list should return a result")
	}

	// Test config
	config := tool.GetConfig()
	if config == nil {
		t.Error("Tool should have config")
	}

	newConfig := map[string]interface{}{
		"max_items": 50,
	}
	if err := tool.SetConfig(newConfig); err != nil {
		t.Errorf("Setting config should not fail: %v", err)
	}
}

func TestShellTool(t *testing.T) {
	tool := NewShellTool()

	if tool.GetName() != "shell" {
		t.Error("Shell tool name should be 'shell'")
	}

	if tool.GetDescription() == "" {
		t.Error("Shell tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "shell" {
		t.Error("Definition function name should be 'shell'")
	}

	// Test validation
	validArgs := `{"command": "echo hello"}`
	if err := tool.Validate(validArgs); err != nil {
		t.Errorf("Valid args should not fail validation: %v", err)
	}

	invalidArgs := `{"command": ""}`
	if err := tool.Validate(invalidArgs); err == nil {
		t.Error("Empty command should fail validation")
	}

	// Test safe execution
	ctx := context.Background()
	args := `{"command": "echo test"}`

	result, err := tool.Execute(ctx, args)
	if err != nil {
		t.Fatalf("Shell execution failed: %v", err)
	}

	if !strings.Contains(result, "test") {
		t.Error("Shell should return command output")
	}
}

func TestHTTPTool(t *testing.T) {
	tool := NewHTTPTool()

	if tool.GetName() != "http_request" {
		t.Error("HTTP tool name should be 'http_request'")
	}

	if tool.GetDescription() == "" {
		t.Error("HTTP tool should have a description")
	}

	// Test definition
	def := tool.GetDefinition()
	if def.Function.Name != "http_request" {
		t.Error("Definition function name should be 'http_request'")
	}

	// Test validation
	validArgs := `{"url": "https://httpbin.org/get", "method": "GET"}`
	if err := tool.Validate(validArgs); err != nil {
		t.Errorf("Valid args should not fail validation: %v", err)
	}

	invalidArgs := `{"url": "", "method": "GET"}`
	if err := tool.Validate(invalidArgs); err == nil {
		t.Error("Empty URL should fail validation")
	}

	// Test config
	config := tool.GetConfig()
	if config == nil {
		t.Error("Tool should have config")
	}

	newConfig := map[string]interface{}{
		"timeout": 30,
	}
	if err := tool.SetConfig(newConfig); err != nil {
		t.Errorf("Setting config should not fail: %v", err)
	}
}

func TestToolDefinitionSerialization(t *testing.T) {
	tool := NewCalculatorTool()
	def := tool.GetDefinition()

	// Test JSON serialization
	jsonData, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("Failed to serialize tool definition: %v", err)
	}

	var deserializedDef llm.ToolDefinition
	if err := json.Unmarshal(jsonData, &deserializedDef); err != nil {
		t.Fatalf("Failed to deserialize tool definition: %v", err)
	}

	if deserializedDef.Function.Name != def.Function.Name {
		t.Error("Deserialized definition should match original")
	}
}

func TestToolRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewToolRegistry()

	// Test concurrent registration and access
	done := make(chan bool, 10)

	// Concurrent registrations
	for i := 0; i < 5; i++ {
		go func(id int) {
			mockTool := &MockTool{
				name:        fmt.Sprintf("mock_tool_%d", id),
				description: fmt.Sprintf("Mock tool %d", id),
			}
			registry.RegisterTool(mockTool)
			done <- true
		}(i)
	}

	// Concurrent access
	for i := 0; i < 5; i++ {
		go func() {
			registry.ListTools()
			registry.GetAllDefinitions()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all tools were registered
	tools := registry.ListTools()
	mockToolCount := 0
	for _, tool := range tools {
		if strings.HasPrefix(tool, "mock_tool_") {
			mockToolCount++
		}
	}

	if mockToolCount != 5 {
		t.Errorf("Expected 5 mock tools, got %d", mockToolCount)
	}
}

// MockTool for testing
type MockTool struct {
	name        string
	description string
}

func (m *MockTool) GetName() string {
	return m.name
}

func (m *MockTool) GetDescription() string {
	return m.description
}

func (m *MockTool) GetDefinition() llm.ToolDefinition {
	return llm.ToolDefinition{
		Type: "function",
		Function: llm.Function{
			Name:        m.name,
			Description: m.description,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]interface{}{
						"type":        "string",
						"description": "Input for mock tool",
					},
				},
				"required": []string{"input"},
			},
		},
	}
}

func (m *MockTool) Execute(ctx context.Context, args string) (string, error) {
	return "mock result", nil
}

func (m *MockTool) Validate(args string) error {
	return nil
}

func (m *MockTool) GetConfig() map[string]interface{} {
	return map[string]interface{}{}
}

func (m *MockTool) SetConfig(config map[string]interface{}) error {
	return nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkToolRegistry_RegisterTool(b *testing.B) {
	registry := NewToolRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockTool := &MockTool{
			name:        fmt.Sprintf("benchmark_tool_%d", i),
			description: "Benchmark tool",
		}
		registry.RegisterTool(mockTool)
	}
}

func BenchmarkToolRegistry_GetTool(b *testing.B) {
	registry := NewToolRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.GetTool("calculator")
	}
}

func BenchmarkToolRegistry_ListTools(b *testing.B) {
	registry := NewToolRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.ListTools()
	}
}

func BenchmarkCalculatorTool_Execute(b *testing.B) {
	tool := NewCalculatorTool()
	ctx := context.Background()
	args := `{"expression": "2 + 3 * 4"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, args)
	}
}

func BenchmarkTimeTool_Execute(b *testing.B) {
	tool := NewTimeTool()
	ctx := context.Background()
	args := `{"format": "RFC3339", "timezone": "UTC"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.Execute(ctx, args)
	}
}
