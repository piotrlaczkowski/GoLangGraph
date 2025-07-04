package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/sirupsen/logrus"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	DatabaseTypePostgres   DatabaseType = "postgres"
	DatabaseTypePostgresQL DatabaseType = "postgresql"
	DatabaseTypePgVector   DatabaseType = "pgvector"
	DatabaseTypeRedis      DatabaseType = "redis"
	DatabaseTypeOpenSearch DatabaseType = "opensearch"
	DatabaseTypeElastic    DatabaseType = "elasticsearch"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
)

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type         DatabaseType `json:"type"` // "postgres", "pgvector", "redis", "opensearch", etc.
	Host         string       `json:"host"`
	Port         int          `json:"port"`
	Database     string       `json:"database"`
	Username     string       `json:"username"`
	Password     string       `json:"password"`
	SSLMode      string       `json:"ssl_mode"`
	MaxOpenConns int          `json:"max_open_conns"`
	MaxIdleConns int          `json:"max_idle_conns"`
	MaxLifetime  string       `json:"max_lifetime"`

	// Vector-specific configuration
	VectorDimension int    `json:"vector_dimension"`
	VectorMetric    string `json:"vector_metric"` // "cosine", "euclidean", "dot_product"

	// OpenSearch/Elasticsearch specific
	Index   string `json:"index"`
	APIKey  string `json:"api_key"`
	CloudID string `json:"cloud_id"`
	CACert  string `json:"ca_cert"`

	// Additional connection parameters
	ConnectionParams map[string]string `json:"connection_params"`

	// RAG-specific settings
	EnableRAG           bool    `json:"enable_rag"`
	EmbeddingModel      string  `json:"embedding_model"`
	EmbeddingDimension  int     `json:"embedding_dimension"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
}

// DatabaseConnection represents a database connection interface
type DatabaseConnection interface {
	Connect() error
	Close() error
	Ping() error
	GetType() DatabaseType
	GetConfig() *DatabaseConfig
	ExecuteQuery(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) interface{}
	QueryRows(ctx context.Context, query string, args ...interface{}) (interface{}, error)
}

// PostgresConnection implements PostgreSQL connection
type PostgresConnection struct {
	db     *sql.DB
	config *DatabaseConfig
	logger *logrus.Logger
}

// NewPostgresConnection creates a new PostgreSQL connection
func NewPostgresConnection(config *DatabaseConfig) (*PostgresConnection, error) {
	conn := &PostgresConnection{
		config: config,
		logger: logrus.New(),
	}

	if err := conn.Connect(); err != nil {
		return nil, err
	}

	return conn, nil
}

// Connect establishes the PostgreSQL connection
func (p *PostgresConnection) Connect() error {
	var dsn string

	// Support different PostgreSQL connection formats
	switch p.config.Type {
	case DatabaseTypePostgres, DatabaseTypePostgresQL:
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			p.config.Host, p.config.Port, p.config.Username, p.config.Password, p.config.Database, p.config.SSLMode)
	case DatabaseTypePgVector:
		// PostgreSQL with pgvector extension
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			p.config.Host, p.config.Port, p.config.Username, p.config.Password, p.config.Database, p.config.SSLMode)
	default:
		return fmt.Errorf("unsupported database type: %s", p.config.Type)
	}

	// Add additional connection parameters
	if p.config.ConnectionParams != nil {
		var params []string
		for k, v := range p.config.ConnectionParams {
			params = append(params, fmt.Sprintf("%s=%s", k, v))
		}
		if len(params) > 0 {
			dsn += " " + strings.Join(params, " ")
		}
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	if p.config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(p.config.MaxOpenConns)
	} else {
		db.SetMaxOpenConns(25) // Default
	}

	if p.config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(p.config.MaxIdleConns)
	} else {
		db.SetMaxIdleConns(5) // Default
	}

	if p.config.MaxLifetime != "" {
		if duration, err := time.ParseDuration(p.config.MaxLifetime); err == nil {
			db.SetConnMaxLifetime(duration)
		}
	} else {
		db.SetConnMaxLifetime(5 * time.Minute) // Default
	}

	p.db = db
	return p.Ping()
}

// Ping tests the database connection
func (p *PostgresConnection) Ping() error {
	if p.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.db.PingContext(ctx)
}

// Close closes the database connection
func (p *PostgresConnection) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// GetType returns the database type
func (p *PostgresConnection) GetType() DatabaseType {
	return p.config.Type
}

// GetConfig returns the database configuration
func (p *PostgresConnection) GetConfig() *DatabaseConfig {
	return p.config
}

// ExecuteQuery executes a query without returning results
func (p *PostgresConnection) ExecuteQuery(ctx context.Context, query string, args ...interface{}) error {
	_, err := p.db.ExecContext(ctx, query, args...)
	return err
}

// QueryRow executes a query that returns a single row
func (p *PostgresConnection) QueryRow(ctx context.Context, query string, args ...interface{}) interface{} {
	return p.db.QueryRowContext(ctx, query, args...)
}

// QueryRows executes a query that returns multiple rows
func (p *PostgresConnection) QueryRows(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// PostgresCheckpointer implements database-based checkpointing with PostgreSQL
type PostgresCheckpointer struct {
	conn   *PostgresConnection
	config *DatabaseConfig
	logger *logrus.Logger
}

// NewPostgresCheckpointer creates a new PostgreSQL checkpointer
func NewPostgresCheckpointer(config *DatabaseConfig) (*PostgresCheckpointer, error) {
	conn, err := NewPostgresConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection: %w", err)
	}

	checkpointer := &PostgresCheckpointer{
		conn:   conn,
		config: config,
		logger: logrus.New(),
	}

	// Initialize schema
	if err := checkpointer.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return checkpointer, nil
}

// initSchema initializes the database schema with enhanced support for RAG and vector operations
func (p *PostgresCheckpointer) initSchema() error {
	// Enable pgvector extension if using pgvector
	if p.config.Type == DatabaseTypePgVector {
		if err := p.conn.ExecuteQuery(context.Background(), "CREATE EXTENSION IF NOT EXISTS vector;"); err != nil {
			p.logger.Warnf("Failed to create vector extension (may not be available): %v", err)
		}
	}

	// Create main tables
	schema := `
	-- Threads table for conversation management
	CREATE TABLE IF NOT EXISTS threads (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255),
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Checkpoints table for state persistence
	CREATE TABLE IF NOT EXISTS checkpoints (
		id VARCHAR(255) PRIMARY KEY,
		thread_id VARCHAR(255) NOT NULL,
		state_data JSONB NOT NULL,
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		node_id VARCHAR(255),
		step_id INTEGER,
		FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
	);

	-- Sessions table for user session management
	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(255) PRIMARY KEY,
		thread_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255),
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE,
		FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
	);

	-- Create indexes for better performance
	CREATE INDEX IF NOT EXISTS idx_checkpoints_thread_id ON checkpoints(thread_id);
	CREATE INDEX IF NOT EXISTS idx_checkpoints_created_at ON checkpoints(created_at);
	CREATE INDEX IF NOT EXISTS idx_checkpoints_node_id ON checkpoints(node_id);
	CREATE INDEX IF NOT EXISTS idx_checkpoints_step_id ON checkpoints(step_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_thread_id ON sessions(thread_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`

	if err := p.conn.ExecuteQuery(context.Background(), schema); err != nil {
		return fmt.Errorf("failed to create basic schema: %w", err)
	}

	// Create RAG-specific tables if enabled
	if p.config.EnableRAG {
		if err := p.initRAGSchema(); err != nil {
			return fmt.Errorf("failed to initialize RAG schema: %w", err)
		}
	}

	return nil
}

// initRAGSchema initializes RAG-specific database schema
func (p *PostgresCheckpointer) initRAGSchema() error {
	var vectorSchema string

	if p.config.Type == DatabaseTypePgVector {
		// Use pgvector for vector storage
		vectorDim := p.config.VectorDimension
		if vectorDim == 0 {
			vectorDim = 1536 // Default OpenAI embedding dimension
		}

		vectorSchema = fmt.Sprintf(`
		-- Documents table for RAG document storage
		CREATE TABLE IF NOT EXISTS documents (
			id VARCHAR(255) PRIMARY KEY,
			thread_id VARCHAR(255),
			content TEXT NOT NULL,
			metadata JSONB,
			embedding vector(%d),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
		);

		-- Memory table for conversational memory with embeddings
		CREATE TABLE IF NOT EXISTS memory (
			id VARCHAR(255) PRIMARY KEY,
			thread_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255),
			content TEXT NOT NULL,
			memory_type VARCHAR(50) DEFAULT 'conversation',
			embedding vector(%d),
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
		);

		-- Vector indexes for similarity search
		CREATE INDEX IF NOT EXISTS idx_documents_embedding ON documents USING ivfflat (embedding vector_cosine_ops);
		CREATE INDEX IF NOT EXISTS idx_memory_embedding ON memory USING ivfflat (embedding vector_cosine_ops);
		CREATE INDEX IF NOT EXISTS idx_documents_thread_id ON documents(thread_id);
		CREATE INDEX IF NOT EXISTS idx_memory_thread_id ON memory(thread_id);
		CREATE INDEX IF NOT EXISTS idx_memory_user_id ON memory(user_id);
		CREATE INDEX IF NOT EXISTS idx_memory_type ON memory(memory_type);
		`, vectorDim, vectorDim)
	} else {
		// Fallback without vector support
		vectorSchema = `
		-- Documents table for RAG document storage (without vectors)
		CREATE TABLE IF NOT EXISTS documents (
			id VARCHAR(255) PRIMARY KEY,
			thread_id VARCHAR(255),
			content TEXT NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
		);

		-- Memory table for conversational memory (without vectors)
		CREATE TABLE IF NOT EXISTS memory (
			id VARCHAR(255) PRIMARY KEY,
			thread_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255),
			content TEXT NOT NULL,
			memory_type VARCHAR(50) DEFAULT 'conversation',
			metadata JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
		);

		-- Indexes for text search
		CREATE INDEX IF NOT EXISTS idx_documents_thread_id ON documents(thread_id);
		CREATE INDEX IF NOT EXISTS idx_documents_content ON documents USING gin(to_tsvector('english', content));
		CREATE INDEX IF NOT EXISTS idx_memory_thread_id ON memory(thread_id);
		CREATE INDEX IF NOT EXISTS idx_memory_user_id ON memory(user_id);
		CREATE INDEX IF NOT EXISTS idx_memory_type ON memory(memory_type);
		CREATE INDEX IF NOT EXISTS idx_memory_content ON memory USING gin(to_tsvector('english', content));
		`
	}

	return p.conn.ExecuteQuery(context.Background(), vectorSchema)
}

// Save saves a checkpoint to PostgreSQL
func (p *PostgresCheckpointer) Save(ctx context.Context, checkpoint *Checkpoint) error {
	stateData, err := json.Marshal(checkpoint.State)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	metadataData, err := json.Marshal(checkpoint.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO checkpoints (id, thread_id, state_data, metadata, created_at, node_id, step_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			state_data = EXCLUDED.state_data,
			metadata = EXCLUDED.metadata,
			created_at = EXCLUDED.created_at,
			node_id = EXCLUDED.node_id,
			step_id = EXCLUDED.step_id
	`

	err = p.conn.ExecuteQuery(ctx, query,
		checkpoint.ID,
		checkpoint.ThreadID,
		stateData,
		metadataData,
		checkpoint.CreatedAt,
		checkpoint.NodeID,
		checkpoint.StepID,
	)

	if err != nil {
		return fmt.Errorf("failed to save checkpoint: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"checkpoint_id": checkpoint.ID,
		"thread_id":     checkpoint.ThreadID,
	}).Info("Checkpoint saved to database")

	return nil
}

// Load loads a checkpoint from PostgreSQL
func (p *PostgresCheckpointer) Load(ctx context.Context, threadID, checkpointID string) (*Checkpoint, error) {
	query := `
		SELECT id, thread_id, state_data, metadata, created_at, node_id, step_id
		FROM checkpoints
		WHERE thread_id = $1 AND id = $2
	`

	row := p.conn.QueryRow(ctx, query, threadID, checkpointID).(*sql.Row)

	var checkpoint Checkpoint
	var stateData, metadataData []byte

	err := row.Scan(
		&checkpoint.ID,
		&checkpoint.ThreadID,
		&stateData,
		&metadataData,
		&checkpoint.CreatedAt,
		&checkpoint.NodeID,
		&checkpoint.StepID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("checkpoint %s not found in thread %s", checkpointID, threadID)
		}
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Unmarshal state
	var state core.BaseState
	if err := json.Unmarshal(stateData, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	checkpoint.State = &state

	// Unmarshal metadata
	if err := json.Unmarshal(metadataData, &checkpoint.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &checkpoint, nil
}

// List lists checkpoints for a thread
func (p *PostgresCheckpointer) List(ctx context.Context, threadID string) ([]*CheckpointMetadata, error) {
	query := `
		SELECT id, thread_id, metadata, created_at, node_id, step_id
		FROM checkpoints
		WHERE thread_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.conn.QueryRows(ctx, query, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}
	defer rows.(*sql.Rows).Close()

	var checkpoints []*CheckpointMetadata
	for rows.(*sql.Rows).Next() {
		var checkpoint CheckpointMetadata
		var metadataData []byte

		err := rows.(*sql.Rows).Scan(
			&checkpoint.ID,
			&checkpoint.ThreadID,
			&metadataData,
			&checkpoint.CreatedAt,
			&checkpoint.NodeID,
			&checkpoint.StepID,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan checkpoint: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataData, &checkpoint.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	return checkpoints, nil
}

// Delete deletes a checkpoint
func (p *PostgresCheckpointer) Delete(ctx context.Context, threadID, checkpointID string) error {
	query := `DELETE FROM checkpoints WHERE thread_id = $1 AND id = $2`

	err := p.conn.ExecuteQuery(ctx, query, threadID, checkpointID)
	if err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}

	return nil
}

// Close closes the PostgreSQL checkpointer
func (p *PostgresCheckpointer) Close() error {
	return p.conn.Close()
}

// RAG-specific methods

// SaveDocument saves a document for RAG
func (p *PostgresCheckpointer) SaveDocument(ctx context.Context, doc *Document) error {
	if !p.config.EnableRAG {
		return fmt.Errorf("RAG is not enabled")
	}

	var query string
	var args []interface{}

	if p.config.Type == DatabaseTypePgVector && doc.Embedding != nil {
		query = `
			INSERT INTO documents (id, thread_id, content, metadata, embedding, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				content = EXCLUDED.content,
				metadata = EXCLUDED.metadata,
				embedding = EXCLUDED.embedding,
				updated_at = EXCLUDED.updated_at
		`
		args = []interface{}{doc.ID, doc.ThreadID, doc.Content, doc.Metadata, doc.Embedding, doc.CreatedAt, doc.UpdatedAt}
	} else {
		query = `
			INSERT INTO documents (id, thread_id, content, metadata, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
				content = EXCLUDED.content,
				metadata = EXCLUDED.metadata,
				updated_at = EXCLUDED.updated_at
		`
		args = []interface{}{doc.ID, doc.ThreadID, doc.Content, doc.Metadata, doc.CreatedAt, doc.UpdatedAt}
	}

	return p.conn.ExecuteQuery(ctx, query, args...)
}

// SearchDocuments performs similarity search on documents
func (p *PostgresCheckpointer) SearchDocuments(ctx context.Context, threadID string, queryEmbedding []float64, limit int) ([]*Document, error) {
	if !p.config.EnableRAG {
		return nil, fmt.Errorf("RAG is not enabled")
	}

	var query string
	var args []interface{}

	if p.config.Type == DatabaseTypePgVector && queryEmbedding != nil {
		query = `
			SELECT id, thread_id, content, metadata, embedding, created_at, updated_at
			FROM documents
			WHERE thread_id = $1
			ORDER BY embedding <-> $2
			LIMIT $3
		`
		args = []interface{}{threadID, queryEmbedding, limit}
	} else {
		// Fallback to text search
		query = `
			SELECT id, thread_id, content, metadata, created_at, updated_at
			FROM documents
			WHERE thread_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{threadID, limit}
	}

	rows, err := p.conn.QueryRows(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}
	defer rows.(*sql.Rows).Close()

	var documents []*Document
	for rows.(*sql.Rows).Next() {
		var doc Document
		var metadataData []byte
		var embedding interface{}

		if p.config.Type == DatabaseTypePgVector {
			err := rows.(*sql.Rows).Scan(&doc.ID, &doc.ThreadID, &doc.Content, &metadataData, &embedding, &doc.CreatedAt, &doc.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to scan document: %w", err)
			}
			// Handle embedding conversion if needed
		} else {
			err := rows.(*sql.Rows).Scan(&doc.ID, &doc.ThreadID, &doc.Content, &metadataData, &doc.CreatedAt, &doc.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to scan document: %w", err)
			}
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataData, &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		documents = append(documents, &doc)
	}

	return documents, nil
}

// RedisCheckpointer implements Redis-based checkpointing
type RedisCheckpointer struct {
	client *redis.Client
	config *DatabaseConfig
	logger *logrus.Logger
	ttl    time.Duration
}

// NewRedisCheckpointer creates a new Redis checkpointer
func NewRedisCheckpointer(config *DatabaseConfig) (*RedisCheckpointer, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCheckpointer{
		client: client,
		config: config,
		logger: logrus.New(),
		ttl:    24 * time.Hour, // Default TTL
	}, nil
}

// Save saves a checkpoint to Redis
func (r *RedisCheckpointer) Save(ctx context.Context, checkpoint *Checkpoint) error {
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	key := fmt.Sprintf("checkpoint:%s:%s", checkpoint.ThreadID, checkpoint.ID)

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save checkpoint to Redis: %w", err)
	}

	// Add to thread index
	threadKey := fmt.Sprintf("thread:%s:checkpoints", checkpoint.ThreadID)
	if err := r.client.SAdd(ctx, threadKey, checkpoint.ID).Err(); err != nil {
		return fmt.Errorf("failed to add checkpoint to thread index: %w", err)
	}

	return nil
}

// Load loads a checkpoint from Redis
func (r *RedisCheckpointer) Load(ctx context.Context, threadID, checkpointID string) (*Checkpoint, error) {
	key := fmt.Sprintf("checkpoint:%s:%s", threadID, checkpointID)

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("checkpoint %s not found in thread %s", checkpointID, threadID)
		}
		return nil, fmt.Errorf("failed to load checkpoint from Redis: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal([]byte(data), &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// List lists checkpoints for a thread
func (r *RedisCheckpointer) List(ctx context.Context, threadID string) ([]*CheckpointMetadata, error) {
	threadKey := fmt.Sprintf("thread:%s:checkpoints", threadID)

	checkpointIDs, err := r.client.SMembers(ctx, threadKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get checkpoint IDs: %w", err)
	}

	var metadata []*CheckpointMetadata
	for _, checkpointID := range checkpointIDs {
		checkpoint, err := r.Load(ctx, threadID, checkpointID)
		if err != nil {
			r.logger.Warnf("Failed to load checkpoint %s: %v", checkpointID, err)
			continue
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
func (r *RedisCheckpointer) Delete(ctx context.Context, threadID, checkpointID string) error {
	key := fmt.Sprintf("checkpoint:%s:%s", threadID, checkpointID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete checkpoint from Redis: %w", err)
	}

	// Remove from thread index
	threadKey := fmt.Sprintf("thread:%s:checkpoints", threadID)
	if err := r.client.SRem(ctx, threadKey, checkpointID).Err(); err != nil {
		return fmt.Errorf("failed to remove checkpoint from thread index: %w", err)
	}

	return nil
}

// Close closes the Redis checkpointer
func (r *RedisCheckpointer) Close() error {
	return r.client.Close()
}

// Document represents a document for RAG
type Document struct {
	ID        string                 `json:"id"`
	ThreadID  string                 `json:"thread_id"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Embedding []float64              `json:"embedding,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SessionManager manages user sessions and threads
type SessionManager struct {
	conn   DatabaseConnection
	logger *logrus.Logger
}

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	ThreadID  string                 `json:"thread_id"`
	UserID    string                 `json:"user_id"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
}

// Thread represents a conversation thread
type Thread struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(conn DatabaseConnection) *SessionManager {
	return &SessionManager{
		conn:   conn,
		logger: logrus.New(),
	}
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, session *Session) error {
	metadataData, err := json.Marshal(session.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO sessions (id, thread_id, user_id, metadata, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	return sm.conn.ExecuteQuery(ctx, query,
		session.ID,
		session.ThreadID,
		session.UserID,
		metadataData,
		session.CreatedAt,
		session.ExpiresAt,
	)
}

// GetSession retrieves a session
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT id, thread_id, user_id, metadata, created_at, expires_at
		FROM sessions
		WHERE id = $1
	`

	row := sm.conn.QueryRow(ctx, query, sessionID).(*sql.Row)

	var session Session
	var metadataData []byte

	err := row.Scan(
		&session.ID,
		&session.ThreadID,
		&session.UserID,
		&metadataData,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session %s not found", sessionID)
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataData, &session.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &session, nil
}

// CreateThread creates a new thread
func (sm *SessionManager) CreateThread(ctx context.Context, thread *Thread) error {
	metadataData, err := json.Marshal(thread.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO threads (id, name, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	return sm.conn.ExecuteQuery(ctx, query,
		thread.ID,
		thread.Name,
		metadataData,
		thread.CreatedAt,
		thread.UpdatedAt,
	)
}

// GetThread retrieves a thread
func (sm *SessionManager) GetThread(ctx context.Context, threadID string) (*Thread, error) {
	query := `
		SELECT id, name, metadata, created_at, updated_at
		FROM threads
		WHERE id = $1
	`

	row := sm.conn.QueryRow(ctx, query, threadID).(*sql.Row)

	var thread Thread
	var metadataData []byte

	err := row.Scan(
		&thread.ID,
		&thread.Name,
		&metadataData,
		&thread.CreatedAt,
		&thread.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("thread %s not found", threadID)
		}
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	// Unmarshal metadata
	if err := json.Unmarshal(metadataData, &thread.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &thread, nil
}

// DatabaseConnectionManager manages multiple database connections
type DatabaseConnectionManager struct {
	connections map[string]DatabaseConnection
	logger      *logrus.Logger
}

// NewDatabaseConnectionManager creates a new connection manager
func NewDatabaseConnectionManager() *DatabaseConnectionManager {
	return &DatabaseConnectionManager{
		connections: make(map[string]DatabaseConnection),
		logger:      logrus.New(),
	}
}

// AddConnection adds a database connection
func (dcm *DatabaseConnectionManager) AddConnection(name string, config *DatabaseConfig) error {
	var conn DatabaseConnection
	var err error

	switch config.Type {
	case DatabaseTypePostgres, DatabaseTypePostgresQL, DatabaseTypePgVector:
		conn, err = NewPostgresConnection(config)
	case DatabaseTypeRedis:
		// Redis connection would be implemented here
		return fmt.Errorf("Redis connection not implemented in this version")
	case DatabaseTypeOpenSearch, DatabaseTypeElastic:
		// OpenSearch/Elasticsearch connections would be implemented here
		return fmt.Errorf("OpenSearch/Elasticsearch connection not implemented in this version")
	case DatabaseTypeMongoDB:
		// MongoDB connection would be implemented here
		return fmt.Errorf("MongoDB connection not implemented in this version")
	case DatabaseTypeMySQL:
		// MySQL connection would be implemented here
		return fmt.Errorf("MySQL connection not implemented in this version")
	case DatabaseTypeSQLite:
		// SQLite connection would be implemented here
		return fmt.Errorf("SQLite connection not implemented in this version")
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create connection for %s: %w", name, err)
	}

	dcm.connections[name] = conn
	dcm.logger.Infof("Added database connection: %s (%s)", name, config.Type)
	return nil
}

// GetConnection retrieves a database connection
func (dcm *DatabaseConnectionManager) GetConnection(name string) (DatabaseConnection, error) {
	conn, exists := dcm.connections[name]
	if !exists {
		return nil, fmt.Errorf("connection %s not found", name)
	}
	return conn, nil
}

// CloseAll closes all database connections
func (dcm *DatabaseConnectionManager) CloseAll() error {
	var errors []string
	for name, conn := range dcm.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close %s: %v", name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing connections: %s", strings.Join(errors, "; "))
	}

	return nil
}

// CreateCheckpointer creates a checkpointer for the specified database
func CreateCheckpointer(config *DatabaseConfig) (Checkpointer, error) {
	switch config.Type {
	case DatabaseTypePostgres, DatabaseTypePostgresQL, DatabaseTypePgVector:
		return NewPostgresCheckpointer(config)
	case DatabaseTypeRedis:
		return NewRedisCheckpointer(config)
	case DatabaseTypeOpenSearch, DatabaseTypeElastic:
		// OpenSearch/Elasticsearch checkpointer would be implemented here
		return nil, fmt.Errorf("OpenSearch/Elasticsearch checkpointer not implemented in this version")
	case DatabaseTypeMongoDB:
		// MongoDB checkpointer would be implemented here
		return nil, fmt.Errorf("MongoDB checkpointer not implemented in this version")
	case DatabaseTypeMySQL:
		// MySQL checkpointer would be implemented here
		return nil, fmt.Errorf("MySQL checkpointer not implemented in this version")
	case DatabaseTypeSQLite:
		// SQLite checkpointer would be implemented here
		return nil, fmt.Errorf("SQLite checkpointer not implemented in this version")
	default:
		return nil, fmt.Errorf("unsupported database type for checkpointer: %s", config.Type)
	}
}

// Helper function to create a default PostgreSQL configuration
func NewPostgresConfig(host string, port int, database, username, password string) *DatabaseConfig {
	return &DatabaseConfig{
		Type:         DatabaseTypePostgres,
		Host:         host,
		Port:         port,
		Database:     database,
		Username:     username,
		Password:     password,
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  "5m",
	}
}

// Helper function to create a PostgreSQL with pgvector configuration
func NewPgVectorConfig(host string, port int, database, username, password string, vectorDim int) *DatabaseConfig {
	return &DatabaseConfig{
		Type:            DatabaseTypePgVector,
		Host:            host,
		Port:            port,
		Database:        database,
		Username:        username,
		Password:        password,
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		MaxLifetime:     "5m",
		VectorDimension: vectorDim,
		VectorMetric:    "cosine",
		EnableRAG:       true,
	}
}

// Helper function to create a Redis configuration
func NewRedisConfig(host string, port int, password string) *DatabaseConfig {
	return &DatabaseConfig{
		Type:     DatabaseTypeRedis,
		Host:     host,
		Port:     port,
		Password: password,
	}
}
