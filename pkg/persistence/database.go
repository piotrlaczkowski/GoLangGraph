package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/piotrlaczkowski/GoLangGraph/pkg/core"
	"github.com/sirupsen/logrus"
)

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type         string `json:"type"` // "postgres", "redis"
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     string `json:"database"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	SSLMode      string `json:"ssl_mode"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxLifetime  string `json:"max_lifetime"`
}

// PostgresCheckpointer implements database-based checkpointing with PostgreSQL
type PostgresCheckpointer struct {
	db     *sql.DB
	config *DatabaseConfig
	logger *logrus.Logger
}

// NewPostgresCheckpointer creates a new PostgreSQL checkpointer
func NewPostgresCheckpointer(config *DatabaseConfig) (*PostgresCheckpointer, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxLifetime != "" {
		if duration, err := time.ParseDuration(config.MaxLifetime); err == nil {
			db.SetConnMaxLifetime(duration)
		}
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	checkpointer := &PostgresCheckpointer{
		db:     db,
		config: config,
		logger: logrus.New(),
	}

	// Initialize schema
	if err := checkpointer.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return checkpointer, nil
}

// initSchema initializes the database schema
func (p *PostgresCheckpointer) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS checkpoints (
		id VARCHAR(255) PRIMARY KEY,
		thread_id VARCHAR(255) NOT NULL,
		state_data JSONB NOT NULL,
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		node_id VARCHAR(255),
		step_id INTEGER,
		INDEX (thread_id),
		INDEX (created_at)
	);

	CREATE TABLE IF NOT EXISTS threads (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255),
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id VARCHAR(255) PRIMARY KEY,
		thread_id VARCHAR(255) NOT NULL,
		user_id VARCHAR(255),
		metadata JSONB,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		expires_at TIMESTAMP WITH TIME ZONE,
		FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
	);
	`

	_, err := p.db.Exec(schema)
	return err
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

	_, err = p.db.ExecContext(ctx, query,
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

	row := p.db.QueryRowContext(ctx, query, threadID, checkpointID)

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

	rows, err := p.db.QueryContext(ctx, query, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}
	defer rows.Close()

	var checkpoints []*CheckpointMetadata
	for rows.Next() {
		var checkpoint CheckpointMetadata
		var metadataData []byte

		err := rows.Scan(
			&checkpoint.ID,
			&checkpoint.ThreadID,
			&metadataData,
			&checkpoint.CreatedAt,
			&checkpoint.NodeID,
			&checkpoint.StepID,
		)

		if err != nil {
			continue // Skip invalid rows
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataData, &checkpoint.Metadata); err != nil {
			checkpoint.Metadata = make(map[string]interface{})
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	return checkpoints, nil
}

// Delete deletes a checkpoint
func (p *PostgresCheckpointer) Delete(ctx context.Context, threadID, checkpointID string) error {
	query := `DELETE FROM checkpoints WHERE thread_id = $1 AND id = $2`

	result, err := p.db.ExecContext(ctx, query, threadID, checkpointID)
	if err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("checkpoint %s not found in thread %s", checkpointID, threadID)
	}

	return nil
}

// Close closes the database connection
func (p *PostgresCheckpointer) Close() error {
	return p.db.Close()
}

// RedisCheckpointer implements Redis-based checkpointing for fast access
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
		DB:       0, // Use default DB
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCheckpointer{
		client: client,
		config: config,
		logger: logrus.New(),
		ttl:    24 * time.Hour, // Default TTL of 24 hours
	}, nil
}

// Save saves a checkpoint to Redis
func (r *RedisCheckpointer) Save(ctx context.Context, checkpoint *Checkpoint) error {
	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	key := fmt.Sprintf("checkpoint:%s:%s", checkpoint.ThreadID, checkpoint.ID)
	err = r.client.Set(ctx, key, data, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save checkpoint to Redis: %w", err)
	}

	// Add to thread index
	threadKey := fmt.Sprintf("thread:%s:checkpoints", checkpoint.ThreadID)
	err = r.client.ZAdd(ctx, threadKey, &redis.Z{
		Score:  float64(checkpoint.CreatedAt.Unix()),
		Member: checkpoint.ID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to add to thread index: %w", err)
	}

	// Set TTL on thread index
	r.client.Expire(ctx, threadKey, r.ttl)

	r.logger.WithFields(logrus.Fields{
		"checkpoint_id": checkpoint.ID,
		"thread_id":     checkpoint.ThreadID,
	}).Info("Checkpoint saved to Redis")

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

	// Get checkpoint IDs ordered by creation time (newest first)
	checkpointIDs, err := r.client.ZRevRange(ctx, threadKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints from Redis: %w", err)
	}

	var checkpoints []*CheckpointMetadata
	for _, checkpointID := range checkpointIDs {
		checkpoint, err := r.Load(ctx, threadID, checkpointID)
		if err != nil {
			continue // Skip invalid checkpoints
		}

		metadata := &CheckpointMetadata{
			ID:        checkpoint.ID,
			ThreadID:  checkpoint.ThreadID,
			Metadata:  checkpoint.Metadata,
			CreatedAt: checkpoint.CreatedAt,
			NodeID:    checkpoint.NodeID,
			StepID:    checkpoint.StepID,
		}

		checkpoints = append(checkpoints, metadata)
	}

	return checkpoints, nil
}

// Delete deletes a checkpoint from Redis
func (r *RedisCheckpointer) Delete(ctx context.Context, threadID, checkpointID string) error {
	key := fmt.Sprintf("checkpoint:%s:%s", threadID, checkpointID)

	// Delete the checkpoint
	deleted, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete checkpoint from Redis: %w", err)
	}

	if deleted == 0 {
		return fmt.Errorf("checkpoint %s not found in thread %s", checkpointID, threadID)
	}

	// Remove from thread index
	threadKey := fmt.Sprintf("thread:%s:checkpoints", threadID)
	r.client.ZRem(ctx, threadKey, checkpointID)

	return nil
}

// Close closes the Redis connection
func (r *RedisCheckpointer) Close() error {
	return r.client.Close()
}

// SessionManager manages user sessions and threads
type SessionManager struct {
	db     *sql.DB
	redis  *redis.Client
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
func NewSessionManager(db *sql.DB, redis *redis.Client) *SessionManager {
	return &SessionManager{
		db:     db,
		redis:  redis,
		logger: logrus.New(),
	}
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, session *Session) error {
	if sm.db != nil {
		query := `
			INSERT INTO sessions (id, thread_id, user_id, metadata, created_at, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`

		metadataData, err := json.Marshal(session.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		_, err = sm.db.ExecContext(ctx, query,
			session.ID,
			session.ThreadID,
			session.UserID,
			metadataData,
			session.CreatedAt,
			session.ExpiresAt,
		)

		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Also store in Redis for fast access
	if sm.redis != nil {
		data, err := json.Marshal(session)
		if err != nil {
			return fmt.Errorf("failed to marshal session: %w", err)
		}

		key := fmt.Sprintf("session:%s", session.ID)
		ttl := 24 * time.Hour
		if session.ExpiresAt != nil {
			ttl = time.Until(*session.ExpiresAt)
		}

		err = sm.redis.Set(ctx, key, data, ttl).Err()
		if err != nil {
			return fmt.Errorf("failed to store session in Redis: %w", err)
		}
	}

	return nil
}

// GetSession retrieves a session
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	// Try Redis first
	if sm.redis != nil {
		key := fmt.Sprintf("session:%s", sessionID)
		data, err := sm.redis.Get(ctx, key).Result()
		if err == nil {
			var session Session
			if err := json.Unmarshal([]byte(data), &session); err == nil {
				return &session, nil
			}
		}
	}

	// Fallback to database
	if sm.db != nil {
		query := `
			SELECT id, thread_id, user_id, metadata, created_at, expires_at
			FROM sessions
			WHERE id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		`

		row := sm.db.QueryRowContext(ctx, query, sessionID)

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
				return nil, fmt.Errorf("session %s not found or expired", sessionID)
			}
			return nil, fmt.Errorf("failed to get session: %w", err)
		}

		// Unmarshal metadata
		if err := json.Unmarshal(metadataData, &session.Metadata); err != nil {
			session.Metadata = make(map[string]interface{})
		}

		return &session, nil
	}

	return nil, fmt.Errorf("no storage backend available")
}

// CreateThread creates a new thread
func (sm *SessionManager) CreateThread(ctx context.Context, thread *Thread) error {
	if sm.db == nil {
		return fmt.Errorf("database not available")
	}

	query := `
		INSERT INTO threads (id, name, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	metadataData, err := json.Marshal(thread.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = sm.db.ExecContext(ctx, query,
		thread.ID,
		thread.Name,
		metadataData,
		thread.CreatedAt,
		thread.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create thread: %w", err)
	}

	return nil
}

// GetThread retrieves a thread
func (sm *SessionManager) GetThread(ctx context.Context, threadID string) (*Thread, error) {
	if sm.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT id, name, metadata, created_at, updated_at
		FROM threads
		WHERE id = $1
	`

	row := sm.db.QueryRowContext(ctx, query, threadID)

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
		thread.Metadata = make(map[string]interface{})
	}

	return &thread, nil
}
