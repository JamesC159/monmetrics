.PHONY: install dev build preview clean setup

# Install dependencies
install:
	@echo "Installing backend dependencies..."
	cd backend && go mod tidy
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

# Start development servers
dev:
	@echo "Starting MongoDB..."
	docker-compose up -d mongodb
	@echo "Starting development servers..."
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:3000"
	@(cd backend && go run cmd/server/main.go) & \
	(cd frontend && npm run dev) & \
	wait

# Build for production
build:
	@echo "Building backend..."
	cd backend && go build -o bin/server cmd/server/main.go
	@echo "Building frontend..."
	cd frontend && npm run build

# Preview production build
preview:
	cd frontend && npm run preview

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	cd backend && rm -rf bin/
	cd frontend && rm -rf dist/
	docker-compose down

# Initial project setup
setup: install
	@echo "Setting up project..."
	@if [ ! -f backend/.env ]; then \
		cp backend/.env.example backend/.env; \
		echo "Created backend/.env from example"; \
	fi
	@if [ ! -f frontend/.env.local ]; then \
		cp frontend/.env.example frontend/.env.local; \
		echo "Created frontend/.env.local from example"; \
	fi
	@echo "Starting MongoDB for initial setup..."
	docker-compose up -d mongodb
	@echo "Waiting for MongoDB to be ready..."
	sleep 5
	@echo "Setup complete! Run 'make dev' to start development servers."

# Test backend
test-backend:
	cd backend && go test ./...

# Test frontend
test-frontend:
	cd frontend && npm test

# Lint frontend
lint-frontend:
	cd frontend && npm run lint

# Type check frontend
type-check:
	cd frontend && npm run type-check

# Production start
start-prod: build
	@echo "Starting production servers..."
	cd frontend && npm run serve

# Docker compose for development
dev-docker:
	docker-compose up --build