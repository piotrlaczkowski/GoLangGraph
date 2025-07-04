// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

func RunDatabasePersistenceDemo() {
	fmt.Println("=== GoLangGraph Database Persistence Demo ===")

	// Demo 1: PostgreSQL Basic Connection
	fmt.Println("\n1. PostgreSQL Basic Connection")
	demoPostgresBasic()

	// Demo 2: PostgreSQL with pgvector for RAG
	fmt.Println("\n2. PostgreSQL with pgvector for RAG")
	demoPostgresPgVector()

	// Demo 3: Redis for Fast Checkpointing
	fmt.Println("\n3. Redis for Fast Checkpointing")
	demoRedis()

	// Demo 4: Database Connection Manager
	fmt.Println("\n4. Database Connection Manager")
	demoConnectionManager()

	// Demo 5: RAG Document Storage
	fmt.Println("\n5. RAG Document Storage")
	demoRAGStorage()
}

// demoPostgresBasic demonstrates basic PostgreSQL connection and checkpointing
func demoPostgresBasic() {
	// Create PostgreSQL configuration
	config := persistence.NewPostgresConfig("localhost", 5432, "golanggraph", "postgres", "password")

	// Create checkpointer
	checkpointer, err := persistence.NewPostgresCheckpointer(config)
	if err != nil {
		log.Printf("Failed to create PostgreSQL checkpointer (this is expected if PostgreSQL is not running): %v", err)
		return
	}
	defer checkpointer.Close()

	// Create a sample checkpoint
	state := core.NewBaseState()
	state.Set("step", 1)
	state.Set("message", "Hello World")

	checkpoint := &persistence.Checkpoint{
		ID:        uuid.New().String(),
		ThreadID:  "thread-123",
		State:     state,
		Metadata:  map[string]interface{}{"agent": "demo", "version": "1.0"},
		CreatedAt: time.Now(),
		NodeID:    "start_node",
		StepID:    1,
	}

	ctx := context.Background()

	// Save checkpoint
	if err := checkpointer.Save(ctx, checkpoint); err != nil {
		log.Printf("Failed to save checkpoint: %v", err)
		return
	}

	fmt.Printf("✓ Saved checkpoint: %s\n", checkpoint.ID)

	// Load checkpoint
	loaded, err := checkpointer.Load(ctx, checkpoint.ThreadID, checkpoint.ID)
	if err != nil {
		log.Printf("Failed to load checkpoint: %v", err)
		return
	}

	fmt.Printf("✓ Loaded checkpoint: %s (Step: %d)\n", loaded.ID, loaded.StepID)

	// List checkpoints
	checkpoints, err := checkpointer.List(ctx, checkpoint.ThreadID)
	if err != nil {
		log.Printf("Failed to list checkpoints: %v", err)
		return
	}

	fmt.Printf("✓ Found %d checkpoints for thread %s\n", len(checkpoints), checkpoint.ThreadID)
}

// demoPostgresPgVector demonstrates PostgreSQL with pgvector for RAG
func demoPostgresPgVector() {
	// Create PostgreSQL with pgvector configuration
	config := persistence.NewPgVectorConfig("localhost", 5432, "golanggraph", "postgres", "password", 1536)

	// Create checkpointer with RAG support
	checkpointer, err := persistence.NewPostgresCheckpointer(config)
	if err != nil {
		log.Printf("Failed to create PostgreSQL+pgvector checkpointer (this is expected if PostgreSQL with pgvector is not running): %v", err)
		return
	}
	defer checkpointer.Close()

	fmt.Printf("✓ Connected to PostgreSQL with pgvector (RAG enabled)\n")

	// Create a sample document for RAG
	doc := &persistence.Document{
		ID:        uuid.New().String(),
		ThreadID:  "thread-456",
		Content:   "This is a sample document for RAG demonstration. It contains important information about the system.",
		Metadata:  map[string]interface{}{"source": "demo", "category": "documentation"},
		Embedding: generateMockEmbedding(1536), // Mock embedding vector
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.Background()

	// Save document
	if err := checkpointer.SaveDocument(ctx, doc); err != nil {
		log.Printf("Failed to save document: %v", err)
		return
	}

	fmt.Printf("✓ Saved document: %s\n", doc.ID)

	// Search documents using vector similarity
	queryEmbedding := generateMockEmbedding(1536)
	documents, err := checkpointer.SearchDocuments(ctx, doc.ThreadID, queryEmbedding, 5)
	if err != nil {
		log.Printf("Failed to search documents: %v", err)
		return
	}

	fmt.Printf("✓ Found %d similar documents\n", len(documents))
}

// demoRedis demonstrates Redis connection for fast checkpointing
func demoRedis() {
	// Create Redis configuration
	config := persistence.NewRedisConfig("localhost", 6379, "")

	// Create Redis checkpointer
	checkpointer, err := persistence.NewRedisCheckpointer(config)
	if err != nil {
		log.Printf("Failed to create Redis checkpointer (this is expected if Redis is not running): %v", err)
		return
	}
	defer checkpointer.Close()

	// Create a sample checkpoint
	state := core.NewBaseState()
	state.Set("step", 1)
	state.Set("cached", true)

	checkpoint := &persistence.Checkpoint{
		ID:        uuid.New().String(),
		ThreadID:  "thread-789",
		State:     state,
		Metadata:  map[string]interface{}{"cache": "redis", "ttl": "24h"},
		CreatedAt: time.Now(),
		NodeID:    "cache_node",
		StepID:    1,
	}

	ctx := context.Background()

	// Save checkpoint to Redis
	if err := checkpointer.Save(ctx, checkpoint); err != nil {
		log.Printf("Failed to save checkpoint to Redis: %v", err)
		return
	}

	fmt.Printf("✓ Saved checkpoint to Redis: %s\n", checkpoint.ID)

	// Load checkpoint from Redis
	loaded, err := checkpointer.Load(ctx, checkpoint.ThreadID, checkpoint.ID)
	if err != nil {
		log.Printf("Failed to load checkpoint from Redis: %v", err)
		return
	}

	fmt.Printf("✓ Loaded checkpoint from Redis: %s\n", loaded.ID)
}

// demoConnectionManager demonstrates managing multiple database connections
func demoConnectionManager() {
	// Create connection manager
	manager := persistence.NewDatabaseConnectionManager()
	defer manager.CloseAll()

	// Add PostgreSQL connection
	postgresConfig := persistence.NewPostgresConfig("localhost", 5432, "golanggraph", "postgres", "password")
	if err := manager.AddConnection("postgres-main", postgresConfig); err != nil {
		log.Printf("Failed to add PostgreSQL connection: %v", err)
	} else {
		fmt.Printf("✓ Added PostgreSQL connection: postgres-main\n")
	}

	// Add pgvector connection
	pgvectorConfig := persistence.NewPgVectorConfig("localhost", 5432, "golanggraph_rag", "postgres", "password", 1536)
	if err := manager.AddConnection("postgres-rag", pgvectorConfig); err != nil {
		log.Printf("Failed to add pgvector connection: %v", err)
	} else {
		fmt.Printf("✓ Added pgvector connection: postgres-rag\n")
	}

	// Try to add unsupported database (will show error message)
	mongoConfig := &persistence.DatabaseConfig{
		Type:     persistence.DatabaseTypeMongoDB,
		Host:     "localhost",
		Port:     27017,
		Database: "golanggraph",
	}
	if err := manager.AddConnection("mongo-main", mongoConfig); err != nil {
		fmt.Printf("✓ Expected error for unsupported database: %v\n", err)
	}

	// Get connection
	if conn, err := manager.GetConnection("postgres-main"); err != nil {
		log.Printf("Failed to get connection: %v", err)
	} else {
		fmt.Printf("✓ Retrieved connection: %s\n", conn.GetType())
	}
}

// demoRAGStorage demonstrates RAG document storage and retrieval
func demoRAGStorage() {
	// Create configuration with RAG enabled
	config := &persistence.DatabaseConfig{
		Type:                persistence.DatabaseTypePostgres,
		Host:                "localhost",
		Port:                5432,
		Database:            "golanggraph_rag",
		Username:            "postgres",
		Password:            "password",
		SSLMode:             "disable",
		EnableRAG:           true,
		EmbeddingDimension:  1536,
		SimilarityThreshold: 0.7,
	}

	// Create checkpointer with RAG support
	checkpointer, err := persistence.NewPostgresCheckpointer(config)
	if err != nil {
		log.Printf("Failed to create RAG checkpointer (this is expected if database is not available): %v", err)
		return
	}
	defer checkpointer.Close()

	ctx := context.Background()
	threadID := "rag-demo-thread"

	// Create sample documents
	documents := []*persistence.Document{
		{
			ID:        uuid.New().String(),
			ThreadID:  threadID,
			Content:   "GoLangGraph is a powerful framework for building stateful AI agents with graph-based workflows.",
			Metadata:  map[string]interface{}{"topic": "framework", "category": "introduction"},
			Embedding: generateMockEmbedding(1536),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			ThreadID:  threadID,
			Content:   "The persistence layer supports PostgreSQL, Redis, and vector databases for comprehensive data management.",
			Metadata:  map[string]interface{}{"topic": "persistence", "category": "architecture"},
			Embedding: generateMockEmbedding(1536),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New().String(),
			ThreadID:  threadID,
			Content:   "RAG (Retrieval-Augmented Generation) enables AI agents to access and utilize external knowledge sources.",
			Metadata:  map[string]interface{}{"topic": "rag", "category": "ai"},
			Embedding: generateMockEmbedding(1536),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Save documents
	for _, doc := range documents {
		if err := checkpointer.SaveDocument(ctx, doc); err != nil {
			log.Printf("Failed to save document %s: %v", doc.ID, err)
			return
		}
		fmt.Printf("✓ Saved RAG document: %s\n", doc.ID)
	}

	// Search for similar documents
	queryEmbedding := generateMockEmbedding(1536)
	results, err := checkpointer.SearchDocuments(ctx, threadID, queryEmbedding, 3)
	if err != nil {
		log.Printf("Failed to search documents: %v", err)
		return
	}

	fmt.Printf("✓ Found %d similar documents for RAG query\n", len(results))
	for i, doc := range results {
		fmt.Printf("  %d. %s (Topic: %v)\n", i+1, doc.Content[:50]+"...", doc.Metadata["topic"])
	}
}

// generateMockEmbedding generates a mock embedding vector for demonstration
func generateMockEmbedding(dimension int) []float64 {
	embedding := make([]float64, dimension)
	for i := 0; i < dimension; i++ {
		embedding[i] = float64(i%100) / 100.0 // Simple pattern for demo
	}
	return embedding
}

// Helper functions for different database configurations

// CreateProductionPostgresConfig creates a production-ready PostgreSQL configuration
func CreateProductionPostgresConfig(host, database, username, password string) *persistence.DatabaseConfig {
	return &persistence.DatabaseConfig{
		Type:         persistence.DatabaseTypePostgres,
		Host:         host,
		Port:         5432,
		Database:     database,
		Username:     username,
		Password:     password,
		SSLMode:      "require",
		MaxOpenConns: 50,
		MaxIdleConns: 10,
		MaxLifetime:  "10m",
		ConnectionParams: map[string]string{
			"application_name": "golanggraph",
			"connect_timeout":  "10",
		},
	}
}

// CreateRAGPostgresConfig creates a PostgreSQL configuration optimized for RAG
func CreateRAGPostgresConfig(host, database, username, password string, embeddingDim int) *persistence.DatabaseConfig {
	return &persistence.DatabaseConfig{
		Type:                persistence.DatabaseTypePgVector,
		Host:                host,
		Port:                5432,
		Database:            database,
		Username:            username,
		Password:            password,
		SSLMode:             "require",
		MaxOpenConns:        50,
		MaxIdleConns:        10,
		MaxLifetime:         "10m",
		VectorDimension:     embeddingDim,
		VectorMetric:        "cosine",
		EnableRAG:           true,
		EmbeddingModel:      "text-embedding-ada-002",
		EmbeddingDimension:  embeddingDim,
		SimilarityThreshold: 0.7,
		ConnectionParams: map[string]string{
			"application_name": "golanggraph-rag",
			"connect_timeout":  "10",
		},
	}
}

// CreateRedisConfig creates a Redis configuration for caching
func CreateRedisConfig(host, password string, port int) *persistence.DatabaseConfig {
	return &persistence.DatabaseConfig{
		Type:     persistence.DatabaseTypeRedis,
		Host:     host,
		Port:     port,
		Password: password,
		ConnectionParams: map[string]string{
			"max_retries":     "3",
			"retry_delay":     "100ms",
			"dial_timeout":    "5s",
			"read_timeout":    "3s",
			"write_timeout":   "3s",
			"pool_size":       "10",
			"pool_timeout":    "4s",
			"idle_timeout":    "5m",
			"idle_check_freq": "1m",
		},
	}
}

// Database setup examples and best practices

func ExampleDatabaseSetup() {
	fmt.Println("=== Database Setup Examples ===")

	// Example 1: Simple PostgreSQL setup
	fmt.Println("\n1. Simple PostgreSQL Setup:")
	fmt.Println("   - Install PostgreSQL")
	fmt.Println("   - Create database: CREATE DATABASE golanggraph;")
	fmt.Println("   - Use NewPostgresConfig() helper")

	// Example 2: PostgreSQL with pgvector for RAG
	fmt.Println("\n2. PostgreSQL with pgvector for RAG:")
	fmt.Println("   - Install PostgreSQL with pgvector extension")
	fmt.Println("   - CREATE EXTENSION vector;")
	fmt.Println("   - Use NewPgVectorConfig() helper")

	// Example 3: Redis for fast caching
	fmt.Println("\n3. Redis for fast caching:")
	fmt.Println("   - Install Redis server")
	fmt.Println("   - Configure persistence and memory limits")
	fmt.Println("   - Use NewRedisConfig() helper")

	// Example 4: OpenSearch for advanced RAG (future)
	fmt.Println("\n4. OpenSearch for advanced RAG (future support):")
	fmt.Println("   - Install OpenSearch cluster")
	fmt.Println("   - Configure vector search plugin")
	fmt.Println("   - Use NewOpenSearchConfig() helper")

	// Example 5: Multi-database setup
	fmt.Println("\n5. Multi-database setup:")
	fmt.Println("   - PostgreSQL for main persistence")
	fmt.Println("   - Redis for fast caching")
	fmt.Println("   - pgvector for RAG document storage")
	fmt.Println("   - Use DatabaseConnectionManager")
}
