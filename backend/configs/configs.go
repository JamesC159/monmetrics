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
	Port               string
	MongoURI           string
	DBName             string
	JWTSecret          []byte
	CORSOrigins        []string
	Environment        string
	RateLimitRequests  int
	RateLimitWindow    time.Duration
	MarketplaceMode    string
	TokenEncryptionKey []byte
	FrontendURL        string
	APIBaseURL         string
	AlertsEnabled      bool
	AlertEvalInterval  time.Duration
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPassword       string
	SMTPFrom           string
}

func Load() *Config {
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
