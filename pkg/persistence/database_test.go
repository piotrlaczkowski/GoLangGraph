package persistence

import (
	"context"
	"testing"
)

func TestNewPostgresConfig(t *testing.T) {
	config := NewPostgresConfig("localhost", 5432, "testdb", "testuser", "testpass")

	if config.Type != DatabaseTypePostgres {
		t.Error("Config type should be PostgreSQL")
	}

	if config.Host != "localhost" {
		t.Error("Config host should be localhost")
	}

	if config.Port != 5432 {
		t.Error("Config port should be 5432")
	}

	if config.Database != "testdb" {
		t.Error("Config database should be testdb")
	}

	if config.Username != "testuser" {
		t.Error("Config username should be testuser")
	}

	if config.Password != "testpass" {
		t.Error("Config password should be testpass")
	}
}

func TestNewPgVectorConfig(t *testing.T) {
	config := NewPgVectorConfig("localhost", 5432, "testdb", "testuser", "testpass", 1536)

	if config.Type != DatabaseTypePgVector {
		t.Error("Config type should be PgVector")
	}

	if config.VectorDimension != 1536 {
		t.Error("Config vector dimension should be 1536")
	}

	if config.EnableRAG != true {
		t.Error("Config should have RAG enabled")
	}
}

func TestNewRedisConfig(t *testing.T) {
	config := NewRedisConfig("localhost", 6379, "testpass")

	if config.Type != DatabaseTypeRedis {
		t.Error("Config type should be Redis")
	}

	if config.Host != "localhost" {
		t.Error("Config host should be localhost")
	}

	if config.Port != 6379 {
		t.Error("Config port should be 6379")
	}

	if config.Password != "testpass" {
		t.Error("Config password should be testpass")
	}
}

func TestNewDatabaseConnectionManager(t *testing.T) {
	manager := NewDatabaseConnectionManager()

	if manager == nil {
		t.Fatal("NewDatabaseConnectionManager() returned nil")
	}

	if manager.connections == nil {
		t.Error("Manager connections should be initialized")
	}
}

func TestDatabaseConnectionManager_AddConnection(t *testing.T) {
	manager := NewDatabaseConnectionManager()

	// Test adding PostgreSQL connection
	config := NewPostgresConfig("localhost", 5432, "testdb", "testuser", "testpass")
	err := manager.AddConnection("test_postgres", config)

	// This will fail without actual database, but we can test the error handling
	if err == nil {
		t.Error("AddConnection should return error when database is not available")
	}
}

func TestDatabaseConnectionManager_GetConnection(t *testing.T) {
	manager := NewDatabaseConnectionManager()

	// Test getting non-existent connection
	_, err := manager.GetConnection("nonexistent")
	if err == nil {
		t.Error("Should return error for non-existent connection")
	}
}

func TestCreateCheckpointer(t *testing.T) {
	// Test creating PostgreSQL checkpointer
	config := NewPostgresConfig("localhost", 5432, "testdb", "testuser", "testpass")

	_, err := CreateCheckpointer(config)
	// This will fail without actual database, but we can test the error handling
	if err == nil {
		t.Error("CreateCheckpointer should return error when database is not available")
	}

	// Test creating Redis checkpointer
	redisConfig := NewRedisConfig("localhost", 6379, "testpass")

	_, err = CreateCheckpointer(redisConfig)
	// This will fail without actual Redis, but we can test the error handling
	if err == nil {
		t.Error("CreateCheckpointer should return error when Redis is not available")
	}

	// Test creating checkpointer with unsupported type
	unsupportedConfig := &DatabaseConfig{
		Type: DatabaseType("unsupported"),
	}

	_, err = CreateCheckpointer(unsupportedConfig)
	if err == nil {
		t.Error("CreateCheckpointer should return error for unsupported type")
	}
}

func TestDocument_Validate(t *testing.T) {
	// Test valid document
	doc := &Document{
		ID:       "test_id",
		ThreadID: "test_thread",
		Content:  "test content",
		Metadata: map[string]interface{}{"key": "value"},
	}

	// Document doesn't have a Validate method, so we'll test its fields
	if doc.ID == "" {
		t.Error("Document ID should not be empty")
	}

	if doc.Content == "" {
		t.Error("Document content should not be empty")
	}

	if doc.Metadata == nil {
		t.Error("Document metadata should be initialized")
	}
}

func TestSession_Fields(t *testing.T) {
	session := &Session{
		ID:       "test_session",
		ThreadID: "test_thread",
		UserID:   "test_user",
		Metadata: map[string]interface{}{"key": "value"},
	}

	if session.ID != "test_session" {
		t.Error("Session ID should be set correctly")
	}

	if session.ThreadID != "test_thread" {
		t.Error("Session ThreadID should be set correctly")
	}

	if session.UserID != "test_user" {
		t.Error("Session UserID should be set correctly")
	}
}

func TestThread_Fields(t *testing.T) {
	thread := &Thread{
		ID:       "test_thread",
		Name:     "Test Thread",
		Metadata: map[string]interface{}{"key": "value"},
	}

	if thread.ID != "test_thread" {
		t.Error("Thread ID should be set correctly")
	}

	if thread.Name != "Test Thread" {
		t.Error("Thread Name should be set correctly")
	}

	if thread.Metadata == nil {
		t.Error("Thread metadata should be initialized")
	}
}

func TestDatabaseTypes(t *testing.T) {
	// Test that all database types are defined correctly
	types := []DatabaseType{
		DatabaseTypePostgres,
		DatabaseTypePostgresQL,
		DatabaseTypePgVector,
		DatabaseTypeRedis,
		DatabaseTypeOpenSearch,
		DatabaseTypeElastic,
		DatabaseTypeMongoDB,
		DatabaseTypeMySQL,
		DatabaseTypeSQLite,
	}

	for _, dbType := range types {
		if string(dbType) == "" {
			t.Errorf("Database type %s should not be empty", dbType)
		}
	}
}

func TestDatabaseConfig_Fields(t *testing.T) {
	config := &DatabaseConfig{
		Type:         DatabaseTypePostgres,
		Host:         "localhost",
		Port:         5432,
		Database:     "testdb",
		Username:     "testuser",
		Password:     "testpass",
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  "5m",
	}

	if config.Type != DatabaseTypePostgres {
		t.Error("Config type should be PostgreSQL")
	}

	if config.Host != "localhost" {
		t.Error("Config host should be localhost")
	}

	if config.Port != 5432 {
		t.Error("Config port should be 5432")
	}

	if config.MaxOpenConns != 25 {
		t.Error("Config MaxOpenConns should be 25")
	}

	if config.MaxIdleConns != 5 {
		t.Error("Config MaxIdleConns should be 5")
	}
}

func TestDatabaseConfig_VectorFields(t *testing.T) {
	config := &DatabaseConfig{
		Type:            DatabaseTypePgVector,
		VectorDimension: 1536,
		VectorMetric:    "cosine",
		EnableRAG:       true,
		EmbeddingModel:  "text-embedding-ada-002",
	}

	if config.VectorDimension != 1536 {
		t.Error("Config VectorDimension should be 1536")
	}

	if config.VectorMetric != "cosine" {
		t.Error("Config VectorMetric should be cosine")
	}

	if config.EnableRAG != true {
		t.Error("Config EnableRAG should be true")
	}

	if config.EmbeddingModel != "text-embedding-ada-002" {
		t.Error("Config EmbeddingModel should be text-embedding-ada-002")
	}
}

// Mock connection for testing (without actual database connection)
type MockConnection struct {
	dbType DatabaseType
	config *DatabaseConfig
}

func (m *MockConnection) Connect() error {
	return nil
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) Ping() error {
	return nil
}

func (m *MockConnection) GetType() DatabaseType {
	return m.dbType
}

func (m *MockConnection) GetConfig() *DatabaseConfig {
	return m.config
}

func (m *MockConnection) ExecuteQuery(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

func (m *MockConnection) QueryRow(ctx context.Context, query string, args ...interface{}) interface{} {
	return nil
}

func (m *MockConnection) QueryRows(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

// Integration tests (these would require actual database connections)
func TestPostgresCheckpointer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real PostgreSQL database
	// For now, we'll skip it unless explicitly running integration tests
	t.Skip("Integration test requires PostgreSQL database")
}

func TestRedisCheckpointer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real Redis instance
	// For now, we'll skip it unless explicitly running integration tests
	t.Skip("Integration test requires Redis instance")
}

// Benchmark tests
func BenchmarkNewPostgresConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewPostgresConfig("localhost", 5432, "testdb", "testuser", "testpass")
	}
}

func BenchmarkNewPgVectorConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewPgVectorConfig("localhost", 5432, "testdb", "testuser", "testpass", 1536)
	}
}

func BenchmarkNewRedisConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewRedisConfig("localhost", 6379, "testpass")
	}
}
