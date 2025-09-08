# MonMetrics - Trading Card Price Analysis Platform

A powerful web application for tracking, analyzing, and predicting trading card prices with advanced technical indicators. Built with **Go** (backend) and **React 19** (frontend) for maximum performance and scalability.

## 🚀 Quick Start (5 minutes)

```bash
# 1. Clone the repository
git clone <your-repo-url>
cd monmetrics

# 2. Complete setup with sample data
make full-setup

# 3. Start development servers
make dev
```

That's it! Visit **http://localhost:3000** to start using MonMetrics.

## 📋 Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Node.js 18+** - [Download](https://nodejs.org/)
- **Docker & Docker Compose** - [Download](https://docker.com/get-started)

### Verify Installation

```bash
go version      # Should show 1.21+
node --version  # Should show 18+
docker --version
docker-compose --version
```

## 🛠 Project Structure

```
monmetrics/
├── backend/                    # Go backend API
│   ├── cmd/
│   │   ├── server/            # Main server application
│   │   └── seeder/            # Database seeder
│   ├── internal/
│   │   ├── handlers/          # HTTP request handlers
│   │   ├── middleware/        # HTTP middleware
│   │   ├── models/            # Data models
│   │   ├── database/          # MongoDB connection
│   │   └── services/          # Business logic
│   ├── configs/               # Configuration management
│   └── go.mod                 # Go dependencies
├── frontend/                  # React 19 frontend
│   ├── src/
│   │   ├── components/        # Reusable components
│   │   ├── pages/             # Page components
│   │   ├── context/           # React contexts
│   │   ├── hooks/             # Custom hooks
│   │   ├── utils/             # Utilities & API client
│   │   └── types/             # TypeScript definitions
│   ├── public/                # Static assets
│   ├── package.json           # Node dependencies
│   └── vite.config.ts         # Vite configuration
├── docker-compose.yml         # Development services
├── Makefile                   # Development commands
└── README.md                  # This file
```

## ⚙️ Detailed Setup

### Step 1: Install Dependencies

```bash
make install
```

This installs:
- Go backend dependencies via `go mod tidy`
- Frontend dependencies via `npm install`

### Step 2: Environment Configuration

```bash
make setup
```

This creates:
- `backend/.env` - Backend configuration
- `frontend/.env.local` - Frontend configuration
- Starts MongoDB container

### Step 3: Populate Database

```bash
make seed
```

This creates sample data:
- **11 trading cards** across Pokemon, Magic, and Yu-Gi-Oh
- **5 years of price history** for each card
- **Sample marketplace listings**
- **Technical indicators** and market data

### Step 4: Start Development

```bash
make dev
```

This starts:
- **Backend server** on http://localhost:8080
- **Frontend server** on http://localhost:3000
- **MongoDB** via Docker

## 🎯 Available Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make full-setup` | Complete setup (install + config + seed) |
| `make dev` | Start development servers |
| `make build` | Build for production |
| `make test-backend` | Run backend tests |
| `make test-frontend` | Run frontend tests |
| `make clean` | Clean build artifacts |
| `make reset` | Reset database and builds |
| `make seed` | Populate database with sample data |
| `make db-status` | Check database status |

## 🧪 Testing the Application

### 1. Search Functionality

Visit http://localhost:3000/search and try searching for:

- **"Charizard"** - Should find Charizard VMAX
- **"Black Lotus"** - Should find the iconic Magic card
- **"Blue-Eyes"** - Should find Blue-Eyes White Dragon
- **"Pokemon"** - Should filter by game
- **"sealed"** - Should show sealed products

### 2. Card Detail Pages

Click on any search result to view:
- **Price history charts** with 5 years of data
- **Current market listings** from eBay and TCGPlayer
- **All-time high/low prices** with dates
- **Interactive time range selection** (1D, 7D, 30D, 90D, 1Y, 5Y)

### 3. User Registration

1. Click "Sign In" → "Sign Up"
2. Create an account
3. Access the dashboard with saved charts functionality

## 🔧 Configuration Options

### Backend Configuration (backend/.env)

```env
PORT=8080                                    # Server port
MONGODB_URI=mongodb://localhost:27017        # Database connection
DB_NAME=monmetrics                          # Database name
JWT_SECRET=your-super-secret-jwt-key        # JWT signing key
CORS_ORIGINS=http://localhost:3000          # Allowed origins
RATE_LIMIT_REQUESTS=100                     # Rate limit
RATE_LIMIT_WINDOW=60s                       # Rate limit window
ENVIRONMENT=development                      # Environment
```

### Frontend Configuration (frontend/.env.local)

```env
VITE_API_BASE_URL=http://localhost:8080     # Backend API URL
VITE_APP_TITLE=MonMetrics                   # App title
VITE_APP_DESCRIPTION=Professional Trading Card Analysis
```

## 📊 Features Overview

### 🔓 Public Features
- **Advanced Search** - Find cards by name, game, set, rarity
- **Price History** - 5 years of historical data with charts
- **Market Listings** - Current eBay and TCGPlayer listings
- **Technical Analysis** - Basic indicators for all users

### 🔒 Premium Features (Registered Users)
- **Dashboard** - Personal analytics and saved charts
- **Advanced Indicators** - Up to 10 technical indicators (vs 3 for free)
- **Price Alerts** - Get notified of price changes
- **Chart Saving** - Save and share your analysis

### 📈 Technical Indicators (Coming Soon)
- **Bollinger Bands** - Volatility analysis
- **RSI** - Relative Strength Index
- **Moving Averages** - SMA, EMA analysis
- **MACD** - Trend momentum
- **Volume Analysis** - Market activity patterns

## 🚀 Production Deployment

### Build for Production

```bash
make build
```

This creates:
- `backend/bin/server` - Compiled Go binary
- `frontend/dist/` - Static assets with SSR

### Environment Variables

Update production environment files:

**Backend (.env):**
```env
PORT=8080
MONGODB_URI=mongodb://your-production-mongodb:27017
DB_NAME=monmetrics_prod
JWT_SECRET=your-super-secure-production-jwt-key
CORS_ORIGINS=https://your-domain.com
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=60s
ENVIRONMENT=production
```

**Frontend (.env.production):**
```env
VITE_API_BASE_URL=https://api.your-domain.com
VITE_APP_TITLE=MonMetrics
VITE_APP_DESCRIPTION=Professional Trading Card Analysis
```

### Start Production Servers

```bash
make start-prod
```

## 🔒 Security Features

MonMetrics implements **OWASP Secure Coding Practices**:

- **CORS Protection** - Configurable allowed origins
- **Rate Limiting** - Prevent abuse and DDoS
- **JWT Authentication** - Secure token-based auth
- **Input Validation** - Prevent injection attacks
- **Security Headers** - CSP, HSTS, X-Frame-Options
- **Password Hashing** - Secure password storage
- **SQL Injection Prevention** - Parameterized queries

## 🐛 Troubleshooting

### Common Issues

**❌ MongoDB Connection Failed**
```bash
make reset
make setup
```

**❌ Port Already in Use**
```bash
make check-ports  # Check what's using ports
# Kill processes using ports 3000, 8080, or 27017
```

**❌ Frontend Build Fails**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

**❌ Backend Won't Start**
```bash
cd backend
go mod tidy
go clean -cache
```

**❌ Database Empty After Seeding**
```bash
make reset        # Reset everything
make full-setup   # Complete setup again
```

### Getting Help

Check system information:
```bash
make info        # Show versions and status
make db-status   # Check database
make logs        # Show container logs
```

## 🎯 API Endpoints

### Public Endpoints
```
GET  /health                    # Health check
POST /api/auth/register         # User registration
POST /api/auth/login            # User login
GET  /api/cards/search          # Search cards
GET  /api/cards/{id}            # Get card details
GET  /api/cards/{id}/prices     # Get price history
```

### Protected Endpoints (Require Authentication)
```
GET  /api/protected/user/dashboard        # User dashboard
POST /api/protected/user/charts           # Save chart
GET  /api/protected/user/charts           # Get saved charts
DEL  /api/protected/user/charts/{id}      # Delete chart
```

### Example API Usage

**Search Cards:**
```bash
curl "http://localhost:8080/api/cards/search?q=charizard&game=Pokemon&limit=10"
```

**Get Card Details:**
```bash
curl "http://localhost:8080/api/cards/CARD_ID"
```

**Get Price History:**
```bash
curl "http://localhost:8080/api/cards/CARD_ID/prices?range=30d"
```

## 🛣️ Roadmap

### Phase 1 (Current)
- ✅ Core search and price tracking
- ✅ Basic technical indicators
- ✅ User authentication
- ✅ Dashboard functionality

### Phase 2 (Next)
- 🔲 Advanced technical indicators
- 🔲 Price alerts and notifications
- 🔲 Mobile responsive design
- 🔲 Data export capabilities

### Phase 3 (Future)
- 🔲 Mobile applications (iOS/Android)
- 🔲 Machine learning price predictions
- 🔲 Social features and community
- 🔲 API for third-party integrations

## 💎 Tech Stack Details

### Backend
- **Language:** Go 1.21+ (pure stdlib, no frameworks)
- **Database:** MongoDB 7.0 with text search indexes
- **Authentication:** JWT with HMAC-SHA256
- **Security:** OWASP compliant middleware stack
- **Performance:** Single binary deployment

### Frontend
- **Framework:** React 19 with native SSR
- **Build Tool:** Vite 5.0 for fast development
- **Styling:** Tailwind CSS for modern design
- **Routing:** React Router v6 with dynamic routes
- **Charts:** Recharts for interactive visualizations
- **State:** React Context + Custom hooks
- **TypeScript:** Full type safety

### Infrastructure
- **Development:** Docker Compose for local setup
- **Database:** MongoDB with replica set support
- **Caching:** Built-in Go caching mechanisms
- **Monitoring:** Health checks and logging

## 📄 License

This project is proprietary software. All rights reserved.

## 📧 Support

For support and questions:
- Create an issue in the repository
- Check the troubleshooting section above
- Review the API documentation

---

**Happy Trading! 🎯**

Built with ❤️ using Go and React for the trading card community.