# MonMetrics Development Makefile
# Provides easy commands for development workflow

# Ensure we use Go 1.24.2 from /usr/local/go and Go tool binaries (air) are on PATH
export PATH := /usr/local/go/bin:$(HOME)/go/bin:$(PATH)

.PHONY: help install install-tools dev build preview clean setup seed test-backend test-frontend lint-frontend type-check start-prod dev-docker

# Default target - show help
help:
	@echo "🎯 MonMetrics Development Commands"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "📦 Setup Commands:"
	@echo "  make install     - Install all dependencies"
	@echo "  make install-tools - Install dev tools (air live reload)"
	@echo "  make setup       - Initial project setup with .env files"
	@echo "  make seed        - Populate database with sample data"
	@echo "  make full-setup  - Complete setup (install + setup + seed)"
	@echo ""
	@echo "🚀 Development Commands:"
	@echo "  make dev         - Start development servers"
	@echo "  make dev-docker  - Start with Docker Compose"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  make build       - Build for production"
	@echo "  make preview     - Preview production build"
	@echo "  make start-prod  - Start production servers"
	@echo ""
	@echo "🧪 Testing Commands:"
	@echo "  make test-backend    - Run backend tests"
	@echo "  make test-frontend   - Run frontend tests"
	@echo "  make lint-frontend   - Lint frontend code"
	@echo "  make type-check      - TypeScript type checking"
	@echo ""
	@echo "🧹 Maintenance Commands:"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make reset       - Reset database and clean build"
	@echo ""
	@echo "📚 Quick Start:"
	@echo "  make full-setup && make dev"

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	@echo "Installing backend dependencies..."
	cd backend && go mod tidy
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@$(MAKE) install-tools
	@echo "✅ Dependencies installed successfully!"

# Install Go dev tooling (live reload)
install-tools:
	@command -v air >/dev/null 2>&1 || { \
		echo "📥 Installing air (Go live reload)..."; \
		cd backend && go install github.com/air-verse/air@latest; \
	}
	@echo "✅ Dev tools ready (air)"

# Start development servers
dev: install-tools
	@echo "🚀 Starting development environment..."
	@echo "Starting MongoDB..."
	@docker-compose up -d mongodb
	@echo "⏳ Waiting for MongoDB to be ready..."
	@sleep 3
	@echo "🎯 Starting development servers..."
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:3000"
	@echo "Health Check: http://localhost:8080/health"
	@echo "🔁 Hot reload: Go (air) + Vite HMR — no restarts needed"
	@echo ""
	@echo "Press Ctrl+C to stop all servers"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@trap 'echo "\n🛑 Stopping servers..."; kill 0; docker-compose stop; exit 0' INT TERM; \
	(cd backend && air -c .air.toml) & \
	(cd frontend && npm run dev) & \
	wait

# Build for production
build:
	@echo "🏗️ Building for production..."
	@echo "Building backend..."
	cd backend && go build -o bin/server cmd/server/main.go
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "✅ Production build completed!"

# Preview production build
preview:
	@echo "👀 Starting production preview..."
	cd frontend && npm run preview

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	cd backend && rm -rf bin/ tmp/
	cd frontend && rm -rf dist/
	@echo "🗑️ Stopping containers..."
	@docker-compose down
	@echo "✅ Clean completed!"

# Reset everything (clean + remove database)
reset: clean
	@echo "🔄 Resetting database..."
	@docker-compose down -v
	@docker volume rm monmetrics_mongodb_data 2>/dev/null || true
	@echo "✅ Full reset completed!"

# Initial project setup
setup: install
	@echo "⚙️ Setting up project configuration..."
	@if [ ! -f backend/.env ]; then \
		cp backend/.env.example backend/.env; \
		echo "✅ Created backend/.env from example"; \
	else \
		echo "ℹ️ backend/.env already exists"; \
	fi
	@if [ ! -f frontend/.env.local ]; then \
		cp frontend/.env.example frontend/.env.local; \
		echo "✅ Created frontend/.env.local from example"; \
	else \
		echo "ℹ️ frontend/.env.local already exists"; \
	fi
	@echo "🐳 Starting MongoDB for initial setup..."
	@docker-compose up -d mongodb
	@echo "⏳ Waiting for MongoDB to be ready..."
	@sleep 5
	@echo "✅ Setup complete!"
	@echo ""
	@echo "📝 Next steps:"
	@echo "1. Run 'make seed' to populate the database"
	@echo "2. Run 'make dev' to start development servers"

# Populate database with sample data
seed:
	@echo "🌱 Seeding database with sample data..."
	@echo "🐳 Ensuring MongoDB is running..."
	@docker-compose up -d mongodb
	@echo "⏳ Waiting for MongoDB to be ready..."
	@sleep 3
	@echo "🔍 Testing MongoDB connection..."
	@if ! docker exec monmetrics_mongo mongosh --eval "db.adminCommand('ping')" --quiet > /dev/null 2>&1; then \
		echo "❌ Error: MongoDB is not responding"; \
		echo "   Try: make reset && make setup"; \
		exit 1; \
	fi
	@echo "📦 Updating Go dependencies..."
	@cd backend && go mod tidy
	@echo "🧹 Cleaning Go build cache..."
	@cd backend && go clean -cache
	@echo "🔨 Building seeder..."
	@cd backend && go build -o bin/seeder ./cmd/seeder
	@echo "🚀 Running database seeder..."
	@cd backend && ./bin/seeder
	@echo ""
	@echo "🎉 Database seeded successfully!"
	@echo ""
	@echo "🎯 You can now search for these sample cards:"
	@echo "  • Charizard VMAX"
	@echo "  • Black Lotus"
	@echo "  • Blue-Eyes White Dragon"
	@echo "  • Pikachu VMAX"
	@echo "  • Pokemon Base Set Booster Box"
	@echo "  • Magic Alpha Starter Deck"
	@echo ""
	@echo "💡 Run 'make dev' to start the application!"

# Complete setup workflow
full-setup: setup seed
	@echo ""
	@echo "🎉 Full setup completed successfully!"
	@echo ""
	@echo "🚀 Ready to start development:"
	@echo "   make dev"
	@echo ""
	@echo "📊 Database contains:"
	@echo "   • 11 sample trading cards"
	@echo "   • 5 years of price history per card"
	@echo "   • Sample marketplace listings"
	@echo ""
	@echo "🌐 Application URLs:"
	@echo "   • Frontend: http://localhost:3000"
	@echo "   • Backend:  http://localhost:8080"
	@echo "   • Health:   http://localhost:8080/health"

# Test backend
test-backend:
	@echo "🧪 Running backend tests..."
	cd backend && go test ./...

# Test frontend
test-frontend:
	@echo "🧪 Running frontend tests..."
	cd frontend && npm test

# Lint frontend
lint-frontend:
	@echo "🔍 Linting frontend code..."
	cd frontend && npm run lint

# Type check frontend
type-check:
	@echo "📝 Running TypeScript type check..."
	cd frontend && npm run type-check

# Production start
start-prod: build
	@echo "🚀 Starting production servers..."
	@echo "Starting MongoDB..."
	@docker-compose up -d mongodb
	@echo "Starting backend server..."
	@cd backend && ./bin/server &
	@echo "Starting frontend server..."
	cd frontend && npm run serve

# Docker compose for development
dev-docker:
	@echo "🐳 Starting development environment with Docker..."
	docker-compose up --build

# Database management commands
db-status:
	@echo "📊 Database Status:"
	@docker exec monmetrics_mongo mongosh --eval "db.adminCommand('ping')" --quiet && echo "✅ MongoDB is running" || echo "❌ MongoDB is not running"
	@docker exec monmetrics_mongo mongosh monmetrics --eval "db.stats()" --quiet 2>/dev/null | grep -E "(collections|dataSize|objects)" || echo "🗃️ Database is empty"

db-backup:
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	@docker exec monmetrics_mongo mongodump --db monmetrics --out /tmp/backup
	@docker cp monmetrics_mongo:/tmp/backup ./backups/backup-$(shell date +%Y%m%d-%H%M%S)
	@echo "✅ Backup created in ./backups/"

db-restore:
	@echo "🔄 Database restore requires a backup directory"
	@echo "Usage: make db-restore BACKUP_DIR=./backups/backup-YYYYMMDD-HHMMSS"
	@if [ -n "$(BACKUP_DIR)" ] && [ -d "$(BACKUP_DIR)" ]; then \
		docker cp $(BACKUP_DIR) monmetrics_mongo:/tmp/restore; \
		docker exec monmetrics_mongo mongorestore --db monmetrics /tmp/restore/monmetrics; \
		echo "✅ Database restored from $(BACKUP_DIR)"; \
	fi

# Logs and monitoring
logs:
	@echo "📋 Showing container logs..."
	docker-compose logs -f

logs-mongo:
	@echo "📋 Showing MongoDB logs..."
	docker-compose logs -f mongodb

# Development helpers
check-ports:
	@echo "🔍 Checking if required ports are available..."
	@if lsof -Pi :3000 -sTCP:LISTEN -t >/dev/null 2>&1; then \
		echo "❌ Port 3000 is already in use"; \
		lsof -Pi :3000 -sTCP:LISTEN; \
	else \
		echo "✅ Port 3000 is available"; \
	fi
	@if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then \
		echo "❌ Port 8080 is already in use"; \
		lsof -Pi :8080 -sTCP:LISTEN; \
	else \
		echo "✅ Port 8080 is available"; \
	fi
	@if lsof -Pi :27017 -sTCP:LISTEN -t >/dev/null 2>&1; then \
		echo "❌ Port 27017 is already in use"; \
		lsof -Pi :27017 -sTCP:LISTEN; \
	else \
		echo "✅ Port 27017 is available"; \
	fi

# Show system information
info:
	@echo "ℹ️ System Information:"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Go version: $(shell go version 2>/dev/null || echo 'Not installed')"
	@echo "Node version: $(shell node --version 2>/dev/null || echo 'Not installed')"
	@echo "Docker version: $(shell docker --version 2>/dev/null || echo 'Not installed')"
	@echo "Docker Compose version: $(shell docker-compose --version 2>/dev/null || echo 'Not installed')"
	@echo ""
	@echo "🐳 Container Status:"
	@docker-compose ps 2>/dev/null || echo "No containers running"