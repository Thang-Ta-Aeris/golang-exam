# Blog API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Endpoints

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "time": "2025-09-11T10:30:00Z"
}
```

### Create Post
```http
POST /posts
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Learning Go Concurrency",
  "content": "Goroutines and channels are powerful features...",
  "tags": ["golang", "concurrency", "programming"]
}
```

**Response (201 Created):**
```json
{
  "message": "Post created successfully",
  "data": {
    "id": 1,
    "title": "Learning Go Concurrency",
    "content": "Goroutines and channels are powerful features...",
    "tags": ["golang", "concurrency", "programming"],
    "created_at": "2025-09-11T10:30:00Z",
    "updated_at": "2025-09-11T10:30:00Z"
  }
}
```

### Get Post by ID
```http
GET /posts/{id}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "title": "Learning Go Concurrency",
    "content": "Goroutines and channels are powerful features...",
    "tags": ["golang", "concurrency", "programming"],
    "created_at": "2025-09-11T10:30:00Z",
    "updated_at": "2025-09-11T10:30:00Z"
  }
}
```

**Response (404 Not Found):**
```json
{
  "error": "Post not found"
}
```

### Update Post
```http
PUT /posts/{id}
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Advanced Go Concurrency Patterns",
  "content": "Updated content about advanced patterns...",
  "tags": ["golang", "concurrency", "advanced", "patterns"]
}
```

**Response (200 OK):**
```json
{
  "message": "Post updated successfully",
  "data": {
    "id": 1,
    "title": "Advanced Go Concurrency Patterns",
    "content": "Updated content about advanced patterns...",
    "tags": ["golang", "concurrency", "advanced", "patterns"],
    "created_at": "2025-09-11T10:30:00Z",
    "updated_at": "2025-09-11T10:30:15Z"
  }
}
```

### Search Posts by Tag
```http
GET /posts/search-by-tag?tag={tag_name}
```

**Example:**
```http
GET /posts/search-by-tag?tag=golang
```

**Response (200 OK):**
```json
{
  "data": {
    "posts": [
      {
        "id": 1,
        "title": "Learning Go Concurrency",
        "content": "Goroutines and channels...",
        "tags": ["golang", "concurrency", "programming"],
        "created_at": "2025-09-11T10:30:00Z",
        "updated_at": "2025-09-11T10:30:00Z"
      }
    ],
    "total": 1,
    "tag": "golang"
  }
}
```

### Full-text Search
```http
GET /posts/search?q={query_string}
```

**Example:**
```http
GET /posts/search?q=concurrency
```

**Response (200 OK):**
```json
{
  "data": {
    "posts": [
      {
        "id": 1,
        "title": "Learning Go Concurrency",
        "content": "Goroutines and channels...",
        "tags": ["golang", "concurrency", "programming"]
      }
    ],
    "total": 1,
    "query": "concurrency"
  }
}
```

## Error Responses

### Validation Error (400 Bad Request)
```json
{
  "error": "Validation failed",
  "details": "Key: 'CreatePostRequest.Title' Error:Field validation for 'Title' failed on the 'required' tag"
}
```

### Internal Server Error (500)
```json
{
  "error": "Failed to create post",
  "details": "database connection error"
}
```

## Features Demonstrated

### 1. PostgreSQL with GIN Index
- **Tag Search Optimization**: Uses GIN index for efficient array queries
- **Query**: `SELECT * FROM posts WHERE tags @> ARRAY['golang']`
- **Performance**: O(log n) lookup time for tag searches

### 2. Redis Cache-Aside Pattern
- **Cache Key Format**: `post:{id}`
- **TTL**: 5 minutes
- **Flow**:
  1. Check Redis cache
  2. On miss, query PostgreSQL
  3. Store result in Redis
  4. Return data

### 3. Cache Invalidation
- **Trigger**: POST/PUT operations
- **Action**: Delete corresponding Redis key
- **Ensures**: Next request gets fresh data

### 4. Transaction Support
- **Operations**: Post creation + Activity log insertion
- **ACID Compliance**: Both operations succeed or both fail
- **Rollback**: Automatic on any failure

### 5. Elasticsearch Full-text Search
- **Index**: `posts`
- **Fields**: `title`, `content`
- **Query Type**: Multi-match query
- **Sync**: Asynchronous indexing after DB operations

## Performance Characteristics

### Database Queries
- **Tag Search**: Uses GIN index for sub-millisecond lookups
- **Primary Key Lookup**: B-tree index for O(log n) performance
- **Connection Pooling**: Efficient connection reuse

### Caching Strategy
- **Cache Hit Ratio**: ~80-90% for read-heavy workloads
- **Response Time**: Sub-millisecond for cached responses
- **Memory Usage**: Configurable TTL prevents memory bloat

### Search Performance
- **Index Size**: Scales with document count
- **Query Time**: Typically < 50ms for simple queries
- **Relevance**: BM25 scoring algorithm

## Example Usage with curl

### Complete Workflow
```bash
# 1. Create a post
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Microservices with Go",
    "content": "Building scalable microservices using Go, gRPC, and containerization.",
    "tags": ["golang", "microservices", "grpc", "docker"]
  }'

# 2. Get the post (cache miss -> database -> cache populate)
curl http://localhost:8080/api/v1/posts/1

# 3. Get the post again (cache hit)
curl http://localhost:8080/api/v1/posts/1

# 4. Search by tag (GIN index)
curl "http://localhost:8080/api/v1/posts/search-by-tag?tag=golang"

# 5. Update the post (cache invalidation)
curl -X PUT http://localhost:8080/api/v1/posts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Advanced Microservices Patterns in Go",
    "content": "Updated content with advanced patterns...",
    "tags": ["golang", "microservices", "patterns", "advanced"]
  }'

# 6. Full-text search (Elasticsearch)
curl "http://localhost:8080/api/v1/posts/search?q=microservices"
```
