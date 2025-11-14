# Models Package Structure

This package contains all data models for the MonMetrics application, organized by domain responsibility.

## File Organization

### `user.go` - User & Dashboard Models

- **User** - User account information
- **UserStats** - User statistics (charts created, indicators used, etc.)
- **Dashboard** - User dashboard aggregated data

### `card.go` - Card & Content Models

- **Card** - Trading card or sealed product
- **SearchResult** - Card search results with pagination
- **GameCardGroup** - Cards grouped by game and category
- **FeaturedContent** - Carousel content (market movers, news, products, etc.)

### `price.go` - Price Data Models

- **PricePoint** - Individual price data point from sources (eBay, TCGPlayer)
- **PriceHistory** - Historical price data with indicators
- **IndicatorPoint** - Calculated technical indicator value

### `chart.go` - Chart Configuration Models

- **SavedChart** - User's saved chart configuration
- **ChartIndicator** - Technical indicator settings (Bollinger, RSI, SMA, etc.)

### `market.go` - Market Data Models

- **MarketData** - Aggregated OHLC market data
- **Listing** - Current marketplace listing

### `api.go` - API Request/Response Models

**Request Models:**

- **RegisterRequest** - User registration
- **LoginRequest** - User login
- **SearchParams** - Search query parameters

**Response Models:**

- **AuthResponse** - Authentication response with token
- **ErrorResponse** - Standard error response
- **HealthResponse** - Health check response

## Design Principles

1. **Single Responsibility** - Each file focuses on one domain area
2. **Clear Separation** - Request/response models separate from domain models
3. **Logical Grouping** - Related models stay together (e.g., Chart + ChartIndicator)
4. **Easy Navigation** - File names clearly indicate contents
5. **Maintainability** - Easy to find and modify specific model types

## Usage

All models are in the same `models` package, so imports remain unchanged:

```go
import "github.com/jamesc159/monmetrics/internal/models"

// Use as before
user := models.User{...}
card := models.Card{...}
```

## Benefits of This Structure

- **Scalability** - Easy to add new models without file bloat
- **Readability** - Smaller files are easier to understand
- **Collaboration** - Reduces merge conflicts in team development
- **Testing** - Focused test files mirror model organization
- **Documentation** - Each file can have targeted package documentation
