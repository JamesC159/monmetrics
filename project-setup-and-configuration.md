# .env.example
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=MonMetrics
VITE_APP_DESCRIPTION=Professional Trading Card Analysis

# .env.local (for development - create this file)
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=MonMetrics
VITE_APP_DESCRIPTION=Professional Trading Card Analysis

# backend/.env.example
PORT=8080
MONGODB_URI=mongodb://localhost:27017
DB_NAME=monmetrics
JWT_SECRET=your-super-secret-jwt-key-change-this
CORS_ORIGINS=http://localhost:3000
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60s
ENVIRONMENT=development

# backend/.env (for development - create this file)
PORT=8080
MONGODB_URI=mongodb://localhost:27017
DB_NAME=monmetrics
JWT_SECRET=your-super-secret-jwt-key-change-this
CORS_ORIGINS=http://localhost:3000
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60s
ENVIRONMENT=development

# .gitignore
# Logs
logs
*.log
npm-debug.log*
yarn-debug.log*
yarn-error.log*
pnpm-debug.log*
lerna-debug.log*

# Dependencies
node_modules
go.sum

# Build outputs
dist
dist-ssr
build
bin/
*.local

# Environment variables
.env
.env.local
.env.production
.env.development

# Editor directories and files
.vscode/*
!.vscode/extensions.json
.idea
.DS_Store
*.suo
*.ntvs*
*.njsproj
*.sln
*.sw?

# Testing
coverage/
.nyc_output

# ESLint cache
.eslintcache

# OS generated files
Thumbs.db
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db

# MongoDB data
data/

# Docker volumes
mongodb_data/

# docker-compose.yml (optional - for development)
version: '3.8'
services:
  mongodb:
    image: mongo:7.0
    container_name: monmetrics_mongo
    restart: unless-stopped
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_DATABASE: monmetrics
    volumes:
      - mongodb_data:/data/db

volumes:
  mongodb_data:

# Makefile for easy development
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

# README.md
# MonMetrics - Professional Trading Card Analysis

MonMetrics is a powerful web application for tracking, analyzing, and predicting trading card prices with advanced technical indicators. Built with Go (backend) and React 19 (frontend) for maximum performance and scalability.

## 🚀 Features

- **Advanced Price Analytics**: Track price movements with 5 years of historical data
- **Technical Indicators**: Apply professional trading indicators (Bollinger Bands, RSI, Moving Averages)
- **Real-time Updates**: Get instant market updates and price alerts
- **Secure & Reliable**: Enterprise-grade security with 99.9% uptime
- **Multi-platform Support**: Pokemon, Magic the Gathering, Yu-Gi-Oh, and more
- **Professional Charts**: Save and share your analysis with advanced charting tools

## 🛠 Tech Stack

### Backend
- **Language**: Go 1.21+
- **Database**: MongoDB 7.0
- **Authentication**: JWT with HMAC-SHA256
- **Security**: OWASP compliant with comprehensive middleware
- **Performance**: Pure Go stdlib, no external frameworks

### Frontend
- **Framework**: React 19 with native SSR
- **Build Tool**: Vite 5.0
- **Styling**: Tailwind CSS
- **Routing**: React Router v6
- **Charts**: Recharts
- **Icons**: Lucide React

## 📋 Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Docker and Docker Compose
- MongoDB (via Docker)

## 🚀 Quick Start

### 1. Clone and Setup
```bash
git clone <repository-url>
cd monmetrics
make setup
```

### 2. Start Development Servers
```bash
make dev
```

This will:
- Start MongoDB via Docker
- Launch the Go backend on http://localhost:8080
- Launch the React frontend on http://localhost:3000

### 3. Open Your Browser
Navigate to http://localhost:3000 to see the landing page.

## 📁 Project Structure

```
monmetrics/
├── backend/                 # Go backend
│   ├── cmd/server/         # Main application entry
│   ├── internal/           # Private application code
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # HTTP middleware
│   │   ├── models/         # Data models
│   │   ├── database/       # Database connection
│   │   └── services/       # Business logic
│   ├── configs/            # Configuration
│   └── scripts/            # Setup scripts
├── frontend/               # React 19 frontend
│   ├── src/
│   │   ├── components/     # Reusable components
│   │   ├── pages/          # Page components
│   │   ├── context/        # React contexts
│   │   ├── hooks/          # Custom hooks
│   │   ├── utils/          # Utilities
│   │   └── types/          # TypeScript types
│   ├── public/             # Static assets
│   └── server.js           # SSR server
├── docker-compose.yml      # Development services
├── Makefile               # Development commands
└── README.md
```

## 🔧 Development Commands

| Command | Description |
|---------|-------------|
| `make setup` | Initial project setup |
| `make dev` | Start development servers |
| `make build` | Build for production |
| `make clean` | Clean build artifacts |
| `make test-backend` | Run backend tests |
| `make test-frontend` | Run frontend tests |

## 🏗 Building for Production

### Backend
```bash
cd backend
go build -o bin/server cmd/server/main.go
```

### Frontend
```bash
cd frontend
npm run build
npm run serve
```

## 🚀 Deployment

The application is designed to run on a single dedicated server:

1. **Backend**: Compile to a single binary
2. **Frontend**: Build static assets with SSR server
3. **Database**: MongoDB (can be containerized)

### Environment Variables

#### Backend (.env)
```bash
PORT=8080
MONGODB_URI=mongodb://localhost:27017
DB_NAME=monmetrics
JWT_SECRET=your-super-secret-jwt-key
CORS_ORIGINS=http://localhost:3000
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60s
ENVIRONMENT=production
```

#### Frontend (.env.local)
```bash
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=MonMetrics
VITE_APP_DESCRIPTION=Professional Trading Card Analysis
```

## 🔒 Security Features

- **OWASP Compliance**: Implements OWASP Secure Coding Practices
- **CSRF Protection**: Cross-site request forgery protection
- **XSS Prevention**: Content Security Policy and input sanitization
- **Rate Limiting**: Configurable rate limiting per IP
- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: Secure password storage
- **HTTPS Ready**: TLS/SSL configuration support

## 📊 Features in Detail

### For Free Users
- Up to 3 technical indicators
- Basic price history
- Limited search results
- Community support

### For Pro Users ($19/month)
- Up to 10 technical indicators
- 5 years price history
- Unlimited searches
- Advanced charting tools
- Price alerts
- Priority support
- Data export capabilities

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 API Documentation

The backend provides a REST API with the following endpoints:

### Public Endpoints
- `GET /health` - Health check
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/cards/search` - Search cards
- `GET /api/cards/{id}` - Get card details
- `GET /api/cards/{id}/prices` - Get price history

### Protected Endpoints (Require Authentication)
- `GET /api/protected/user/dashboard` - User dashboard
- `POST /api/protected/user/charts` - Save chart
- `GET /api/protected/user/charts` - Get saved charts
- `DELETE /api/protected/user/charts/{id}` - Delete chart

## 🐛 Troubleshooting

### Common Issues

1. **MongoDB Connection Failed**
   ```bash
   docker-compose up -d mongodb
   ```

2. **Frontend Build Fails**
   ```bash
   cd frontend && rm -rf node_modules && npm install
   ```

3. **Backend Won't Start**
   ```bash
   cd backend && go mod tidy
   ```

4. **Port Already in Use**
   - Change ports in `.env` files
   - Kill existing processes

## 📄 License

This project is proprietary software. All rights reserved.

## 📧 Support

For support, email contact@monmetrics.com or visit our support portal.

## 🎯 Roadmap

- [ ] Mobile applications (iOS/Android)
- [ ] Advanced portfolio tracking
- [ ] Machine learning price predictions
- [ ] Social features and community
- [ ] API for third-party integrations
- [ ] Advanced alerting system