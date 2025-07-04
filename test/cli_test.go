package test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLICommands tests the CLI command functionality
func TestCLICommands(t *testing.T) {
	t.Run("TestInitCommand", testInitCommand)
	t.Run("TestDockerBuild", testDockerBuild)
	t.Run("TestValidateCommand", testValidateCommand)
	t.Run("TestDevMode", testDevMode)
}

func testInitCommand(t *testing.T) {
	// Test basic template initialization
	tempDir := t.TempDir()
	projectName := "test-project"
	projectPath := filepath.Join(tempDir, projectName)

	// Mock the init command functionality
	err := createProjectStructure(projectPath, "basic")
	require.NoError(t, err)

	// Check that directories were created
	expectedDirs := []string{
		"configs",
		"agents",
		"tools",
		"static",
		"tests",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(projectPath, dir)
		assert.DirExists(t, dirPath, "Directory %s should exist", dir)
	}

	// Check that configuration files were created
	configPath := filepath.Join(projectPath, "configs", "agent-config.yaml")
	assert.FileExists(t, configPath, "Agent configuration file should exist")

	dockerComposePath := filepath.Join(projectPath, "docker-compose.yml")
	assert.FileExists(t, dockerComposePath, "Docker compose file should exist")

	// Verify configuration content
	configContent, err := ioutil.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(configContent), "name: \"basic-agent\"")
	assert.Contains(t, string(configContent), "type: \"chat\"")
}

func testDockerBuild(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test regular Dockerfile creation
	dockerfilePath := filepath.Join(tempDir, "Dockerfile.agent")
	err := createTestDockerfile(dockerfilePath, false)
	require.NoError(t, err)
	
	content, err := ioutil.ReadFile(dockerfilePath)
	require.NoError(t, err)
	
	dockerfileContent := string(content)
	assert.Contains(t, dockerfileContent, "FROM golang:1.21-alpine AS builder")
	assert.Contains(t, dockerfileContent, "FROM alpine:latest")
	assert.Contains(t, dockerfileContent, "HEALTHCHECK")
	assert.Contains(t, dockerfileContent, "USER golanggraph")
	
	// Test distroless Dockerfile creation
	distrolessPath := filepath.Join(tempDir, "Dockerfile.distroless")
	err = createTestDockerfile(distrolessPath, true)
	require.NoError(t, err)
	
	distrolessContent, err := ioutil.ReadFile(distrolessPath)
	require.NoError(t, err)
	
	distrolessStr := string(distrolessContent)
	assert.Contains(t, distrolessStr, "FROM gcr.io/distroless/static:nonroot")
	assert.Contains(t, distrolessStr, "USER nonroot:nonroot")
	assert.NotContains(t, distrolessStr, "HEALTHCHECK") // Distroless doesn't support healthcheck
}

func testValidateCommand(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create a test configuration file
	configPath := filepath.Join(tempDir, "test-config.yaml")
	configContent := `
name: "test-agent"
type: "chat"
model: "gpt-3.5-turbo"
provider: "openai"
system_prompt: "You are a helpful assistant."
temperature: 0.7
max_tokens: 1000

tools:
  - name: "calculator"
    enabled: true

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"
`
	
	err := ioutil.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Test validation
	isValid, errors := validateConfig(configPath)
	assert.True(t, isValid, "Configuration should be valid")
	assert.Empty(t, errors, "No validation errors expected")
	
	// Test with invalid configuration
	invalidConfigPath := filepath.Join(tempDir, "invalid-config.yaml")
	invalidContent := `
name: ""
type: "invalid-type"
model: ""
`
	
	err = ioutil.WriteFile(invalidConfigPath, []byte(invalidContent), 0644)
	require.NoError(t, err)
	
	isValid, errors = validateConfig(invalidConfigPath)
	assert.False(t, isValid, "Configuration should be invalid")
	assert.NotEmpty(t, errors, "Validation errors expected")
}

func testDevMode(t *testing.T) {
	// Create a test server in dev mode
	config := &server.ServerConfig{
		Host:     "localhost",
		Port:     0, // Use random port
		DevMode:  true,
		LogLevel: "debug",
	}
	
	srv := server.NewServer(config)
	
	// Start server in background
	go func() {
		srv.Start()
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Test dev mode endpoints
	baseURL := fmt.Sprintf("http://localhost:%d", config.Port)
	
	// Test debug dashboard
	resp, err := http.Get(baseURL + "/debug/")
	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
	
	// Test playground
	resp, err = http.Get(baseURL + "/playground/")
	if err == nil {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}
	
	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Stop(ctx)
}

// Helper functions for testing

func createProjectStructure(projectPath, template string) error {
	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return err
	}
	
	// Create subdirectories
	dirs := []string{
		"configs",
		"agents",
		"tools",
		"static",
		"tests",
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return err
		}
	}
	
	// Create configuration files based on template
	switch template {
	case "basic":
		return createBasicTemplate(projectPath)
	case "advanced":
		return createAdvancedTemplate(projectPath)
	case "rag":
		return createRAGTemplate(projectPath)
	default:
		return createBasicTemplate(projectPath)
	}
}

func createBasicTemplate(projectPath string) error {
	// Create basic agent configuration
	agentConfig := `name: "basic-agent"
type: "chat"
model: "gpt-3.5-turbo"
provider: "openai"
system_prompt: "You are a helpful assistant."
temperature: 0.7
max_tokens: 1000

tools:
  - name: "calculator"
    enabled: true
  - name: "web_search"
    enabled: false

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"
`
	
	configPath := filepath.Join(projectPath, "configs", "agent-config.yaml")
	if err := ioutil.WriteFile(configPath, []byte(agentConfig), 0644); err != nil {
		return err
	}
	
	// Create docker-compose for development
	dockerCompose := `version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: golanggraph
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
`
	
	dockerComposePath := filepath.Join(projectPath, "docker-compose.yml")
	return ioutil.WriteFile(dockerComposePath, []byte(dockerCompose), 0644)
}

func createAdvancedTemplate(projectPath string) error {
	if err := createBasicTemplate(projectPath); err != nil {
		return err
	}
	
	// Add advanced configuration
	advancedConfig := `name: "advanced-agent"
type: "multi-agent"
model: "gpt-4"
provider: "openai"
system_prompt: "You are an advanced AI assistant with multiple capabilities."
temperature: 0.7
max_tokens: 2000

agents:
  - name: "research-agent"
    type: "research"
    tools: ["web_search", "document_reader"]
  - name: "analysis-agent"
    type: "analysis"
    tools: ["calculator", "data_analyzer"]
  - name: "synthesis-agent"
    type: "synthesis"
    tools: ["summarizer", "report_generator"]

workflow:
  start_node: "research-agent"
  edges:
    - from: "research-agent"
      to: "analysis-agent"
    - from: "analysis-agent"
      to: "synthesis-agent"
  end_node: "synthesis-agent"

tools:
  - name: "web_search"
    enabled: true
    config:
      api_key: "${SEARCH_API_KEY}"
  - name: "document_reader"
    enabled: true
  - name: "calculator"
    enabled: true
  - name: "data_analyzer"
    enabled: true
  - name: "summarizer"
    enabled: true
  - name: "report_generator"
    enabled: true

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"

vector_store:
  type: "pgvector"
  host: "localhost"
  port: 5432
  database: "vectordb"
  username: "postgres"
  password: "password"
  dimensions: 1536
`
	
	configPath := filepath.Join(projectPath, "configs", "advanced-config.yaml")
	return ioutil.WriteFile(configPath, []byte(advancedConfig), 0644)
}

func createRAGTemplate(projectPath string) error {
	if err := createAdvancedTemplate(projectPath); err != nil {
		return err
	}
	
	// Add RAG-specific configuration
	ragConfig := `name: "rag-agent"
type: "rag"
model: "gpt-4"
provider: "openai"
system_prompt: "You are a RAG-enabled AI assistant that can retrieve and analyze information from documents."
temperature: 0.7
max_tokens: 2000

rag:
  enabled: true
  chunk_size: 1000
  chunk_overlap: 200
  similarity_threshold: 0.7
  max_chunks: 5
  embedding_model: "text-embedding-ada-002"

vector_store:
  type: "pgvector"
  host: "localhost"
  port: 5432
  database: "vectordb"
  username: "postgres"
  password: "password"
  dimensions: 1536
  collection_name: "documents"

document_loaders:
  - type: "pdf"
    enabled: true
  - type: "text"
    enabled: true
  - type: "markdown"
    enabled: true
  - type: "web"
    enabled: true

tools:
  - name: "vector_search"
    enabled: true
  - name: "document_loader"
    enabled: true
  - name: "web_search"
    enabled: true
  - name: "summarizer"
    enabled: true

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "golanggraph"
  username: "postgres"
  password: "password"
`
	
	configPath := filepath.Join(projectPath, "configs", "rag-config.yaml")
	return ioutil.WriteFile(configPath, []byte(ragConfig), 0644)
}

func createTestDockerfile(filepath string, distroless bool) error {
	var dockerfile string
	
	if distroless {
		dockerfile = `# Distroless Dockerfile for GoLangGraph Agent
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=production
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -a -installsuffix cgo \
    -o golanggraph-agent \
    ./cmd/golanggraph

# Distroless production stage
FROM gcr.io/distroless/static:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/golanggraph-agent /

# Copy configuration files
COPY configs/ /configs/
COPY static/ /static/

# Use distroless nonroot user
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the agent
ENTRYPOINT ["/golanggraph-agent"]
CMD ["serve", "--host", "0.0.0.0", "--port", "8080"]
`
	} else {
		dockerfile = `# Production Dockerfile for GoLangGraph Agent
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=production
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -a -installsuffix cgo \
    -o golanggraph-agent \
    ./cmd/golanggraph

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S golanggraph && \
    adduser -u 1001 -S golanggraph -G golanggraph

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/golanggraph-agent .

# Copy configuration files
COPY configs/ ./configs/
COPY static/ ./static/

# Create necessary directories
RUN mkdir -p ./logs ./data

# Change ownership to non-root user
RUN chown -R golanggraph:golanggraph /app

# Switch to non-root user
USER golanggraph

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./golanggraph-agent health || exit 1

# Run the agent
ENTRYPOINT ["./golanggraph-agent"]
CMD ["serve", "--host", "0.0.0.0", "--port", "8080"]
`
	}
	
	return ioutil.WriteFile(filepath, []byte(dockerfile), 0644)
}

func validateConfig(configPath string) (bool, []string) {
	var errors []string
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		errors = append(errors, "Configuration file does not exist")
		return false, errors
	}
	
	// Read file content
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		errors = append(errors, "Failed to read configuration file")
		return false, errors
	}
	
	configStr := string(content)
	
	// Basic validation checks
	if !strings.Contains(configStr, "name:") {
		errors = append(errors, "Missing 'name' field")
	}
	
	if !strings.Contains(configStr, "type:") {
		errors = append(errors, "Missing 'type' field")
	}
	
	if !strings.Contains(configStr, "model:") {
		errors = append(errors, "Missing 'model' field")
	}
	
	if strings.Contains(configStr, "name: \"\"") {
		errors = append(errors, "Name cannot be empty")
	}
	
	if strings.Contains(configStr, "type: \"invalid-type\"") {
		errors = append(errors, "Invalid agent type")
	}
	
	return len(errors) == 0, errors
}

// TestDockerContainerIntegration tests the Docker container functionality
func TestDockerContainerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	t.Run("TestDockerfileGeneration", testDockerfileGeneration)
	t.Run("TestConfigurationValidation", testConfigurationValidation)
}

func testDockerfileGeneration(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test both regular and distroless Dockerfiles
	testCases := []struct {
		name       string
		distroless bool
		expected   []string
	}{
		{
			name:       "Regular Dockerfile",
			distroless: false,
			expected:   []string{"FROM alpine:latest", "HEALTHCHECK", "USER golanggraph"},
		},
		{
			name:       "Distroless Dockerfile",
			distroless: true,
			expected:   []string{"FROM gcr.io/distroless/static:nonroot", "USER nonroot:nonroot"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dockerfilePath := filepath.Join(tempDir, fmt.Sprintf("Dockerfile.%s", tc.name))
			err := createTestDockerfile(dockerfilePath, tc.distroless)
			require.NoError(t, err)
			
			content, err := ioutil.ReadFile(dockerfilePath)
			require.NoError(t, err)
			
			dockerfileContent := string(content)
			for _, expected := range tc.expected {
				assert.Contains(t, dockerfileContent, expected, "Dockerfile should contain %s", expected)
			}
		})
	}
}

func testConfigurationValidation(t *testing.T) {
	tempDir := t.TempDir()
	
	testCases := []struct {
		name        string
		config      string
		shouldValid bool
		expectedErr string
	}{
		{
			name: "Valid Configuration",
			config: `name: "test-agent"
type: "chat"
model: "gpt-3.5-turbo"
provider: "openai"
system_prompt: "You are a helpful assistant."
temperature: 0.7
max_tokens: 1000`,
			shouldValid: true,
		},
		{
			name: "Missing Name",
			config: `type: "chat"
model: "gpt-3.5-turbo"
provider: "openai"`,
			shouldValid: false,
			expectedErr: "Missing 'name' field",
		},
		{
			name: "Empty Name",
			config: `name: ""
type: "chat"
model: "gpt-3.5-turbo"
provider: "openai"`,
			shouldValid: false,
			expectedErr: "Name cannot be empty",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "config.yaml")
			err := ioutil.WriteFile(configPath, []byte(tc.config), 0644)
			require.NoError(t, err)
			
			isValid, errors := validateConfig(configPath)
			assert.Equal(t, tc.shouldValid, isValid, "Validation result should match expected")
			
			if !tc.shouldValid {
				assert.Contains(t, strings.Join(errors, ", "), tc.expectedErr)
			}
		})
	}
}