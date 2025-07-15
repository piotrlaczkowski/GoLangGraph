-- GoLangGraph Stateful Agents Database Initialization Script
-- This script sets up PostgreSQL with pgvector extension and all required tables

-- Enable pgvector extension for vector operations
CREATE EXTENSION IF NOT EXISTS vector;

-- Create enhanced user if not exists (with proper permissions)
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'golanggraph') THEN
      CREATE ROLE golanggraph LOGIN PASSWORD 'stateful_password_2024'; -- pragma: allowlist secret
   END IF;
END
$do$;

-- Grant necessary permissions
ALTER USER golanggraph CREATEDB;
GRANT ALL PRIVILEGES ON DATABASE golanggraph_stateful TO golanggraph;

-- Create core tables for GoLangGraph persistence

-- Threads table for conversation management
CREATE TABLE IF NOT EXISTS threads (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Checkpoints table for state persistence
CREATE TABLE IF NOT EXISTS checkpoints (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    state_data JSONB NOT NULL,
    metadata JSONB DEFAULT '{}',
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
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

-- Documents table for RAG document storage with vector embeddings
CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255),
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    embedding vector(1536), -- OpenAI embedding dimension
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
    embedding vector(1536),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

-- Stateful agents specific tables

-- User preferences for learning and adaptation
CREATE TABLE IF NOT EXISTS user_preferences (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    thread_id VARCHAR(255),
    agent_type VARCHAR(100) NOT NULL,
    preferences JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

-- Design iterations for the Designer agent
CREATE TABLE IF NOT EXISTS design_iterations (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    design_concept TEXT,
    feedback TEXT,
    rating FLOAT DEFAULT 0,
    improvements TEXT[],
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

-- Conversation analysis for the Highlighter agent
CREATE TABLE IF NOT EXISTS conversation_analysis (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    analysis_type VARCHAR(100),
    insights TEXT[],
    themes JSONB DEFAULT '[]',
    sentiment_score FLOAT DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

-- Story creations for the Storymaker agent
CREATE TABLE IF NOT EXISTS story_creations (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    title VARCHAR(500),
    story_content TEXT,
    characters JSONB DEFAULT '[]',
    themes TEXT[],
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

-- Interview sessions for the Interviewer agent
CREATE TABLE IF NOT EXISTS interview_sessions (
    id VARCHAR(255) PRIMARY KEY,
    thread_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    interview_phase VARCHAR(100),
    topics_covered TEXT[],
    user_profile JSONB DEFAULT '{}',
    requirements JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
);

-- Create performance indexes

-- Basic indexes for performance
CREATE INDEX IF NOT EXISTS idx_checkpoints_thread_id ON checkpoints(thread_id);
CREATE INDEX IF NOT EXISTS idx_checkpoints_created_at ON checkpoints(created_at);
CREATE INDEX IF NOT EXISTS idx_checkpoints_node_id ON checkpoints(node_id);
CREATE INDEX IF NOT EXISTS idx_checkpoints_step_id ON checkpoints(step_id);

CREATE INDEX IF NOT EXISTS idx_sessions_thread_id ON sessions(thread_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Vector indexes for similarity search (using ivfflat)
CREATE INDEX IF NOT EXISTS idx_documents_embedding ON documents USING ivfflat (embedding vector_cosine_ops);
CREATE INDEX IF NOT EXISTS idx_memory_embedding ON memory USING ivfflat (embedding vector_cosine_ops);

-- Content indexes for text search
CREATE INDEX IF NOT EXISTS idx_documents_thread_id ON documents(thread_id);
CREATE INDEX IF NOT EXISTS idx_documents_content ON documents USING gin(to_tsvector('english', content));

CREATE INDEX IF NOT EXISTS idx_memory_thread_id ON memory(thread_id);
CREATE INDEX IF NOT EXISTS idx_memory_user_id ON memory(user_id);
CREATE INDEX IF NOT EXISTS idx_memory_type ON memory(memory_type);
CREATE INDEX IF NOT EXISTS idx_memory_content ON memory USING gin(to_tsvector('english', content));

-- Agent-specific indexes
CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_user_preferences_agent_type ON user_preferences(agent_type);

CREATE INDEX IF NOT EXISTS idx_design_iterations_thread_id ON design_iterations(thread_id);
CREATE INDEX IF NOT EXISTS idx_design_iterations_user_id ON design_iterations(user_id);

CREATE INDEX IF NOT EXISTS idx_conversation_analysis_thread_id ON conversation_analysis(thread_id);
CREATE INDEX IF NOT EXISTS idx_conversation_analysis_user_id ON conversation_analysis(user_id);

CREATE INDEX IF NOT EXISTS idx_story_creations_thread_id ON story_creations(thread_id);
CREATE INDEX IF NOT EXISTS idx_story_creations_user_id ON story_creations(user_id);

CREATE INDEX IF NOT EXISTS idx_interview_sessions_thread_id ON interview_sessions(thread_id);
CREATE INDEX IF NOT EXISTS idx_interview_sessions_user_id ON interview_sessions(user_id);

-- Grant permissions on all tables
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO golanggraph;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO golanggraph;

-- Insert initial data for testing

-- Create a default thread for testing
INSERT INTO threads (id, name, metadata, created_at, updated_at)
VALUES ('test_thread_001', 'Default Test Thread', '{"purpose": "testing", "version": "1.0"}', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Insert sample user preferences
INSERT INTO user_preferences (id, user_id, thread_id, agent_type, preferences, created_at, updated_at)
VALUES
    ('pref_001', 'test_user_123', 'test_thread_001', 'designer', '{"style": "modern", "sustainability": "high"}', NOW(), NOW()),
    ('pref_002', 'test_user_123', 'test_thread_001', 'interviewer', '{"language": "french", "depth": "detailed"}', NOW(), NOW()),
    ('pref_003', 'test_user_123', 'test_thread_001', 'highlighter', '{"focus": "insights", "format": "structured"}', NOW(), NOW()),
    ('pref_004', 'test_user_123', 'test_thread_001', 'storymaker', '{"genre": "sci-fi", "length": "medium"}', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Verify installation
SELECT 'Database initialization completed successfully!' as status,
       version() as postgres_version,
       extversion as pgvector_version
FROM pg_extension WHERE extname = 'vector';
