// Copyright (c) 2024 GoLangGraph Team
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// Package: GoLangGraph - A powerful Go framework for building AI agent workflows

package database

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/piotrlaczkowski/GoLangGraph/pkg/persistence"
)

// EnhancedDatabaseConfig extends the base persistence config with additional features
type EnhancedDatabaseConfig struct {
	// Main database configuration using GoLangGraph's persistence layer
	Primary *persistence.DatabaseConfig `json:"primary"`
	Cache   *persistence.DatabaseConfig `json:"cache"`
	Vector  *persistence.DatabaseConfig `json:"vector"`

	// Memory management configuration
	Memory *MemoryConfig `json:"memory"`

	// Session management configuration
	Sessions *SessionConfig `json:"sessions"`

	// RAG configuration
	RAG *RAGConfig `json:"rag"`

	// Performance configuration
	Performance *PerformanceConfig `json:"performance"`
}

// MemoryConfig defines memory management settings
type MemoryConfig struct {
	EnableEmbeddings    bool    `json:"enable_embeddings"`
	EmbeddingModel      string  `json:"embedding_model"`
	EmbeddingDimension  int     `json:"embedding_dimension"`
	MaxMemorySize       int     `json:"max_memory_size"`
	CleanupInterval     string  `json:"cleanup_interval"`
	SimilarityThreshold float64 `json:"similarity_threshold"`

	// Conversation memory settings
	MaxConversationHistory int `json:"max_conversation_history"`
	ContextWindowSize      int `json:"context_window_size"`
	SummaryThreshold       int `json:"summary_threshold"`

	// User preference learning
	PreferenceDecayFactor float64 `json:"preference_decay_factor"`
	PreferenceUpdateRate  float64 `json:"preference_update_rate"`
}

// SessionConfig defines session management settings
type SessionConfig struct {
	DefaultTTL           string `json:"default_ttl"`
	ExtendedTTL          string `json:"extended_ttl"`
	CleanupInterval      string `json:"cleanup_interval"`
	MaxActiveSessions    int    `json:"max_active_sessions"`
	EnableSessionMetrics bool   `json:"enable_session_metrics"`

	// Thread management
	MaxThreadsPerSession int    `json:"max_threads_per_session"`
	ThreadIdleTimeout    string `json:"thread_idle_timeout"`
	AutoArchiveThreads   bool   `json:"auto_archive_threads"`
}

// RAGConfig defines RAG (Retrieval-Augmented Generation) settings
type RAGConfig struct {
	Enabled            bool    `json:"enabled"`
	ChunkSize          int     `json:"chunk_size"`
	ChunkOverlap       int     `json:"chunk_overlap"`
	MaxRetrievalDocs   int     `json:"max_retrieval_docs"`
	RetrievalThreshold float64 `json:"retrieval_threshold"`

	// Document processing
	SupportedFormats    []string `json:"supported_formats"`
	ProcessingBatchSize int      `json:"processing_batch_size"`

	// Vector search optimization
	IndexUpdateInterval string `json:"index_update_interval"`
	SearchAlgorithm     string `json:"search_algorithm"`
}

// PerformanceConfig defines performance optimization settings
type PerformanceConfig struct {
	ConnectionPoolSize int    `json:"connection_pool_size"`
	QueryTimeout       string `json:"query_timeout"`
	BatchSize          int    `json:"batch_size"`
	EnableQueryLogging bool   `json:"enable_query_logging"`

	// Caching settings
	EnableResultCaching bool   `json:"enable_result_caching"`
	CacheTTL            string `json:"cache_ttl"`
	MaxCacheSize        int    `json:"max_cache_size"`

	// Monitoring
	EnableMetrics   bool   `json:"enable_metrics"`
	MetricsInterval string `json:"metrics_interval"`
}

// DatabaseManager coordinates multiple database connections and services
type DatabaseManager struct {
	Primary persistence.DatabaseConnection
	Cache   persistence.DatabaseConnection
	Vector  persistence.DatabaseConnection

	// GoLangGraph components
	SessionManager *persistence.SessionManager
	Checkpointer   persistence.Checkpointer

	// Enhanced components
	MemoryManager *EnhancedMemoryManager
	RAGManager    *EnhancedRAGManager

	config *EnhancedDatabaseConfig
	logger *logrus.Logger
}

// EnhancedMemoryManager provides advanced memory management capabilities
type EnhancedMemoryManager struct {
	checkpointer persistence.Checkpointer
	config       *MemoryConfig
	logger       *logrus.Logger

	// Memory stores
	conversationMemory map[string][]*MemoryItem
	userPreferences    map[string]*UserPreferences
	designHistory      map[string][]*DesignIteration
}

// MemoryItem represents a stored memory with embeddings
type MemoryItem struct {
	ID          string                 `json:"id"`
	ThreadID    string                 `json:"thread_id"`
	UserID      string                 `json:"user_id"`
	Content     string                 `json:"content"`
	MemoryType  string                 `json:"memory_type"`
	Embedding   []float64              `json:"embedding"`
	Metadata    map[string]interface{} `json:"metadata"`
	Importance  float64                `json:"importance"`
	CreatedAt   time.Time              `json:"created_at"`
	LastAccess  time.Time              `json:"last_access"`
	AccessCount int                    `json:"access_count"`
}

// UserPreferences stores learned user preferences
type UserPreferences struct {
	UserID               string                 `json:"user_id"`
	DesignStyles         map[string]float64     `json:"design_styles"`
	MaterialPrefs        map[string]float64     `json:"material_preferences"`
	BudgetRange          *BudgetRange           `json:"budget_range"`
	SustainabilityWeight float64                `json:"sustainability_weight"`
	Metadata             map[string]interface{} `json:"metadata"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// BudgetRange represents user budget preferences
type BudgetRange struct {
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Preferred float64 `json:"preferred"`
}

// DesignIteration stores design evolution information
type DesignIteration struct {
	ID            string                 `json:"id"`
	ThreadID      string                 `json:"thread_id"`
	UserID        string                 `json:"user_id"`
	DesignConcept string                 `json:"design_concept"`
	Feedback      string                 `json:"feedback"`
	Rating        float64                `json:"rating"`
	Improvements  []string               `json:"improvements"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
}

// EnhancedRAGManager provides advanced RAG capabilities
type EnhancedRAGManager struct {
	checkpointer persistence.Checkpointer
	config       *RAGConfig
	logger       *logrus.Logger

	// Document stores
	documentStore map[string]*persistence.Document
	vectorIndex   map[string][]float64
}

// NewEnhancedDatabaseConfig creates a new enhanced database configuration
func NewEnhancedDatabaseConfig() *EnhancedDatabaseConfig {
	return &EnhancedDatabaseConfig{
		Primary: &persistence.DatabaseConfig{
			Type:                persistence.DatabaseTypePgVector,
			Host:                getEnv("POSTGRES_HOST", "localhost"),
			Port:                getEnvAsInt("POSTGRES_PORT", 5432),
			Database:            getEnv("POSTGRES_DB", "golanggraph_stateful"),
			Username:            getEnv("POSTGRES_USER", "postgres"),
			Password:            getEnv("POSTGRES_PASSWORD", "password"),
			SSLMode:             getEnv("POSTGRES_SSLMODE", "disable"),
			MaxOpenConns:        getEnvAsInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns:        getEnvAsInt("POSTGRES_MAX_IDLE_CONNS", 5),
			MaxLifetime:         getEnv("POSTGRES_MAX_LIFETIME", "5m"),
			VectorDimension:     getEnvAsInt("VECTOR_DIMENSION", 1536),
			VectorMetric:        getEnv("VECTOR_METRIC", "cosine"),
			EnableRAG:           true,
			EmbeddingModel:      getEnv("EMBEDDING_MODEL", "text-embedding-ada-002"),
			EmbeddingDimension:  getEnvAsInt("EMBEDDING_DIMENSION", 1536),
			SimilarityThreshold: getEnvAsFloat("SIMILARITY_THRESHOLD", 0.7),
		},
		Cache: &persistence.DatabaseConfig{
			Type:     persistence.DatabaseTypeRedis,
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			Database: getEnv("REDIS_DB", "golanggraph_cache"),
		},
		Vector: &persistence.DatabaseConfig{
			Type:            persistence.DatabaseTypePgVector,
			Host:            getEnv("VECTOR_HOST", "localhost"),
			Port:            getEnvAsInt("VECTOR_PORT", 5432),
			Database:        getEnv("VECTOR_DB", "golanggraph_vectors"),
			Username:        getEnv("VECTOR_USER", "postgres"),
			Password:        getEnv("VECTOR_PASSWORD", "password"),
			SSLMode:         "disable",
			VectorDimension: getEnvAsInt("VECTOR_DIMENSION", 1536),
			VectorMetric:    "cosine",
			EnableRAG:       true,
		},
		Memory: &MemoryConfig{
			EnableEmbeddings:       true,
			EmbeddingModel:         "text-embedding-ada-002",
			EmbeddingDimension:     1536,
			MaxMemorySize:          10000,
			CleanupInterval:        "1h",
			SimilarityThreshold:    0.7,
			MaxConversationHistory: 50,
			ContextWindowSize:      4000,
			SummaryThreshold:       20,
			PreferenceDecayFactor:  0.95,
			PreferenceUpdateRate:   0.1,
		},
		Sessions: &SessionConfig{
			DefaultTTL:           "24h",
			ExtendedTTL:          "7d",
			CleanupInterval:      "1h",
			MaxActiveSessions:    1000,
			EnableSessionMetrics: true,
			MaxThreadsPerSession: 10,
			ThreadIdleTimeout:    "2h",
			AutoArchiveThreads:   true,
		},
		RAG: &RAGConfig{
			Enabled:             true,
			ChunkSize:           500,
			ChunkOverlap:        50,
			MaxRetrievalDocs:    5,
			RetrievalThreshold:  0.7,
			SupportedFormats:    []string{"txt", "md", "pdf", "docx"},
			ProcessingBatchSize: 10,
			IndexUpdateInterval: "5m",
			SearchAlgorithm:     "ivfflat",
		},
		Performance: &PerformanceConfig{
			ConnectionPoolSize:  25,
			QueryTimeout:        "30s",
			BatchSize:           100,
			EnableQueryLogging:  false,
			EnableResultCaching: true,
			CacheTTL:            "1h",
			MaxCacheSize:        1000,
			EnableMetrics:       true,
			MetricsInterval:     "1m",
		},
	}
}

// NewDatabaseManager creates a new database manager using GoLangGraph's infrastructure
func NewDatabaseManager(config *EnhancedDatabaseConfig) (*DatabaseManager, error) {
	logger := logrus.New()

	// Create primary database connection using GoLangGraph
	primaryConn, err := persistence.NewPostgresConnection(config.Primary)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary database connection: %w", err)
	}

	// Create checkpointer with RAG support
	checkpointer, err := persistence.NewPostgresCheckpointer(config.Primary)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpointer: %w", err)
	}

	// Create session manager using GoLangGraph
	sessionManager := persistence.NewSessionManager(primaryConn)

	// Create enhanced components
	memoryManager, err := NewEnhancedMemoryManager(checkpointer, config.Memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory manager: %w", err)
	}

	ragManager, err := NewEnhancedRAGManager(checkpointer, config.RAG)
	if err != nil {
		return nil, fmt.Errorf("failed to create RAG manager: %w", err)
	}

	return &DatabaseManager{
		Primary:        primaryConn,
		SessionManager: sessionManager,
		Checkpointer:   checkpointer,
		MemoryManager:  memoryManager,
		RAGManager:     ragManager,
		config:         config,
		logger:         logger,
	}, nil
}

// NewEnhancedMemoryManager creates a new enhanced memory manager
func NewEnhancedMemoryManager(checkpointer persistence.Checkpointer, config *MemoryConfig) (*EnhancedMemoryManager, error) {
	return &EnhancedMemoryManager{
		checkpointer:       checkpointer,
		config:             config,
		logger:             logrus.New(),
		conversationMemory: make(map[string][]*MemoryItem),
		userPreferences:    make(map[string]*UserPreferences),
		designHistory:      make(map[string][]*DesignIteration),
	}, nil
}

// NewEnhancedRAGManager creates a new enhanced RAG manager
func NewEnhancedRAGManager(checkpointer persistence.Checkpointer, config *RAGConfig) (*EnhancedRAGManager, error) {
	return &EnhancedRAGManager{
		checkpointer:  checkpointer,
		config:        config,
		logger:        logrus.New(),
		documentStore: make(map[string]*persistence.Document),
		vectorIndex:   make(map[string][]float64),
	}, nil
}

// StoreMemory stores a memory item with embeddings
func (emm *EnhancedMemoryManager) StoreMemory(ctx context.Context, memory *MemoryItem) error {
	// Store in PostgreSQL using GoLangGraph's persistence layer
	doc := &persistence.Document{
		ID:        memory.ID,
		ThreadID:  memory.ThreadID,
		Content:   memory.Content,
		Metadata:  memory.Metadata,
		Embedding: memory.Embedding,
		CreatedAt: memory.CreatedAt,
		UpdatedAt: time.Now(),
	}

	// Use PostgreSQL checkpointer to save document (which handles RAG storage)
	if pgCheckpointer, ok := emm.checkpointer.(*persistence.PostgresCheckpointer); ok {
		return pgCheckpointer.SaveDocument(ctx, doc)
	}

	// Fallback to in-memory storage
	if emm.conversationMemory[memory.ThreadID] == nil {
		emm.conversationMemory[memory.ThreadID] = make([]*MemoryItem, 0)
	}
	emm.conversationMemory[memory.ThreadID] = append(emm.conversationMemory[memory.ThreadID], memory)

	return nil
}

// RetrieveMemories retrieves relevant memories using vector similarity
func (emm *EnhancedMemoryManager) RetrieveMemories(ctx context.Context, threadID string, queryEmbedding []float64, limit int) ([]*MemoryItem, error) {
	// Use PostgreSQL checkpointer for vector search
	if pgCheckpointer, ok := emm.checkpointer.(*persistence.PostgresCheckpointer); ok {
		docs, err := pgCheckpointer.SearchDocuments(ctx, threadID, queryEmbedding, limit)
		if err != nil {
			return nil, err
		}

		// Convert documents to memory items
		var memories []*MemoryItem
		for _, doc := range docs {
			memory := &MemoryItem{
				ID:        doc.ID,
				ThreadID:  doc.ThreadID,
				Content:   doc.Content,
				Embedding: doc.Embedding,
				Metadata:  doc.Metadata,
				CreatedAt: doc.CreatedAt,
			}
			memories = append(memories, memory)
		}

		return memories, nil
	}

	// Fallback to in-memory search
	memories := emm.conversationMemory[threadID]
	if len(memories) == 0 {
		return []*MemoryItem{}, nil
	}

	// Simple similarity search (placeholder implementation)
	var relevant []*MemoryItem
	for _, memory := range memories {
		if len(memory.Embedding) == len(queryEmbedding) {
			similarity := calculateCosineSimilarity(queryEmbedding, memory.Embedding)
			if similarity >= emm.config.SimilarityThreshold {
				relevant = append(relevant, memory)
			}
		}
	}

	// Return up to limit items
	if len(relevant) > limit {
		relevant = relevant[:limit]
	}

	return relevant, nil
}

// UpdateUserPreferences updates user preferences with learning
func (emm *EnhancedMemoryManager) UpdateUserPreferences(userID string, feedback map[string]interface{}) error {
	prefs, exists := emm.userPreferences[userID]
	if !exists {
		prefs = &UserPreferences{
			UserID:        userID,
			DesignStyles:  make(map[string]float64),
			MaterialPrefs: make(map[string]float64),
			BudgetRange:   &BudgetRange{},
			Metadata:      make(map[string]interface{}),
			UpdatedAt:     time.Now(),
		}
		emm.userPreferences[userID] = prefs
	}

	// Apply decay factor to existing preferences
	for style := range prefs.DesignStyles {
		prefs.DesignStyles[style] *= emm.config.PreferenceDecayFactor
	}

	for material := range prefs.MaterialPrefs {
		prefs.MaterialPrefs[material] *= emm.config.PreferenceDecayFactor
	}

	// Update with new feedback
	if styles, ok := feedback["design_styles"].(map[string]float64); ok {
		for style, weight := range styles {
			currentWeight := prefs.DesignStyles[style]
			prefs.DesignStyles[style] = currentWeight + (weight * emm.config.PreferenceUpdateRate)
		}
	}

	if materials, ok := feedback["material_preferences"].(map[string]float64); ok {
		for material, weight := range materials {
			currentWeight := prefs.MaterialPrefs[material]
			prefs.MaterialPrefs[material] = currentWeight + (weight * emm.config.PreferenceUpdateRate)
		}
	}

	prefs.UpdatedAt = time.Now()
	return nil
}

// Close closes all database connections
func (dm *DatabaseManager) Close() error {
	var errors []string

	if dm.Primary != nil {
		if err := dm.Primary.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("primary: %v", err))
		}
	}

	if dm.Cache != nil {
		if err := dm.Cache.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("cache: %v", err))
		}
	}

	if dm.Vector != nil {
		if err := dm.Vector.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("vector: %v", err))
		}
	}

	if dm.Checkpointer != nil {
		if err := dm.Checkpointer.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("checkpointer: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errors)
	}

	return nil
}

// Utility functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// calculateCosineSimilarity calculates cosine similarity between two vectors
func calculateCosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (normA * normB)
}
