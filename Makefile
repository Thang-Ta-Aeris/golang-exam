.PHONY: help build run test clean logs stop restart local-setup local-test

# Default target
help:
	@echo "🚀 Blog API Development Commands"
	@echo "================================"
	@echo ""
	@echo "🐳 Docker Commands (recommended):"
	@echo "make build     - Build the Docker image"
	@echo "make run       - Start all services with Docker Compose"
	@echo "make test      - Run API tests against Docker services"
	@echo "make logs      - View logs from all services"
	@echo "make stop      - Stop all services"
	@echo "make restart   - Restart all services"
	@echo "make clean     - Clean up Docker containers and volumes"
	@echo "make dev       - Full development workflow (clean, build, run, test)"
	@echo ""
	@echo "💻 Local Development Commands:"
	@echo "make local-setup  - Set up local development environment"
	@echo "make local-build  - Build the Go application locally"
	@echo "make local-run    - Run the API server locally"
	@echo "make local-test   - Test the API against local services"
	@echo ""
	@echo "📊 Individual Docker service logs:"
	@echo "make logs-api  - View API logs only"
	@echo "make logs-db   - View PostgreSQL logs only"
	@echo "make logs-redis - View Redis logs only"
	@echo "make logs-es   - View Elasticsearch logs only"
	@echo ""
	@echo "🏥 Health and Status:"
	@echo "make health    - Check all services health"
	@echo "make status    - Show service status"

# Docker Commands
build:
	@echo "🔨 Building Docker image..."
	docker compose build

run:
	@echo "🚀 Starting all services..."
	docker compose up -d
	@echo "✅ Services started! API available at http://localhost:8080"
	@echo "📊 Health check: curl http://localhost:8080/health"

test:
	@echo "🧪 Running API tests..."
	./test-api.sh

logs:
	docker compose logs -f

logs-api:
	docker compose logs -f blog-api

logs-db:
	docker compose logs -f postgres

logs-redis:
	docker compose logs -f redis

logs-es:
	docker compose logs -f elasticsearch

stop:
	@echo "🛑 Stopping all services..."
	docker compose down

restart: stop run

clean:
	@echo "🧹 Cleaning up Docker containers and volumes..."
	docker compose down -v --remove-orphans
	docker system prune -f

dev: clean build run
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@make test
	@echo "🎉 Development environment ready!"

# Local Development Commands
local-setup:
	@echo "🔧 Setting up local development environment..."
	./setup-local.sh

local-build:
	@echo "🔨 Building Go application..."
	go mod tidy
	go build -o blog-api cmd/api/main.go
	@echo "✅ Build successful!"

local-run: local-build
	@echo "🚀 Starting API server locally..."
	@echo "Make sure PostgreSQL, Redis, and Elasticsearch are running!"
	@echo "Use Ctrl+C to stop the server"
	./blog-api

local-test:
	@echo "🧪 Testing local API..."
	./test-local.sh

# Health and Status Commands
health:
	@echo "🏥 Checking service health..."
	@echo "API Health:"
	@curl -s http://localhost:8080/health | jq . || echo "❌ API not responding"
	@echo ""
	@echo "PostgreSQL:"
	@docker compose exec postgres pg_isready -U blog_user -d blog_db || echo "❌ PostgreSQL not ready"
	@echo ""
	@echo "Redis:"
	@docker compose exec redis redis-cli ping || echo "❌ Redis not responding"
	@echo ""
	@echo "Elasticsearch:"
	@curl -s http://localhost:9200/_cluster/health | jq .status || echo "❌ Elasticsearch not responding"

status:
	@echo "📊 Service Status:"
	docker compose ps

# Demo and Documentation
demo: dev
	@echo ""
	@echo "🎬 Demo Commands You Can Try:"
	@echo "==============================================="
	@echo "# Create a post:"
	@echo "curl -X POST http://localhost:8080/api/v1/posts \\"
	@echo "  -H 'Content-Type: application/json' \\"
	@echo "  -d '{\"title\":\"My Post\",\"content\":\"Content here\",\"tags\":[\"demo\"]}'"
	@echo ""
	@echo "# Get a post (demonstrates caching):"
	@echo "curl http://localhost:8080/api/v1/posts/1"
	@echo ""
	@echo "# Search by tag (demonstrates GIN index):"
	@echo "curl 'http://localhost:8080/api/v1/posts/search-by-tag?tag=golang'"
	@echo ""
	@echo "# Full-text search (demonstrates Elasticsearch):"
	@echo "curl 'http://localhost:8080/api/v1/posts/search?q=concurrency'"
	@echo ""
