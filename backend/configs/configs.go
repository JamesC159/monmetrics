package configs

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                string
	MongoURI            string
	DBName              string
	JWTSecret           []byte
	CORSOrigins         []string
	Environment         string
	RateLimitRequests   int
	RateLimitWindow     time.Duration
	MarketplaceMode     string
	TokenEncryptionKey  []byte
	FrontendURL         string
	APIBaseURL          string
	AlertsEnabled       bool
	AlertEvalInterval   time.Duration
	ScraperEnabled      bool
	ScraperInterval     time.Duration
	ScraperRequestDelay time.Duration
	ScraperMaxPages     int
	ScraperUserAgent    string
	ScraperBaseURL      string
	ScraperDailyBudget  int
	ScrapeTierIntervals map[int]time.Duration
	CatalogEnabled      bool
	CatalogInterval     time.Duration
	CatalogResumeWindow time.Duration
	CatalogBaseURL      string
	CatalogUserAgent    string
	CatalogRequestDelay time.Duration
	CatalogCategories   []int
	CatalogMaxProducts  int
	// PriceRequestDelay spaces out price-file requests. TCGCSV asks for ~100ms between
	// calls and bans applications exceeding 10k requests per 24h.
	PriceRequestDelay time.Duration
	// EnabledGames limits which TCGs the product exposes. Matched case-insensitively as
	// substrings: seeded cards say "Magic The Gathering", catalog cards say "Magic: The
	// Gathering". Empty exposes every game.
	EnabledGames []string
	// HiddenImageGames withholds card images for these games (matched like EnabledGames)
	// while keeping their prices, so a publisher's art can be pulled on request.
	HiddenImageGames []string
	// AdminEmails gates /api/admin. Privilege comes from server config, never from the
	// request, so revoking an admin takes effect on restart rather than at token expiry.
	AdminEmails  []string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

func Load() *Config {
	// Real environment variables take precedence; the file fills in what is unset.
	// Checked from the backend folder and from the repo root.
	loadDotEnv(".env", "backend/.env")

	config := &Config{
		Port:            getEnv("PORT", "8080"),
		MongoURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		DBName:          getEnv("DB_NAME", "monmetrics"),
		JWTSecret:       []byte(getEnv("JWT_SECRET", "change-this-super-secret-key")),
		Environment:     getEnv("ENVIRONMENT", "development"),
		MarketplaceMode: getEnv("MARKETPLACE_MODE", "mock"),
		FrontendURL:     strings.TrimRight(getEnv("FRONTEND_URL", "http://localhost:3000"), "/"),
	}
	config.APIBaseURL = strings.TrimRight(getEnv("API_BASE_URL", "http://localhost:"+config.Port), "/")

	config.TokenEncryptionKey = loadTokenKey(config)

	// Parse CORS origins
	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:3000")
	config.CORSOrigins = strings.Split(corsOrigins, ",")

	// Parse rate limiting config
	rateLimitRequests, err := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	if err != nil {
		rateLimitRequests = 100
	}
	config.RateLimitRequests = rateLimitRequests

	rateLimitWindow, err := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "60s"))
	if err != nil {
		rateLimitWindow = 60 * time.Second
	}
	config.RateLimitWindow = rateLimitWindow

	config.AlertsEnabled = getEnv("ALERTS_ENABLED", "true") != "false"
	interval, err := time.ParseDuration(getEnv("ALERT_EVAL_INTERVAL", "15m"))
	if err != nil || interval < 5*time.Second {
		interval = 15 * time.Minute
	}
	config.AlertEvalInterval = interval

	config.ScraperEnabled = getEnv("SCRAPER_ENABLED", "false") == "true"
	scraperInterval, err := time.ParseDuration(getEnv("SCRAPER_INTERVAL", "24h"))
	if err != nil || scraperInterval < 5*time.Second {
		scraperInterval = 24 * time.Hour
	}
	config.ScraperInterval = scraperInterval
	scraperDelay, err := time.ParseDuration(getEnv("SCRAPER_REQUEST_DELAY", "8s"))
	if err != nil || scraperDelay < 0 {
		scraperDelay = 8 * time.Second
	}
	config.ScraperRequestDelay = scraperDelay
	scraperMaxPages, err := strconv.Atoi(getEnv("SCRAPER_MAX_PAGES", "3"))
	if err != nil || scraperMaxPages < 1 || scraperMaxPages > 20 {
		scraperMaxPages = 3
	}
	config.ScraperMaxPages = scraperMaxPages
	config.ScraperUserAgent = getEnv("SCRAPER_USER_AGENT", "MonMetrics/1.0")
	config.ScraperBaseURL = strings.TrimRight(getEnv("SCRAPER_BASE_URL", "https://www.ebay.com"), "/")

	scraperBudget, err := strconv.Atoi(getEnv("SCRAPER_DAILY_BUDGET", "500"))
	if err != nil || scraperBudget < 1 {
		scraperBudget = 500
	}
	config.ScraperDailyBudget = scraperBudget
	config.ScrapeTierIntervals = map[int]time.Duration{
		1: parseDuration(getEnv("SCRAPER_TIER1_INTERVAL", "24h"), 24*time.Hour),
		2: parseDuration(getEnv("SCRAPER_TIER2_INTERVAL", "168h"), 7*24*time.Hour),
		3: parseDuration(getEnv("SCRAPER_TIER3_INTERVAL", "720h"), 30*24*time.Hour),
	}

	config.CatalogEnabled = getEnv("CATALOG_ENABLED", "false") == "true"
	config.CatalogInterval = parseDuration(getEnv("CATALOG_INTERVAL", "24h"), 24*time.Hour)
	config.CatalogResumeWindow = parseDuration(getEnv("CATALOG_RESUME_WINDOW", ""), config.CatalogInterval)
	config.CatalogBaseURL = strings.TrimRight(getEnv("CATALOG_BASE_URL", "https://tcgcsv.com"), "/")
	config.CatalogUserAgent = getEnv("CATALOG_USER_AGENT", config.ScraperUserAgent)
	config.CatalogRequestDelay = parseDuration(getEnv("CATALOG_REQUEST_DELAY", "150ms"), 150*time.Millisecond)
	config.CatalogCategories = parseIntList(os.Getenv("CATALOG_CATEGORIES"))
	catalogMax, err := strconv.Atoi(getEnv("CATALOG_MAX_PRODUCTS", "0"))
	if err != nil || catalogMax < 0 {
		catalogMax = 0
	}
	config.CatalogMaxProducts = catalogMax
	config.PriceRequestDelay = parseDuration(getEnv("PRICE_REQUEST_DELAY", "150ms"), 150*time.Millisecond)

	// Set ENABLED_GAMES="" to expose every game again.
	config.EnabledGames = parseStringList(getEnv("ENABLED_GAMES", "Magic"))
	config.HiddenImageGames = parseStringList(os.Getenv("HIDDEN_IMAGE_GAMES"))
	config.AdminEmails = normalizeEmails(parseStringList(os.Getenv("ADMIN_EMAILS")))

	config.SMTPHost = os.Getenv("SMTP_HOST")
	config.SMTPPort = getEnv("SMTP_PORT", "587")
	config.SMTPUser = os.Getenv("SMTP_USER")
	config.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	config.SMTPFrom = getEnv("SMTP_FROM", "MonMetrics <alerts@monmetrics.local>")

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed < time.Second {
		return fallback
	}
	return parsed
}

func parseIntList(value string) []int {
	values := make([]int, 0)
	for _, part := range strings.Split(value, ",") {
		number, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		values = append(values, number)
	}
	return values
}

func parseStringList(value string) []string {
	values := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func normalizeEmails(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, strings.ToLower(strings.TrimSpace(v)))
	}
	return out
}

// loadTokenKey reads TOKEN_ENCRYPTION_KEY (base64, 32 bytes). Outside production a
// key derived from the JWT secret is used so local setups work without extra config.
func loadTokenKey(config *Config) []byte {
	if raw := os.Getenv("TOKEN_ENCRYPTION_KEY"); raw != "" {
		key, err := base64.StdEncoding.DecodeString(raw)
		if err != nil || len(key) != 32 {
			log.Fatal("TOKEN_ENCRYPTION_KEY must be base64-encoded 32 bytes")
		}
		return key
	}
	if config.Environment == "production" {
		log.Fatal("TOKEN_ENCRYPTION_KEY is required in production")
	}
	sum := sha256.Sum256(append([]byte("monmetrics-token-key:"), config.JWTSecret...))
	return sum[:]
}
