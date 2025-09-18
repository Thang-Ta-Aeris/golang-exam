-- Initialize database schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create posts table
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create activity_logs table
CREATE TABLE activity_logs (
    id SERIAL PRIMARY KEY,
    action VARCHAR(100) NOT NULL,
    post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
    logged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create GIN index on tags for efficient array search
CREATE INDEX idx_posts_tags_gin ON posts USING GIN (tags);

-- Create index on created_at for sorting
CREATE INDEX idx_posts_created_at ON posts (created_at DESC);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_posts_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data
INSERT INTO posts (title, content, tags) VALUES
('Introduction to Go', 'Go is a programming language developed by Google...', ARRAY['golang', 'programming', 'tutorial']),
('Redis Caching Strategies', 'Redis is an in-memory data structure store...', ARRAY['redis', 'caching', 'performance']),
('Elasticsearch Full-text Search', 'Elasticsearch provides powerful search capabilities...', ARRAY['elasticsearch', 'search', 'database']);
