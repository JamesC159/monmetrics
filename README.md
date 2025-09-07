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